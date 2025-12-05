// Package anysync provides Any-Sync integration components.
package anysync

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/anyproto/any-sync/commonspace/object/accountdata"
	"github.com/anyproto/any-sync/commonspace/object/tree/objecttree"
	"github.com/anyproto/any-sync/commonspace/objecttreebuilder"
)

// DocumentMetadata holds application-level document metadata.
// This is stored separately from ObjectTree to enable efficient querying.
type DocumentMetadata struct {
	DocumentID string            `json:"document_id"` // ObjectTree ID
	SpaceID    string            `json:"space_id"`
	Title      string            `json:"title"`
	Tags       []string          `json:"tags"`
	Metadata   map[string]string `json:"metadata"`
	CreatedAt  int64             `json:"created_at"`
	UpdatedAt  int64             `json:"updated_at"`
}

// DocumentManager manages documents within spaces using ObjectTree.
// Each document is an ObjectTree with changes stored as a DAG.
type DocumentManager struct {
	mu           sync.RWMutex
	spaceManager *SpaceManager
	keys         *accountdata.AccountKeys
	eventManager *EventManager
	// Document metadata cache for efficient querying
	// Key: spaceID -> documentID -> metadata
	metadata map[string]map[string]*DocumentMetadata
}

// DocumentManager constructor
func NewDocumentManager(spaceManager *SpaceManager, keys *accountdata.AccountKeys, eventManager *EventManager) (*DocumentManager, error) {
	if spaceManager == nil {
		return nil, fmt.Errorf("space manager required")
	}
	if keys == nil {
		return nil, fmt.Errorf("account keys required")
	}
	if eventManager == nil {
		return nil, fmt.Errorf("event manager required")
	}

	dm := &DocumentManager{
		spaceManager: spaceManager,
		keys:         keys,
		eventManager: eventManager,
		metadata:     make(map[string]map[string]*DocumentMetadata),
	}

	// Load metadata from all spaces
	if err := dm.loadAllMetadata(); err != nil {
		return nil, fmt.Errorf("failed to load metadata: %w", err)
	}

	return dm, nil
}

// CreateDocument creates a new document in a space.
// The document data is stored as the root change in an ObjectTree.
func (dm *DocumentManager) CreateDocument(spaceID, title string, data []byte, metadata map[string]string) (string, error) {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	// Get the space object
	space, err := dm.spaceManager.GetSpaceObject(spaceID)
	if err != nil {
		return "", fmt.Errorf("failed to get space: %w", err)
	}

	// Get TreeBuilder from space
	treeBuilder := space.TreeBuilder()
	if treeBuilder == nil {
		return "", fmt.Errorf("tree builder not available")
	}

	// Create ObjectTree payload
	ctx := context.Background()
	createPayload := objecttree.ObjectTreeCreatePayload{
		PrivKey:       dm.keys.SignKey,
		ChangeType:    "document",
		ChangePayload: data,
		SpaceId:       spaceID,
		IsEncrypted:   false, // TODO: Add encryption support
		Timestamp:     time.Now().Unix(),
	}

	// Create the tree storage payload
	treePayload, err := treeBuilder.CreateTree(ctx, createPayload)
	if err != nil {
		return "", fmt.Errorf("failed to create tree: %w", err)
	}

	// Build and store the ObjectTree
	tree, err := treeBuilder.PutTree(ctx, treePayload, nil)
	if err != nil {
		return "", fmt.Errorf("failed to put tree: %w", err)
	}

	documentID := tree.Id()

	// Store document metadata
	now := time.Now().Unix()
	docMeta := &DocumentMetadata{
		DocumentID: documentID,
		SpaceID:    spaceID,
		Title:      title,
		Tags:       []string{},
		Metadata:   metadata,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	if dm.metadata[spaceID] == nil {
		dm.metadata[spaceID] = make(map[string]*DocumentMetadata)
	}
	dm.metadata[spaceID][documentID] = docMeta

	if err := dm.saveMetadata(spaceID); err != nil {
		return "", fmt.Errorf("failed to save metadata: %w", err)
	}

	// Emit document.created event
	dm.eventManager.EmitEvent(EventDocumentCreated, spaceID, map[string]string{
		"document_id": documentID,
		"collection":  "", // TODO: Add collection support
	})

	return documentID, nil
}

// GetDocument retrieves a document by ID from a space.
func (dm *DocumentManager) GetDocument(spaceID, documentID string) ([]byte, *DocumentMetadata, error) {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	// Get metadata
	docMeta, err := dm.getMetadata(spaceID, documentID)
	if err != nil {
		return nil, nil, err
	}

	// Get the space object
	space, err := dm.spaceManager.GetSpaceObject(spaceID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get space: %w", err)
	}

	// Get TreeBuilder from space
	treeBuilder := space.TreeBuilder()
	if treeBuilder == nil {
		return nil, nil, fmt.Errorf("tree builder not available")
	}

	// Build the ObjectTree
	ctx := context.Background()
	tree, err := treeBuilder.BuildTree(ctx, documentID, objecttreebuilder.BuildTreeOpts{})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to build tree: %w", err)
	}
	defer tree.Close()

	// Get the latest head (most recent change) instead of the root
	heads := tree.Heads()
	if len(heads) == 0 {
		return nil, nil, fmt.Errorf("document has no heads")
	}

	// Get the latest change from the head
	latestChange, err := tree.GetChange(heads[0])
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get latest change: %w", err)
	}

	// The change.Data contains a simple protobuf with:
	// field 1: changeType (string)
	// field 2: changePayload (bytes) - our actual document data
	// We need to extract field 2.
	data := latestChange.Data
	if len(data) == 0 {
		return nil, nil, fmt.Errorf("root change has no data")
	}

	// Simple protobuf parser to extract field 2 (our document payload)
	// Format: [field_tag][length][data]...
	// We're looking for field 2 with wire type 2 (0x12)
	extracted, err := extractProtobufField(data, 2)
	if err != nil || extracted == nil {
		// If extraction fails, return raw data (backward compatibility)
		return data, docMeta, nil
	}

	return extracted, docMeta, nil
}

// UpdateDocument updates an existing document by adding a new change to its ObjectTree.
func (dm *DocumentManager) UpdateDocument(spaceID, documentID string, data []byte, metadata map[string]string) error {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	// Verify document exists
	if _, err := dm.getMetadata(spaceID, documentID); err != nil {
		return err
	}

	// Get the space object
	space, err := dm.spaceManager.GetSpaceObject(spaceID)
	if err != nil {
		return fmt.Errorf("failed to get space: %w", err)
	}

	// Get TreeBuilder from space
	treeBuilder := space.TreeBuilder()
	if treeBuilder == nil {
		return fmt.Errorf("tree builder not available")
	}

	// Build the ObjectTree
	ctx := context.Background()
	tree, err := treeBuilder.BuildTree(ctx, documentID, objecttreebuilder.BuildTreeOpts{})
	if err != nil {
		return fmt.Errorf("failed to build tree: %w", err)
	}
	defer tree.Close()

	// Add new content to the tree
	// Note: AddContent expects raw data and will wrap it appropriately
	changeContent := objecttree.SignableChangeContent{
		Data:              data,
		Key:               dm.keys.SignKey,
		IsSnapshot:        false,
		ShouldBeEncrypted: false,
		Timestamp:         time.Now().Unix(),
		DataType:          "document",
	}

	_, err = tree.AddContent(ctx, changeContent)
	if err != nil {
		return fmt.Errorf("failed to add content: %w", err)
	}

	// Update metadata
	now := time.Now().Unix()
	if dm.metadata[spaceID] != nil && dm.metadata[spaceID][documentID] != nil {
		docMeta := dm.metadata[spaceID][documentID]
		docMeta.UpdatedAt = now

		// Replace metadata entirely with provided metadata
		// Frontend should send complete metadata map to preserve fields
		if metadata != nil {
			// Update title if provided
			if title, ok := metadata["title"]; ok {
				docMeta.Title = title
			}

			// Full replacement, application controls what's kept
			docMeta.Metadata = metadata
		}

		if err := dm.saveMetadata(spaceID); err != nil {
			return fmt.Errorf("failed to save metadata: %w", err)
		}
	}

	// Emit document.updated event
	dm.eventManager.EmitEvent(EventDocumentUpdated, spaceID, map[string]string{
		"document_id": documentID,
	})

	return nil
}

// DeleteDocument marks a document as deleted.
func (dm *DocumentManager) DeleteDocument(spaceID, documentID string) error {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	// Verify document exists
	if _, err := dm.getMetadata(spaceID, documentID); err != nil {
		return err
	}

	// Get the space object
	space, err := dm.spaceManager.GetSpaceObject(spaceID)
	if err != nil {
		return fmt.Errorf("failed to get space: %w", err)
	}

	// Delete the tree via space
	ctx := context.Background()
	if err := space.DeleteTree(ctx, documentID); err != nil {
		return fmt.Errorf("failed to delete tree: %w", err)
	}

	// Remove metadata
	if dm.metadata[spaceID] != nil {
		delete(dm.metadata[spaceID], documentID)
		if err := dm.saveMetadata(spaceID); err != nil {
			return fmt.Errorf("failed to save metadata: %w", err)
		}
	}

	// Emit document.deleted event
	dm.eventManager.EmitEvent(EventDocumentDeleted, spaceID, map[string]string{
		"document_id": documentID,
	})

	return nil
}

// ListDocuments returns all documents in a space.
func (dm *DocumentManager) ListDocuments(spaceID string) ([]*DocumentMetadata, error) {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	spaceMeta, exists := dm.metadata[spaceID]
	if !exists {
		return []*DocumentMetadata{}, nil
	}

	documents := make([]*DocumentMetadata, 0, len(spaceMeta))
	for _, doc := range spaceMeta {
		// Create a copy to avoid mutation
		docCopy := *doc
		documents = append(documents, &docCopy)
	}

	return documents, nil
}

// QueryDocuments returns documents matching the given query.
// For now, this is a simple tag-based filter, but can be extended.
func (dm *DocumentManager) QueryDocuments(spaceID string, tags []string) ([]*DocumentMetadata, error) {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	spaceMeta, exists := dm.metadata[spaceID]
	if !exists {
		return []*DocumentMetadata{}, nil
	}

	if len(tags) == 0 {
		// No filter, return all
		return dm.ListDocuments(spaceID)
	}

	// Filter by tags
	var matched []*DocumentMetadata
	for _, doc := range spaceMeta {
		if hasAllTags(doc.Tags, tags) {
			docCopy := *doc
			matched = append(matched, &docCopy)
		}
	}

	return matched, nil
}

// Helper functions

func (dm *DocumentManager) getMetadata(spaceID, documentID string) (*DocumentMetadata, error) {
	spaceMeta, exists := dm.metadata[spaceID]
	if !exists {
		return nil, fmt.Errorf("space not found: %s", spaceID)
	}

	docMeta, exists := spaceMeta[documentID]
	if !exists {
		return nil, fmt.Errorf("document not found: %s", documentID)
	}

	return docMeta, nil
}

func (dm *DocumentManager) loadAllMetadata() error {
	// Load metadata for all spaces
	spaces := dm.spaceManager.ListSpaces()
	for _, space := range spaces {
		if err := dm.loadMetadata(space.SpaceID); err != nil {
			// Log error but continue loading other spaces
			continue
		}
	}
	return nil
}

func (dm *DocumentManager) loadMetadata(spaceID string) error {
	// Load metadata from JSON file
	dataDir := dm.spaceManager.GetDataDir()
	metadataPath := filepath.Join(dataDir, "documents", spaceID+".json")

	data, err := os.ReadFile(metadataPath)
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist yet, initialize empty
			if dm.metadata[spaceID] == nil {
				dm.metadata[spaceID] = make(map[string]*DocumentMetadata)
			}
			return nil
		}
		return fmt.Errorf("failed to read metadata file: %w", err)
	}

	var spaceMeta map[string]*DocumentMetadata
	if err := json.Unmarshal(data, &spaceMeta); err != nil {
		return fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	dm.metadata[spaceID] = spaceMeta
	return nil
}

func (dm *DocumentManager) saveMetadata(spaceID string) error {
	// Save metadata to JSON file
	spaceMeta, exists := dm.metadata[spaceID]
	if !exists {
		return nil
	}

	dataDir := dm.spaceManager.GetDataDir()
	documentsDir := filepath.Join(dataDir, "documents")

	// Ensure documents directory exists
	if err := os.MkdirAll(documentsDir, 0755); err != nil {
		return fmt.Errorf("failed to create documents directory: %w", err)
	}

	metadataPath := filepath.Join(documentsDir, spaceID+".json")
	data, err := json.MarshalIndent(spaceMeta, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	if err := os.WriteFile(metadataPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write metadata file: %w", err)
	}

	return nil
}

func hasAllTags(documentTags, queryTags []string) bool {
	tagSet := make(map[string]bool)
	for _, tag := range documentTags {
		tagSet[tag] = true
	}

	for _, tag := range queryTags {
		if !tagSet[tag] {
			return false
		}
	}

	return true
}

// extractProtobufField extracts a field value from a simple protobuf message.
// This is a minimal parser that works for length-delimited fields (wire type 2).
func extractProtobufField(data []byte, fieldNumber int) ([]byte, error) {
	i := 0
	for i < len(data) {
		if i >= len(data) {
			break
		}

		// Read tag
		tag := data[i]
		i++

		wireType := tag & 0x07
		currentField := int(tag >> 3)

		if wireType != 2 {
			// Skip non-length-delimited fields (not supported in this simple parser)
			return nil, fmt.Errorf("unsupported wire type: %d", wireType)
		}

		// Read length (varint)
		if i >= len(data) {
			return nil, fmt.Errorf("unexpected end of data")
		}
		length := int(data[i])
		i++

		// Check if this is our target field
		if currentField == fieldNumber {
			if i+length > len(data) {
				return nil, fmt.Errorf("field length exceeds data bounds")
			}
			return data[i : i+length], nil
		}

		// Skip this field's data
		i += length
	}

	return nil, fmt.Errorf("field %d not found", fieldNumber)
}

// Close closes the DocumentManager and releases resources.
func (dm *DocumentManager) Close() error {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	// Clear metadata cache
	dm.metadata = make(map[string]map[string]*DocumentMetadata)

	// No other cleanup needed as we don't hold any persistent resources
	return nil
}

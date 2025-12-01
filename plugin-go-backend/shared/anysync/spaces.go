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

	"github.com/anyproto/any-sync/accountservice"
	"github.com/anyproto/any-sync/app"
	"github.com/anyproto/any-sync/commonspace"
	"github.com/anyproto/any-sync/commonspace/object/accountdata"
	"github.com/anyproto/any-sync/commonspace/object/keyvalue/keyvaluestorage"
	"github.com/anyproto/any-sync/commonspace/spacepayloads"
	"github.com/anyproto/any-sync/commonspace/spacestorage"
	"github.com/anyproto/any-sync/commonspace/syncstatus"
	"github.com/anyproto/any-sync/util/crypto"
)

// SpaceMetadata holds application-level space metadata.
// This is stored separately from Any-Sync's internal space structure.
type SpaceMetadata struct {
	SpaceID   string            `json:"space_id"`
	Name      string            `json:"name"`
	Metadata  map[string]string `json:"metadata"`
	CreatedAt int64             `json:"created_at"`
	UpdatedAt int64             `json:"updated_at"`
}

// SpaceManager manages local spaces with full Any-Sync structure.
//
// Phase 2D Implementation:
// - Uses app.App with minimal components for local-only operation
// - SpaceService creates and manages Space objects
// - Space objects provide TreeBuilder for document operations
// - Maintains backward-compatible metadata storage
type SpaceManager struct {
	mu           sync.RWMutex
	dataDir      string
	keys         *accountdata.AccountKeys
	spaces       map[string]*SpaceMetadata    // Application-level metadata
	spaceObjects map[string]commonspace.Space // Any-Sync Space objects
	storageDir   string                       // Directory for space storage databases

	// Any-Sync components
	app             *app.App
	spaceService    commonspace.SpaceService
	storageProvider spacestorage.SpaceStorageProvider
}

// NewSpaceManager creates a new SpaceManager with full Any-Sync integration.
func NewSpaceManager(dataDir string, keys *accountdata.AccountKeys) (*SpaceManager, error) {
	if keys == nil {
		return nil, fmt.Errorf("account keys required")
	}

	storageDir := filepath.Join(dataDir, "spaces")
	if err := os.MkdirAll(storageDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %w", err)
	}

	sm := &SpaceManager{
		dataDir:      dataDir,
		keys:         keys,
		spaces:       make(map[string]*SpaceMetadata),
		spaceObjects: make(map[string]commonspace.Space),
		storageDir:   storageDir,
	}

	// Initialize Any-Sync components
	if err := sm.initializeAnySync(); err != nil {
		return nil, fmt.Errorf("failed to initialize Any-Sync: %w", err)
	}

	// Load existing space metadata
	if err := sm.loadMetadata(); err != nil {
		return nil, fmt.Errorf("failed to load space metadata: %w", err)
	}

	return sm, nil
}

// initializeAnySync sets up the minimal app.App with components needed for local-only space operations.
func (sm *SpaceManager) initializeAnySync() error {
	// Create app.App instance
	sm.app = new(app.App)

	// Create and register storage provider
	sm.storageProvider = newLocalSpaceStorageProvider(sm.storageDir)
	sm.app.Register(sm.storageProvider)

	// Create and register account service
	accountSvc := newLocalAccountService(sm.keys)
	sm.app.Register(accountSvc)

	// Register minimal mock components for local-only operation
	sm.app.Register(newNoOpTreeManager())
	sm.app.Register(newNoOpNodeConf())
	sm.app.Register(newNoOpPeerManagerProvider())
	sm.app.Register(newNoOpPool())
	sm.app.Register(newNoOpConfig())

	// Create and register SpaceService
	sm.spaceService = commonspace.New()
	sm.app.Register(sm.spaceService)

	// Start the app (initializes all components)
	ctx := context.Background()
	if err := sm.app.Start(ctx); err != nil {
		return fmt.Errorf("failed to start app: %w", err)
	}

	return nil
}

// CreateSpace creates a new space with full Any-Sync structure using SpaceService.
func (sm *SpaceManager) CreateSpace(referenceName, name string, metadata map[string]string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	ctx := context.Background()

	// Generate cryptographic keys for the space
	masterKey, _, err := crypto.GenerateRandomEd25519KeyPair()
	if err != nil {
		return fmt.Errorf("failed to generate master key: %w", err)
	}

	metadataKey, _, err := crypto.GenerateRandomEd25519KeyPair()
	if err != nil {
		return fmt.Errorf("failed to generate metadata key: %w", err)
	}

	readKey := crypto.NewAES()

	// Create space payload using Any-Sync's helper
	createPayload := spacepayloads.SpaceCreatePayload{
		SigningKey:     sm.keys.SignKey,
		SpaceType:      "syncspace", // Our space type identifier
		ReplicationKey: 10,          // Arbitrary value for local-only operation
		SpacePayload:   nil,         // No custom payload
		MasterKey:      masterKey,
		ReadKey:        readKey,
		MetadataKey:    metadataKey,
		Metadata:       []byte(name), // Store name in Any-Sync metadata
	}

	// Convert to storage payload
	storagePayload, err := spacepayloads.StoragePayloadForSpaceCreate(createPayload)
	if err != nil {
		return fmt.Errorf("failed to create storage payload: %w", err)
	}

	// Create space storage via provider
	storage, err := sm.storageProvider.CreateSpaceStorage(ctx, storagePayload)
	if err != nil {
		return fmt.Errorf("failed to create space storage: %w", err)
	}

	// Extract the space ID from the space header
	actualSpaceID := storagePayload.SpaceHeaderWithId.Id

	// Create Space object using SpaceService
	spaceDeps := sm.createSpaceDeps()
	space, err := sm.spaceService.NewSpace(ctx, actualSpaceID, spaceDeps)
	if err != nil {
		storage.Close(ctx)
		return fmt.Errorf("failed to create space object: %w", err)
	}

	// Note: We don't call space.Init() here. NewSpace() handles initialization internally.
	// Calling Init() would start sync services which need network components we don't have.

	// Store application metadata
	now := time.Now().Unix()
	spaceMeta := &SpaceMetadata{
		SpaceID:   actualSpaceID,
		Name:      name,
		Metadata:  metadata,
		CreatedAt: now,
		UpdatedAt: now,
	}

	sm.spaces[actualSpaceID] = spaceMeta
	sm.spaceObjects[actualSpaceID] = space

	if err := sm.saveMetadata(); err != nil {
		// Rollback
		space.Close()
		delete(sm.spaces, actualSpaceID)
		delete(sm.spaceObjects, actualSpaceID)
		return fmt.Errorf("failed to save metadata: %w", err)
	}

	return nil
}

// createSpaceDeps creates the dependencies needed for Space creation.
func (sm *SpaceManager) createSpaceDeps() commonspace.Deps {
	return commonspace.Deps{
		SyncStatus:     syncstatus.NewNoOpSyncStatus(),
		TreeSyncer:     newNoOpTreeSyncer(),
		AccountService: sm.app.MustComponent(accountservice.CName).(accountservice.Service),
		Indexer:        keyvaluestorage.NoOpIndexer{},
	}
}

// GetSpaceObject retrieves or initializes a Space object by ID.
// This is the method that Phase 2D will use to access TreeBuilder.
func (sm *SpaceManager) GetSpaceObject(spaceID string) (commonspace.Space, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Check if space metadata exists
	if _, exists := sm.spaces[spaceID]; !exists {
		return nil, fmt.Errorf("space not found: %s", spaceID)
	}

	// Check if already initialized
	if space, exists := sm.spaceObjects[spaceID]; exists {
		return space, nil
	}

	// Initialize the Space object
	ctx := context.Background()
	spaceDeps := sm.createSpaceDeps()
	space, err := sm.spaceService.NewSpace(ctx, spaceID, spaceDeps)
	if err != nil {
		return nil, fmt.Errorf("failed to create space object: %w", err)
	}

	// Note: NewSpace() handles initialization internally
	sm.spaceObjects[spaceID] = space
	return space, nil
}

// ListSpaces returns all spaces.
func (sm *SpaceManager) ListSpaces() []*SpaceMetadata {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	spaces := make([]*SpaceMetadata, 0, len(sm.spaces))
	for _, space := range sm.spaces {
		// Create a copy to avoid mutation
		spaceCopy := *space
		spaces = append(spaces, &spaceCopy)
	}

	return spaces
}

// GetSpace retrieves space metadata by ID.
func (sm *SpaceManager) GetSpace(spaceID string) (*SpaceMetadata, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	space, exists := sm.spaces[spaceID]
	if !exists {
		return nil, fmt.Errorf("space not found: %s", spaceID)
	}

	// Return a copy
	spaceCopy := *space
	return &spaceCopy, nil
}

// DeleteSpace removes a space and its storage.
func (sm *SpaceManager) DeleteSpace(spaceID string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Check if space exists
	if _, exists := sm.spaces[spaceID]; !exists {
		return fmt.Errorf("space not found: %s", spaceID)
	}

	// Close Space object if open (catch panics from partially initialized spaces)
	if space, ok := sm.spaceObjects[spaceID]; ok {
		func() {
			defer func() {
				if r := recover(); r != nil {
					// Ignore panics from closing partially initialized spaces
				}
			}()
			space.Close()
		}()
		delete(sm.spaceObjects, spaceID)
	}

	// Remove storage database file
	dbPath := filepath.Join(sm.storageDir, spaceID+".db")
	if err := os.RemoveAll(dbPath); err != nil {
		return fmt.Errorf("failed to remove space storage: %w", err)
	}

	// Remove from metadata
	delete(sm.spaces, spaceID)

	if err := sm.saveMetadata(); err != nil {
		return fmt.Errorf("failed to save metadata: %w", err)
	}

	return nil
}

// loadMetadata loads space metadata from disk.
func (sm *SpaceManager) loadMetadata() error {
	metadataPath := filepath.Join(sm.dataDir, "spaces_metadata.json")

	data, err := os.ReadFile(metadataPath)
	if err != nil {
		if os.IsNotExist(err) {
			// No metadata file yet, that's fine
			return nil
		}
		return fmt.Errorf("failed to read metadata file: %w", err)
	}

	var spaces []*SpaceMetadata
	if err := json.Unmarshal(data, &spaces); err != nil {
		return fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	for _, space := range spaces {
		sm.spaces[space.SpaceID] = space
	}

	return nil
}

// saveMetadata persists space metadata to disk.
func (sm *SpaceManager) saveMetadata() error {
	metadataPath := filepath.Join(sm.dataDir, "spaces_metadata.json")

	spaces := make([]*SpaceMetadata, 0, len(sm.spaces))
	for _, space := range sm.spaces {
		spaces = append(spaces, space)
	}

	data, err := json.Marshal(spaces)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	if err := os.WriteFile(metadataPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write metadata file: %w", err)
	}

	return nil
}

// Close closes all open spaces and shuts down the Any-Sync app.
func (sm *SpaceManager) Close() error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Close all Space objects (catch panics from partially initialized spaces in local-only mode)
	for _, space := range sm.spaceObjects {
		func() {
			defer func() {
				if r := recover(); r != nil {
					// Ignore panics from closing partially initialized spaces
					// This can happen in local-only mode where some components aren't fully set up
				}
			}()
			space.Close()
		}()
	}
	sm.spaceObjects = make(map[string]commonspace.Space)

	// Close the app (which closes all components)
	if sm.app != nil {
		ctx := context.Background()
		if err := sm.app.Close(ctx); err != nil {
			return fmt.Errorf("failed to close app: %w", err)
		}
	}

	return nil
}

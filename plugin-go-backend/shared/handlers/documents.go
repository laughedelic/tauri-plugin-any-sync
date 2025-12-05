package handlers

import (
	"context"
	"fmt"

	pb "anysync-backend/shared/proto/syncspace/v1"

	"google.golang.org/protobuf/proto"
)

// CreateDocument handles document creation.
func CreateDocument(ctx context.Context, req proto.Message) (proto.Message, error) {
	if err := ensureInitialized(); err != nil {
		return nil, err
	}

	docReq := req.(*pb.CreateDocumentRequest)

	globalState.mu.RLock()
	docManager := globalState.documentManager
	globalState.mu.RUnlock()

	if docManager == nil {
		return nil, fmt.Errorf("document manager not initialized")
	}

	// Extract title from metadata if present, otherwise use empty string
	title := ""
	if docReq.Metadata != nil {
		if t, ok := docReq.Metadata["title"]; ok {
			title = t
		}
	}

	// Create document using DocumentManager
	documentID, err := docManager.CreateDocument(
		docReq.SpaceId,
		title,
		docReq.Data,
		docReq.Metadata,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create document: %w", err)
	}

	return &pb.CreateDocumentResponse{
		DocumentId: documentID,
		Version:    1, // First version is always 1
	}, nil
}

// GetDocument retrieves a document.
func GetDocument(ctx context.Context, req proto.Message) (proto.Message, error) {
	if err := ensureInitialized(); err != nil {
		return nil, err
	}

	getReq := req.(*pb.GetDocumentRequest)

	globalState.mu.RLock()
	docManager := globalState.documentManager
	globalState.mu.RUnlock()

	if docManager == nil {
		return nil, fmt.Errorf("document manager not initialized")
	}

	// Get document using DocumentManager
	data, metadata, err := docManager.GetDocument(getReq.SpaceId, getReq.DocumentId)
	if err != nil {
		// Document not found or other error
		return &pb.GetDocumentResponse{
			Found: false,
		}, nil
	}

	return &pb.GetDocumentResponse{
		Found: true,
		Document: &pb.Document{
			DocumentId: metadata.DocumentID,
			SpaceId:    metadata.SpaceID,
			Collection: "", // TODO: Add collection support
			Data:       data,
			Metadata:   metadata.Metadata,
			Version:    1, // TODO: Add version tracking
			CreatedAt:  metadata.CreatedAt,
			UpdatedAt:  metadata.UpdatedAt,
		},
	}, nil
}

// UpdateDocument updates a document.
func UpdateDocument(ctx context.Context, req proto.Message) (proto.Message, error) {
	if err := ensureInitialized(); err != nil {
		return nil, err
	}

	updateReq := req.(*pb.UpdateDocumentRequest)

	globalState.mu.RLock()
	docManager := globalState.documentManager
	globalState.mu.RUnlock()

	if docManager == nil {
		return nil, fmt.Errorf("document manager not initialized")
	}

	// Update document using DocumentManager
	err := docManager.UpdateDocument(
		updateReq.SpaceId,
		updateReq.DocumentId,
		updateReq.Data,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update document: %w", err)
	}

	// TODO: Update metadata if provided in request
	// TODO: Implement version checking for optimistic locking

	return &pb.UpdateDocumentResponse{
		Version: 2, // TODO: Return actual version
	}, nil
}

// DeleteDocument deletes a document.
func DeleteDocument(ctx context.Context, req proto.Message) (proto.Message, error) {
	if err := ensureInitialized(); err != nil {
		return nil, err
	}

	deleteReq := req.(*pb.DeleteDocumentRequest)

	globalState.mu.RLock()
	docManager := globalState.documentManager
	globalState.mu.RUnlock()

	if docManager == nil {
		return nil, fmt.Errorf("document manager not initialized")
	}

	// Check if document exists before deletion
	_, _, err := docManager.GetDocument(deleteReq.SpaceId, deleteReq.DocumentId)
	existed := err == nil

	// Delete document using DocumentManager
	err = docManager.DeleteDocument(deleteReq.SpaceId, deleteReq.DocumentId)
	if err != nil {
		return nil, fmt.Errorf("failed to delete document: %w", err)
	}

	return &pb.DeleteDocumentResponse{
		Existed: existed,
	}, nil
}

// ListDocuments lists documents.
func ListDocuments(ctx context.Context, req proto.Message) (proto.Message, error) {
	if err := ensureInitialized(); err != nil {
		return nil, err
	}

	listReq := req.(*pb.ListDocumentsRequest)

	globalState.mu.RLock()
	docManager := globalState.documentManager
	globalState.mu.RUnlock()

	if docManager == nil {
		return nil, fmt.Errorf("document manager not initialized")
	}

	// List documents using DocumentManager
	metadataList, err := docManager.ListDocuments(listReq.SpaceId)
	if err != nil {
		return nil, fmt.Errorf("failed to list documents: %w", err)
	}

	// Convert to protobuf DocumentInfo
	documents := make([]*pb.DocumentInfo, 0, len(metadataList))
	for _, metadata := range metadataList {
		// TODO: Apply collection filter when collection support is implemented
		// For now, ignore collection filter to return all documents

		// Apply limit if specified
		if listReq.Limit > 0 && len(documents) >= int(listReq.Limit) {
			break
		}

		documents = append(documents, &pb.DocumentInfo{
			DocumentId: metadata.DocumentID,
			Collection: listReq.Collection, // Echo back the requested collection
			Metadata:   metadata.Metadata,
			Version:    1, // TODO: Add version tracking
			CreatedAt:  metadata.CreatedAt,
			UpdatedAt:  metadata.UpdatedAt,
		})
	}

	return &pb.ListDocumentsResponse{
		Documents:  documents,
		NextCursor: "", // TODO: Add pagination support
	}, nil
}

// QueryDocuments queries documents.
func QueryDocuments(ctx context.Context, req proto.Message) (proto.Message, error) {
	if err := ensureInitialized(); err != nil {
		return nil, err
	}

	queryReq := req.(*pb.QueryDocumentsRequest)

	globalState.mu.RLock()
	docManager := globalState.documentManager
	globalState.mu.RUnlock()

	if docManager == nil {
		return nil, fmt.Errorf("document manager not initialized")
	}

	// Extract tags from filters (simple implementation for now)
	// TODO: Support full QueryFilter operators
	var tags []string
	for _, filter := range queryReq.Filters {
		if filter.Field == "tags" && filter.Operator == "contains" {
			tags = append(tags, filter.Value)
		}
	}

	// Query documents using DocumentManager
	metadataList, err := docManager.QueryDocuments(queryReq.SpaceId, tags)
	if err != nil {
		return nil, fmt.Errorf("failed to query documents: %w", err)
	}

	// Convert to protobuf DocumentInfo
	documents := make([]*pb.DocumentInfo, 0, len(metadataList))
	for _, metadata := range metadataList {
		// TODO: Apply collection filter when collection support is implemented
		// For now, ignore collection filter to return all documents

		// Apply limit if specified
		if queryReq.Limit > 0 && len(documents) >= int(queryReq.Limit) {
			break
		}

		documents = append(documents, &pb.DocumentInfo{
			DocumentId: metadata.DocumentID,
			Collection: queryReq.Collection, // Echo back the requested collection
			Metadata:   metadata.Metadata,
			Version:    1, // TODO: Add version tracking
			CreatedAt:  metadata.CreatedAt,
			UpdatedAt:  metadata.UpdatedAt,
		})
	}

	return &pb.QueryDocumentsResponse{
		Documents:  documents,
		NextCursor: "", // TODO: Add pagination support
	}, nil
}

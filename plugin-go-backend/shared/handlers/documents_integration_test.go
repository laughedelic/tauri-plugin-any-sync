package handlers

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	pb "anysync-backend/shared/proto/syncspace/v1"

	"google.golang.org/protobuf/proto"
)

// TestDocumentHandlers_Integration tests document operations through handlers.
// NOTE: Uses global state - run separately or exclude E2E tests from regular test runs.
func TestDocumentHandlers_Integration(t *testing.T) {
	// Setup: Initialize system
	tempDir := t.TempDir()
	dataDir := filepath.Join(tempDir, "test_data")
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		t.Fatalf("failed to create data dir: %v", err)
	}

	ctx := context.Background()

	// Initialize
	initReq := &pb.InitRequest{
		DataDir:   dataDir,
		NetworkId: "test-network",
		DeviceId:  "test-device",
	}
	_, err := Init(ctx, initReq)
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	defer func() {
		Shutdown(ctx, &pb.ShutdownRequest{})
	}()

	// Create space
	createSpaceReq := &pb.CreateSpaceRequest{
		Name:     "Test Space",
		Metadata: map[string]string{"purpose": "testing"},
	}
	createSpaceResp, err := CreateSpace(ctx, createSpaceReq)
	if err != nil {
		t.Fatalf("CreateSpace failed: %v", err)
	}
	spaceID := createSpaceResp.(*pb.CreateSpaceResponse).SpaceId

	t.Run("CreateDocument_Success", func(t *testing.T) {
		req := &pb.CreateDocumentRequest{
			SpaceId: spaceID,
			Data:    []byte("test document content"),
			Metadata: map[string]string{
				"title": "Test Document",
				"tag":   "test",
			},
		}

		resp, err := CreateDocument(ctx, req)
		if err != nil {
			t.Fatalf("CreateDocument failed: %v", err)
		}

		createResp := resp.(*pb.CreateDocumentResponse)
		if createResp.DocumentId == "" {
			t.Error("Expected non-empty document ID")
		}
		if createResp.Version != 1 {
			t.Errorf("Expected version 1, got %d", createResp.Version)
		}
	})

	t.Run("CreateDocument_InvalidSpace", func(t *testing.T) {
		req := &pb.CreateDocumentRequest{
			SpaceId: "invalid-space-id",
			Data:    []byte("test"),
		}

		_, err := CreateDocument(ctx, req)
		if err == nil {
			t.Error("Expected error for invalid space ID")
		}
	})

	t.Run("GetDocument_Success", func(t *testing.T) {
		// Create a document first
		createReq := &pb.CreateDocumentRequest{
			SpaceId: spaceID,
			Data:    []byte("get test content"),
			Metadata: map[string]string{
				"title": "Get Test",
			},
		}
		createResp, err := CreateDocument(ctx, createReq)
		if err != nil {
			t.Fatalf("CreateDocument failed: %v", err)
		}
		documentID := createResp.(*pb.CreateDocumentResponse).DocumentId

		// Get the document
		getReq := &pb.GetDocumentRequest{
			SpaceId:    spaceID,
			DocumentId: documentID,
		}
		resp, err := GetDocument(ctx, getReq)
		if err != nil {
			t.Fatalf("GetDocument failed: %v", err)
		}

		getResp := resp.(*pb.GetDocumentResponse)
		if !getResp.Found {
			t.Error("Expected document to be found")
		}
		if string(getResp.Document.Data) != "get test content" {
			t.Errorf("Expected 'get test content', got '%s'", string(getResp.Document.Data))
		}
		if getResp.Document.Metadata["title"] != "Get Test" {
			t.Errorf("Expected title 'Get Test', got '%s'", getResp.Document.Metadata["title"])
		}
	})

	t.Run("GetDocument_NotFound", func(t *testing.T) {
		req := &pb.GetDocumentRequest{
			SpaceId:    spaceID,
			DocumentId: "nonexistent-doc-id",
		}

		resp, err := GetDocument(ctx, req)
		if err != nil {
			t.Fatalf("GetDocument failed: %v", err)
		}

		getResp := resp.(*pb.GetDocumentResponse)
		if getResp.Found {
			t.Error("Expected document not to be found")
		}
	})

	t.Run("UpdateDocument_Success", func(t *testing.T) {
		// Create a document first
		createReq := &pb.CreateDocumentRequest{
			SpaceId: spaceID,
			Data:    []byte("original content"),
		}
		createResp, err := CreateDocument(ctx, createReq)
		if err != nil {
			t.Fatalf("CreateDocument failed: %v", err)
		}
		documentID := createResp.(*pb.CreateDocumentResponse).DocumentId

		// Update the document
		updateReq := &pb.UpdateDocumentRequest{
			SpaceId:    spaceID,
			DocumentId: documentID,
			Data:       []byte("updated content"),
		}
		resp, err := UpdateDocument(ctx, updateReq)
		if err != nil {
			t.Fatalf("UpdateDocument failed: %v", err)
		}

		updateResp := resp.(*pb.UpdateDocumentResponse)
		if updateResp.Version < 2 {
			t.Errorf("Expected version >= 2, got %d", updateResp.Version)
		}

		// Verify the update
		getReq := &pb.GetDocumentRequest{
			SpaceId:    spaceID,
			DocumentId: documentID,
		}
		getResp, err := GetDocument(ctx, getReq)
		if err != nil {
			t.Fatalf("GetDocument failed: %v", err)
		}

		doc := getResp.(*pb.GetDocumentResponse).Document
		if string(doc.Data) != "updated content" {
			t.Errorf("Expected 'updated content', got '%s'", string(doc.Data))
		}
	})

	t.Run("DeleteDocument_Success", func(t *testing.T) {
		// Create a document first
		createReq := &pb.CreateDocumentRequest{
			SpaceId: spaceID,
			Data:    []byte("to be deleted"),
		}
		createResp, err := CreateDocument(ctx, createReq)
		if err != nil {
			t.Fatalf("CreateDocument failed: %v", err)
		}
		documentID := createResp.(*pb.CreateDocumentResponse).DocumentId

		// Delete the document
		deleteReq := &pb.DeleteDocumentRequest{
			SpaceId:    spaceID,
			DocumentId: documentID,
		}
		resp, err := DeleteDocument(ctx, deleteReq)
		if err != nil {
			t.Fatalf("DeleteDocument failed: %v", err)
		}

		deleteResp := resp.(*pb.DeleteDocumentResponse)
		if !deleteResp.Existed {
			t.Error("Expected document to have existed")
		}

		// Verify deletion
		getReq := &pb.GetDocumentRequest{
			SpaceId:    spaceID,
			DocumentId: documentID,
		}
		getResp, err := GetDocument(ctx, getReq)
		if err != nil {
			t.Fatalf("GetDocument failed: %v", err)
		}

		if getResp.(*pb.GetDocumentResponse).Found {
			t.Error("Expected document to not be found after deletion")
		}
	})

	t.Run("ListDocuments_Success", func(t *testing.T) {
		// Create multiple documents
		for i := 0; i < 3; i++ {
			createReq := &pb.CreateDocumentRequest{
				SpaceId: spaceID,
				Data:    []byte(fmt.Sprintf("list test %d", i)),
				Metadata: map[string]string{
					"index": fmt.Sprintf("%d", i),
				},
			}
			_, err := CreateDocument(ctx, createReq)
			if err != nil {
				t.Fatalf("CreateDocument failed: %v", err)
			}
		}

		// List documents
		listReq := &pb.ListDocumentsRequest{
			SpaceId: spaceID,
		}
		resp, err := ListDocuments(ctx, listReq)
		if err != nil {
			t.Fatalf("ListDocuments failed: %v", err)
		}

		listResp := resp.(*pb.ListDocumentsResponse)
		if len(listResp.Documents) < 3 {
			t.Errorf("Expected at least 3 documents, got %d", len(listResp.Documents))
		}
	})

	t.Run("ListDocuments_WithLimit", func(t *testing.T) {
		listReq := &pb.ListDocumentsRequest{
			SpaceId: spaceID,
			Limit:   2,
		}
		resp, err := ListDocuments(ctx, listReq)
		if err != nil {
			t.Fatalf("ListDocuments failed: %v", err)
		}

		listResp := resp.(*pb.ListDocumentsResponse)
		if len(listResp.Documents) > 2 {
			t.Errorf("Expected at most 2 documents, got %d", len(listResp.Documents))
		}
	})

	t.Run("QueryDocuments_Success", func(t *testing.T) {
		// Query documents with empty filters (should return all)
		queryReq := &pb.QueryDocumentsRequest{
			SpaceId: spaceID,
			Filters: []*pb.QueryFilter{},
		}
		resp, err := QueryDocuments(ctx, queryReq)
		if err != nil {
			t.Fatalf("QueryDocuments failed: %v", err)
		}

		queryResp := resp.(*pb.QueryDocumentsResponse)
		// Should return at least some documents from previous tests
		if len(queryResp.Documents) == 0 {
			t.Error("Expected some documents, got 0")
		}

		// Note: Full tag-based querying requires proper tags field support
		// which will be implemented when we add collection/tag support to CreateDocument
	})
}

// TestDocumentHandlers_NotInitialized tests error handling when not initialized.
func TestDocumentHandlers_NotInitialized(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name    string
		handler func(context.Context, proto.Message) (proto.Message, error)
		req     proto.Message
	}{
		{
			name:    "CreateDocument",
			handler: CreateDocument,
			req:     &pb.CreateDocumentRequest{SpaceId: "test", Data: []byte("test")},
		},
		{
			name:    "GetDocument",
			handler: GetDocument,
			req:     &pb.GetDocumentRequest{SpaceId: "test", DocumentId: "test"},
		},
		{
			name:    "UpdateDocument",
			handler: UpdateDocument,
			req:     &pb.UpdateDocumentRequest{SpaceId: "test", DocumentId: "test", Data: []byte("test")},
		},
		{
			name:    "DeleteDocument",
			handler: DeleteDocument,
			req:     &pb.DeleteDocumentRequest{SpaceId: "test", DocumentId: "test"},
		},
		{
			name:    "ListDocuments",
			handler: ListDocuments,
			req:     &pb.ListDocumentsRequest{SpaceId: "test"},
		},
		{
			name:    "QueryDocuments",
			handler: QueryDocuments,
			req:     &pb.QueryDocumentsRequest{SpaceId: "test"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.handler(ctx, tt.req)
			if err == nil {
				t.Error("Expected error when not initialized")
			}
			if err.Error() != "not initialized: call Init first" {
				t.Errorf("Expected 'not initialized' error, got: %v", err)
			}
		})
	}
}

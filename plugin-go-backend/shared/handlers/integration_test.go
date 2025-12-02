package handlers

import (
	"testing"

	pb "anysync-backend/shared/proto/syncspace/v1"
)

// TestIntegration_DocumentHandlers tests document CRUD operations using the TestContext helper.
// All sub-tests run within a single Init/Shutdown cycle for efficiency.
func TestIntegration_DocumentHandlers(t *testing.T) {
	tc := SetupIntegrationTest(t)

	t.Run("CreateDocument", func(t *testing.T) {
		req := &pb.CreateDocumentRequest{
			SpaceId: tc.SpaceID(),
			Data:    []byte("test document content"),
			Metadata: map[string]string{
				"title": "Test Document",
				"tag":   "test",
			},
		}

		resp, err := CreateDocument(tc.Context(), req)
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

		_, err := CreateDocument(tc.Context(), req)
		if err == nil {
			t.Error("Expected error for invalid space ID")
		}
	})

	t.Run("GetDocument", func(t *testing.T) {
		// Create a document first
		docID := tc.CreateDocument([]byte("get test content"), map[string]string{
			"title": "Get Test",
		})

		// Now get it
		req := &pb.GetDocumentRequest{
			SpaceId:    tc.SpaceID(),
			DocumentId: docID,
		}

		resp, err := GetDocument(tc.Context(), req)
		if err != nil {
			t.Fatalf("GetDocument failed: %v", err)
		}

		getResp := resp.(*pb.GetDocumentResponse)
		if !getResp.Found {
			t.Error("Expected Found=true")
		}
		if getResp.Document == nil {
			t.Fatal("Expected non-nil document")
		}
		if string(getResp.Document.Data) != "get test content" {
			t.Errorf("Expected 'get test content', got '%s'", string(getResp.Document.Data))
		}
	})

	t.Run("GetDocument_NotFound", func(t *testing.T) {
		req := &pb.GetDocumentRequest{
			SpaceId:    tc.SpaceID(),
			DocumentId: "nonexistent-doc-id",
		}

		resp, err := GetDocument(tc.Context(), req)
		if err != nil {
			t.Fatalf("GetDocument failed: %v", err)
		}

		getResp := resp.(*pb.GetDocumentResponse)
		if getResp.Found {
			t.Error("Expected Found=false for nonexistent document")
		}
	})

	t.Run("UpdateDocument", func(t *testing.T) {
		// Create a document first
		docID := tc.CreateDocument([]byte("original content"), map[string]string{
			"title": "Update Test",
		})

		// Update it
		updateReq := &pb.UpdateDocumentRequest{
			SpaceId:    tc.SpaceID(),
			DocumentId: docID,
			Data:       []byte("updated content"),
		}

		updateResp, err := UpdateDocument(tc.Context(), updateReq)
		if err != nil {
			t.Fatalf("UpdateDocument failed: %v", err)
		}

		updateResult := updateResp.(*pb.UpdateDocumentResponse)
		if updateResult.Version != 2 {
			t.Errorf("Expected version 2, got %d", updateResult.Version)
		}

		// Verify the update
		getReq := &pb.GetDocumentRequest{
			SpaceId:    tc.SpaceID(),
			DocumentId: docID,
		}

		getResp, err := GetDocument(tc.Context(), getReq)
		if err != nil {
			t.Fatalf("GetDocument failed: %v", err)
		}

		getResult := getResp.(*pb.GetDocumentResponse)
		if getResult.Document == nil {
			t.Fatal("Expected non-nil document")
		}
		if string(getResult.Document.Data) != "updated content" {
			t.Errorf("Expected 'updated content', got '%s'", string(getResult.Document.Data))
		}
	})

	t.Run("DeleteDocument", func(t *testing.T) {
		// Create a document first
		docID := tc.CreateDocument([]byte("delete test content"), map[string]string{
			"title": "Delete Test",
		})

		// Delete it
		deleteReq := &pb.DeleteDocumentRequest{
			SpaceId:    tc.SpaceID(),
			DocumentId: docID,
		}

		_, err := DeleteDocument(tc.Context(), deleteReq)
		if err != nil {
			t.Fatalf("DeleteDocument failed: %v", err)
		}

		// Verify it's gone
		getReq := &pb.GetDocumentRequest{
			SpaceId:    tc.SpaceID(),
			DocumentId: docID,
		}

		getResp, err := GetDocument(tc.Context(), getReq)
		if err != nil {
			t.Fatalf("GetDocument failed: %v", err)
		}

		getResult := getResp.(*pb.GetDocumentResponse)
		if getResult.Found {
			t.Error("Expected Found=false after delete")
		}
	})

	t.Run("ListDocuments", func(t *testing.T) {
		// Create multiple documents
		tc.CreateDocument([]byte("list test 1"), map[string]string{"title": "List Test 1"})
		tc.CreateDocument([]byte("list test 2"), map[string]string{"title": "List Test 2"})
		tc.CreateDocument([]byte("list test 3"), map[string]string{"title": "List Test 3"})

		// List them
		req := &pb.ListDocumentsRequest{
			SpaceId: tc.SpaceID(),
		}

		resp, err := ListDocuments(tc.Context(), req)
		if err != nil {
			t.Fatalf("ListDocuments failed: %v", err)
		}

		listResp := resp.(*pb.ListDocumentsResponse)
		// At least 3 documents (might be more from previous sub-tests)
		if len(listResp.Documents) < 3 {
			t.Errorf("Expected at least 3 documents, got %d", len(listResp.Documents))
		}
	})

	t.Run("QueryDocuments", func(t *testing.T) {
		// Query documents
		req := &pb.QueryDocumentsRequest{
			SpaceId: tc.SpaceID(),
		}

		resp, err := QueryDocuments(tc.Context(), req)
		if err != nil {
			t.Fatalf("QueryDocuments failed: %v", err)
		}

		queryResp := resp.(*pb.QueryDocumentsResponse)
		if queryResp == nil {
			t.Error("Expected non-nil query response")
		}
	})
}

// TestIntegration_MultipleSpaces tests creating and managing multiple spaces.
func TestIntegration_MultipleSpaces(t *testing.T) {
	tc := SetupIntegrationTest(t)

	t.Run("CreateMultipleSpaces", func(t *testing.T) {
		space1 := tc.CreateSpace("Space 1", map[string]string{"order": "1"})
		space2 := tc.CreateSpace("Space 2", map[string]string{"order": "2"})
		space3 := tc.CreateSpace("Space 3", map[string]string{"order": "3"})

		if space1 == "" || space2 == "" || space3 == "" {
			t.Error("Expected non-empty space IDs")
		}

		// Verify they're all different
		if space1 == space2 || space2 == space3 || space1 == space3 {
			t.Error("Expected unique space IDs")
		}
	})

	t.Run("ListAllSpaces", func(t *testing.T) {
		req := &pb.ListSpacesRequest{}
		resp, err := ListSpaces(tc.Context(), req)
		if err != nil {
			t.Fatalf("ListSpaces failed: %v", err)
		}

		listResp := resp.(*pb.ListSpacesResponse)
		// At least 4 spaces (1 default + 3 from CreateMultipleSpaces)
		if len(listResp.Spaces) < 4 {
			t.Errorf("Expected at least 4 spaces, got %d", len(listResp.Spaces))
		}
	})
}

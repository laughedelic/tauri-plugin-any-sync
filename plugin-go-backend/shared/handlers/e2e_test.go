package handlers

// End-to-End Integration Tests
//
// These tests use global state (globalState variable) and can interfere with other
// tests when run in parallel. They are designed to test the complete system lifecycle
// and validate that data persists correctly across Init/Shutdown cycles.
//
// Running E2E tests:
//   - Separately:  go test ./handlers -run TestE2E -v
//   - Via task:    task shared:test-e2e  (from plugin-go-backend directory)
//
// Regular tests (excluding E2E):
//   - Via task:    task shared:test  (default, excludes E2E tests)
//
// All E2E tests pass when run as a group. They validate:
//   1. Full lifecycle (Init → CreateSpace → CreateDocument → Shutdown)
//   2. Data persistence across restarts
//   3. Multiple spaces with documents
//   4. Error handling (operations before init, invalid IDs, etc.)
//   5. Document versioning

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	pb "anysync-backend/shared/proto/syncspace/v1"
)

// TestE2E_FullLifecycle tests the complete lifecycle: Init → CreateSpace → CreateDocument → GetDocument → Shutdown
func TestE2E_FullLifecycle(t *testing.T) {
	tempDir := t.TempDir()
	dataDir := filepath.Join(tempDir, "test_data")
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		t.Fatalf("failed to create data dir: %v", err)
	}

	ctx := context.Background()

	// Step 1: Initialize
	t.Log("Step 1: Initialize")
	initReq := &pb.InitRequest{
		DataDir:   dataDir,
		NetworkId: "test-network-e2e",
		DeviceId:  "test-device-e2e",
		Config:    map[string]string{"test": "e2e"},
	}
	initResp, err := Init(ctx, initReq)
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	if !initResp.(*pb.InitResponse).Success {
		t.Fatal("Expected Init success to be true")
	}

	// Step 2: Create Space
	t.Log("Step 2: Create Space")
	createSpaceReq := &pb.CreateSpaceRequest{
		Name: "E2E Test Space",
		Metadata: map[string]string{
			"test":    "e2e",
			"purpose": "full lifecycle test",
		},
	}
	createSpaceResp, err := CreateSpace(ctx, createSpaceReq)
	if err != nil {
		t.Fatalf("CreateSpace failed: %v", err)
	}
	spaceID := createSpaceResp.(*pb.CreateSpaceResponse).SpaceId
	if spaceID == "" {
		t.Fatal("Expected non-empty space ID")
	}

	// Step 3: Create Document
	t.Log("Step 3: Create Document")
	createDocReq := &pb.CreateDocumentRequest{
		SpaceId: spaceID,
		Data:    []byte("E2E test document content"),
		Metadata: map[string]string{
			"title": "E2E Test Document",
			"type":  "test",
		},
	}
	createDocResp, err := CreateDocument(ctx, createDocReq)
	if err != nil {
		t.Fatalf("CreateDocument failed: %v", err)
	}
	documentID := createDocResp.(*pb.CreateDocumentResponse).DocumentId
	if documentID == "" {
		t.Fatal("Expected non-empty document ID")
	}

	// Step 4: Get Document
	t.Log("Step 4: Get Document")
	getDocReq := &pb.GetDocumentRequest{
		SpaceId:    spaceID,
		DocumentId: documentID,
	}
	getDocResp, err := GetDocument(ctx, getDocReq)
	if err != nil {
		t.Fatalf("GetDocument failed: %v", err)
	}
	if !getDocResp.(*pb.GetDocumentResponse).Found {
		t.Fatal("Expected document to be found")
	}
	doc := getDocResp.(*pb.GetDocumentResponse).Document
	if string(doc.Data) != "E2E test document content" {
		t.Errorf("Expected document content 'E2E test document content', got '%s'", string(doc.Data))
	}

	// Step 5: List Spaces
	t.Log("Step 5: List Spaces")
	listSpacesReq := &pb.ListSpacesRequest{}
	listSpacesResp, err := ListSpaces(ctx, listSpacesReq)
	if err != nil {
		t.Fatalf("ListSpaces failed: %v", err)
	}
	if len(listSpacesResp.(*pb.ListSpacesResponse).Spaces) == 0 {
		t.Fatal("Expected at least one space")
	}

	// Step 6: List Documents
	t.Log("Step 6: List Documents")
	listDocsReq := &pb.ListDocumentsRequest{SpaceId: spaceID}
	listDocsResp, err := ListDocuments(ctx, listDocsReq)
	if err != nil {
		t.Fatalf("ListDocuments failed: %v", err)
	}
	if len(listDocsResp.(*pb.ListDocumentsResponse).Documents) == 0 {
		t.Fatal("Expected at least one document")
	}

	// Step 7: Shutdown
	t.Log("Step 7: Shutdown")
	shutdownReq := &pb.ShutdownRequest{}
	shutdownResp, err := Shutdown(ctx, shutdownReq)
	if err != nil {
		t.Fatalf("Shutdown failed: %v", err)
	}
	if !shutdownResp.(*pb.ShutdownResponse).Success {
		t.Fatal("Expected Shutdown success to be true")
	}
}

// TestE2E_Persistence tests that data survives Init → Shutdown → Init cycles
func TestE2E_Persistence(t *testing.T) {
	tempDir := t.TempDir()
	dataDir := filepath.Join(tempDir, "test_data")
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		t.Fatalf("failed to create data dir: %v", err)
	}

	ctx := context.Background()
	var spaceID, documentID string

	// First session: Create space and document
	t.Log("Session 1: Create data")
	{
		initReq := &pb.InitRequest{
			DataDir:   dataDir,
			NetworkId: "test-network-persist",
			DeviceId:  "test-device-persist",
		}
		_, err := Init(ctx, initReq)
		if err != nil {
			t.Fatalf("Init failed: %v", err)
		}

		createSpaceReq := &pb.CreateSpaceRequest{
			Name:     "Persistence Test Space",
			Metadata: map[string]string{"test": "persistence"},
		}
		createSpaceResp, err := CreateSpace(ctx, createSpaceReq)
		if err != nil {
			t.Fatalf("CreateSpace failed: %v", err)
		}
		spaceID = createSpaceResp.(*pb.CreateSpaceResponse).SpaceId

		createDocReq := &pb.CreateDocumentRequest{
			SpaceId:  spaceID,
			Data:     []byte("Persistent document content"),
			Metadata: map[string]string{"title": "Persistent Doc"},
		}
		createDocResp, err := CreateDocument(ctx, createDocReq)
		if err != nil {
			t.Fatalf("CreateDocument failed: %v", err)
		}
		documentID = createDocResp.(*pb.CreateDocumentResponse).DocumentId

		_, err = Shutdown(ctx, &pb.ShutdownRequest{})
		if err != nil {
			t.Fatalf("Shutdown failed: %v", err)
		}
	}

	// Wait a bit to ensure filesystem sync
	time.Sleep(100 * time.Millisecond)

	// Second session: Verify data survived
	t.Log("Session 2: Verify persistence")
	{
		initReq := &pb.InitRequest{
			DataDir:   dataDir,
			NetworkId: "test-network-persist",
			DeviceId:  "test-device-persist",
		}
		_, err := Init(ctx, initReq)
		if err != nil {
			t.Fatalf("Init failed after restart: %v", err)
		}
		defer Shutdown(ctx, &pb.ShutdownRequest{})

		// Verify space still exists
		listSpacesReq := &pb.ListSpacesRequest{}
		listSpacesResp, err := ListSpaces(ctx, listSpacesReq)
		if err != nil {
			t.Fatalf("ListSpaces failed: %v", err)
		}
		spaces := listSpacesResp.(*pb.ListSpacesResponse).Spaces
		if len(spaces) == 0 {
			t.Fatal("Expected space to persist, got 0 spaces")
		}

		found := false
		for _, space := range spaces {
			if space.SpaceId == spaceID {
				found = true
				if space.Name != "Persistence Test Space" {
					t.Errorf("Expected space name 'Persistence Test Space', got '%s'", space.Name)
				}
				break
			}
		}
		if !found {
			t.Error("Expected space to be found after restart")
		}

		// Verify document still exists
		getDocReq := &pb.GetDocumentRequest{
			SpaceId:    spaceID,
			DocumentId: documentID,
		}
		getDocResp, err := GetDocument(ctx, getDocReq)
		if err != nil {
			t.Fatalf("GetDocument failed: %v", err)
		}
		getResp := getDocResp.(*pb.GetDocumentResponse)
		if !getResp.Found {
			t.Fatalf("Expected document to persist (spaceID=%s, docID=%s)", spaceID, documentID)
		}
		doc := getResp.Document
		if string(doc.Data) != "Persistent document content" {
			t.Errorf("Expected 'Persistent document content', got '%s'", string(doc.Data))
		}
	}
}

// TestE2E_MultipleSpaces tests creating and managing multiple spaces with documents
func TestE2E_MultipleSpaces(t *testing.T) {
	tempDir := t.TempDir()
	dataDir := filepath.Join(tempDir, "test_data")
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		t.Fatalf("failed to create data dir: %v", err)
	}

	ctx := context.Background()

	initReq := &pb.InitRequest{
		DataDir:   dataDir,
		NetworkId: "test-network-multi",
		DeviceId:  "test-device-multi",
	}
	_, err := Init(ctx, initReq)
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	defer Shutdown(ctx, &pb.ShutdownRequest{})

	// Create multiple spaces
	spaceCount := 3
	spaceIDs := make([]string, spaceCount)
	for i := 0; i < spaceCount; i++ {
		createSpaceReq := &pb.CreateSpaceRequest{
			Name: fmt.Sprintf("Multi-Space %d", i),
			Metadata: map[string]string{
				"index": fmt.Sprintf("%d", i),
			},
		}
		createSpaceResp, err := CreateSpace(ctx, createSpaceReq)
		if err != nil {
			t.Fatalf("CreateSpace %d failed: %v", i, err)
		}
		spaceIDs[i] = createSpaceResp.(*pb.CreateSpaceResponse).SpaceId
	}

	// Create documents in each space
	for i, spaceID := range spaceIDs {
		for j := 0; j < 2; j++ {
			createDocReq := &pb.CreateDocumentRequest{
				SpaceId: spaceID,
				Data:    []byte(fmt.Sprintf("Space %d, Doc %d", i, j)),
				Metadata: map[string]string{
					"space": fmt.Sprintf("%d", i),
					"doc":   fmt.Sprintf("%d", j),
				},
			}
			_, err := CreateDocument(ctx, createDocReq)
			if err != nil {
				t.Fatalf("CreateDocument failed for space %d, doc %d: %v", i, j, err)
			}
		}
	}

	// Verify each space has its documents
	for i, spaceID := range spaceIDs {
		listDocsReq := &pb.ListDocumentsRequest{SpaceId: spaceID}
		listDocsResp, err := ListDocuments(ctx, listDocsReq)
		if err != nil {
			t.Fatalf("ListDocuments failed for space %d: %v", i, err)
		}
		docs := listDocsResp.(*pb.ListDocumentsResponse).Documents
		if len(docs) != 2 {
			t.Errorf("Expected 2 documents in space %d, got %d", i, len(docs))
		}
	}

	// Verify total space count
	listSpacesReq := &pb.ListSpacesRequest{}
	listSpacesResp, err := ListSpaces(ctx, listSpacesReq)
	if err != nil {
		t.Fatalf("ListSpaces failed: %v", err)
	}
	if len(listSpacesResp.(*pb.ListSpacesResponse).Spaces) != spaceCount {
		t.Errorf("Expected %d spaces, got %d", spaceCount, len(listSpacesResp.(*pb.ListSpacesResponse).Spaces))
	}
}

// TestE2E_ErrorHandling tests various error conditions
func TestE2E_ErrorHandling(t *testing.T) {
	tempDir := t.TempDir()
	dataDir := filepath.Join(tempDir, "test_data")
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		t.Fatalf("failed to create data dir: %v", err)
	}

	ctx := context.Background()

	t.Run("Operations before Init", func(t *testing.T) {
		// Ensure clean state
		Shutdown(ctx, &pb.ShutdownRequest{})

		// Try creating space before init
		_, err := CreateSpace(ctx, &pb.CreateSpaceRequest{Name: "Test"})
		if err == nil {
			t.Error("Expected error when creating space before Init")
		}
	})

	t.Run("Double Init", func(t *testing.T) {
		// First init
		initReq := &pb.InitRequest{
			DataDir:   dataDir,
			NetworkId: "test-network-error",
			DeviceId:  "test-device-error",
		}
		_, err := Init(ctx, initReq)
		if err != nil {
			t.Fatalf("First Init failed: %v", err)
		}

		// Try second init
		_, err = Init(ctx, initReq)
		if err == nil {
			t.Error("Expected error on double Init")
		}

		Shutdown(ctx, &pb.ShutdownRequest{})
	})

	t.Run("Invalid Space ID", func(t *testing.T) {
		initReq := &pb.InitRequest{
			DataDir:   dataDir,
			NetworkId: "test-network-error",
			DeviceId:  "test-device-error",
		}
		_, err := Init(ctx, initReq)
		if err != nil {
			t.Fatalf("Init failed: %v", err)
		}
		defer Shutdown(ctx, &pb.ShutdownRequest{})

		// Try to create document in nonexistent space
		_, err = CreateDocument(ctx, &pb.CreateDocumentRequest{
			SpaceId: "invalid-space-id-12345",
			Data:    []byte("test"),
		})
		if err == nil {
			t.Error("Expected error for invalid space ID")
		}
	})

	t.Run("Shutdown before Init", func(t *testing.T) {
		// Ensure clean state
		Shutdown(ctx, &pb.ShutdownRequest{})

		// Try shutdown before init
		_, err := Shutdown(ctx, &pb.ShutdownRequest{})
		if err == nil {
			t.Error("Expected error when shutting down before Init")
		}
	})
}

// TestE2E_DocumentVersioning tests that document updates create new versions
func TestE2E_DocumentVersioning(t *testing.T) {
	tempDir := t.TempDir()
	dataDir := filepath.Join(tempDir, "test_data")
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		t.Fatalf("failed to create data dir: %v", err)
	}

	ctx := context.Background()

	initReq := &pb.InitRequest{
		DataDir:   dataDir,
		NetworkId: "test-network-version",
		DeviceId:  "test-device-version",
	}
	_, err := Init(ctx, initReq)
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	defer Shutdown(ctx, &pb.ShutdownRequest{})

	// Create space
	createSpaceReq := &pb.CreateSpaceRequest{Name: "Version Test Space"}
	createSpaceResp, err := CreateSpace(ctx, createSpaceReq)
	if err != nil {
		t.Fatalf("CreateSpace failed: %v", err)
	}
	spaceID := createSpaceResp.(*pb.CreateSpaceResponse).SpaceId

	// Create document
	createDocReq := &pb.CreateDocumentRequest{
		SpaceId: spaceID,
		Data:    []byte("version 1"),
	}
	createDocResp, err := CreateDocument(ctx, createDocReq)
	if err != nil {
		t.Fatalf("CreateDocument failed: %v", err)
	}
	documentID := createDocResp.(*pb.CreateDocumentResponse).DocumentId

	// Get initial document
	getDocReq := &pb.GetDocumentRequest{
		SpaceId:    spaceID,
		DocumentId: documentID,
	}
	getDocResp, err := GetDocument(ctx, getDocReq)
	if err != nil {
		t.Fatalf("GetDocument failed: %v", err)
	}
	doc := getDocResp.(*pb.GetDocumentResponse).Document
	if string(doc.Data) != "version 1" {
		t.Errorf("Expected 'version 1', got '%s'", string(doc.Data))
	}

	// Update document multiple times
	for i := 2; i <= 5; i++ {
		updateDocReq := &pb.UpdateDocumentRequest{
			SpaceId:    spaceID,
			DocumentId: documentID,
			Data:       []byte(fmt.Sprintf("version %d", i)),
		}
		_, err := UpdateDocument(ctx, updateDocReq)
		if err != nil {
			t.Fatalf("UpdateDocument (v%d) failed: %v", i, err)
		}

		// Verify update
		getDocResp, err := GetDocument(ctx, getDocReq)
		if err != nil {
			t.Fatalf("GetDocument (v%d) failed: %v", i, err)
		}
		doc := getDocResp.(*pb.GetDocumentResponse).Document
		expected := fmt.Sprintf("version %d", i)
		if string(doc.Data) != expected {
			t.Errorf("Expected '%s', got '%s'", expected, string(doc.Data))
		}
	}
}

package handlers

import (
	"context"
	"testing"

	pb "anysync-backend/shared/proto/syncspace/v1"

	"google.golang.org/protobuf/proto"
)

// TestUnit_DocumentHandlers_NotInitialized tests error handling when handlers are called before Init.
func TestUnit_DocumentHandlers_NotInitialized(t *testing.T) {
	resetGlobalState()

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

func TestUnit_Documents_CreateDocumentNotInitialized(t *testing.T) {
	resetGlobalState()

	req := &pb.CreateDocumentRequest{
		SpaceId:    "space1",
		DocumentId: "doc1",
		Collection: "notes",
		Data:       []byte("test data"),
	}

	_, err := CreateDocument(context.Background(), req)
	if err == nil {
		t.Fatal("Expected error when not initialized")
	}
}

func TestUnit_Documents_GetDocumentNotInitialized(t *testing.T) {
	resetGlobalState()

	req := &pb.GetDocumentRequest{
		SpaceId:    "space1",
		DocumentId: "doc1",
	}

	_, err := GetDocument(context.Background(), req)
	if err == nil {
		t.Fatal("Expected error when not initialized")
	}
}

func TestUnit_Documents_GetDocumentNotFound(t *testing.T) {
	resetGlobalState()
	initReq := &pb.InitRequest{
		DataDir:   t.TempDir(),
		NetworkId: "test-network",
		DeviceId:  "test-device",
	}
	Init(context.Background(), initReq)
	t.Cleanup(func() {
		Shutdown(context.Background(), &pb.ShutdownRequest{})
	})

	req := &pb.GetDocumentRequest{
		SpaceId:    "space1",
		DocumentId: "nonexistent",
	}

	resp, err := GetDocument(context.Background(), req)
	if err != nil {
		t.Fatalf("GetDocument failed: %v", err)
	}

	getResp := resp.(*pb.GetDocumentResponse)
	if getResp.Found {
		t.Error("Expected Found=false for nonexistent document")
	}
}

func TestUnit_Documents_UpdateDocumentNotInitialized(t *testing.T) {
	resetGlobalState()

	req := &pb.UpdateDocumentRequest{
		SpaceId:    "space1",
		DocumentId: "doc1",
		Data:       []byte("updated data"),
	}

	_, err := UpdateDocument(context.Background(), req)
	if err == nil {
		t.Fatal("Expected error when not initialized")
	}
}

func TestUnit_Documents_DeleteDocumentNotInitialized(t *testing.T) {
	resetGlobalState()

	req := &pb.DeleteDocumentRequest{
		SpaceId:    "space1",
		DocumentId: "doc1",
	}

	_, err := DeleteDocument(context.Background(), req)
	if err == nil {
		t.Fatal("Expected error when not initialized")
	}
}

func TestUnit_Documents_DeleteDocumentNotFound(t *testing.T) {
	resetGlobalState()
	initReq := &pb.InitRequest{
		DataDir:   t.TempDir(),
		NetworkId: "test-network",
		DeviceId:  "test-device",
	}
	Init(context.Background(), initReq)
	t.Cleanup(func() {
		Shutdown(context.Background(), &pb.ShutdownRequest{})
	})

	// Test with invalid space ID - should error
	req := &pb.DeleteDocumentRequest{
		SpaceId:    "invalid-space-id",
		DocumentId: "nonexistent",
	}

	_, err := DeleteDocument(context.Background(), req)
	if err == nil {
		t.Error("Expected error for invalid space ID")
	}
}

func TestUnit_Documents_ListDocumentsNotInitialized(t *testing.T) {
	resetGlobalState()

	req := &pb.ListDocumentsRequest{
		SpaceId: "space1",
	}

	_, err := ListDocuments(context.Background(), req)
	if err == nil {
		t.Fatal("Expected error when not initialized")
	}
}

func TestUnit_Documents_ListDocumentsEmpty(t *testing.T) {
	resetGlobalState()
	initReq := &pb.InitRequest{
		DataDir:   t.TempDir(),
		NetworkId: "test-network",
		DeviceId:  "test-device",
	}
	Init(context.Background(), initReq)
	t.Cleanup(func() {
		Shutdown(context.Background(), &pb.ShutdownRequest{})
	})

	req := &pb.ListDocumentsRequest{
		SpaceId: "space1",
	}

	resp, err := ListDocuments(context.Background(), req)
	if err != nil {
		t.Fatalf("ListDocuments failed: %v", err)
	}

	listResp := resp.(*pb.ListDocumentsResponse)
	if len(listResp.Documents) != 0 {
		t.Errorf("Expected empty list, got %d documents", len(listResp.Documents))
	}
}

func TestUnit_Documents_QueryDocumentsNotInitialized(t *testing.T) {
	resetGlobalState()

	req := &pb.QueryDocumentsRequest{
		SpaceId:    "space1",
		Collection: "notes",
	}

	_, err := QueryDocuments(context.Background(), req)
	if err == nil {
		t.Fatal("Expected error when not initialized")
	}
}

func TestUnit_Documents_QueryDocumentsEmpty(t *testing.T) {
	resetGlobalState()
	initReq := &pb.InitRequest{
		DataDir:   t.TempDir(),
		NetworkId: "test-network",
		DeviceId:  "test-device",
	}
	Init(context.Background(), initReq)
	t.Cleanup(func() {
		Shutdown(context.Background(), &pb.ShutdownRequest{})
	})

	req := &pb.QueryDocumentsRequest{
		SpaceId:    "space1",
		Collection: "notes",
	}

	resp, err := QueryDocuments(context.Background(), req)
	if err != nil {
		t.Fatalf("QueryDocuments failed: %v", err)
	}

	queryResp := resp.(*pb.QueryDocumentsResponse)
	if len(queryResp.Documents) != 0 {
		t.Errorf("Expected empty list, got %d documents", len(queryResp.Documents))
	}
}

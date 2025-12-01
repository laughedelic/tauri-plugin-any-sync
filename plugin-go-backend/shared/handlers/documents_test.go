package handlers

import (
	"context"
	"testing"

	pb "anysync-backend/shared/proto/syncspace/v1"
)

func TestCreateDocument_NotInitialized(t *testing.T) {
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

func TestGetDocument_NotInitialized(t *testing.T) {
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

func TestGetDocument_NotFound(t *testing.T) {
	resetGlobalState()
	initReq := &pb.InitRequest{
		DataDir:   "/tmp/test",
		NetworkId: "test-network",
		DeviceId:  "test-device",
	}
	Init(context.Background(), initReq)

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

func TestUpdateDocument_NotInitialized(t *testing.T) {
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

func TestDeleteDocument_NotInitialized(t *testing.T) {
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

func TestDeleteDocument_NotFound(t *testing.T) {
	resetGlobalState()
	initReq := &pb.InitRequest{
		DataDir:   "/tmp/test",
		NetworkId: "test-network",
		DeviceId:  "test-device",
	}
	Init(context.Background(), initReq)

	req := &pb.DeleteDocumentRequest{
		SpaceId:    "space1",
		DocumentId: "nonexistent",
	}

	resp, err := DeleteDocument(context.Background(), req)
	if err != nil {
		t.Fatalf("DeleteDocument failed: %v", err)
	}

	deleteResp := resp.(*pb.DeleteDocumentResponse)
	if deleteResp.Existed {
		t.Error("Expected Existed=false for nonexistent document")
	}
}

func TestListDocuments_NotInitialized(t *testing.T) {
	resetGlobalState()

	req := &pb.ListDocumentsRequest{
		SpaceId: "space1",
	}

	_, err := ListDocuments(context.Background(), req)
	if err == nil {
		t.Fatal("Expected error when not initialized")
	}
}

func TestListDocuments_Empty(t *testing.T) {
	resetGlobalState()
	initReq := &pb.InitRequest{
		DataDir:   "/tmp/test",
		NetworkId: "test-network",
		DeviceId:  "test-device",
	}
	Init(context.Background(), initReq)

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

func TestQueryDocuments_NotInitialized(t *testing.T) {
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

func TestQueryDocuments_Empty(t *testing.T) {
	resetGlobalState()
	initReq := &pb.InitRequest{
		DataDir:   "/tmp/test",
		NetworkId: "test-network",
		DeviceId:  "test-device",
	}
	Init(context.Background(), initReq)

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

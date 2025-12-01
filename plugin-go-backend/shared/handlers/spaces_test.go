package handlers

import (
	"context"
	"testing"

	pb "anysync-backend/shared/proto/syncspace/v1"
)

func TestCreateSpace_NotInitialized(t *testing.T) {
	resetGlobalState()

	req := &pb.CreateSpaceRequest{
		SpaceId: "space1",
		Name:    "Test Space",
	}

	_, err := CreateSpace(context.Background(), req)
	if err == nil {
		t.Fatal("Expected error when not initialized")
	}
}

func TestCreateSpace_Success(t *testing.T) {
	// Initialize first
	resetGlobalState()
	initReq := &pb.InitRequest{
		DataDir:   t.TempDir(),
		NetworkId: "test-network",
		DeviceId:  "test-device",
	}
	Init(context.Background(), initReq)

	req := &pb.CreateSpaceRequest{
		SpaceId: "space1",
		Name:    "Test Space",
		Metadata: map[string]string{
			"type": "personal",
		},
	}

	_, err := CreateSpace(context.Background(), req)
	// Currently returns "not implemented yet" error
	if err == nil {
		t.Fatal("Expected not implemented error")
	}
}

func TestJoinSpace_NotInitialized(t *testing.T) {
	resetGlobalState()

	req := &pb.JoinSpaceRequest{
		SpaceId: "space1",
	}

	_, err := JoinSpace(context.Background(), req)
	if err == nil {
		t.Fatal("Expected error when not initialized")
	}
}

func TestLeaveSpace_NotInitialized(t *testing.T) {
	resetGlobalState()

	req := &pb.LeaveSpaceRequest{
		SpaceId: "space1",
	}

	_, err := LeaveSpace(context.Background(), req)
	if err == nil {
		t.Fatal("Expected error when not initialized")
	}
}

func TestListSpaces_NotInitialized(t *testing.T) {
	resetGlobalState()

	req := &pb.ListSpacesRequest{}

	_, err := ListSpaces(context.Background(), req)
	if err == nil {
		t.Fatal("Expected error when not initialized")
	}
}

func TestListSpaces_Empty(t *testing.T) {
	resetGlobalState()
	initReq := &pb.InitRequest{
		DataDir:   t.TempDir(),
		NetworkId: "test-network",
		DeviceId:  "test-device",
	}
	Init(context.Background(), initReq)

	req := &pb.ListSpacesRequest{}

	resp, err := ListSpaces(context.Background(), req)
	if err != nil {
		t.Fatalf("ListSpaces failed: %v", err)
	}

	listResp := resp.(*pb.ListSpacesResponse)
	if len(listResp.Spaces) != 0 {
		t.Errorf("Expected empty list, got %d spaces", len(listResp.Spaces))
	}
}

func TestDeleteSpace_NotInitialized(t *testing.T) {
	resetGlobalState()

	req := &pb.DeleteSpaceRequest{
		SpaceId: "space1",
	}

	_, err := DeleteSpace(context.Background(), req)
	if err == nil {
		t.Fatal("Expected error when not initialized")
	}
}

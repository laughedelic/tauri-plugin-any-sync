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
	_, err := Init(context.Background(), initReq)
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	req := &pb.CreateSpaceRequest{
		SpaceId: "ref1", // Used as reference name
		Name:    "Test Space",
		Metadata: map[string]string{
			"type": "personal",
		},
	}

	resp, err := CreateSpace(context.Background(), req)
	if err != nil {
		t.Fatalf("CreateSpace failed: %v", err)
	}

	createResp := resp.(*pb.CreateSpaceResponse)
	if createResp.SpaceId == "" {
		t.Fatal("Expected non-empty space ID")
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
	_, err := Init(context.Background(), initReq)
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

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

func TestListSpaces_WithSpaces(t *testing.T) {
	resetGlobalState()
	initReq := &pb.InitRequest{
		DataDir:   t.TempDir(),
		NetworkId: "test-network",
		DeviceId:  "test-device",
	}
	_, err := Init(context.Background(), initReq)
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	// Create multiple spaces
	createReq1 := &pb.CreateSpaceRequest{
		SpaceId:  "ref1",
		Name:     "Space 1",
		Metadata: map[string]string{"type": "work"},
	}
	_, err = CreateSpace(context.Background(), createReq1)
	if err != nil {
		t.Fatalf("CreateSpace 1 failed: %v", err)
	}

	createReq2 := &pb.CreateSpaceRequest{
		SpaceId:  "ref2",
		Name:     "Space 2",
		Metadata: map[string]string{"type": "personal"},
	}
	_, err = CreateSpace(context.Background(), createReq2)
	if err != nil {
		t.Fatalf("CreateSpace 2 failed: %v", err)
	}

	// List spaces
	listReq := &pb.ListSpacesRequest{}
	resp, err := ListSpaces(context.Background(), listReq)
	if err != nil {
		t.Fatalf("ListSpaces failed: %v", err)
	}

	listResp := resp.(*pb.ListSpacesResponse)
	if len(listResp.Spaces) != 2 {
		t.Errorf("Expected 2 spaces, got %d", len(listResp.Spaces))
	}

	// Verify space details
	for _, space := range listResp.Spaces {
		if space.Name != "Space 1" && space.Name != "Space 2" {
			t.Errorf("Unexpected space name: %s", space.Name)
		}
		if space.SpaceId == "" {
			t.Error("Expected non-empty space ID")
		}
		if space.CreatedAt <= 0 {
			t.Error("Expected positive CreatedAt timestamp")
		}
	}
}

func TestDeleteSpace_Success(t *testing.T) {
	resetGlobalState()
	initReq := &pb.InitRequest{
		DataDir:   t.TempDir(),
		NetworkId: "test-network",
		DeviceId:  "test-device",
	}
	_, err := Init(context.Background(), initReq)
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	// Create a space
	createReq := &pb.CreateSpaceRequest{
		SpaceId: "ref1",
		Name:    "Test Space",
	}
	createResp, err := CreateSpace(context.Background(), createReq)
	if err != nil {
		t.Fatalf("CreateSpace failed: %v", err)
	}

	spaceID := createResp.(*pb.CreateSpaceResponse).SpaceId

	// Delete the space
	deleteReq := &pb.DeleteSpaceRequest{
		SpaceId: spaceID,
	}
	deleteResp, err := DeleteSpace(context.Background(), deleteReq)
	if err != nil {
		t.Fatalf("DeleteSpace failed: %v", err)
	}

	if !deleteResp.(*pb.DeleteSpaceResponse).Success {
		t.Error("Expected Success=true")
	}

	// Verify space is gone
	listReq := &pb.ListSpacesRequest{}
	listResp, err := ListSpaces(context.Background(), listReq)
	if err != nil {
		t.Fatalf("ListSpaces failed: %v", err)
	}

	if len(listResp.(*pb.ListSpacesResponse).Spaces) != 0 {
		t.Error("Expected no spaces after deletion")
	}
}

func TestDeleteSpace_NotFound(t *testing.T) {
	resetGlobalState()
	initReq := &pb.InitRequest{
		DataDir:   t.TempDir(),
		NetworkId: "test-network",
		DeviceId:  "test-device",
	}
	_, err := Init(context.Background(), initReq)
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	// Try to delete non-existent space
	deleteReq := &pb.DeleteSpaceRequest{
		SpaceId: "non-existent-space",
	}
	_, err = DeleteSpace(context.Background(), deleteReq)
	if err == nil {
		t.Fatal("Expected error when deleting non-existent space")
	}
}

func TestJoinSpace_NotImplemented(t *testing.T) {
	resetGlobalState()
	initReq := &pb.InitRequest{
		DataDir:   t.TempDir(),
		NetworkId: "test-network",
		DeviceId:  "test-device",
	}
	_, err := Init(context.Background(), initReq)
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	joinReq := &pb.JoinSpaceRequest{
		SpaceId: "some-space",
	}
	_, err = JoinSpace(context.Background(), joinReq)
	if err == nil {
		t.Fatal("Expected JoinSpace to return not implemented error")
	}
}

func TestLeaveSpace_NotImplemented(t *testing.T) {
	resetGlobalState()
	initReq := &pb.InitRequest{
		DataDir:   t.TempDir(),
		NetworkId: "test-network",
		DeviceId:  "test-device",
	}
	_, err := Init(context.Background(), initReq)
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	leaveReq := &pb.LeaveSpaceRequest{
		SpaceId: "some-space",
	}
	_, err = LeaveSpace(context.Background(), leaveReq)
	if err == nil {
		t.Fatal("Expected LeaveSpace to return not implemented error")
	}
}

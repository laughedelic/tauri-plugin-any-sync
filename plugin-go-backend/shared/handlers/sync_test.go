package handlers

import (
	"context"
	"testing"

	pb "anysync-backend/shared/proto/syncspace/v1"
)

func TestStartSync_NotInitialized(t *testing.T) {
	resetGlobalState()

	req := &pb.StartSyncRequest{
		SpaceId: "space1",
	}

	_, err := StartSync(context.Background(), req)
	if err == nil {
		t.Fatal("Expected error when not initialized")
	}
}

func TestPauseSync_NotInitialized(t *testing.T) {
	resetGlobalState()

	req := &pb.PauseSyncRequest{
		SpaceId: "space1",
	}

	_, err := PauseSync(context.Background(), req)
	if err == nil {
		t.Fatal("Expected error when not initialized")
	}
}

func TestGetSyncStatus_NotInitialized(t *testing.T) {
	resetGlobalState()

	req := &pb.GetSyncStatusRequest{
		SpaceId: "space1",
	}

	_, err := GetSyncStatus(context.Background(), req)
	if err == nil {
		t.Fatal("Expected error when not initialized")
	}
}

func TestGetSyncStatus_Empty(t *testing.T) {
	resetGlobalState()
	initReq := &pb.InitRequest{
		DataDir:   t.TempDir(),
		NetworkId: "test-network",
		DeviceId:  "test-device",
	}
	Init(context.Background(), initReq)

	req := &pb.GetSyncStatusRequest{
		SpaceId: "",
	}

	resp, err := GetSyncStatus(context.Background(), req)
	if err != nil {
		t.Fatalf("GetSyncStatus failed: %v", err)
	}

	statusResp := resp.(*pb.GetSyncStatusResponse)
	if len(statusResp.Statuses) != 0 {
		t.Errorf("Expected empty status list, got %d statuses", len(statusResp.Statuses))
	}
}

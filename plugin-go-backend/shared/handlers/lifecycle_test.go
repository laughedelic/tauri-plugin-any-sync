package handlers

import (
	"context"
	"testing"

	pb "anysync-backend/shared/proto/syncspace/v1"
)

func TestInit_Success(t *testing.T) {
	resetGlobalState()

	req := &pb.InitRequest{
		DataDir:   "/tmp/test-data",
		NetworkId: "test-network",
		DeviceId:  "test-device",
		Config: map[string]string{
			"key1": "value1",
			"key2": "value2",
		},
	}

	resp, err := Init(context.Background(), req)
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	initResp := resp.(*pb.InitResponse)
	if !initResp.Success {
		t.Error("Expected success=true")
	}

	globalState.mu.RLock()
	defer globalState.mu.RUnlock()

	if !globalState.initialized {
		t.Error("Expected initialized=true")
	}
	if globalState.dataDir != req.DataDir {
		t.Errorf("Expected dataDir=%s, got %s", req.DataDir, globalState.dataDir)
	}
	if globalState.networkID != req.NetworkId {
		t.Errorf("Expected networkID=%s, got %s", req.NetworkId, globalState.networkID)
	}
	if globalState.deviceID != req.DeviceId {
		t.Errorf("Expected deviceID=%s, got %s", req.DeviceId, globalState.deviceID)
	}
}

func TestInit_AlreadyInitialized(t *testing.T) {
	resetGlobalState()
	globalState.mu.Lock()
	globalState.initialized = true
	globalState.mu.Unlock()

	req := &pb.InitRequest{
		DataDir:   "/tmp/test-data",
		NetworkId: "test-network",
		DeviceId:  "test-device",
	}

	_, err := Init(context.Background(), req)
	if err == nil {
		t.Fatal("Expected error when already initialized")
	}
}

func TestShutdown_Success(t *testing.T) {
	resetGlobalState()
	globalState.mu.Lock()
	globalState.initialized = true
	globalState.dataDir = "/tmp/test"
	globalState.networkID = "test-network"
	globalState.deviceID = "test-device"
	globalState.config = map[string]string{"key": "value"}
	globalState.mu.Unlock()

	req := &pb.ShutdownRequest{}

	resp, err := Shutdown(context.Background(), req)
	if err != nil {
		t.Fatalf("Shutdown failed: %v", err)
	}

	shutdownResp := resp.(*pb.ShutdownResponse)
	if !shutdownResp.Success {
		t.Error("Expected success=true")
	}

	globalState.mu.RLock()
	defer globalState.mu.RUnlock()

	if globalState.initialized {
		t.Error("Expected initialized=false")
	}
	if globalState.dataDir != "" {
		t.Errorf("Expected empty dataDir, got %s", globalState.dataDir)
	}
}

func TestShutdown_NotInitialized(t *testing.T) {
	resetGlobalState()

	req := &pb.ShutdownRequest{}

	_, err := Shutdown(context.Background(), req)
	if err == nil {
		t.Fatal("Expected error when not initialized")
	}
}

func resetGlobalState() {
	globalState.mu.Lock()
	defer globalState.mu.Unlock()

	globalState.initialized = false
	globalState.dataDir = ""
	globalState.networkID = ""
	globalState.deviceID = ""
	globalState.config = nil
}

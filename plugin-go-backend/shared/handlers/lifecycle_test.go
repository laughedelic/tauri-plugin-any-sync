package handlers

import (
	"context"
	"testing"

	pb "anysync-backend/shared/proto/syncspace/v1"
)

func TestUnit_Lifecycle_InitSuccess(t *testing.T) {
	resetGlobalState()

	tmpDir := t.TempDir()
	req := &pb.InitRequest{
		DataDir:   tmpDir,
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

	// Verify account manager is initialized
	if globalState.accountManager == nil {
		t.Fatal("Expected accountManager to be initialized")
	}
	if !globalState.accountManager.HasKeys() {
		t.Fatal("Expected keys to be loaded/generated")
	}
	if globalState.accountManager.GetKeys() == nil {
		t.Fatal("Expected GetKeys() to return non-nil")
	}
}

func TestUnit_Lifecycle_InitAlreadyInitialized(t *testing.T) {
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

func TestUnit_Lifecycle_ShutdownSuccess(t *testing.T) {
	resetGlobalState()

	tmpDir := t.TempDir()

	// First initialize
	initReq := &pb.InitRequest{
		DataDir:   tmpDir,
		NetworkId: "test-network",
		DeviceId:  "test-device",
	}
	_, err := Init(context.Background(), initReq)
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	// Verify keys are loaded
	globalState.mu.RLock()
	if globalState.accountManager == nil || !globalState.accountManager.HasKeys() {
		globalState.mu.RUnlock()
		t.Fatal("Keys should be loaded after Init")
	}
	globalState.mu.RUnlock()

	// Now shutdown
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

	// Verify keys are cleared
	if globalState.accountManager != nil {
		t.Error("Expected accountManager to be nil after Shutdown")
	}
}

func TestUnit_Lifecycle_ShutdownNotInitialized(t *testing.T) {
	resetGlobalState()

	req := &pb.ShutdownRequest{}

	_, err := Shutdown(context.Background(), req)
	if err == nil {
		t.Fatal("Expected error when not initialized")
	}
}

func TestUnit_Lifecycle_InitKeyPersistenceAcrossRestarts(t *testing.T) {
	resetGlobalState()

	tmpDir := t.TempDir()
	req := &pb.InitRequest{
		DataDir:   tmpDir,
		NetworkId: "test-network",
		DeviceId:  "test-device",
	}

	// First Init: generates and stores keys
	resp1, err := Init(context.Background(), req)
	if err != nil {
		t.Fatalf("First Init failed: %v", err)
	}
	if !resp1.(*pb.InitResponse).Success {
		t.Fatal("First Init should succeed")
	}

	// Get the peer ID from first initialization
	globalState.mu.RLock()
	firstPeerId := globalState.accountManager.GetKeys().PeerId
	globalState.mu.RUnlock()

	// Shutdown
	_, err = Shutdown(context.Background(), &pb.ShutdownRequest{})
	if err != nil {
		t.Fatalf("Shutdown failed: %v", err)
	}

	// Second Init: should load existing keys
	resp2, err := Init(context.Background(), req)
	if err != nil {
		t.Fatalf("Second Init failed: %v", err)
	}
	if !resp2.(*pb.InitResponse).Success {
		t.Fatal("Second Init should succeed")
	}

	// Verify the peer ID is the same
	globalState.mu.RLock()
	secondPeerId := globalState.accountManager.GetKeys().PeerId
	globalState.mu.RUnlock()

	if firstPeerId != secondPeerId {
		t.Fatalf("PeerId should persist across restarts: first=%s, second=%s", firstPeerId, secondPeerId)
	}
}

func resetGlobalState() {
	globalState.mu.Lock()
	defer globalState.mu.Unlock()

	// Close DocumentManager if it exists
	if globalState.documentManager != nil {
		globalState.documentManager.Close()
		globalState.documentManager = nil
	}

	// Close SpaceManager if it exists
	if globalState.spaceManager != nil {
		globalState.spaceManager.Close()
		globalState.spaceManager = nil
	}

	// Close EventManager if it exists
	if globalState.eventManager != nil {
		globalState.eventManager.Close()
		globalState.eventManager = nil
	}

	// Clear keys from memory if accountManager exists
	if globalState.accountManager != nil {
		globalState.accountManager.ClearKeys()
		globalState.accountManager = nil
	}

	globalState.initialized = false
	globalState.dataDir = ""
	globalState.networkID = ""
	globalState.deviceID = ""
	globalState.config = nil
}

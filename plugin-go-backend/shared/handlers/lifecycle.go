// Package handlers provides handler functions for SyncSpace operations.
package handlers

import (
	"context"
	"fmt"
	"sync"

	pb "anysync-backend/shared/proto/syncspace/v1"
	"google.golang.org/protobuf/proto"
)

// State holds the global state for the SyncSpace backend.
type State struct {
	mu          sync.RWMutex
	dataDir     string
	networkID   string
	deviceID    string
	config      map[string]string
	initialized bool
}

var globalState = &State{}

// Init handles the Init operation.
func Init(ctx context.Context, req proto.Message) (proto.Message, error) {
	initReq := req.(*pb.InitRequest)

	globalState.mu.Lock()
	defer globalState.mu.Unlock()

	if globalState.initialized {
		return nil, fmt.Errorf("already initialized")
	}

	// Store configuration
	globalState.dataDir = initReq.DataDir
	globalState.networkID = initReq.NetworkId
	globalState.deviceID = initReq.DeviceId
	globalState.config = initReq.Config
	globalState.initialized = true

	// TODO: Initialize Any-Sync components (SpaceService, ObjectTree)

	return &pb.InitResponse{Success: true}, nil
}

// Shutdown handles the Shutdown operation.
func Shutdown(ctx context.Context, req proto.Message) (proto.Message, error) {
	globalState.mu.Lock()
	defer globalState.mu.Unlock()

	if !globalState.initialized {
		return nil, fmt.Errorf("not initialized")
	}

	// TODO: Cleanup Any-Sync components

	globalState.initialized = false
	globalState.dataDir = ""
	globalState.networkID = ""
	globalState.deviceID = ""
	globalState.config = nil

	return &pb.ShutdownResponse{Success: true}, nil
}

// ensureInitialized checks if the system is initialized.
func ensureInitialized() error {
	globalState.mu.RLock()
	defer globalState.mu.RUnlock()

	if !globalState.initialized {
		return fmt.Errorf("not initialized: call Init first")
	}
	return nil
}

// Package handlers provides handler functions for SyncSpace operations.
package handlers

import (
	"context"
	"fmt"
	"sync"

	"anysync-backend/shared/anysync"
	pb "anysync-backend/shared/proto/syncspace/v1"

	"google.golang.org/protobuf/proto"
)

// State holds the global state for the SyncSpace backend.
type State struct {
	mu              sync.RWMutex
	dataDir         string
	networkID       string
	deviceID        string
	config          map[string]string
	accountManager  *anysync.AccountManager
	spaceManager    *anysync.SpaceManager
	documentManager *anysync.DocumentManager
	eventManager    *anysync.EventManager
	initialized     bool
}

var globalState = &State{}

// Init handles the Init operation.
func Init(ctx context.Context, req proto.Message) (proto.Message, error) {
	initReq := req.(*pb.InitRequest)

	globalState.mu.Lock()
	defer globalState.mu.Unlock()

	// If already initialized with same data directory, return success (idempotent)
	if globalState.initialized {
		// FIXME: WHAT IF other parameters differ?
		if globalState.dataDir == initReq.DataDir {
			return &pb.InitResponse{Success: true}, nil
		}
		return nil, fmt.Errorf("already initialized with different data directory")
	}

	// Store configuration
	globalState.dataDir = initReq.DataDir
	globalState.networkID = initReq.NetworkId
	globalState.deviceID = initReq.DeviceId
	globalState.config = initReq.Config

	// Initialize AccountManager
	globalState.accountManager = anysync.NewAccountManager(initReq.DataDir)

	// Check if keys already exist on disk
	if globalState.accountManager.KeysExist() {
		// Load existing keys
		if err := globalState.accountManager.LoadKeys(); err != nil {
			return nil, fmt.Errorf("failed to load existing keys: %w", err)
		}
	} else {
		// Generate new keys
		if err := globalState.accountManager.GenerateKeys(); err != nil {
			return nil, fmt.Errorf("failed to generate keys: %w", err)
		}

		// Store keys to disk
		if err := globalState.accountManager.StoreKeys(); err != nil {
			return nil, fmt.Errorf("failed to store keys: %w", err)
		}
	}

	// Verify keys are loaded
	if !globalState.accountManager.HasKeys() {
		return nil, fmt.Errorf("keys not loaded after initialization")
	}

	// Initialize EventManager
	globalState.eventManager = anysync.NewEventManager()

	// Initialize SpaceManager with loaded keys
	spaceManager, err := anysync.NewSpaceManager(initReq.DataDir, globalState.accountManager.GetKeys(), globalState.eventManager)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize space manager: %w", err)
	}
	globalState.spaceManager = spaceManager

	// Initialize DocumentManager
	documentManager, err := anysync.NewDocumentManager(globalState.spaceManager, globalState.accountManager.GetKeys(), globalState.eventManager)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize document manager: %w", err)
	}
	globalState.documentManager = documentManager

	globalState.initialized = true

	return &pb.InitResponse{Success: true}, nil
}

// Shutdown handles the Shutdown operation.
func Shutdown(ctx context.Context, req proto.Message) (proto.Message, error) {
	globalState.mu.Lock()
	defer globalState.mu.Unlock()

	if !globalState.initialized {
		return nil, fmt.Errorf("not initialized")
	}

	// Close DocumentManager
	if globalState.documentManager != nil {
		if err := globalState.documentManager.Close(); err != nil {
			// Log error but continue shutdown
			fmt.Printf("Warning: failed to close document manager: %v\n", err)
		}
		globalState.documentManager = nil
	}

	// Close SpaceManager (closes all space storages)
	if globalState.spaceManager != nil {
		if err := globalState.spaceManager.Close(); err != nil {
			// Log error but continue shutdown
			fmt.Printf("Warning: failed to close space manager: %v\n", err)
		}
		globalState.spaceManager = nil
	}

	// Close EventManager
	if globalState.eventManager != nil {
		if err := globalState.eventManager.Close(); err != nil {
			// Log error but continue shutdown
			fmt.Printf("Warning: failed to close event manager: %v\n", err)
		}
		globalState.eventManager = nil
	}

	// Clear keys from memory
	if globalState.accountManager != nil {
		globalState.accountManager.ClearKeys()
		globalState.accountManager = nil
	}

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

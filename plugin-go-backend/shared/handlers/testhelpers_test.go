package handlers

// Test Helpers - Integration Test Pattern
//
// This file provides a cleaner pattern for writing integration tests.
// The TestContext helper manages the Init/Shutdown lifecycle and provides
// convenience methods for common test operations.
//
// Pattern Benefits:
//   - Automatic lifecycle management (Init/Shutdown)
//   - Default test space created automatically
//   - Convenience methods reduce boilerplate
//   - Single Init/Shutdown per test (efficient)
//   - Cleanup registered automatically via t.Cleanup()
//
// Example:
//   func TestMyFeature(t *testing.T) {
//       tc := SetupIntegrationTest(t)
//       
//       // Create document in default space
//       docID := tc.CreateDocument([]byte("test data"), nil)
//       
//       // Or use context for custom operations
//       resp, err := GetDocument(tc.Context(), &pb.GetDocumentRequest{
//           SpaceId: tc.SpaceID(),
//           DocumentId: docID,
//       })
//   }
//
// See integration_refactored_test.go for complete examples.

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	pb "anysync-backend/shared/proto/syncspace/v1"
)

// TestContext holds shared test state for integration tests.
// This enables running multiple integration tests within a single Init/Shutdown cycle,
// which better matches real-world usage and avoids test isolation issues.
type TestContext struct {
	t       *testing.T
	ctx     context.Context
	dataDir string
	spaceID string // Default test space
}

// SetupIntegrationTest creates a test context with initialized system.
// Returns a TestContext with:
// - Initialized global state
// - A default test space
// - Cleanup registered via t.Cleanup()
//
// Usage:
//
//	func TestMyFeature(t *testing.T) {
//	    tc := SetupIntegrationTest(t)
//	    // Use tc.spaceID for operations
//	    // System will be automatically shut down after test
//	}
func SetupIntegrationTest(t *testing.T) *TestContext {
	t.Helper()

	// Ensure clean state before initialization
	resetGlobalState()

	tempDir := t.TempDir()
	dataDir := filepath.Join(tempDir, "test_data")
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		t.Fatalf("failed to create data dir: %v", err)
	}

	ctx := context.Background()

	// Initialize system
	initReq := &pb.InitRequest{
		DataDir:   dataDir,
		NetworkId: "test-network",
		DeviceId:  "test-device",
	}
	_, err := Init(ctx, initReq)
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	// Create a default test space
	createSpaceReq := &pb.CreateSpaceRequest{
		Name:     "Test Space",
		Metadata: map[string]string{"purpose": "integration-testing"},
	}
	createSpaceResp, err := CreateSpace(ctx, createSpaceReq)
	if err != nil {
		t.Fatalf("CreateSpace failed: %v", err)
	}
	spaceID := createSpaceResp.(*pb.CreateSpaceResponse).SpaceId

	// Register cleanup
	t.Cleanup(func() {
		Shutdown(ctx, &pb.ShutdownRequest{})
	})

	return &TestContext{
		t:       t,
		ctx:     ctx,
		dataDir: dataDir,
		spaceID: spaceID,
	}
}

// Context returns the test context.
func (tc *TestContext) Context() context.Context {
	return tc.ctx
}

// DataDir returns the test data directory.
func (tc *TestContext) DataDir() string {
	return tc.dataDir
}

// SpaceID returns the default test space ID.
func (tc *TestContext) SpaceID() string {
	return tc.spaceID
}

// CreateSpace creates a new space and returns its ID.
func (tc *TestContext) CreateSpace(name string, metadata map[string]string) string {
	tc.t.Helper()

	req := &pb.CreateSpaceRequest{
		Name:     name,
		Metadata: metadata,
	}
	resp, err := CreateSpace(tc.ctx, req)
	if err != nil {
		tc.t.Fatalf("CreateSpace failed: %v", err)
	}
	return resp.(*pb.CreateSpaceResponse).SpaceId
}

// CreateDocument creates a document in the default space.
func (tc *TestContext) CreateDocument(data []byte, metadata map[string]string) string {
	tc.t.Helper()

	req := &pb.CreateDocumentRequest{
		SpaceId:  tc.spaceID,
		Data:     data,
		Metadata: metadata,
	}
	resp, err := CreateDocument(tc.ctx, req)
	if err != nil {
		tc.t.Fatalf("CreateDocument failed: %v", err)
	}
	return resp.(*pb.CreateDocumentResponse).DocumentId
}

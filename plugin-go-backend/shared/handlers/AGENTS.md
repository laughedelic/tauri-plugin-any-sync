# Testing Guide

## Overview

The handlers package contains 97 tests covering all backend functionality. Tests use a singleton global state pattern with proper isolation via `resetGlobalState()` and `t.Cleanup()`.

## Test Organization

**Unit Tests** (`TestUnit_*`) - Individual handlers in isolation:
- `dispatcher_test.go` - Command routing
- `lifecycle_test.go` - Init/Shutdown
- `documents_test.go` - Document error cases, not-initialized scenarios
- `spaces_test.go` - Space validation
- `sync_test.go` - Sync stubs
- `events_test.go` - Event system

**Integration Tests** (`TestIntegration_*`) - Full backend stack with TestContext helper:
- `integration_test.go` - Document CRUD, space management with real storage

**E2E Tests** (`TestE2E_*`) - Complete lifecycle scenarios:
- `e2e_test.go` - Full lifecycle, persistence, versioning, error handling

## Naming Conventions

Tests follow Go idiomatic prefix naming:
- **`TestUnit_*`** - Fast, isolated unit tests (error cases, validation)
- **`TestIntegration_*`** - Tests with real backend dependencies
- **`TestE2E_*`** - Complete system lifecycle scenarios

**Benefits:**
- Clear intent and categorization
- Selective test runs: `go test -run ^TestIntegration_`
- IDE-friendly (all Go editors understand prefixes)

## Writing Tests

### Integration Tests (Recommended Pattern)

Use the `TestContext` helper from `testhelpers_test.go` for automatic Init/Shutdown:

```go
func TestIntegration_MyFeature(t *testing.T) {
    tc := SetupIntegrationTest(t)  // Auto Init/Shutdown + default space
    
    t.Run("CreateDocument", func(t *testing.T) {
        docID := tc.CreateDocument([]byte("data"), map[string]string{"key": "val"})
        // tc.Context(), tc.SpaceID(), tc.CreateSpace() available
    })
}
```

### Unit Tests

Manual Init/Shutdown with proper cleanup:

```go
func TestUnit_MyHandler_Success(t *testing.T) {
    resetGlobalState()
    
    initReq := &pb.InitRequest{DataDir: t.TempDir(), NetworkId: "test", DeviceId: "test"}
    _, err := Init(context.Background(), initReq)
    if err != nil {
        t.Fatalf("Init failed: %v", err)
    }
    t.Cleanup(func() {
        Shutdown(context.Background(), &pb.ShutdownRequest{})
    })
    
    // Test code...
}
```

### Error Cases

Test without initialization:

```go
func TestUnit_MyHandler_NotInitialized(t *testing.T) {
    resetGlobalState()
    
    _, err := MyHandler(context.Background(), req)
    if err == nil {
        t.Fatal("Expected error when not initialized")
    }
}
```

## Best Practices

1. **Use `t.Cleanup()`** instead of `defer` - runs even on panic, proper ordering
2. **Call `resetGlobalState()`** before Init() - ensures clean state
3. **Prefer TestContext** for integration tests - less boilerplate
4. **Use `t.TempDir()`** for data directories - automatic cleanup
5. **Test with `-count=3`** during development - catches cleanup issues

## Running Tests

```bash
cd plugin-go-backend/shared
go test ./handlers                      # All tests
go test ./handlers -v                   # Verbose
go test ./handlers -count=3             # Multiple iterations

# By category (prefix-based)
go test ./handlers -run ^TestUnit_      # Unit tests only (fast)
go test ./handlers -run ^TestIntegration_ # Integration tests only
go test ./handlers -run ^TestE2E_       # E2E tests only

# Coverage
go test ./handlers -cover
```

## Test Helpers

**`SetupIntegrationTest(t)`** - Returns TestContext with:
- Automatic Init/Shutdown lifecycle
- Default test space created
- `tc.Context()`, `tc.SpaceID()`, `tc.CreateDocument()`, `tc.CreateSpace()`

**`resetGlobalState()`** - Cleans global state:
- Closes DocumentManager, SpaceManager, EventManager
- Clears AccountManager keys
- Resets all state variables

## Troubleshooting

**"already initialized" errors** - Add `resetGlobalState()` before Init() and `t.Cleanup()` with Shutdown()

**Tests fail together but pass individually** - Verify all Init() calls have matching Shutdown() in t.Cleanup()

**Panic in cleanup** - Check cleanup order, ensure managers exist before closing

## Architecture

Global state singleton contains:
- `initialized` flag
- `accountManager` - Cryptographic keys
- `spaceManager` - Space management  
- `documentManager` - Document storage
- `eventManager` - Event broadcasting

Cleanup order: DocumentManager → SpaceManager → EventManager → AccountManager → state variables

## Test Stability

✅ All tests pass individually  
✅ All tests pass together  
✅ All tests pass with multiple iterations (`-count=3`)  
✅ No race conditions (verified with `-race`)  
✅ Clean state isolation

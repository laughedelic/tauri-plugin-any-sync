# Any-Sync Plugin: Aggressive Migration Plan

## Overview

This plan transforms the current per-operation architecture into a single-dispatch pattern with protobuf as the single source of truth. No backward compatibility considerations—we delete and replace.

Plugin name: **any-bridge** (or `tauri-plugin-any-bridge`)
API name: **SyncSpace API**

---

## Phase 1: Define the New API

### 1.1 Create the Plugin Protobuf Schema

Create a single `syncspace.proto` file that defines the entire plugin API. This becomes the source of truth for all layers.

The API should include:
- **Lifecycle**: Init, Shutdown
- **Spaces**: Create, Join, Leave, List, Delete
- **Documents**: Create, Get, Update, Delete, List, Query (with `bytes data` for opaque payloads)
- **Sync Control**: Start, Pause, Status
- **Events**: Subscribe (server streaming)

All document operations use opaque `bytes` for the payload—the plugin doesn't know or care about the app's data model. Include a `collection` field for logical grouping and `metadata` map for indexable fields.

### 1.2 Set Up Protobuf Tooling

Configure `buf` for protobuf management:
- Linting and breaking change detection
- Code generation for Go, TypeScript, and optionally Rust
- Single `buf generate` command produces all artifacts

---

## Phase 2: Rebuild Go Backend

### 2.1 Delete Existing Go Code

Remove:
- All existing protobuf definitions
- All existing gRPC service implementations
- All per-operation mobile exports
- Desktop gRPC server implementation

Keep:
- Any-Store/Any-Sync integration code (will be refactored)
- Build scripts and Makefile structure

### 2.2 Implement Dispatcher Pattern

Create a central dispatcher that:
- Maintains a registry of command name → handler function
- Handles protobuf unmarshaling/marshaling
- Provides unified error handling
- Routes based on command string

### 2.3 Implement Service Handlers

Create handler functions for each operation defined in the protobuf schema. Each handler:
- Receives raw bytes
- Unmarshals to the appropriate request type
- Executes business logic (wrapping Any-Sync APIs)
- Marshals response to bytes
- Returns bytes or error

### 2.4 Reduce Mobile Exports to Four Functions

The entire gomobile API becomes:
- `Init(dataPath string) error`
- `Command(cmd string, data []byte) ([]byte, error)`
- `SetEventHandler(handler func([]byte))`
- `Shutdown() error`

Delete all other exports.

### 2.5 Simplify Desktop Entry Point

The gRPC server for desktop should expose the same dispatcher interface. Consider whether gRPC is even necessary—a simple stdin/stdout protocol or Unix socket with the same `Command(cmd, data)` pattern may be simpler.

### 2.6 Write Go Tests

- Unit tests for each handler (mock Any-Sync dependencies)
- Dispatcher routing tests
- Test fixtures for common request/response patterns

This is Tier 1 testing—do it now, not later.

---

## Phase 3: Rebuild Rust Plugin

### 3.1 Delete Existing Rust Code

Remove:
- All per-operation Tauri commands
- All per-operation permission files
- Platform-specific service implementations with per-operation methods
- Generated protobuf types (will regenerate)

Keep:
- Plugin initialization boilerplate
- Sidecar management code (refactor to be simpler)
- Mobile FFI bridge code (refactor to single function)

### 3.2 Implement Single Command Handler

Create one Tauri command: `command(cmd: String, data: Vec<u8>) -> Result<Vec<u8>, Error>`

This command:
- Takes command name and raw bytes
- Forwards to backend (sidecar on desktop, FFI on mobile)
- Returns raw bytes or error

### 3.3 Implement Backend Trait

Define a simple trait with three methods:
- `command(&self, cmd: &str, data: &[u8]) -> Result<Vec<u8>, Error>`
- `set_event_handler(&self, handler: Box<dyn Fn(Vec<u8>)>)`
- `shutdown(&self) -> Result<(), Error>`

Implement for desktop (calls sidecar) and mobile (calls native FFI).

### 3.4 Simplify Permissions

Replace all per-operation permission files with a single permission that allows the `command` handler. One permission file instead of N.

### 3.5 Update Build Script

Modify `build.rs` to:
- Generate Rust types from protobuf (optional—only if you want typed Rust APIs)
- Handle binary bundling as before

### 3.6 Write Rust Passthrough Tests

Minimal tests verifying:
- Bytes pass through without corruption
- Errors propagate correctly
- Sidecar lifecycle works (desktop)

Keep these minimal—the Rust layer has no logic.

---

## Phase 4: Rebuild TypeScript API

### 4.1 Delete Existing TypeScript Code

Remove all hand-written API functions.

### 4.2 Generate TypeScript from Protobuf

Use `protobuf-ts` or similar to generate:
- TypeScript interfaces for all request/response messages
- Encode/decode functions for each message type

### 4.3 Create Typed Client

Build a thin client class that:
- Provides a typed method for each operation
- Internally calls the generic `command(name, bytes)` function
- Handles encoding requests and decoding responses

This client can be generated or hand-written (it's mechanical).

### 4.4 Export Raw Command Function

Also export the raw `command(cmd: string, data: Uint8Array)` function for advanced use cases where apps want to bypass the typed layer.

---

## Phase 5: Simplify Native Shims

### 5.1 iOS

Reduce to minimal Swift code that:
- Calls Go's exported C functions
- Bridges Tauri plugin system to Go library
- Contains zero business logic

Target: ~30 lines, never changes after initial implementation.

### 5.2 Android

Reduce to minimal Kotlin code that:
- Calls Go's JNI bindings
- Bridges Tauri plugin system to Go library
- Contains zero business logic

Target: ~30 lines, never changes after initial implementation.

---

## Phase 6: Integrate Any-Sync

### 6.1 Replace Any-Store Direct Usage

Remove direct Any-Store calls. Instead, use Any-Sync's higher-level APIs:
- `SpaceService` for space management
- `ObjectTree` for document storage
- Built-in sync mechanisms

Any-Store is used internally by Any-Sync—you don't interact with it directly.

### 6.2 Implement Space Operations

Wire up space handlers to Any-Sync's SpaceService:
- Create space → `SpaceService.CreateSpace()`
- Join space → `SpaceService.JoinSpace()` (handle invites)
- Leave/delete → corresponding Any-Sync methods

### 6.3 Implement Document Operations

Wire up document handlers to Any-Sync's ObjectTree:
- Create document → create new ObjectTree with initial change
- Update document → add change to ObjectTree
- Get document → read from ObjectTree
- Delete document → mark for deletion (Any-Sync's deletion mechanism)

### 6.4 Implement Sync Control

Expose Any-Sync's sync mechanisms:
- Start/pause sync → control HeadSync and peer connections
- Sync status → query current sync state

### 6.5 Implement Event Streaming

Bridge Any-Sync's internal events to the plugin's event system:
- Document changes (local and remote)
- Sync status changes
- Conflict notifications

### 6.6 Write Integration Tests

Now that handlers use real Any-Sync, write integration tests:
- Create space, add documents, verify persistence
- Test sync between two in-process instances
- Test restart scenarios (data survives)

This is where integration bugs surface—budget time for debugging.

---

## Phase 7: Update Example App

### 7.1 Delete Old Integration Code

Remove all code using the old per-operation API.

### 7.2 Build Domain Service Layer

Create an example service (e.g., `NotesService`) that:
- Uses the plugin's generic document API
- Defines its own data model
- Handles serialization/deserialization
- Demonstrates the intended usage pattern

### 7.3 Update Integration Tests

Rewrite tests to use the new single-command API. Tests should verify:
- Dispatch routing works correctly
- Each operation produces expected results
- Events flow through correctly
- Error handling works across layers

### 7.4 Add E2E Smoke Tests

Create E2E tests that exercise the full stack:
- TS calls plugin, plugin calls Go, Go calls Any-Sync
- Cover happy paths for all SyncSpace operations
- At least one error path test

These tests also serve as executable documentation of the API.

---

## Testing Strategy

### Testing Layers

**Layer 1: Go Handler Unit Tests (Critical)**

Test each SyncSpace API handler in isolation:
- Request deserialization from bytes
- Response serialization to bytes
- Business logic correctness
- Error conditions and error message propagation
- Edge cases (empty inputs, invalid IDs, etc.)

Mock Any-Sync dependencies at this layer to test handler logic independently.

**Layer 2: Go Dispatcher Tests (Critical)**

Test the dispatch mechanism:
- Correct routing of command names to handlers
- Unknown command handling (proper error response)
- Malformed input handling
- Concurrent dispatch behavior

**Layer 3: Go Integration Tests (Important)**

Test handlers with real Any-Sync (not mocked):
- Actual space creation and persistence
- Document CRUD with real ObjectTrees
- Data survives restart (persistence verification)
- Sync behavior between two in-process instances

These require more setup but catch integration bugs that unit tests miss.

**Layer 4: Rust Passthrough Tests (Minimal)**

Test that the Rust layer correctly passes data through:
- Bytes go in, bytes come out (no corruption)
- Errors propagate correctly
- Sidecar lifecycle (start, communicate, shutdown) on desktop
- FFI bridge works on mobile

These should be few and focused—the Rust layer has no logic to test.

**Layer 5: End-to-End Tests (Important)**

Test the full stack from TypeScript:
- TS → Rust → Go → Any-Sync → back
- Run against actual sidecar (desktop) or embedded library (mobile)
- Verify the complete roundtrip works

Use the example app's test harness for these.

**Layer 6: Cross-Device Sync Tests (Deferred)**

Test actual sync between multiple instances:
- Two devices syncing through Any-Sync network
- Conflict creation and resolution
- Offline changes merging when reconnected
- Network interruption handling

These are complex to set up and can be deferred until core functionality is solid.

**Layer 7: Platform-Specific Tests (Deferred)**

Test mobile-specific behavior:
- iOS framework integration
- Android AAR integration
- App lifecycle events (backgrounding, termination)
- Platform storage paths and permissions

---

### Priority Tiers

**Tier 1: Implement Immediately (Safety Net for Iteration)**

- Go handler unit tests (write alongside each handler)
- Go dispatcher tests (write when dispatcher is implemented)
- Basic E2E smoke test (one happy path through full stack)

These give you confidence to iterate quickly. If handlers are tested and dispatch works, you can refactor freely.

**Tier 2: Implement Before Shipping**

- Go integration tests with real Any-Sync
- Rust passthrough tests (basic coverage)
- E2E tests for all operations (happy paths)
- E2E error handling tests

These catch integration issues before users do.

**Tier 3: Defer Until Stable**

- Cross-device sync tests (complex setup, slow)
- Platform-specific mobile tests (require device/emulator)
- Performance/stress tests
- Conflict resolution edge cases

These are important but shouldn't block initial development.

---

### Testing Integration Into Phases

**Phase 2 (Go Backend):**
- Write dispatcher tests when implementing dispatcher
- Write handler unit tests alongside each handler implementation
- Set up test fixtures and mocks for Any-Sync

**Phase 3 (Rust Plugin):**
- Write minimal passthrough tests
- Verify error propagation works

**Phase 6 (Any-Sync Integration):**
- Write Go integration tests with real Any-Sync
- This is where most bugs will surface—budget time for it

**Phase 7 (Example App):**
- Write E2E tests using the example app
- These serve as both tests and usage documentation

---

### Test Infrastructure

**Go Tests:**
- Use standard `go test`
- Create test fixtures for common scenarios
- Mock Any-Sync interfaces for unit tests
- Use temporary directories for integration tests

**Rust Tests:**
- Use `cargo test`
- Test against a mock backend or real sidecar
- Integration tests may need the Go binary built first

**E2E Tests:**
- Use the existing integration test setup in example-app
- Run through Tauri's test harness
- Can be run in CI

---

## Final State

### Go Backend
- Single protobuf schema defining all operations
- Dispatcher with registered handlers
- Four mobile exports total
- Thin CRUD abstraction over Any-Sync

### Rust Plugin
- One Tauri command (`command`)
- One permission
- Simple backend trait with two implementations
- No per-operation code

### TypeScript API
- Generated types from protobuf
- Thin typed client (generated or mechanical)
- Raw command function for advanced use

### Native Shims
- ~30 lines each
- Pure passthrough, no logic
- Never change after initial implementation

### Adding New Operations
1. Add to `syncspace.proto`
2. Implement handler in Go
3. Write unit test for handler
4. Run `buf generate`
5. Done (TypeScript types update automatically, E2E tests catch integration issues)

---

## What Gets Deleted

- All per-operation Go mobile exports
- All per-operation Rust commands
- All per-operation TypeScript functions
- All per-operation permission files
- Direct Any-Store usage
- Complex native shim code
- Scattered protobuf definitions

## What Gets Created

- Single `syncspace.proto` as source of truth (SyncSpace API)
- Go dispatcher + handlers
- Go unit tests for all handlers
- Go integration tests with real Any-Sync
- Rust single-command architecture
- Rust passthrough tests (minimal)
- Generated TypeScript client
- Minimal native shims
- Any-Sync integration layer
- E2E tests via example app

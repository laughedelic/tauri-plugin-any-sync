# Implementation Tasks

## Phase 1: Define the New API ✅ COMPLETED

- [x] 1.1 Create unified `buf/proto/syncspace-api/syncspace/v1/syncspace.proto` schema defining complete SyncSpace API
- [x] 1.2 Define lifecycle operations (Init, Shutdown) in protobuf
- [x] 1.3 Define space operations (Create, Join, Leave, List, Delete) in protobuf
- [x] 1.4 Define document operations (Create, Get, Update, Delete, List, Query) with opaque `bytes data` in protobuf
- [x] 1.5 Define sync control operations (Start, Pause, Status) in protobuf
- [x] 1.6 Define event streaming (Subscribe with server streaming) in protobuf
- [x] 1.7 Set up `buf` for protobuf tooling (or configure `protoc` with generation scripts)
- [x] 1.8 Configure code generation for Go, TypeScript, and optionally Rust
- [x] 1.9 Validate protobuf schema compiles and generates artifacts

**Notes:** Created `buf/proto/syncspace-api/syncspace/v1/syncspace.proto` with complete SyncSpace API. Set up `buf` tooling with `buf.yaml` and `buf.gen.yaml`.

## Phase 2: Rebuild Go Backend (Local-First Any-Sync Integration)

### Integration Reality Check

**Original approach (REJECTED)**: Try to use spacestorage directly without full Any-Sync structure.

**Why it won't work**:
- `spacestorage.Create()` requires full `SpaceStorageCreatePayload` (ACL root, space header, settings)
- `ObjectTree` (for documents) requires `AclList` even for single-owner local operation
- Cannot store syncable documents without crypto infrastructure (keys, signing)

**Correct approach**: Initialize full Any-Sync structure in **local-only mode** (matching anytype-heart pattern).

**What's unavoidable** (must implement early):
- Account key generation and storage
- ACL structure and space payload generation
- Full ObjectTree for document storage

**What CAN be deferred** (Phase 6):
- Network registration (coordinator client)
- Peer connections (peermanager with network)
- JoinSpace/LeaveSpace (network operations)
- HeadSync/ObjectSync (network synchronization)

**Key insight**: Spaces created locally with full Any-Sync structure will sync automatically once network is enabled. No data migration needed.

### Phase 2A: Foundation - Dispatcher & Stubs ✅ COMPLETED

- [x] 2.1 Delete existing `plugin-go-backend/desktop/proto/*.proto` files (storage.proto, health.proto)
- [x] 2.2 Delete existing gRPC service implementations in `plugin-go-backend/desktop/api/server/`
- [x] 2.3 Delete all per-operation mobile exports in `plugin-go-backend/mobile/storage.go`
- [x] 2.4 Delete direct Any-Store integration in `plugin-go-backend/shared/storage/`
- [x] 2.5 Create `plugin-go-backend/shared/dispatcher/dispatcher.go` with command registry and routing
- [x] 2.6 Write unit tests for dispatcher (5 tests passing)
- [x] 2.7 Create `plugin-go-backend/shared/handlers/` directory structure
- [x] 2.8 Implement lifecycle handler stubs (Init, Shutdown) - 4 tests passing
- [x] 2.9 Implement space handler stubs (all return "not implemented") - 6 tests passing
- [x] 2.10 Implement document handler stubs (all return "not implemented") - 10 tests passing
- [x] 2.11 Implement sync handler stubs (all return "not implemented") - 4 tests passing
- [x] 2.12 Rewrite `plugin-go-backend/mobile/main.go` with 4-function API
- [x] 2.13 Update desktop entry point with TransportService
- [x] 2.14 Validate Go backend builds for all platforms

**Status**: Foundation complete with 29 passing stub tests (5 dispatcher + 4 lifecycle + 20 handler stubs)

### Phase 2B: Account & Identity Foundation ✅ COMPLETED

**Goal**: Establish cryptographic identity required by all Any-Sync operations.

- [x] 2.15 Create `plugin-go-backend/shared/anysync/account.go`
  - `AccountManager` struct with `accountdata.AccountKeys`
  - `GenerateKeys()` - creates new account keys
  - `StoreKeys(path)` - securely persists keys
  - `LoadKeys(path)` - loads existing keys
  - `GetKeys()` - returns current keys
- [x] 2.16 Update lifecycle handlers to use AccountManager
  - Init: Load or generate keys, store in global state
  - Shutdown: Clear keys from memory
- [x] 2.17 Write unit tests for AccountManager (9 tests)
  - Generate new keys
  - Store and load keys
  - Load keys persistence across restarts
  - Error handling for missing/corrupted keys
  - KeysExist, ClearKeys, StoreWithoutGeneration
- [x] 2.18 Update lifecycle handler tests to verify key management
  - Init verifies keys are loaded/generated
  - Shutdown verifies keys are cleared
  - Key persistence across Init/Shutdown cycles

**Dependencies**: `github.com/anyproto/any-sync@v0.11.5` (commonspace/object/accountdata, util/crypto)

**Enabled operations**: Init (with key generation/loading), Shutdown (with key cleanup)

**Test Results**: 40 tests passing total (9 AccountManager + 5 dispatcher + 26 handler tests)

### Phase 2C: Local Space Creation 🔄 NEXT

**Goal**: Create spaces with full Any-Sync structure, no network.

- [ ] 2.19 Create `plugin-go-backend/shared/anysync/spaces.go`
  - `SpaceManager` struct managing spaces
  - `CreateSpace(name, metadata)` using `spacepayloads.StoragePayloadForSpaceCreate()`
  - `ListSpaces()` enumerating spacestorage + metadata
  - `DeleteSpace(spaceId)` removing spacestorage
  - `GetSpace(spaceId)` retrieving space info
- [ ] 2.20 Implement space metadata storage
  - Separate anystore collection for app-level metadata (name, created_at, user prefs)
  - Not part of Any-Sync space structure
- [ ] 2.21 Implement ACL generation for single-owner spaces
  - Use `list.NewInMemoryDerivedAcl()` for local-only operation
  - Owner = account keys from Phase 2B
- [ ] 2.22 Update space handlers (CreateSpace, ListSpaces, DeleteSpace)
  - CreateSpace: Call SpaceManager.CreateSpace()
  - ListSpaces: Call SpaceManager.ListSpaces()
  - DeleteSpace: Call SpaceManager.DeleteSpace()
  - JoinSpace/LeaveSpace: Keep as "not implemented" (requires network)
- [ ] 2.23 Write unit tests for SpaceManager (8 tests)
  - Create space with valid payload
  - List spaces (empty, multiple spaces)
  - Get space by ID
  - Delete space
  - Persistence across manager restart
  - Duplicate space ID handling
- [ ] 2.24 Update space handler tests (6 tests)
  - CreateSpace success
  - ListSpaces after creating spaces
  - DeleteSpace success
  - JoinSpace returns not implemented
  - LeaveSpace returns not implemented

**Dependencies**: `github.com/anyproto/any-sync/commonspace/spacepayloads`, `github.com/anyproto/any-sync/commonspace/object/acl/list`

**Enabled operations**: CreateSpace, ListSpaces, DeleteSpace

### Phase 2D: Document Operations via ObjectTree

**Goal**: Store and retrieve documents using ObjectTree within spaces.

- [ ] 2.25 Create `plugin-go-backend/shared/anysync/documents.go`
  - `DocumentManager` struct wrapping ObjectTree operations
  - `InitSpace(spaceId)` - initializes Space with TreeBuilder
  - `CreateDocument(spaceId, docId, collection, data, metadata)` - creates ObjectTree
  - `GetDocument(spaceId, docId)` - reads ObjectTree state
  - `UpdateDocument(spaceId, docId, data, metadata)` - adds change to ObjectTree
  - `DeleteDocument(spaceId, docId)` - marks as deleted in ObjectTree
  - `ListDocuments(spaceId, collection)` - enumerates ObjectTrees
  - `QueryDocuments(spaceId, collection, filters)` - filters by metadata
- [ ] 2.26 Implement document metadata storage
  - Space's key-value store for indexable metadata
  - Collection grouping via metadata field
- [ ] 2.27 Implement ObjectTree change builder
  - Wrap document data as change payload
  - Sign changes with account keys
  - Handle change history
- [ ] 2.28 Update document handlers (Create, Get, Update, Delete, List, Query)
  - Replace stub implementations with DocumentManager calls
- [ ] 2.29 Write unit tests for DocumentManager (12 tests)
  - Create document in space
  - Get document (found, not found)
  - Update document (new version)
  - Delete document
  - List documents (empty, filtered by collection)
  - Query documents (metadata filters)
  - Document persistence across manager restart
  - Multiple documents in same space
- [ ] 2.30 Update document handler tests (10 tests)
  - All document operations with real ObjectTree
  - Verify change history tracking

**Dependencies**: `github.com/anyproto/any-sync/commonspace/object/tree/objecttree`, `github.com/anyproto/any-sync/commonspace/object/tree/treebuilder`

**Enabled operations**: CreateDocument, GetDocument, UpdateDocument, DeleteDocument, ListDocuments, QueryDocuments

### Phase 2E: Local Event System

**Goal**: Stream local changes to subscribers.

- [ ] 2.31 Create `plugin-go-backend/shared/anysync/events.go`
  - `EventManager` with subscriber registry
  - `Subscribe(eventTypes, spaceIds)` - registers subscriber
  - `Unsubscribe(subscriberId)` - removes subscriber
  - `emitEvent(event)` - broadcasts to matching subscribers
- [ ] 2.32 Hook ObjectTree change events
  - document.created
  - document.updated
  - document.deleted
- [ ] 2.33 Hook Space lifecycle events
  - space.created
  - space.deleted
- [ ] 2.34 Implement Subscribe handler (server streaming)
  - Create event stream channel
  - Register subscriber with EventManager
  - Stream events to client
  - Cleanup on disconnect
- [ ] 2.35 Write unit tests for EventManager (6 tests)
  - Subscribe with filtering
  - Multiple subscribers
  - Event emission and delivery
  - Unsubscribe cleanup
- [ ] 2.36 Write integration tests for event streaming (4 tests)
  - Document change triggers event
  - Space deletion triggers event
  - Event filtering works
  - Multiple concurrent subscribers

**Dependencies**: None (uses existing components)

**Enabled operations**: Subscribe (local events only)

### Phase 2F: Validation & Integration Testing

- [ ] 2.37 Write end-to-end integration tests (10 tests)
  - Full lifecycle: Init → CreateSpace → CreateDocument → GetDocument → Shutdown
  - Persistence: Init → CreateSpace → Shutdown → Init → ListSpaces (verify survival)
  - Document CRUD within space
  - Event streaming during operations
  - Multiple spaces with documents
  - Error handling (uninitialized, invalid IDs)
- [ ] 2.38 Validate builds for all platforms
  - Desktop: macOS, Linux, Windows
  - Mobile: Android AAR, iOS xcframework
- [ ] 2.39 Run full test suite
  - All unit tests passing (target: ~60 tests)
  - All integration tests passing (target: ~10 tests)
  - No memory leaks (run with race detector)

**Test Coverage Target**: 
- Dispatcher: 5 tests ✅
- Lifecycle handlers: 4 tests ✅
- Account management: 6 tests (Phase 2B)
- Space management: 8 tests (Phase 2C)
- Space handlers: 6 tests (Phase 2C)
- Document management: 12 tests (Phase 2D)
- Document handlers: 10 tests (Phase 2D)
- Event management: 6 tests (Phase 2E)
- Integration tests: 10 tests (Phase 2F)
- **Total target: ~67 tests**

**Architecture Notes:**
- **Two Proto Files:**
  1. `syncspace.proto` - SyncSpace API definitions (messages + service for docs)
  2. `transport.proto` - 4-method transport layer (Init, Command, Subscribe, Shutdown)
- **Shared Go Code:** Dispatcher + handlers in `shared/`
- **Platform Entry Points:**
  - Mobile: Direct Go exports (Init, Command, SetEventHandler, Shutdown) via gomobile
  - Desktop: gRPC server implementing TransportService (same 4 methods) called by Rust gRPC client
- **Command Flow:** Both platforms → dispatcher.Dispatch(cmd, bytes) → handlers

**What's NOT Implemented** (deferred to Phase 6):
- Network registration (no coordinator client)
- Peer connections (no peermanager with network)
- JoinSpace/LeaveSpace operations (network-only)
- HeadSync/ObjectSync (network synchronization)
- Sync control (StartSync, PauseSync, GetSyncStatus)
- Network-related events (sync.started, sync.completed, etc.)

## Phase 3: Rebuild Rust Plugin

- [ ] 3.1 Delete existing per-operation commands from `plugin-rust-core/src/commands.rs`
- [ ] 3.2 Delete existing per-operation service methods from `plugin-rust-core/src/desktop.rs` and `mobile.rs`
- [ ] 3.3 Delete all per-operation permission files in `plugin-rust-core/permissions/` (keep directory structure)
- [ ] 3.4 Define simplified `AnySyncBackend` trait with 3 methods: `command()`, `set_event_handler()`, `shutdown()`
- [ ] 3.5 Implement single `command(cmd: String, data: Vec<u8>) -> Result<Vec<u8>>` Tauri command
- [ ] 3.6 Implement `AnySyncBackend` for desktop (calls sidecar via gRPC or simplified IPC)
- [ ] 3.7 Implement `AnySyncBackend` for mobile (calls native FFI)
- [ ] 3.8 Update iOS Swift shim to ~30 lines (pure passthrough to Go C exports)
- [ ] 3.9 Update Android Kotlin shim to ~30 lines (pure passthrough to Go JNI)
- [ ] 3.10 Create single permission file `plugin-rust-core/permissions/default.toml` for `command` handler
- [ ] 3.11 Update `plugin-rust-core/build.rs` to handle new binary structure (if needed)
- [ ] 3.12 Write minimal Rust passthrough tests (bytes pass through, errors propagate)
- [ ] 3.13 Validate Rust plugin builds for desktop and mobile

## Phase 4: Rebuild TypeScript API

- [ ] 4.1 Delete all hand-written API functions from `plugin-js-api/src/index.ts`
- [ ] 4.2 Set up protobuf TypeScript code generation (protobuf-ts or similar)
- [ ] 4.3 Generate TypeScript types from `syncspace.proto`
- [ ] 4.4 Generate encode/decode functions for all messages
- [ ] 4.5 Create typed client class with method for each operation (generated or mechanical)
- [ ] 4.6 Implement raw `command(cmd: string, data: Uint8Array)` function calling Tauri invoke
- [ ] 4.7 Export typed client as default export
- [ ] 4.8 Export raw command function for advanced use cases
- [ ] 4.9 Add JSDoc documentation to generated/typed client
- [ ] 4.10 Validate TypeScript API builds and type-checks

## Phase 5: Update Native Shims

- [ ] 5.1 Simplify iOS Swift code in `plugin-rust-core/ios/Sources/` to minimal bridge
- [ ] 5.2 Remove per-operation Swift methods, keep only plugin initialization and command forwarding
- [ ] 5.3 Validate iOS builds with simplified shim
- [ ] 5.4 Simplify Android Kotlin code in `plugin-rust-core/android/src/` to minimal bridge
- [ ] 5.5 Remove per-operation Kotlin methods, keep only plugin initialization and command forwarding
- [ ] 5.6 Validate Android builds with simplified shim

## Phase 6: Network Sync Layer

**Goal**: Enable synchronization with Any-Sync network.

### Phase 6A: Network Configuration

- [ ] 6.1 Create `plugin-go-backend/shared/anysync/network.go`
  - `NetworkConfig` struct (coordinator address, node addresses, mode)
  - `NetworkMode` enum (LocalOnly, NetworkEnabled)
  - Configuration parsing and validation
- [ ] 6.2 Update Init operation to accept network configuration
  - Optional network config in InitRequest
  - Default to LocalOnly mode
- [ ] 6.3 Write unit tests for network configuration (4 tests)
  - Parse valid config
  - Validate required fields
  - Default to LocalOnly
  - Error handling for invalid config

### Phase 6B: Coordinator Integration

- [ ] 6.4 Implement coordinator client initialization
  - `coordinatorclient.CoordinatorClient` setup
  - Space registration with coordinator
  - Handle registration errors gracefully
- [ ] 6.5 Update SpaceManager for network registration
  - Register new spaces with coordinator (if NetworkEnabled)
  - Handle offline mode (queue registration attempts)
- [ ] 6.6 Write unit tests for coordinator integration (6 tests)
  - Space registration success
  - Registration failure handling
  - Offline mode behavior
  - Re-registration on reconnect

### Phase 6C: Peer Management

- [ ] 6.7 Initialize peer manager with network support
  - `peermanager.PeerManager` setup
  - `nodeconf.Service` for node discovery
  - Connection pool configuration
- [ ] 6.8 Implement peer connection lifecycle
  - Discover peers for space
  - Establish connections
  - Handle disconnections
  - Reconnection logic
- [ ] 6.9 Write unit tests for peer management (6 tests)
  - Peer discovery
  - Connection establishment
  - Disconnect handling
  - Reconnection attempts

### Phase 6D: Sync Tree Wrapper

- [ ] 6.10 Wrap ObjectTree with SyncTree
  - Convert existing ObjectTrees to SyncTrees
  - Automatic sync on changes
  - Conflict detection
- [ ] 6.11 Implement HeadSync integration
  - Exchange heads with peers
  - Detect missing changes
  - Request missing data
- [ ] 6.12 Implement ObjectSync integration
  - Stream changes to/from peers
  - Merge remote changes
  - Handle conflicts (last-write-wins initially)
- [ ] 6.13 Write unit tests for sync wrapper (8 tests)
  - SyncTree wraps ObjectTree correctly
  - Local changes trigger sync
  - Remote changes received
  - Conflict resolution
  - Multiple peers
  - Offline/online transitions

### Phase 6E: Network Space Operations

- [ ] 6.14 Implement JoinSpace operation
  - Accept invite token
  - Register with coordinator
  - Fetch space data from peers
  - Initialize local space storage
  - Join ACL as member
- [ ] 6.15 Implement LeaveSpace operation
  - Notify coordinator
  - Remove from ACL
  - Optionally delete local data
- [ ] 6.16 Update space handlers
  - Replace "not implemented" with real implementations
- [ ] 6.17 Write unit tests for network space operations (8 tests)
  - JoinSpace with valid invite
  - JoinSpace with invalid invite
  - LeaveSpace success
  - Space data fetch from peers
  - Handle peer unavailability

### Phase 6F: Sync Control Operations

- [ ] 6.18 Implement StartSync operation
  - Enable HeadSync/ObjectSync for space(s)
  - Begin synchronization loops
  - Emit sync.started event
- [ ] 6.19 Implement PauseSync operation
  - Disable sync loops
  - Maintain connections
  - Emit sync.paused event
- [ ] 6.20 Implement GetSyncStatus operation
  - Query sync state per space
  - Report pending changes count
  - Report last sync timestamp
  - Report sync errors
- [ ] 6.21 Update sync control handlers
  - Replace stub implementations
- [ ] 6.22 Write unit tests for sync control (8 tests)
  - StartSync enables sync
  - PauseSync stops sync
  - GetSyncStatus returns accurate data
  - Sync status per space
  - Error reporting

### Phase 6G: Network Events

- [ ] 6.23 Extend EventManager with network events
  - sync.started
  - sync.completed
  - sync.paused
  - sync.error
  - sync.conflict
  - peer.connected
  - peer.disconnected
- [ ] 6.24 Hook network event sources
  - Sync state changes
  - Peer connection events
  - Conflict detection events
- [ ] 6.25 Write unit tests for network events (6 tests)
  - Sync events emitted correctly
  - Peer events emitted
  - Conflict events emitted
  - Event filtering works

### Phase 6H: Integration Testing

- [ ] 6.26 Write two-node sync tests (12 tests)
  - Create space on node A, sync to node B
  - Create document on node A, appears on node B
  - Concurrent edits on both nodes (conflict resolution)
  - Offline node catches up when reconnecting
  - Multiple spaces syncing
  - Join space from invite
  - Leave space
  - Sync pause and resume
- [ ] 6.27 Write network failure tests (6 tests)
  - Handle coordinator unavailable
  - Handle peer unavailable
  - Handle intermittent network
  - Recover from sync errors
- [ ] 6.28 Validate full network stack builds
  - All platforms compile with network code
  - No regression in local-only mode

**Test Coverage Target (Phase 6)**:
- Network config: 4 tests
- Coordinator integration: 6 tests
- Peer management: 6 tests
- Sync wrapper: 8 tests
- Network space ops: 8 tests
- Sync control: 8 tests
- Network events: 6 tests
- Two-node integration: 12 tests
- Network failure: 6 tests
- **Phase 6 total: ~64 tests**
- **Grand total with Phase 2: ~131 tests**

## Phase 7: Update Example App

- [ ] 7.1 Delete old storage integration code from `example-app/src/App.svelte`
- [ ] 7.2 Create example domain service `example-app/src/services/notes.ts` using SyncSpace API
- [ ] 7.3 Implement NotesService with create, get, update, delete, list methods
- [ ] 7.4 Update App.svelte to use NotesService instead of direct storage API
- [ ] 7.5 Add UI for space management (create, join, list spaces)
- [ ] 7.6 Add UI for document operations (CRUD notes)
- [ ] 7.7 Add UI for sync control (start, pause, status)
- [ ] 7.8 Implement event subscription and display (document changes, sync status)
- [ ] 7.9 Write E2E tests covering happy paths for all SyncSpace operations
- [ ] 7.10 Write E2E tests for at least one error path
- [ ] 7.11 Validate example app works on desktop
- [ ] 7.12 Validate example app works on Android (emulator)
- [ ] 7.13 Validate example app works on iOS (simulator)

## Phase 8: Documentation and Cleanup

- [ ] 8.1 Update root README.md with new architecture overview
- [ ] 8.2 Document single-dispatch pattern and design decisions
- [ ] 8.3 Create migration guide from old API to new SyncSpace API
- [ ] 8.4 Document SyncSpace API operations with examples
- [ ] 8.5 Document domain service pattern (how apps should use the plugin)
- [ ] 8.6 Update component-specific READMEs (Go backend, Rust plugin, TypeScript API)
- [ ] 8.7 Update AGENTS.md with new workflow for adding operations
- [ ] 8.8 Clean up old documentation references to per-operation pattern
- [ ] 8.9 Update build/test documentation for new structure
- [ ] 8.10 Archive or update openspec specs to reflect new architecture

## Phase 9: Final Validation

- [ ] 9.1 Run full test suite (unit + integration + E2E) on all platforms
- [ ] 9.2 Verify binary builds for all platforms (desktop x3 + mobile x2)
- [ ] 9.3 Test example app on all 5 platforms (macOS, Linux, Windows, Android, iOS)
- [ ] 9.4 Validate performance (single IPC call per operation)
- [ ] 9.5 Review and validate all documentation is accurate
- [ ] 9.6 Create release notes documenting breaking changes
- [ ] 9.7 Tag release as breaking version (e.g., v2.0.0)

## Dependencies and Parallelization

**Critical Path:**
Phase 1 → Phase 2A → 2B → 2C → 2D → 2E → 2F → Phase 3 → Phase 4 → Phase 7 → Phase 9

**Phase 2 Must Be Sequential:**
- Phase 2B (Account) must complete before 2C (Spaces need keys)
- Phase 2C (Spaces) must complete before 2D (Documents need spaces)
- Phase 2D (Documents) must complete before 2E (Events need documents)
- Phase 2E (Events) must complete before 2F (Integration tests need events)

**Can Parallelize After Phase 2F:**
- Phase 3 (Rust Plugin) can start
- Phase 4 (TypeScript API) can start once Phase 3.5 complete
- Phase 5 (Native Shims) can start once Phase 2.12 complete
- Phase 8 (Documentation) can be done incrementally throughout

**Phase 6 (Network Sync):**
- Can start anytime after Phase 2F complete
- Independent of Phases 3-5 (frontend layers)
- Should be done before Phase 7 (Example App needs to demo sync)

**High Priority (Tier 1 - Local-First):**
- Phase 2B-2F (Account, Spaces, Documents, Events, Integration)
- Phase 3 (Rust Plugin)
- Phase 4 (TypeScript API)
- Phase 7 (Example App with local-only mode)

**Medium Priority (Tier 2 - Network Sync):**
- Phase 6 (Network Sync Layer)
- Phase 7 updated to demo sync

**Lower Priority (Tier 3):**
- Platform-specific validation
- Performance optimization
- Additional test coverage

**Implementation Milestones:**

1. **Milestone 1** (Phase 2A): Dispatcher & Stubs Complete ✅
   - 29 tests passing
   - Foundation ready for Any-Sync integration

2. **Milestone 2** (Phase 2B): Account Keys Working
   - ~35 tests passing
   - Can sign/encrypt operations

3. **Milestone 3** (Phase 2C): Spaces Working Locally
   - ~49 tests passing
   - Can create/list/delete spaces

4. **Milestone 4** (Phase 2D): Documents Working
   - ~71 tests passing
   - Full CRUD on documents within spaces

5. **Milestone 5** (Phase 2E): Events Working
   - ~83 tests passing
   - Real-time change notifications

6. **Milestone 6** (Phase 2F): Integration Tests Pass
   - ~93 tests passing
   - End-to-end local operations validated

7. **Milestone 7** (Phase 3-5): Frontend Complete
   - Rust plugin, TypeScript API, Native shims
   - Example app works in local-only mode

8. **Milestone 8** (Phase 6): Network Sync Working
   - ~157 tests passing
   - Full Any-Sync network synchronization

9. **Milestone 9** (Phase 7-9): Production Ready
   - Example app demos sync
   - All platforms validated
   - Documentation complete

# Implementation Tasks

## Phase 1: Define the New API ✅ COMPLETED

- [x] 1.1 Create unified `plugin-go-backend/proto/syncspace.proto` schema defining complete SyncSpace API
- [x] 1.2 Define lifecycle operations (Init, Shutdown) in protobuf
- [x] 1.3 Define space operations (Create, Join, Leave, List, Delete) in protobuf
- [x] 1.4 Define document operations (Create, Get, Update, Delete, List, Query) with opaque `bytes data` in protobuf
- [x] 1.5 Define sync control operations (Start, Pause, Status) in protobuf
- [x] 1.6 Define event streaming (Subscribe with server streaming) in protobuf
- [x] 1.7 Set up `buf` for protobuf tooling (or configure `protoc` with generation scripts)
- [x] 1.8 Configure code generation for Go, TypeScript, and optionally Rust
- [x] 1.9 Validate protobuf schema compiles and generates artifacts

**Notes:** Created `plugin-go-backend/proto/syncspace/v1/syncspace.proto` with complete SyncSpace API. Set up `buf` tooling with `buf.yaml` and `buf.gen.yaml`. Generated Go code in `proto/syncspace/v1/syncspace.pb.go` and TypeScript code in `plugin-js-api/src/generated/syncspace/v1/syncspace_pb.ts`.

## Phase 2: Rebuild Go Backend ⚠️ PARTIALLY COMPLETE

- [ ] 2.1 Delete existing `plugin-go-backend/desktop/proto/*.proto` files (storage.proto, health.proto)
- [ ] 2.2 Delete existing gRPC service implementations in `plugin-go-backend/desktop/api/server/` (storage.go, health.go)
- [x] 2.3 Delete all per-operation mobile exports in `plugin-go-backend/mobile/storage.go`
- [ ] 2.4 Delete direct Any-Store integration in `plugin-go-backend/shared/storage/` (or mark deprecated)
- [x] 2.5 Create `plugin-go-backend/shared/dispatcher/dispatcher.go` with command registry and routing
- [x] 2.6 Write unit tests for dispatcher (routing, unknown commands, malformed input)
- [x] 2.7 Create `plugin-go-backend/shared/handlers/` directory structure
- [x] 2.8 Implement lifecycle handlers (Init, Shutdown) with unit tests ✅
- [x] 2.9 Implement space handlers (Create, Join, Leave, List, Delete) - **Stubs only, need unit tests** ⚠️
- [x] 2.10 Implement document handlers (Create, Get, Update, Delete, List, Query) - **Stubs only, need unit tests** ⚠️
- [x] 2.11 Implement sync control handlers (Start, Pause, Status) - **Stubs only, need unit tests** ⚠️
- [ ] 2.12 Create `plugin-go-backend/shared/anysync/` integration layer - **NOT STARTED**
- [ ] 2.13 Integrate Any-Sync SpaceService for space management - **NOT STARTED**
- [ ] 2.14 Integrate Any-Sync ObjectTree for document storage - **NOT STARTED**
- [ ] 2.15 Integrate Any-Sync sync mechanisms (HeadSync, ObjectSync) - **NOT STARTED**
- [ ] 2.16 Write integration tests with real Any-Sync (space + document + persistence) - **NOT STARTED**
- [x] 2.17 Rewrite `plugin-go-backend/mobile/main.go` with 4-function API: Init, Command, SetEventHandler, Shutdown
- [ ] 2.18 Update desktop entry point to use dispatcher (keep or simplify gRPC) - **NOT STARTED**
- [x] 2.19 Validate Go backend builds for all platforms (mobile only, desktop needs 2.18)

**Critical Issues:**
1. **NO HANDLER UNIT TESTS**: Handlers are stub implementations without any unit tests
2. **OLD CODE NOT REMOVED**: Desktop gRPC server, old proto files, old storage layer still present
3. **NO ANY-SYNC INTEGRATION**: Handlers return "not implemented yet" errors
4. **DESKTOP NOT UPDATED**: Desktop entry point still uses old gRPC server
5. **NO INTEGRATION TESTS**: No tests verifying the full command flow

**What Actually Works:**
- Dispatcher: ✅ Has unit tests, routing works
- Mobile API: ✅ Compiles with 4-function interface
- Protobuf: ✅ Schema defined and generates code
- Handler stubs: ✅ Compile but return errors

**What's Needed Before Phase 3:**
1. Add unit tests for ALL handlers (lifecycle, spaces, documents, sync)
2. Test error handling and validation in handlers
3. Clean up old code to avoid confusion
4. Either implement basic Any-Sync integration OR use in-memory mocks for testing

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

## Phase 6: Write Integration Tests

- [ ] 6.1 Set up integration test infrastructure in Go with real Any-Sync
- [ ] 6.2 Write integration tests for space operations (create, join, leave)
- [ ] 6.3 Write integration tests for document CRUD with persistence verification
- [ ] 6.4 Write integration tests for sync behavior (two in-process instances)
- [ ] 6.5 Write integration tests for restart scenarios (data survives)
- [ ] 6.6 Create test fixtures and helper functions for common scenarios
- [ ] 6.7 Validate all Go integration tests pass

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
Phase 1 → Phase 2 → Phase 3 → Phase 4 → Phase 7 → Phase 9

**Can Parallelize:**
- Phase 5 (Native Shims) can start once Phase 2.17 (mobile exports) is complete
- Phase 4 (TypeScript) can start once Phase 3.5 (single Rust command) is complete
- Phase 6 (Integration Tests) can start once Phase 2.16 (Go integration) is complete
- Phase 8 (Documentation) can be done incrementally throughout

**High Priority (Tier 1 Testing):**
- 2.6 (Dispatcher tests)
- 2.8-2.11 (Handler unit tests)
- 2.16 (Go integration tests)
- 7.9 (E2E happy path tests)

**Lower Priority (Tier 2-3 Testing):**
- 3.12 (Rust passthrough tests)
- 7.10 (E2E error path tests)
- Platform-specific validation can be done last

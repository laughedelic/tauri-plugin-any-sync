# Change: Refactor to Single-Dispatch SyncSpace Architecture

## Why

The current plugin architecture requires modifying ~11 files for each new operation: protobuf definitions, Go handlers, Go mobile exports, Rust commands, Rust service methods, Rust permissions (3 files), and TypeScript functions. This creates significant maintenance burden and slows down feature development.

After analyzing the Anytype ecosystem (anytype-heart, anytype-ts, anytype-kotlin), we identified a **single-dispatch pattern** that reduces boilerplate by 80%. Instead of per-operation exports, Anytype uses a single `Command(cmd, data)` function that dispatches to handlers based on command name, with protobuf as the single source of truth.

Additionally, the current architecture uses Any-Store directly, but research shows Any-Sync provides the necessary higher-level abstractions for spaces, documents, and synchronization. Any-Store is an internal implementation detail of Any-Sync, not a direct API surface.

This change transforms the plugin to use the **SyncSpace API**: a generic, opaque-data document storage and synchronization system that lets host applications define their own data models while the plugin handles the complexity of local-first sync. The plugin retains the name `tauri-plugin-any-sync`, while the internal API is called `syncspace` to distinguish it from the raw Any-Sync protocol.

## Implementation Status

**Completed (Phases 1-5, 7)**: Local-first foundation with protobuf API, single-dispatch architecture, Any-Sync integration for local storage, generated TypeScript client, simplified mobile shims (~30-50 lines), example app with domain service pattern. 97 integration tests passing.

**Deferred (Phase 6)**: Network synchronization layer including network coordinator, peer management, JoinSpace/LeaveSpace operations, sync control (StartSync/PauseSync/GetSyncStatus), and network events. The local-first foundation is complete and production-ready with full Any-Sync data structures.

**Deferred (Phase 9)**: E2E testing suite for happy/error paths across platforms.

The plugin fully supports local-first operations. All spaces and documents use proper cryptographic keys and Any-Sync structures, enabling future network sync without data migration or API changes.

## What Changes

### Core Architectural Changes

- **Single dispatch pattern**: Replace per-operation exports with one `Command(cmd string, data []byte) ([]byte, error)` function
- **Protobuf as source of truth**: All operations defined in unified `syncspace.proto` schema
- **Any-Sync integration**: Replace Any-Store direct usage with Any-Sync's higher-level APIs (SpaceService, ObjectTree, HeadSync)
- **Opaque data model**: Documents store `bytes data` instead of structured JSON, making the plugin data-model agnostic
- **Generated TypeScript client**: Auto-generate typed client from protobuf instead of hand-written functions

### Component Changes

**Go Backend:**
- Create unified `syncspace.proto` defining the complete SyncSpace API (lifecycle, spaces, documents, sync control, events)
- Implement dispatcher pattern with command name → handler registry
- Reduce mobile exports from N functions to 4: `Init()`, `Command()`, `SetEventHandler()`, `Shutdown()`
- Keep gRPC desktop server but adapt to dispatcher interface (gRPC provides streaming for events and complexity already paid)
- Integrate Any-Sync's SpaceService for space management
- Integrate Any-Sync's ObjectTree for document storage
- Integrate Any-Sync's sync mechanisms (HeadSync, ObjectSync) for synchronization

**Rust Plugin:**
- Replace per-operation commands with single `command(cmd: String, data: Vec<u8>) -> Result<Vec<u8>>`
- Simplify backend trait to 3 methods: `command()`, `set_event_handler()`, `shutdown()`
- Replace all per-operation permissions with single permission for `command` handler
- Reduce native shims (iOS/Android) to ~30 lines of pure passthrough code

**TypeScript API:**
- Use `buf` for protobuf code generation (cleaner setup, built-in linting, breaking change detection)
- Generate TypeScript types from protobuf using `protobuf-ts` or similar
- Generate or create mechanical typed client with methods for each operation
- Export both typed client and raw `command(cmd, data)` function
- Remove all hand-written per-operation functions

### API Surface

The SyncSpace API provides:

**Lifecycle**: Init, Shutdown  
**Spaces**: Create, Join, Leave, List, Delete  
**Documents**: Create, Get, Update, Delete, List, Query (with opaque `bytes data`)  
**Sync Control**: Start, Pause, Status  
**Events**: Subscribe (server streaming for document changes, sync status)

Documents use `bytes data` for payload (app serializes its own format) plus `metadata` map for indexable fields and `collection` string for logical grouping.

## Impact

### Affected Specs

This is a **breaking change** that affects most existing capabilities:

- **plugin-communication** (RENAMED → `single-dispatch-pattern`) - New single-command architecture
- **storage-api** (REPLACED → `syncspace-api`) - From JSON documents to opaque bytes + Any-Sync integration
- **go-backend-scaffolding** (MODIFIED) - Dispatcher pattern, Any-Sync integration, reduced mobile exports
- **desktop-integration** (MODIFIED) - Single command handler, simplified sidecar (potentially non-gRPC)
- **mobile-backend-api** (MODIFIED) - Four mobile exports instead of N functions
- **android-plugin-integration** (MODIFIED) - Minimal shim (~30 lines)
- **binaries-distribution** (MODIFIED) - May need updated checksums and build process
- **example-app-update** (MODIFIED) - Use new SyncSpace API with domain service layer

New specs:
- **syncspace-api** (NEW) - Generic spaces + documents + sync API
- **any-sync-integration** (NEW) - SpaceService, ObjectTree, sync protocol integration

### Affected Code

Nearly all code will be deleted and rewritten:

**Deleted:**
- All per-operation Go mobile exports (`plugin-go-backend/mobile/storage.go`)
- All per-operation Rust commands (`plugin-rust-core/src/commands.rs`)
- All per-operation TypeScript functions (`plugin-js-api/src/index.ts`)
- All per-operation permission files (`plugin-rust-core/permissions/*.toml`)
- Current protobuf definitions (`plugin-go-backend/desktop/proto/storage.proto`)
- Direct Any-Store integration (`plugin-go-backend/shared/storage/`)
- Complex native shim code

**Created:**
- `plugin-go-backend/proto/syncspace.proto` - Unified SyncSpace API definition
- `plugin-go-backend/shared/dispatcher/` - Command dispatcher with handler registry
- `plugin-go-backend/shared/handlers/` - Handler functions for each SyncSpace operation
- `plugin-go-backend/shared/anysync/` - Any-Sync integration layer (SpaceService, ObjectTree, sync)
- `plugin-go-backend/mobile/main.go` - New 4-function mobile API
- `plugin-rust-core/src/commands.rs` - Single `command()` handler
- `plugin-rust-core/permissions/default.toml` - Single permission
- `plugin-js-api/src/generated/` - Generated types and client
- `example-app/src/services/notes.ts` - Example domain service using SyncSpace API

**Line count estimate:**
- Delete: ~3000 lines (existing implementation + boilerplate)
- Add: ~2500 lines (dispatcher + handlers + Any-Sync integration + tests)
- Net: -500 lines despite adding full Any-Sync integration

### Not Affected

- Build system and binary distribution (structure remains same)
- Tauri plugin initialization boilerplate
- Sidecar management on desktop (gRPC IPC retained)
- Overall project structure (folders remain same)

### Migration Path

This is a **breaking change with no backward compatibility**. Since the plugin has no external users yet, this is a clean break with no migration guide needed.

### Risks

- **Large scope**: Touches nearly every file in the project
- **Any-Sync learning curve**: Team needs to understand SpaceService, ObjectTree, sync protocol
- **Testing burden**: Need comprehensive tests at each layer (Go handlers, dispatcher, integration, E2E)
- **Breaking change**: Acceptable since plugin has no external users yet
- **Documentation**: Extensive docs needed to explain the new architecture and patterns

### Benefits

- **Drastically reduced boilerplate**: Adding new operation goes from 11 files → 2 files
- **Type safety from protobuf**: Single source of truth generates all types
- **Proper Any-Sync usage**: Leverage higher-level abstractions instead of low-level Any-Store
- **Generic and reusable**: Plugin is data-model agnostic, works for any application
- **Simplified mobile**: 4 functions instead of N, easier to maintain
- **Clearer architecture**: Thin passthrough layers, logic centralized in Go

## Decisions

1. **IPC mechanism for desktop**: **Keep gRPC**. It provides streaming for free (needed for event subscriptions), the complexity is already paid, and the dispatch pattern isolates this decision for future changes if needed.

2. **Protobuf tooling**: **Use buf**. Single config file (buf.gen.yaml) instead of shell scripts, built-in linting, breaking change detection, and cleaner multi-language generation setup.

3. **Testing strategy**: **Tier 1 is sufficient** (Go unit tests for handlers, dispatcher tests, basic E2E). Higher test coverage can be added incrementally as the API stabilizes.

4. **Migration support**: **Clean break, no migration guide**. The plugin has no external users yet, so no backward compatibility or migration paths are needed.

5. **Plugin naming**: **Keep `tauri-plugin-any-sync`**. The internal API (defined in proto) is called `syncspace` to distinguish it from the raw Any-Sync protocol.

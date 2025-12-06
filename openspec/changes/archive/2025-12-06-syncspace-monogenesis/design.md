# Design: Single-Dispatch SyncSpace Architecture

## Context

The current architecture evolved from a simple storage plugin into a more complex system, but retained the per-operation pattern from the initial implementation. Each new operation requires changes to protobuf definitions, Go handlers, Go mobile exports, Rust commands, Rust service traits, Rust permissions, and TypeScript functions.

Research into the Anytype ecosystem (which also uses Any-Sync and gomobile) revealed a mature pattern that dramatically reduces boilerplate: the **single-dispatch pattern** with protobuf as the single source of truth. This pattern is battle-tested in production with over 200 operations.

Additionally, current direct Any-Store usage bypasses Any-Sync's higher-level abstractions. Any-Sync provides the necessary APIs for spaces, documents, and synchronization—Any-Store is an internal implementation detail.

**Constraints:**
- Must maintain cross-platform support (desktop + mobile)
- Must remain gomobile-compatible (Android/iOS FFI)
- Must minimize runtime performance impact
- Must support opaque application data models
- Breaking change is acceptable (plugin not widely adopted yet)

**Stakeholders:**
- Plugin maintainers (reduced maintenance burden)
- Application developers (cleaner API, better type safety)
- Any-Sync integration (proper usage of higher-level APIs)

## Goals / Non-Goals

**Goals:**
- Reduce per-operation code changes from 11 files to 2 files
- Establish protobuf as single source of truth for type definitions
- Properly integrate Any-Sync's SpaceService and ObjectTree APIs
- Make plugin data-model agnostic (opaque bytes for documents)
- Maintain single-command IPC across all platforms
- Generate TypeScript types instead of hand-writing them
- Comprehensive test coverage at critical layers

**Non-Goals:**
- Backward compatibility (clean break is acceptable)
- Supporting multiple sync protocols (Any-Sync only)
- Built-in conflict resolution UI (application responsibility)
- Web/WASM support (deferred)
- Full Any-Sync feature parity in Phase 1 (focus on core operations)

## Architectural Options Considered

### Option A: Abstraction in Go (CHOSEN)

**Pattern:**
```
TypeScript → Rust → Go Dispatcher → Handler → Any-Sync
   (1 call)     (passthrough)    (routing)      (logic)
```

**Characteristics:**
- ✅ Single IPC call per operation (efficient)
- ✅ Go handles all ObjectTree/sync complexity
- ✅ Rust/TypeScript are pure passthrough (no logic)
- ✅ Logic changes only require Go modification
- ✅ Matches proven Anytype pattern

**Why chosen:** Best balance of performance, maintainability, and proper layer separation. Keeps complex Any-Sync integration in Go where it belongs.

### Option B: Abstraction in Rust

**Pattern:**
```
TypeScript → Rust Orchestration → Multiple Go calls → Any-Sync
   (1 call)    (complex logic)      (4+ round trips)
```

**Characteristics:**
- ❌ Multiple IPC calls per operation (performance hit)
- ❌ Rust needs to understand Any-Sync internals
- ❌ Complex error handling across multiple calls
- ❌ Duplicates Go's existing logic

**Why rejected:** Poor performance from multiple IPC round trips. Rust layer should remain thin and simple.

### Option C: Abstraction in TypeScript

**Pattern:**
```
TypeScript Orchestration → Multiple Rust calls → Multiple Go calls → Any-Sync
  (complex logic)            (passthrough)          (raw operations)
```

**Characteristics:**
- ❌ Many IPC calls (worst latency)
- ❌ TypeScript needs to understand Any-Sync internals
- ❌ Every app reinvents the wheel
- ✅ Maximum flexibility for host app

**Why rejected:** Terrible performance. Wrong layer for business logic. Defeats purpose of having a plugin.

### Option D: Raw Any-Sync Passthrough

**Pattern:**
```
TypeScript → Rust → Go → Raw Any-Sync Sync Protocol
```

**Characteristics:**
- ❌ Exposes low-level sync protocol (HeadSync, ObjectSyncStream, etc.)
- ❌ Applications must understand CRDTs, heads, diffs
- ❌ Complex operations like "create document" require multiple protocol steps
- ❌ Wrong abstraction level

**Why rejected:** Any-Sync's protobuf services are the sync *protocol*, not a CRUD API. It's like asking apps to write TCP packets. Applications need higher-level operations like "create document" and "update document" that internally manage ObjectTrees, heads, and sync.

## Key Decisions

### Decision 1: Single Dispatcher Function

**What:** Replace per-operation exports with one `Command(cmd string, data []byte) ([]byte, error)` function.

**Why:**
- Reduces mobile exports from N to 4 functions total
- Simplifies FFI boundary (fewer function signatures)
- Command routing is internal implementation detail
- Adding operations doesn't change mobile API surface
- Matches proven Anytype pattern

**Implementation:**
```go
// The ENTIRE mobile API
func Init(dataPath string) error
func Command(cmd string, data []byte) ([]byte, error)
func SetEventHandler(handler func([]byte))
func Shutdown() error
```

Desktop uses the same pattern (gRPC or simpler IPC).

**Alternatives considered:**
- Keep per-operation exports: Creates N functions, gomobile binding complexity
- Reflection-based dispatch: Performance overhead, less explicit
- HTTP REST API: Too heavyweight for embedded use case

### Decision 2: Protobuf as Single Source of Truth

**What:** Define all operations in unified `syncspace.proto`, generate types for all languages.

**Why:**
- Eliminates hand-written type definitions (error-prone)
- Ensures type consistency across layers
- Changes to schema propagate automatically
- Industry-standard approach for cross-language APIs
- Enables easy addition of new operations

**Generated artifacts:**
- Go: Server stubs and types
- TypeScript: Request/Response types and encode/decode functions
- Rust: Types (optional, for typed internal use)

**Alternatives considered:**
- Hand-written types in each language: Error-prone, inconsistent
- JSON Schema: Less tooling support, no code generation
- Keep current scattered protobuf files: Doesn't solve boilerplate problem

**Tooling choice:** Use `buf` for protobuf management. Single config file (buf.gen.yaml) instead of shell scripts with protoc flags, built-in linting catches proto design issues early, breaking change detection (useful once API stabilizes), and cleaner multi-language generation setup.

### Decision 3: Opaque Data Model (bytes)

**What:** Documents store `bytes data` instead of structured JSON strings or protobuf messages.

**Why:**
- Plugin is data-model agnostic (reusable for any app)
- Applications define their own domain models
- No rigid schema enforcement at plugin level
- Supports future features like encryption (encrypted bytes)
- Matches real-world usage patterns

**API design:**
```protobuf
message DocCreateRequest {
  string space_id = 1;
  string collection = 2;   // Logical grouping
  string id = 3;           // Optional
  bytes data = 4;          // Opaque - app serializes
  map<string, string> metadata = 5;  // Indexable fields
}
```

**Usage pattern:**
```typescript
// App defines its own model
interface Note { title: string; content: string; }

// App serializes its own data
const note: Note = { title: "...", content: "..." };
await anysync.docCreate({
  spaceId: "...",
  collection: "notes",
  data: new TextEncoder().encode(JSON.stringify(note)),
  metadata: { title: note.title }  // For indexing/search
});
```

**Alternatives considered:**
- JSON strings: Forces JSON, no encryption support, still needs parsing
- Predefined protobuf schemas: Too rigid, limits reusability
- Key-value only: Not enough structure for real apps

### Decision 4: Any-Sync Integration Layer

**What:** Use Any-Sync's SpaceService and ObjectTree instead of direct Any-Store.

**Why:**
- Any-Sync provides the abstractions we need (spaces, documents, sync)
- Any-Store is internal to Any-Sync, not a public API surface
- SpaceService handles space lifecycle (create, join, leave)
- ObjectTree handles document CRUD and change tracking
- Built-in sync mechanisms (HeadSync, ObjectSync) work correctly
- Proper CRDT semantics and conflict handling

**Integration points:**
- **SpaceService**: Create space, Join space, Leave space
- **ObjectTree**: Create/update documents (add changes to tree)
- **HeadSync**: Discover remote changes, exchange heads
- **ObjectSync**: Fetch missing changes, merge trees
- **Events**: Bridge Any-Sync internal events to plugin events

**Alternatives considered:**
- Keep direct Any-Store usage: Wrong abstraction, missing sync logic
- Build custom CRDT layer: Reinventing the wheel, bug-prone
- Use only Any-Store and implement sync manually: Complex, error-prone

### Decision 5: Minimal Native Shims

**What:** Reduce iOS/Android native code to ~30 lines of pure passthrough.

**Why:**
- Native layer has zero business logic
- Just bridges Tauri plugin system to Go exports
- Simplifies maintenance (rarely needs changes)
- Clear separation of concerns

**Target implementation:**

**iOS (Swift):**
```swift
// ~30 lines total
class AnySyncPlugin: Plugin {
  func command(_ cmd: String, _ data: Data) -> Data {
    return GoCommand(cmd, data)  // Direct C FFI call
  }
}
```

**Android (Kotlin):**
```kotlin
// ~30 lines total
class AnySyncPlugin : Plugin {
  fun command(cmd: String, data: ByteArray): ByteArray {
    return Anysync.command(cmd, data)  // JNI call
  }
}
```

**Alternatives considered:**
- Complex native logic: Wrong layer, duplicates work
- Eliminate entirely (Rust FFI): Tauri requires native plugin layer
- Keep current per-operation methods: Creates maintenance burden

### Decision 6: Generated TypeScript Client

**What:** Auto-generate typed client from protobuf instead of hand-written functions.

**Why:**
- Eliminates manual type definition errors
- Automatically stays in sync with protobuf schema
- Reduces code review burden (generated code)
- Industry-standard pattern (gRPC-web, etc.)

**Implementation:**
- Use `protobuf-ts` or similar for type generation
- Generate encode/decode functions for each message
- Create thin typed client wrapper (can be generated or mechanical)
- Export both typed client and raw `command()` function

**Generated structure:**
```typescript
// Generated types
export interface DocCreateRequest { ... }
export interface DocCreateResponse { ... }

// Generated or mechanical client
export class AnySyncClient {
  async docCreate(req: DocCreateRequest): Promise<DocCreateResponse> {
    const data = DocCreateRequest.encode(req).finish();
    const response = await command('DocCreate', data);
    return DocCreateResponse.decode(response);
  }
}
```

**Alternatives considered:**
- Hand-written TypeScript: Error-prone, out of sync with protobuf
- Only expose raw command function: Less developer-friendly
- Keep current per-operation functions: Defeats purpose of refactor

### Decision 7: Comprehensive Go Testing

**What:** Write unit tests for each handler and dispatcher tests before integration testing.

**Why:**
- Go layer contains all business logic (must be correct)
- Handlers are testable in isolation (mock Any-Sync)
- Dispatcher routing is critical (wrong handler = wrong behavior)
- Unit tests enable confident refactoring
- Integration tests are expensive (require full stack)

**Test layers:**
1. **Handler unit tests** (critical): Test request/response handling, business logic, error cases
2. **Dispatcher tests** (critical): Test routing, unknown commands, malformed input
3. **Any-Sync integration tests** (important): Test with real SpaceService/ObjectTree
4. **Rust passthrough tests** (minimal): Verify bytes pass through correctly
5. **E2E tests** (important): Full stack validation via example app

**Alternatives considered:**
- Only integration tests: Slow feedback, hard to debug, expensive
- No Go tests (rely on E2E): Catches bugs too late
- Skip integration tests: Miss real Any-Sync behavior

**Decision:** Tier 1 testing (handler unit tests + dispatcher tests + basic E2E) is sufficient for initial implementation. Higher coverage can be added incrementally as the API stabilizes.

### Decision 8: Desktop IPC Mechanism

**What:** Keep gRPC for desktop sidecar communication.

**Why:**
- gRPC provides streaming for free (needed for event subscriptions via Subscribe RPC)
- The "complexity" is already paid—switching costs time for zero user-facing benefit
- If we later want to simplify, the dispatch pattern isolates this decision (change one Rust file)
- The architecture specifically isolates this decision—don't optimize what isn't a problem

**Alternatives considered:**
- Unix socket or stdin/stdout with custom protocol: Simpler in theory, but loses gRPC streaming and tooling, requires custom protocol implementation, and provides no immediate benefit
- REST/HTTP: Too heavyweight, no streaming support

**Implementation note:** Desktop gRPC server will use the same dispatcher that mobile uses, maintaining consistency across platforms.

## Implementation Sequencing

The migration plan defines 7 phases, but key sequencing decisions:

1. **Phase 1-2 (Go Backend)** must complete first - foundation for everything else
2. **Phase 3 (Rust Plugin)** depends on Go being complete
3. **Phase 4 (TypeScript API)** can start once Rust has single command interface
4. **Phase 5 (Native Shims)** can be done in parallel with TypeScript
5. **Phase 6 (Any-Sync Integration)** is interleaved with Phase 2 (handlers use Any-Sync)
6. **Phase 7 (Example App)** validates the full stack

**Critical path:** Go Backend → Rust Plugin → TypeScript API → E2E Tests

**Parallelizable:** Native shims can be done separately once mobile exports are defined

## Success Criteria

1. Adding new operation requires only 2 file changes (protobuf + handler)
2. TypeScript types generated automatically from protobuf
3. Mobile exports total exactly 4 functions
4. Go handler unit tests cover all operations
5. E2E test validates full stack for at least 3 operations
6. Native shims under 50 lines each
7. Example app demonstrates domain service pattern

## Risks and Mitigations

| Risk                                | Impact | Mitigation                                                              |
| ----------------------------------- | ------ | ----------------------------------------------------------------------- |
| Any-Sync learning curve             | High   | Study anytype-heart code, create integration tests early                |
| Large scope, integration bugs       | High   | Implement tests at each phase, validate incrementally                   |
| Breaking change impacts early users | Medium | Acceptable - plugin not widely adopted yet                              |
| Testing burden delays completion    | Medium | Prioritize Tier 1 tests (handlers, dispatcher, basic E2E), defer others |
| Performance regression              | Low    | Single IPC call per operation maintains current performance             |
| Mobile FFI issues                   | Medium | Keep FFI boundary identical to current (gomobile types)                 |

# Any-Sync Plugin Architecture: Deep Analysis

## Reducing Boilerplate: Lessons from Anytype Ecosystem

After analyzing anytype-heart, anytype-ts, and anytype-kotlin, I found a **key architectural pattern** that dramatically reduces boilerplate: the **single dispatch pattern** combined with **protobuf as the single source of truth**.

Our current approach requires modifying multiple files for each new operation. The Anytype pattern reduces this to essentially **2 touches**: define in protobuf, implement the handler.

## How Anytype Does It

### The Single `Command` Function Pattern

Instead of exporting individual functions for each operation, anytype-heart exports just **3 functions** via gomobile:

```go
// clientlibrary/clib/main.go - THE ENTIRE MOBILE API

//export Command
func Command(cmd *C.char, data *C.char, dataLen C.int, callback C.proxyFunc, callbackContext unsafe.Pointer) {
    goCmd := C.GoString(cmd)
    goData := C.GoBytes(unsafe.Pointer(data), dataLen)
    
    service.CommandAsync(goCmd, goData, func(response []byte) {
        // Marshal response back to C and invoke callback
    })
}

//export SetEventHandler
func SetEventHandler(callback C.proxyFunc, callbackContext unsafe.Pointer) { ... }

//export Shutdown
func Shutdown() { ... }
```

The `Command` function takes:
- `cmd`: A command name string (e.g., "ObjectCreate", "SpaceDelete")
- `data`: Serialized protobuf request bytes
- `callback`: Async response handler

Internally, `service.CommandAsync` dispatches to the appropriate handler based on the command name.

### Generated Dispatch Logic

The dispatch logic is **generated from protobuf**. The `ClientCommands` service in protobuf defines all operations:

```protobuf
// pb/protos/service/service.proto
service ClientCommands {
  rpc AppGetVersion(Rpc.App.GetVersion.Request) returns (Rpc.App.GetVersion.Response);
  rpc ObjectCreate(Rpc.Object.Create.Request) returns (Rpc.Object.Create.Response);
  rpc SpaceDelete(Rpc.Space.Delete.Request) returns (Rpc.Space.Delete.Response);
  // ... 200+ operations, all defined here
}
```

Code generation produces:
1. **Go handlers** - Each RPC method becomes a handler function
2. **TypeScript types** - Request/Response interfaces
3. **Java/Kotlin types** - For Android
4. **Swift types** - For iOS

### Frontend Consumption (TypeScript)

The TypeScript frontend uses generated grpc-web clients:

```typescript
// Generated command interface
const C = {
  ObjectCreate: (request: Rpc.Object.Create.Request) => 
    dispatcher.request('ObjectCreate', request),
  SpaceDelete: (request: Rpc.Space.Delete.Request) =>
    dispatcher.request('SpaceDelete', request),
  // ... generated for all operations
};

// Usage
const response = await C.ObjectCreate({
  spaceId: 'abc',
  objectType: 'note'
});
```

### Mobile Consumption (Kotlin)

```kotlin
// Middleware.kt - Thin wrapper over generated types
class Middleware(private val service: MiddlewareService) {
    
    fun objectCreate(command: Command.ObjectCreate): ObjectInfo {
        // Convert to protobuf
        val request = Rpc.Object.Create.Request.newBuilder()
            .setSpaceId(command.spaceId)
            .build()
        
        // Call native (single entry point)
        val response = service.objectCreate(request)
        
        // Convert from protobuf
        return response.toObjectInfo()
    }
}

// MiddlewareServiceImplementation.kt - Generated-ish layer
class MiddlewareServiceImplementation : MiddlewareService {
    override fun objectCreate(request: Rpc.Object.Create.Request): Rpc.Object.Create.Response {
        val bytes = request.toByteArray()
        val responseBytes = Service.objectCreate(bytes) // Native call
        return Rpc.Object.Create.Response.parseFrom(responseBytes)
    }
}
```

---


## Questions Answered

### Question 1: Can we eliminate Kotlin/Swift entirely?

**Short answer: Partially, with tradeoffs.**

#### How gomobile works

gomobile has two outputs:
- **iOS**: `.xcframework` - Contains C-compatible symbols (via `//export` in Go)
- **Android**: `.aar` - Contains JNI wrappers + `.so` native library

#### iOS: Yes, Swift can be eliminated

On iOS, gomobile's `//export` directive creates C-ABI compatible symbols:

```go
//export Command
func Command(cmd *C.char, data unsafe.Pointer, dataLen C.int) *C.char
```

These CAN be called directly from Rust via FFI without Swift:

```rust
// Rust calling Go directly on iOS
extern "C" {
    fn Command(cmd: *const c_char, data: *const u8, len: c_int) -> *const c_char;
}
```

**Current Tauri limitation**: Tauri's iOS plugin system expects Swift plugins. You'd need to either:
1. Keep a minimal Swift shim (just 3 lines that call C functions)
2. Modify how Tauri loads the plugin

#### Android: Harder to eliminate Kotlin

On Android, gomobile generates JNI bindings. The Go code expects to be called via JNI method signatures:

```
Java_com_package_Anysync_command(JNIEnv*, jobject, jstring, jbyteArray)
```

To call Go directly from Rust on Android, you'd need to:
1. **Extract the raw `.so`** from the `.aar`
2. **Use cgo exports** instead of gomobile's JNI bindings
3. This means NOT using `gomobile bind` for Android

**Alternative approach**: Build Go as a pure C library using cgo (not gomobile) for Android. This gives you C-ABI symbols that Rust can call directly.

#### Recommendation

| Platform | Eliminate Native Layer?      | Effort | Risk                  |
| -------- | ---------------------------- | ------ | --------------------- |
| iOS      | Yes, via C FFI               | Low    | Low (proven pattern)  |
| Android  | Possible, needs custom build | Medium | Medium (non-standard) |

**Pragmatic choice**: Keep minimal shims (~30 lines each) rather than fight the tooling. The complexity savings aren't worth the build system gymnastics.

---

### Question 2: Can we just passthrough Any-Sync's raw API?

**Short answer: No, it's too low-level. But your intuition about thin dispatch is correct.**

#### What is Any-Sync's "raw" API?

Any-Sync's protobuf services are the **sync protocol**, not a CRUD API:

```protobuf
service SpaceSync {
  rpc HeadSync(HeadSyncRequest) returns (HeadSyncResponse);      // Compare heads
  rpc StoreDiff(StoreDiffRequest) returns (StoreDiffResponse);   // Compare diffs
  rpc StoreElements(stream StoreKeyValue) returns (stream ...);  // Exchange data
  rpc ObjectSyncStream(stream ObjectSyncMessage) returns (...);  // Sync objects
  rpc SpacePush(SpacePushRequest) returns (SpacePushResponse);   // Push space
  rpc SpacePull(SpacePullRequest) returns (SpacePullResponse);   // Pull space
}
```

#### What does "create a document" actually require?

From my research, creating and syncing an object requires:

1. **Create ObjectTree** with initial change
2. **Update HeadStorage** with new head
3. **Add to space's DiffContainer**
4. **Broadcast ObjectSyncMessage** to peers
5. **Periodic HeadSync** discovers changes on other devices
6. **ObjectSync** exchanges missing changes

This is NOT a single API call - it's a complex orchestration of internal components.

#### Why Anytype-heart exists

Anytype-heart provides ~200 high-level operations that wrap this complexity:
- `ObjectCreate` → internally manages ObjectTree, heads, sync
- `ObjectOpen` → loads from storage, subscribes to changes
- `ObjectUpdate` → adds change to tree, broadcasts

**The raw sync protocol is like TCP - you don't want your app developer writing TCP packets.**

---

### Question 3: Where should the abstraction live?

This is the key architectural question. Let me explore all options:

#### Option A: Abstraction in Go (Current Anytype pattern)

```
TypeScript                    Rust                         Go
─────────────────────────────────────────────────────────────────────
client.documentCreate(req)    
  → invoke('command', ...)    
                              → dispatch(cmd, bytes)
                                  → native_call(...)
                                                             → Dispatcher.handle(cmd)
                                                                 → DocumentService.Create(req)
                                                                     → space.CreateTree()
                                                                     → tree.AddChange()
                                                                     → headStorage.Update()
                                                                     → sync.Broadcast()
                                                             ← response
                              ← bytes
  ← decoded response
```

**Characteristics**:
- ✅ Single IPC call per operation
- ✅ Go handles all ObjectTree/sync complexity
- ✅ Efficient - minimal round trips
- ❌ Any logic change requires Go modification

#### Option B: Abstraction in Rust

```
TypeScript                    Rust                              Go
───────────────────────────────────────────────────────────────────────
client.documentCreate(req)    
  → invoke('command', ...)    
                              → DocumentService.create(req)
                                  → native_call("CreateTree", ...)  ──→ raw tree ops
                                  ← tree_id
                                  → native_call("AddChange", ...) ───→ raw change ops
                                  ← ok
                                  → native_call("UpdateHead", ...) ──→ raw head ops
                                  ← ok
                                  → native_call("Broadcast", ...) ───→ raw sync ops
                                  ← ok
                              ← response
  ← decoded response
```

**Characteristics**:
- ❌ Multiple IPC calls per operation (4+ round trips shown above)
- ❌ Rust needs to understand Any-Sync internals
- ❌ Complex error handling across multiple calls
- ✅ Logic changes can be in Rust
- ❌ Duplicates Go's existing logic

#### Option C: Abstraction in TypeScript

```
TypeScript                           Rust                    Go
─────────────────────────────────────────────────────────────────────
async documentCreate(req) {
  const tree = await invoke('createTree')    → passthrough ──→ raw
  await invoke('addChange', tree, change)    → passthrough ──→ raw
  await invoke('updateHead', tree)           → passthrough ──→ raw
  await invoke('broadcast', tree)            → passthrough ──→ raw
  return { id: tree.id }
}
```

**Characteristics**:
- ❌ Many IPC calls (JS→Rust→Go, each way)
- ❌ Highest latency
- ❌ TypeScript needs to understand Any-Sync internals
- ✅ Maximum flexibility for host app
- ❌ Every app reinvents the wheel

#### Option D: Thin Go abstraction + Passthrough layers (RECOMMENDED)

```
TypeScript                    Rust                         Go
─────────────────────────────────────────────────────────────────────
// Host app defines its OWN higher-level API
class NotesService {
  async createNote(title, content) {
    // Uses plugin's generic document API
    return await anysync.command('DocumentCreate', {
      spaceId: this.spaceId,
      type: 'note',
      data: encode({ title, content })
    });
  }
}

// Plugin provides generic dispatch
anysync.command(cmd, req)
  → invoke('command', cmd, encode(req))
                              → service.dispatch(cmd, bytes)
                                  → native_call(cmd, bytes)
                                                             → Dispatcher.Dispatch(cmd, bytes)
                                                                 → handler(req)
                                                                     // Thin abstraction here
                                                             ← response
                              ← bytes
  ← decode(bytes)
```

**This is the sweet spot**:
- ✅ Single IPC call per operation
- ✅ Rust/TS layers are pure passthrough (no logic)
- ✅ Go layer provides minimal CRUD abstraction
- ✅ Host app builds its own domain model on top
- ✅ Plugin is data-model agnostic

---

## The Recommended Architecture

### Layer 1: Go Backend (Minimal Abstraction)

The Go layer provides a **thin CRUD abstraction** over Any-Sync's internals:

```protobuf
// plugin-go-backend/proto/syncspace.proto
syntax = "proto3";
package anysync.plugin;

// This is the ONLY service - generic enough for any app
service PluginService {
  // Lifecycle
  rpc Init(InitRequest) returns (InitResponse);
  rpc Shutdown(ShutdownRequest) returns (ShutdownResponse);
  
  // Spaces
  rpc SpaceCreate(SpaceCreateRequest) returns (SpaceCreateResponse);
  rpc SpaceJoin(SpaceJoinRequest) returns (SpaceJoinResponse);  
  rpc SpaceLeave(SpaceLeaveRequest) returns (SpaceLeaveResponse);
  rpc SpaceList(SpaceListRequest) returns (SpaceListResponse);
  rpc SpaceDelete(SpaceDeleteRequest) returns (SpaceDeleteResponse);
  
  // Generic Documents (opaque bytes payload)
  rpc DocCreate(DocCreateRequest) returns (DocCreateResponse);
  rpc DocGet(DocGetRequest) returns (DocGetResponse);
  rpc DocUpdate(DocUpdateRequest) returns (DocUpdateResponse);
  rpc DocDelete(DocDeleteRequest) returns (DocDeleteResponse);
  rpc DocList(DocListRequest) returns (DocListResponse);
  rpc DocQuery(DocQueryRequest) returns (DocQueryResponse);
  
  // Sync
  rpc SyncStart(SyncStartRequest) returns (SyncStartResponse);
  rpc SyncPause(SyncPauseRequest) returns (SyncPauseResponse);
  rpc SyncStatus(SyncStatusRequest) returns (SyncStatusResponse);
  
  // Events
  rpc Subscribe(SubscribeRequest) returns (stream Event);
}

message DocCreateRequest {
  string space_id = 1;
  string collection = 2;   // Logical grouping (e.g., "notes", "tasks")
  string id = 3;           // Optional, generated if empty
  bytes data = 4;          // Opaque - app serializes its own format
  map<string, string> metadata = 5;  // Indexable fields
}

message DocCreateResponse {
  string id = 1;
  string version = 2;
}
```

This API is:
- **Generic**: `bytes data` means any app data model works
- **Thin**: Just CRUD + Sync control, no domain logic
- **Complete**: Enough for any local-first app

### Layer 2: Go Dispatch (Single Entry Point)

```go
// plugin-go-backend/shared/dispatcher.go

type Dispatcher struct {
    handlers map[string]func([]byte) ([]byte, error)
}

func NewDispatcher(svc *Service) *Dispatcher {
    d := &Dispatcher{handlers: make(map[string]func([]byte) ([]byte, error))}
    
    // Auto-register handlers (could be generated from protobuf)
    d.Register("Init", svc.handleInit)
    d.Register("SpaceCreate", svc.handleSpaceCreate)
    d.Register("DocCreate", svc.handleDocCreate)
    d.Register("DocGet", svc.handleDocGet)
    // ... etc
    
    return d
}

func (d *Dispatcher) Dispatch(cmd string, data []byte) ([]byte, error) {
    handler, ok := d.handlers[cmd]
    if !ok {
        return nil, fmt.Errorf("unknown command: %s", cmd)
    }
    return handler(data)
}
```

```go
// plugin-go-backend/mobile/exports.go - THE ENTIRE MOBILE API

package mobile

import "github.com/user/plugin/shared"

var (
    dispatcher *shared.Dispatcher
    eventChan  chan []byte
)

// Init initializes the plugin with a data path
func Init(dataPath string) error {
    svc, err := shared.NewService(dataPath)
    if err != nil {
        return err
    }
    dispatcher = shared.NewDispatcher(svc)
    eventChan = make(chan []byte, 100)
    return nil
}

// Command dispatches any command - single entry point
func Command(cmd string, data []byte) ([]byte, error) {
    if dispatcher == nil {
        return nil, errors.New("not initialized")
    }
    return dispatcher.Dispatch(cmd, data)
}

// SetEventHandler registers callback for events
func SetEventHandler(handler func([]byte)) {
    go func() {
        for event := range eventChan {
            handler(event)
        }
    }()
}

// Shutdown cleans up
func Shutdown() error {
    close(eventChan)
    return dispatcher.Shutdown()
}
```

**That's it. 4 functions for the entire API.**

### Layer 3: Rust Plugin (Pure Passthrough)

```rust
// plugin-rust-core/src/lib.rs

mod commands;
mod desktop;
mod mobile;

pub trait AnySyncBackend: Send + Sync {
    fn command(&self, cmd: &str, data: &[u8]) -> Result<Vec<u8>, Error>;
    fn set_event_handler(&self, handler: Box<dyn Fn(Vec<u8>) + Send>);
    fn shutdown(&self) -> Result<(), Error>;
}

pub fn init<R: Runtime>() -> TauriPlugin<R> {
    Builder::new("any-sync")
        .invoke_handler(tauri::generate_handler![
            commands::command,  // SINGLE command handler
        ])
        .setup(|app, api| {
            #[cfg(desktop)]
            let backend = desktop::DesktopBackend::new(app)?;
            
            #[cfg(mobile)]
            let backend = mobile::MobileBackend::new(app)?;
            
            app.manage(Box::new(backend) as Box<dyn AnySyncBackend>);
            Ok(())
        })
        .build()
}
```

```rust
// plugin-rust-core/src/commands.rs

#[tauri::command]
pub async fn command(
    backend: tauri::State<'_, Box<dyn AnySyncBackend>>,
    cmd: String,
    data: Vec<u8>,
) -> Result<Vec<u8>, Error> {
    backend.command(&cmd, &data)
}

// That's it. ONE command.
```

```rust
// plugin-rust-core/src/desktop.rs

pub struct DesktopBackend {
    // gRPC client to sidecar
}

impl AnySyncBackend for DesktopBackend {
    fn command(&self, cmd: &str, data: &[u8]) -> Result<Vec<u8>, Error> {
        // Single gRPC call to sidecar
        let response = self.client.command(cmd, data)?;
        Ok(response)
    }
}
```

```rust
// plugin-rust-core/src/mobile.rs

pub struct MobileBackend {
    // FFI handle
}

impl AnySyncBackend for MobileBackend {
    fn command(&self, cmd: &str, data: &[u8]) -> Result<Vec<u8>, Error> {
        // Single FFI call to Go library
        #[cfg(target_os = "ios")]
        unsafe { ios_command(cmd, data) }
        
        #[cfg(target_os = "android")]
        unsafe { android_command(cmd, data) }
    }
}
```

### Layer 4: TypeScript API (Generated + Passthrough)

```typescript
// plugin-js-api/src/index.ts

import { invoke } from '@tauri-apps/api/core';

// Low-level: direct command passthrough
export async function command(cmd: string, data: Uint8Array): Promise<Uint8Array> {
  const response = await invoke<number[]>('plugin:any-sync|command', {
    cmd,
    data: Array.from(data),
  });
  return new Uint8Array(response);
}

// Generated typed wrappers (from protobuf)
export * from './generated/types';
export * from './generated/client';
```

```typescript
// plugin-js-api/src/generated/client.ts (GENERATED from protobuf)

import { command } from '../index';
import * as pb from './types';

export class AnySyncClient {
  async spaceCreate(req: pb.SpaceCreateRequest): Promise<pb.SpaceCreateResponse> {
    const data = pb.SpaceCreateRequest.encode(req).finish();
    const response = await command('SpaceCreate', data);
    return pb.SpaceCreateResponse.decode(response);
  }
  
  async docCreate(req: pb.DocCreateRequest): Promise<pb.DocCreateResponse> {
    const data = pb.DocCreateRequest.encode(req).finish();
    const response = await command('DocCreate', data);
    return pb.DocCreateResponse.decode(response);
  }
  
  // ... generated for all operations
}

export const anysync = new AnySyncClient();
```

### Layer 5: Host App (Builds on Plugin)

```typescript
// example-app/src/services/notes.ts

import { anysync, DocCreateRequest } from 'tauri-plugin-any-sync-api';

// App defines its OWN data model
interface Note {
  id: string;
  title: string;
  content: string;
  createdAt: Date;
  updatedAt: Date;
}

// App builds higher-level service on plugin
class NotesService {
  private spaceId: string;
  private encoder = new TextEncoder();
  private decoder = new TextDecoder();
  
  async createNote(title: string, content: string): Promise<Note> {
    const note: Note = {
      id: '', // Will be assigned
      title,
      content,
      createdAt: new Date(),
      updatedAt: new Date(),
    };
    
    // Use plugin's generic document API
    const response = await anysync.docCreate({
      spaceId: this.spaceId,
      collection: 'notes',
      data: this.encoder.encode(JSON.stringify(note)),
      metadata: { title }, // Searchable
    });
    
    note.id = response.id;
    return note;
  }
  
  async getNote(id: string): Promise<Note | null> {
    const response = await anysync.docGet({
      spaceId: this.spaceId,
      collection: 'notes',
      id,
    });
    
    if (!response.data) return null;
    return JSON.parse(this.decoder.decode(response.data));
  }
}
```

---

## Summary: What Changes Per New Operation?

### Before (Current)

| Layer            | Files to Modify |
| ---------------- | --------------- |
| Go protobuf      | 1               |
| Go handler       | 1               |
| Go mobile export | 1               |
| Rust command     | 1               |
| Rust lib.rs      | 1               |
| Rust desktop     | 1               |
| Rust mobile      | 1               |
| Rust permissions | 3               |
| TypeScript       | 1               |
| **Total**        | **~11 files**   |

### After (Proposed)

| Layer       | Files to Modify |
| ----------- | --------------- |
| Go protobuf | 1               |
| Go handler  | 1               |
| **Total**   | **2 files**     |

Everything else is either:
- **Generated** (TypeScript types/client)
- **Unchanged** (dispatch mechanism, mobile exports, Rust commands)

---

## Key Decisions Summary

| Question                  | Answer                                               |
| ------------------------- | ---------------------------------------------------- |
| Eliminate Kotlin/Swift?   | Keep minimal (~30 lines), not worth build complexity |
| Raw Any-Sync passthrough? | No, sync protocol too low-level                      |
| Where is abstraction?     | Go layer (thin CRUD over Any-Sync internals)         |
| What about Rust/TS?       | Pure passthrough, no logic                           |
| Host app data model?      | App builds on plugin's generic `bytes data` API      |
| Code generation?          | Yes, for TypeScript types and client                 |

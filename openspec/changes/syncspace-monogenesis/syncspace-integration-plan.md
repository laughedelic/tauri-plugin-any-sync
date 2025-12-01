# SyncSpace API: Phased Any-Sync Integration Plan

## The Core Reality

Your original plan tried to avoid Any-Sync's dependency tree by using spacestorage/anystore directly. **This doesn't work** because:

1. `spacestorage.Create()` requires `SpaceStorageCreatePayload` containing:
   - ACL root record (with owner keys, permissions)
   - Space header (with identity, signing)
   - Space settings root (with ACL head reference)

2. `ObjectTree` (which stores documents) requires:
   - AclList (even for single-owner local operation)
   - Account keys (for signing changes)
   - TreeStorage (backed by anystore, but within Space context)

**You cannot meaningfully store documents without the ACL/key infrastructure.**

However, you CAN defer network operations:
- Coordinator registration
- Remote peer connections
- JoinSpace/LeaveSpace
- HeadSync/ObjectSync

## The Right Mental Model

Any-Sync is designed for **local-first** operation. The architecture is:

```
┌─────────────────────────────────────────────────────────┐
│                    Your App (SyncSpace API)             │
├─────────────────────────────────────────────────────────┤
│  SpaceService (orchestrates spaces)                     │
│    └── Space (contains ObjectTrees)                     │
│          └── ObjectTree (stores documents as DAG)       │
│                └── TreeStorage (persists to anystore)   │
├─────────────────────────────────────────────────────────┤
│  AclList (permissions) ◄── AccountKeys (identity)       │
├─────────────────────────────────────────────────────────┤
│  anystore.DB (SQLite persistence)                       │
└─────────────────────────────────────────────────────────┘

Network layer (can be disabled):
┌─────────────────────────────────────────────────────────┐
│  SyncTree wrapper ──► PeerManager ──► Remote nodes      │
│  Coordinator client ──► Network registration            │
└─────────────────────────────────────────────────────────┘
```

**Anytype-heart's approach**: Initialize the FULL stack, but set `NetworkMode = LocalOnly`. This disables sync while keeping all local operations functional.

---

## Revised Phased Plan

### Phase 1: Account & Identity Foundation

**Goal**: Establish cryptographic identity required by all Any-Sync operations.

**What to implement**:
- Account key generation and storage (`AccountKeys`)
- Device key derivation
- Secure key storage (keychain on mobile, encrypted file on desktop)
- Key loading on Init, cleanup on Shutdown

**SyncSpace operations enabled**:
- ✅ Init (generate or load keys)
- ✅ Shutdown (cleanup)

**Why this first**: Every subsequent operation requires signing/encryption. This is the unavoidable foundation.

**Any-Sync components used**:
- `accountdata.NewRandom()` for key generation
- `accountdata.AccountKeys` struct

---

### Phase 2: Local Space Creation

**Goal**: Create spaces with full Any-Sync structure, but no network.

**What to implement**:
- Generate `SpaceStorageCreatePayload` using `spacepayloads.StoragePayloadForSpaceCreate()`
- Create space storage via `SpaceStorageProvider.CreateSpaceStorage()`
- Initialize in-memory AclList for local-only operation
- Store space metadata (name, created_at) in a dedicated anystore collection (separate from space internals)

**SyncSpace operations enabled**:
- ✅ CreateSpace (full Any-Sync space with ACL)
- ✅ ListSpaces (query metadata collection + enumerate space storages)
- ✅ DeleteSpace (remove space storage)
- ❌ JoinSpace (requires network)
- ❌ LeaveSpace (requires network)

**Key insight**: CreateSpace does NOT require network. The space exists locally and CAN sync later when network is enabled.

**Any-Sync components used**:
- `spacepayloads.StoragePayloadForSpaceCreate()`
- `spacestorage.Create()`
- `list.NewInMemoryDerivedAcl()` (for local ACL)

**Space metadata storage**: Use a separate anystore collection for app-level metadata (space name, user preferences, etc.) that isn't part of Any-Sync's space structure.

---

### Phase 3: Document Operations via ObjectTree ✅ COMPLETED

**Goal**: Store and retrieve documents using ObjectTree within spaces.

**What to implement**:
- Initialize Space with TreeBuilder
- Create ObjectTrees for documents (`TreeBuilder.CreateTree()`)
- Add document content as changes (`ObjectTree.AddContent()`)
- Retrieve documents by reading ObjectTree HEAD state (NOT Root!)
- Store document metadata separately (JSON file, not in ObjectTree)

**SyncSpace operations enabled**:
- ✅ CreateDocument (create ObjectTree + initial change)
- ✅ GetDocument (read ObjectTree HEAD state)
- ✅ UpdateDocument (add change to ObjectTree)
- ✅ DeleteDocument (mark as deleted in metadata)
- ✅ ListDocuments (enumerate from metadata)
- ✅ QueryDocuments (filter by metadata)

**Document structure**:
- Each document = one ObjectTree identified by tree ID
- Document data stored as change payload (opaque bytes)
- Metadata stored in `documents_metadata.json` (separate from ObjectTree)
- Document ID = ObjectTree ID (generated by Any-Sync)

**Any-Sync components used**:
- `objecttreebuilder.TreeBuilder`
- `objecttree.ObjectTree`
- `objecttree.ObjectTreeCreatePayload` (initial creation)
- `objecttree.SignableChangeContent` (updates)

**Critical Implementation Details**:

1. **Version retrieval**: Use `tree.Heads()[0]` + `tree.GetChange(headId)`, NOT `tree.Root()`
   - `Root()` returns the FIRST version (initial creation)
   - `Heads()` returns array of HEAD ids (latest versions)
   - Must call `GetChange()` with head ID to get latest content

2. **Data wrapping**: Any-Sync wraps document data in simple protobuf:
   ```
   Field 1 (tag 0x0a): changeType (string, e.g., "document")
   Field 2 (tag 0x12): changePayload (bytes, actual document data)
   ```
   - NOT the full `RootChange` proto structure
   - Requires custom parsing to extract field 2
   - Implemented as `extractProtobufField(data, 2)` helper

3. **Tree builder patterns**:
   - **Create**: `CreateTree(ObjectTreeCreatePayload)` + `PutTree()` 
   - **Update**: `BuildTree(treeId)` + `AddContent(SignableChangeContent)`
   - **Read**: `BuildTree(treeId)` + `Heads()` + `GetChange()`

4. **Change history**: ObjectTree maintains full version DAG
   - Each update creates new head linked to previous
   - Can traverse history via change parent links
   - Multiple heads possible (conflict branches)

---

### Phase 4: Event System

**Goal**: Stream local changes to subscribers.

**What to implement**:
- Hook into ObjectTree change events
- Hook into Space lifecycle events
- Implement subscriber registry
- Broadcast events to registered handlers

**SyncSpace operations enabled**:
- ✅ Subscribe (local events only at this phase)
  - document.created
  - document.updated
  - document.deleted

**Why before sync**: Apps need change notifications for reactive UI. This works purely locally.

**Implementation approach**: ObjectTree has hooks for change observation. Wrap these and broadcast to Go channel, which feeds the event stream.

---

### Phase 5: Network Sync

**Goal**: Enable synchronization with Any-Sync network.

**What to implement**:
- Network configuration (coordinator address, node addresses)
- Initialize full SpaceService with network components
- Wrap ObjectTree with SyncTree for automatic sync
- Implement peer management
- Handle sync status reporting

**SyncSpace operations enabled**:
- ✅ JoinSpace (network join via invite)
- ✅ LeaveSpace (network departure)
- ✅ StartSync (enable HeadSync/ObjectSync)
- ✅ PauseSync (disable sync)
- ✅ GetSyncStatus (query sync state)
- ✅ Subscribe (now includes sync events)
  - sync.started
  - sync.completed
  - sync.error
  - sync.conflict

**Any-Sync components used**:
- `coordinatorclient.CoordinatorClient`
- `peermanager.PeerManager`
- `nodeconf.Service`
- `synctree.SyncTree` wrapper
- Full SpaceService initialization

**Key change**: Spaces created in Phase 2 now sync automatically when network is enabled. No migration needed—they were created with full Any-Sync structure.

---

## Dependency Summary

| Phase        | Requires Previous | Any-Sync Depth                 | Network Required |
| ------------ | ----------------- | ------------------------------ | ---------------- |
| 1. Account   | -                 | Shallow (keys only)            | No               |
| 2. Spaces    | Phase 1           | Medium (spacestorage, ACL)     | No               |
| 3. Documents | Phase 2           | Deep (ObjectTree, TreeBuilder) | No               |
| 4. Events    | Phase 3           | Same as Phase 3                | No               |
| 5. Sync      | Phase 1-4         | Full stack                     | Yes              |

---

## What CAN'T Be Deferred

- **Account keys**: Required for all signing/encryption (Phase 1)
- **ACL structure**: Required for ObjectTree creation (Phase 2)
- **Space payload**: Can't create "simple" spaces—must use full payload (Phase 2)
- **Understanding ObjectTree versioning**: Must use Heads() not Root() for latest content (Phase 3)
- **Custom protobuf parsing**: Any-Sync's data wrapping requires field extraction (Phase 3)

## What CAN Be Deferred

- **Network registration**: Spaces work locally without coordinator
- **Peer connections**: No sync, but local operations work
- **SyncTree wrapper**: ObjectTree works without it
- **JoinSpace/LeaveSpace**: These are network-only operations
- **Space's key-value store**: Metadata can be stored separately in JSON files

---

## Testing Implications

| Phase | Key Tests                                                                  | Status        |
| ----- | -------------------------------------------------------------------------- | ------------- |
| 1     | Key generation, persistence, loading across restarts                       | ✅ 9 tests     |
| 2     | Space creation with valid payload, space enumeration, deletion             | ✅ 25 tests    |
| 3     | Document CRUD, ObjectTree change history, HEAD retrieval, data unwrapping  | ✅ 15 tests    |
| 4     | Event emission, subscriber filtering, event ordering                       | ⏳ Not started |
| 5     | Sync between two instances, conflict detection, offline/online transitions | ⏳ Not started |

**Current status**: 37 tests passing (Phases 1-3 complete)

Phase 5 tests are the most complex (require multiple instances). Defer until Phases 1-4 are solid.

**Critical test discoveries**:
- Testing HEAD vs Root retrieval required understanding Any-Sync's versioning model
- Data format tests revealed protobuf wrapping (not documented in Any-Sync examples)
- Concurrent test exposed proper Space cleanup requirements
- Test timing important: need 1-second delay between operations for timestamp tests

---

## Summary

**Don't fight Any-Sync's architecture—embrace it in local-only mode.**

The crypto/ACL setup is unavoidable but not complex. The `NewInMemoryDerivedAcl` and `StoragePayloadForSpaceCreate` helpers do the heavy lifting. Once you accept that Phase 1-2 must establish the full local structure, the path becomes clear:

1. **Phase 1**: Keys (small, isolated) ✅ COMPLETE
2. **Phase 2**: Spaces (pulls in ACL, but still no network) ✅ COMPLETE
3. **Phase 3**: Documents (pulls in ObjectTree, the main complexity) ✅ COMPLETE
4. **Phase 4**: Events (hooks into Phase 3) ⏳ Next
5. **Phase 5**: Sync (adds network layer on top) ⏳ Future

Each phase builds on the previous, and Phases 1-4 work entirely offline.

## Key Insights Gained

### ObjectTree Versioning Model
Any-Sync's ObjectTree is a **version DAG** (directed acyclic graph), not a simple key-value store:

- **Root**: The FIRST change ever made (immutable history anchor)
- **Heads**: Current tips of version branches (one or more for conflicts)
- **Changes**: Linked list of modifications with parent pointers

**Critical**: To get the latest document content, you MUST:
```go
heads := tree.Heads()           // Get latest version IDs
change, _ := tree.GetChange(heads[0])  // Retrieve latest change
data := change.Data             // This is the latest content
```

**Don't use** `tree.Root().Data` — that's always the first version!

### Data Wrapping Discovery
Any-Sync wraps all change data in a minimal protobuf structure:

```
Byte 0x0a: Field 1 tag (changeType, wire type 2 = length-delimited)
Byte 0x08: Length of changeType (8 bytes)
Bytes: "document" (the change type string)
Byte 0x12: Field 2 tag (changePayload, wire type 2)
Byte 0x0d: Length of payload (13 bytes in example)
Bytes: [actual document data]
```

This is **NOT** the `RootChange` protobuf structure from `treechangeproto`. It's a simpler wrapper that Any-Sync applies internally. You must parse it manually or use a custom helper.

### Tree Builder Pattern
Two distinct patterns for ObjectTree operations:

1. **Initial Creation** (document doesn't exist yet):
   ```go
   payload := ObjectTreeCreatePayload{
       PrivKey: keys.SignKey,
       ChangeType: "document",
       ChangePayload: data,  // Your raw bytes
       SpaceId: spaceId,
       Timestamp: time.Now().Unix(),
   }
   treePayload, _ := treeBuilder.CreateTree(ctx, payload)
   tree, _ := treeBuilder.PutTree(ctx, treePayload, nil)
   documentId := tree.Id()  // Generated by Any-Sync
   ```

2. **Updates** (document exists, adding new version):
   ```go
   tree, _ := treeBuilder.BuildTree(ctx, documentId, opts)
   content := SignableChangeContent{
       Data: newData,
       Key: keys.SignKey,
       Timestamp: time.Now().Unix(),
       DataType: "document",
   }
   tree.AddContent(ctx, content)
   ```

### Metadata Storage Strategy
**Lesson learned**: Don't try to use Space's key-value store for app-level metadata. Instead:

- Store metadata in separate JSON files (one per space)
- Keep metadata indexable and queryable outside Any-Sync
- Metadata includes: title, tags, created_at, updated_at, custom fields
- ObjectTree stores ONLY the document content (opaque bytes)

This separation provides:
- Fast queries without building trees
- Flexibility in metadata schema
- No Any-Sync overhead for metadata-only operations

### Test Development Insights
1. **Timing matters**: Operations within same second get same timestamp. Add 1-second delays in timestamp assertion tests.
2. **Space cleanup critical**: Must properly close spaces to avoid "tree when unlocked" errors.
3. **Build caching**: Go test cache can hide issues. Use `-count=1` or modify test content to force rebuild.
4. **Hex dumps essential**: When debugging data format issues, print hex dumps to see actual byte structure.

### Architecture Validation
The phased approach WORKS:
- ✅ Phases 1-3 implemented successfully without network
- ✅ Full Any-Sync structure (ACL, Space, ObjectTree) works locally
- ✅ 37 comprehensive tests passing
- ✅ Ready for Phase 4 (Events) without touching network code

**Next steps**: Event system (Phase 4) can now hook into working document operations.

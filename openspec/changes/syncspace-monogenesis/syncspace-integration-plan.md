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

### Phase 3: Document Operations via ObjectTree

**Goal**: Store and retrieve documents using ObjectTree within spaces.

**What to implement**:
- Initialize Space with TreeBuilder
- Create ObjectTrees for documents (`TreeBuilder.CreateTree()`)
- Add document content as changes (`ObjectTree.AddContent()`)
- Retrieve documents by reading ObjectTree state
- Store document metadata in space's key-value store

**SyncSpace operations enabled**:
- ✅ CreateDocument (create ObjectTree + initial change)
- ✅ GetDocument (read ObjectTree state)
- ✅ UpdateDocument (add change to ObjectTree)
- ✅ DeleteDocument (mark as deleted in ObjectTree)
- ✅ ListDocuments (enumerate ObjectTrees in space)
- ✅ QueryDocuments (filter by metadata)

**Document structure**:
- Each document = one ObjectTree
- Document data stored as change payload (your opaque bytes)
- Metadata stored in space's key-value collection (indexable)
- Collection grouping via metadata field

**Any-Sync components used**:
- `objecttreebuilder.TreeBuilder`
- `objecttree.ObjectTree`
- `objecttree.ChangeBuilder`
- Space's key-value store for metadata

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

| Phase | Requires Previous | Any-Sync Depth | Network Required |
|-------|-------------------|----------------|------------------|
| 1. Account | - | Shallow (keys only) | No |
| 2. Spaces | Phase 1 | Medium (spacestorage, ACL) | No |
| 3. Documents | Phase 2 | Deep (ObjectTree, TreeBuilder) | No |
| 4. Events | Phase 3 | Same as Phase 3 | No |
| 5. Sync | Phase 1-4 | Full stack | Yes |

---

## What CAN'T Be Deferred

- **Account keys**: Required for all signing/encryption (Phase 1)
- **ACL structure**: Required for ObjectTree creation (Phase 2)
- **Space payload**: Can't create "simple" spaces—must use full payload (Phase 2)

## What CAN Be Deferred

- **Network registration**: Spaces work locally without coordinator
- **Peer connections**: No sync, but local operations work
- **SyncTree wrapper**: ObjectTree works without it
- **JoinSpace/LeaveSpace**: These are network-only operations

---

## Alternative: anystore-Only Prototype

If you want to validate the API design before committing to full Any-Sync integration, you COULD build a throwaway prototype:

1. Use raw anystore for document storage
2. Implement simple collections for spaces/documents
3. No ACL, no ObjectTree, no sync capability
4. **Explicit limitation**: Data created here CANNOT be migrated to Any-Sync later

**When this makes sense**:
- Testing TypeScript API design
- Validating dispatch pattern
- Building UI before backend is ready

**When to avoid**:
- If you want sync eventually (data won't migrate)
- If you're close to Phase 1-2 anyway (just do it right)

This is a **prototype path**, not a production path.

---

## Testing Implications

| Phase | Key Tests |
|-------|-----------|
| 1 | Key generation, persistence, loading across restarts |
| 2 | Space creation with valid payload, space enumeration, deletion |
| 3 | Document CRUD, ObjectTree change history, metadata queries |
| 4 | Event emission, subscriber filtering, event ordering |
| 5 | Sync between two instances, conflict detection, offline/online transitions |

Phase 5 tests are the most complex (require multiple instances). Defer until Phases 1-4 are solid.

---

## Summary

**Don't fight Any-Sync's architecture—embrace it in local-only mode.**

The crypto/ACL setup is unavoidable but not complex. The `NewInMemoryDerivedAcl` and `StoragePayloadForSpaceCreate` helpers do the heavy lifting. Once you accept that Phase 1-2 must establish the full local structure, the path becomes clear:

1. **Phase 1**: Keys (small, isolated)
2. **Phase 2**: Spaces (pulls in ACL, but still no network)
3. **Phase 3**: Documents (pulls in ObjectTree, the main complexity)
4. **Phase 4**: Events (hooks into Phase 3)
5. **Phase 5**: Sync (adds network layer on top)

Each phase builds on the previous, and Phases 1-4 work entirely offline.

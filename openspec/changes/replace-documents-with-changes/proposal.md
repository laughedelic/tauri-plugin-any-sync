# Change: Replace Document API with Change-Based API

## Why

The current Document API (DocCreate, DocGet, DocUpdate, DocDelete) creates a semantic mismatch with Any-Sync's ObjectTree model:

1. **Wasteful Storage:** Each `DocUpdate` stores a full document snapshot as a new Change node. A 100KB document edited 100 times stores ~10MB instead of deltas.

2. **Hidden History:** ObjectTree maintains a full DAG of changes, but `DocGet` returns only the latest snapshot. Applications cannot access change history, diffs, or merge points.

3. **Broken Conflict Model:** When sync produces multiple heads (concurrent edits), LWW silently picks one. Applications cannot detect, present, or merge conflicts because change structure is opaque.

4. **Wrong Abstraction:** Document CRUD implies "replace the whole thing" semantics, but ObjectTree is designed for "append changes and derive state" semantics. The API fights the storage model.

The document API is essentially a worse key-value store. If you want full-document replacement, use SQLite. ObjectTree's value is in its Merkle-DAG structure for sync—we should expose that, not hide it.

## What Changes

**BREAKING: Remove Document Operations**
- Remove `CreateDocument`, `GetDocument`, `UpdateDocument`, `DeleteDocument`
- Remove `ListDocuments`, `QueryDocuments`
- Remove `Document`, `DocumentInfo`, `DocumentCreatedEvent`, etc.

**Add Object Lifecycle Operations**
- `CreateObject(spaceId, objectType, initialData)` → `objectId, rootChangeId`
- `DeleteObject(spaceId, objectId)` → `existed`
- `ListObjects(spaceId, objectType?, limit?, cursor?)` → `ObjectInfo[], nextCursor`
- `QueryObjects(spaceId, objectType?, filters[], limit?, cursor?)` → `ObjectInfo[], nextCursor`

**Add Change Operations (Core API)**
- `AppendChange(spaceId, objectId, data, dataType?, isSnapshot?, parentIds?)` → `changeId, headIds`
- `GetChanges(spaceId, objectId, sinceChangeId?, fromHeads?, limit?, includeData?)` → `Change[], headIds, hasMore`
- `GetHeads(spaceId, objectId)` → `headIds, hasMultipleHeads`

**Metadata Sync Model**
- **Immutable metadata** (`objectType`): Set at creation, stored in local index, doesn't need to sync
- **Mutable metadata** (title, tags, etc.): Stored as changes with `dataType: "metadata"`, syncs like any other change
- **Local index**: Derived cache rebuilt from changes on startup, used for ListObjects/QueryObjects

**Add Change-Aware Events**
- `ChangeReceivedEvent(objectId, changeId, isLocal, dataType)`
- `HeadsChangedEvent(objectId, oldHeads, newHeads, hasMultipleHeads)`

**Terminology Change**
- "Document" → "Object" (an ObjectTree instance)
- "Version" → "Change" (a node in the ObjectTree DAG)
- "Collection" retained for grouping (becomes `objectType`)

## Impact

**Affected Specs:**
- `syncspace-api` (MODIFIED) - Replace document operations with object/change operations
- `any-sync-integration` (MODIFIED) - Update ObjectTree integration patterns

**Affected Code:**

| Component | Changes |
|-----------|---------|
| `syncspace.proto` | Replace document messages with object/change messages |
| `plugin-go-backend/shared/handlers/documents.go` | Delete, replace with `objects.go` and `changes.go` |
| `plugin-go-backend/shared/anysync/documents.go` | Refactor to expose ObjectTree operations directly |
| `plugin-go-backend/shared/handlers/documents_test.go` | Delete, replace with new test files |
| `plugin-go-backend/shared/anysync/documents_test.go` | Delete, replace with new test files |
| `plugin-js-api/src/generated/` | Auto-regenerated from proto |
| `example-app/src/services/notes.ts` | Rewrite to use change-based operations |
| `example-app/src/store/useNotesStore.ts` | Update to work with change history |

**Not Affected:**
- Space management operations (unchanged)
- Sync control operations (unchanged)
- Single-dispatch architecture (unchanged)
- Rust plugin core (unchanged, uses single command dispatcher)
- Mobile FFI (unchanged, 4-function API)
- Build system (unchanged)

**Relationship to add-network-sync:**
- This change should be implemented BEFORE add-network-sync
- Network sync needs change-level granularity for proper conflict detection
- The `HeadsChangedEvent` provides the hook for conflict UI that add-network-sync defers to apps

**Note:** No existing users, so no migration path or deprecation period needed.

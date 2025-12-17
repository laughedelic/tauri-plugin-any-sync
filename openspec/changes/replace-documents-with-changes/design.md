# Design: Replace Document API with Change-Based API

## Context

The plugin wraps Any-Sync's ObjectTree, a Merkle-DAG structure designed for append-only changes with automatic conflict detection via multiple heads. The current Document API hides this structure behind CRUD semantics, storing full snapshots as change payloads and returning only the latest head.

This design document explains how to expose ObjectTree's native capabilities while remaining simple enough for applications that don't need full DAG access.

**Constraints:**
- Must work with existing single-dispatch architecture
- Must be gomobile-compatible (no complex Go types in FFI)
- Must not require applications to understand Merkle-DAGs to use basic features
- Must enable advanced applications to access full change history
- Should coordinate with add-network-sync proposal (conflict detection via multiple heads)

**Stakeholders:**
- Application developers wanting simple storage (need easy on-ramp)
- Application developers wanting CRDTs/collaboration (need change history)
- Plugin maintainers (keep complexity in Go layer)

## Goals / Non-Goals

**Goals:**
- Expose ObjectTree's change-based model to applications
- Enable access to change history and DAG structure
- Support conflict detection via multiple heads
- Provide efficient incremental sync (changes, not snapshots)
- Allow applications to define their own change formats (deltas, ops, or snapshots)
- Maintain simple API for basic use cases

**Non-Goals:**
- Implement specific CRDT algorithms (application responsibility)
- Provide automatic state reconstruction (application/bridge responsibility)
- Define change payload formats (opaque bytes, app-defined)
- Build conflict resolution UI (application responsibility)

## Decisions

### 1. Objects vs Documents Terminology

**Decision:** Rename "Document" to "Object" throughout the API.

**Why:**
- "Document" implies a file/record that you read and write atomically
- "Object" is neutral and aligns with Any-Sync's "ObjectTree" naming
- Avoids confusion with document-database semantics (MongoDB, CouchDB)
- "Object" can represent any data structure: document, canvas, database, etc.

**Alternatives:**
- Keep "Document": Familiar but misleading about actual semantics
- Use "Tree": Too implementation-focused, confusing for app developers

### 2. Change Payload Semantics

**Decision:** Changes contain opaque `bytes data` with optional `dataType` hint. Plugin never interprets payload content.

**Why:**
- Maximum flexibility for applications (JSON, protobuf, CRDT ops, snapshots)
- Clean separation: plugin handles DAG/sync, app handles data format
- Matches current document data model (already opaque bytes)
- No need to version or evolve plugin-defined formats

**Application patterns:**
```
Pattern A: Full snapshots (simple, like current API)
  Change1: {type: "note", title: "Hello", body: "World"}   // 50 bytes
  Change2: {type: "note", title: "Hello", body: "World!"} // 51 bytes
  → Wasteful but works, easy migration from document API

Pattern B: Operation log (efficient, needs reducer)
  Change1: INIT {title: "Hello", body: "World"}           // 50 bytes
  Change2: PATCH {path: "body", op: "append", value: "!"} // 30 bytes
  → Efficient storage, app must replay to reconstruct state

Pattern C: CRDT updates (Yjs, Automerge, etc.)
  Change1: Y.encodeStateAsUpdate(initialDoc)              // varies
  Change2: Y.encodeUpdate(delta)                          // small
  → Automatic merge, app uses CRDT library
```

**Alternatives:**
- Define structured change format: Limits flexibility, creates versioning burden
- Require deltas only: Breaks simple use cases, complicates initial state

### 3. Snapshot Hint for Optimization

**Decision:** Add `isSnapshot` boolean hint to AppendChange. Plugin stores it in Change metadata but doesn't interpret payload.

**Why:**
- Enables skip-to-snapshot optimization when replaying changes
- App can periodically snapshot to bound replay time
- No plugin logic change—just metadata passthrough
- Useful hint for sync (can skip old changes if snapshot exists)

**Usage:**
```typescript
// Every N changes, create a snapshot
if (changeCount % 100 === 0) {
  await appendChange({
    objectId,
    data: serializeFullState(currentState),
    isSnapshot: true,
    dataType: "myapp.note.snapshot"
  });
}

// When loading, start from latest snapshot
const changes = await getChanges({
  objectId,
  sinceChangeId: latestSnapshotId
});
```

**Alternatives:**
- No snapshot support: Unbounded replay time for long-lived objects
- Plugin-managed snapshots: Complex, requires understanding app data format

### 4. Multiple Heads = Conflict

**Decision:** Expose head count directly. When `headIds.length > 1`, concurrent edits exist and app should handle.

**Why:**
- Direct exposure of ObjectTree semantics
- No hidden LWW that loses data
- App can choose resolution strategy: auto-merge, prompt user, queue for later
- `HeadsChangedEvent` with `hasMultipleHeads` enables real-time conflict UI

**How apps handle conflicts:**
```typescript
const { headIds, hasMultipleHeads } = await getHeads({ objectId });

if (hasMultipleHeads) {
  // Option A: Auto-merge if using CRDT
  const states = await Promise.all(headIds.map(h => getChanges({ fromHeads: [h] })));
  const merged = myCrdt.merge(states);
  await appendChange({ objectId, data: merged, parentIds: headIds });

  // Option B: Prompt user
  showConflictDialog(headIds);

  // Option C: Defer (keep working, merge later)
  continueWithHead(headIds[0]);
}
```

**Alternatives:**
- Hide conflicts with LWW: Loses data, surprises users
- Require immediate resolution: Blocks workflow, bad offline UX
- Plugin-level merge: Requires understanding app data format

### 5. Parent IDs for Explicit DAG Control

**Decision:** AppendChange accepts optional `parentIds`. If empty, append to current heads (automatic). If provided, create change with specific parents.

**Why:**
- Simple case (empty): just append, works like before
- Advanced case (explicit): enables merge commits, branch resolution
- Necessary for conflict resolution (merge commit with multiple parents)
- Matches Git semantics (implicit HEAD vs explicit parent)

**Usage:**
```typescript
// Simple: append to current heads (most common)
await appendChange({ objectId, data });

// Advanced: create merge commit resolving conflict
await appendChange({
  objectId,
  data: mergedState,
  parentIds: [head1, head2]  // Explicit merge
});
```

**Alternatives:**
- No parent control: Can't create merge commits, stuck with conflicts
- Require parents always: Verbose for simple case, error-prone

### 6. GetChanges Pagination and Filtering

**Decision:** GetChanges supports `sinceChangeId`, `fromHeads`, `limit`, and `includeData` parameters for flexible retrieval.

**Why:**
- `sinceChangeId`: Efficient incremental loading (only new changes)
- `fromHeads`: Get path from specific heads to current (for conflict resolution)
- `limit`: Pagination for large histories
- `includeData`: Metadata-only mode for sync negotiation (saves bandwidth)

**Common patterns:**
```typescript
// Initial load: all changes (or from latest snapshot)
const all = await getChanges({ objectId, includeData: true });

// Incremental: changes since last known
const new = await getChanges({
  objectId,
  sinceChangeId: lastKnownChangeId,
  includeData: true
});

// Conflict analysis: what's on each head?
const branch1 = await getChanges({ objectId, fromHeads: [head1] });
const branch2 = await getChanges({ objectId, fromHeads: [head2] });

// Sync negotiation: just IDs
const metadata = await getChanges({ objectId, includeData: false });
```

**Alternatives:**
- Single retrieval mode: Forces loading entire history always
- Separate endpoints: More API surface, harder to discover

### 7. Mutable Metadata Must Live in ObjectTree

**Decision:** All mutable metadata (title, tags, properties) MUST be stored as changes in ObjectTree, not in a separate cache. The local metadata index is a derived cache rebuilt from changes.

**Why:**
- Any-Sync syncs ObjectTree changes, nothing else
- Metadata stored outside ObjectTree won't sync across devices
- The `dataType` field distinguishes metadata changes from content changes
- Local index is rebuilt on load by scanning metadata changes

**Two types of metadata:**

1. **Immutable (set at creation, doesn't change):**
   - `objectType` - stored in plugin index, doesn't need to sync
   - Set via `CreateObject`, never updated

2. **Mutable (user can change, must sync):**
   - Title, tags, properties - stored as changes in ObjectTree
   - Updated via `AppendChange` with `dataType: "metadata"`
   - Syncs like any other change

**How mutable metadata works:**
```typescript
// Create object - objectType is immutable, set once
const { objectId } = await createObject({
  spaceId,
  objectType: "note",
  initialData: encode({ title: "Hello", content: "World" })
});

// Update title - append a metadata change
await appendChange({
  spaceId,
  objectId,
  data: encode({ title: "Updated Title" }),
  dataType: "metadata"  // Distinguishes from content changes
});

// Both changes sync to other devices
// App reconstructs current title by replaying metadata changes
```

**Local index as derived cache:**
```
ObjectTree (source of truth, syncs)     Local Index (derived, local only)
├─ Change1: { title: "Hello", ... }     ├─ objectId → {
├─ Change2: { title: "Updated" }        │    objectType: "note",
├─ Change3: content update              │    latestTitle: "Updated",  ← derived
└─ ...                                  │    latestTags: ["work"],    ← derived
                                        │ }
On startup/sync: rebuild index by       Used for: ListObjects, QueryObjects
scanning changes with dataType="metadata"   (efficient local queries)
```

**No dedicated UpdateMetadata RPC:**
- Apps use `AppendChange` with `dataType: "metadata"`
- Keeps API thin - metadata is just another change type
- Metadata format is app-defined (plugin doesn't interpret)

**Change metadata (system-managed, always syncs):**
```typescript
const changes = await getChanges({ objectId });
changes.forEach(c => {
  c.changeId;        // CID (content-addressed)
  c.timestamp;       // When created
  c.authorIdentity;  // Signing key
  c.parentIds;       // DAG structure
  c.isSnapshot;      // Optimization hint
  c.dataType;        // App hint: "content", "metadata", "yjs.update"
  c.orderIndex;      // Deterministic ordering
});
```

**Alternatives considered:**
- Separate metadata sync: Would require a second sync mechanism, complexity
- Metadata in plugin index only: Won't sync, broken multi-device experience
- Dedicated UpdateMetadata RPC: More API surface, no real benefit over AppendChange

### 8. Event Model for Changes

**Decision:** Two events: `ChangeReceivedEvent` (per change) and `HeadsChangedEvent` (state transition).

**Why:**
- `ChangeReceivedEvent`: React to each change (update CRDT state incrementally)
- `HeadsChangedEvent`: React to head transitions, especially multiple heads (conflict)
- Separation enables both incremental updates and conflict detection
- `isLocal` flag distinguishes local edits from sync-received changes

**Event handling:**
```typescript
subscribe({
  eventTypes: ["change.received", "heads.changed"],
  spaceIds: [spaceId]
}).on("event", (event) => {
  if (event.type === "change.received") {
    const { objectId, changeId, isLocal, dataType } = event.payload;
    if (!isLocal) {
      // Remote change arrived, update local state
      const change = await getChange(changeId);
      applyChange(change);
    }
  }
  if (event.type === "heads.changed") {
    const { objectId, hasMultipleHeads, oldHeads, newHeads } = event.payload;
    if (hasMultipleHeads) {
      showConflictIndicator(objectId);
    }
  }
});
```

**Alternatives:**
- Single event type: Can't distinguish change arrival from conflict state
- No events: Requires polling, misses real-time updates

### 9. QueryObjects for Metadata Filtering

**Decision:** Provide `QueryObjects` that filters the derived local index using simple operators. The index is rebuilt from ObjectTree changes on startup and kept in sync via change events.

**Why:**
- Enables efficient local queries without scanning all changes every time
- Index is derived from ObjectTree (source of truth), not a separate data store
- Keeps the thin API pattern: plugin manages index, Any-Sync manages changes
- Enables discovery patterns ("all notes with tag X") efficiently

**Implementation:**
- Index stored in `{dataDir}/objects/{spaceId}.json` (local cache)
- Index rebuilt on startup by scanning changes with `dataType: "metadata"`
- Index updated when `ChangeReceivedEvent` arrives with metadata dataType
- Query filters apply to indexed fields
- Simple operators: `eq`, `ne`, `contains`, `startsWith`
- No complex query language (that's application territory)

**Index rebuild flow:**
```
Startup:
  1. Load all objects in space
  2. For each object, scan changes where dataType="metadata"
  3. Apply metadata changes in order to build current state
  4. Store in local index

On ChangeReceivedEvent (dataType="metadata"):
  1. Parse metadata from change payload
  2. Update index entry for that object
  3. Index now reflects latest state
```

**Usage:**
```typescript
// Find objects by indexed metadata
const results = await queryObjects({
  spaceId,
  objectType: "note",
  filters: [
    { field: "status", operator: "eq", value: "published" },
    { field: "tags", operator: "contains", value: "work" }
  ],
  limit: 20
});
```

**Note on queryable fields:**
- Applications define which metadata fields exist (plugin doesn't interpret)
- Plugin indexes whatever metadata changes contain
- Query filters match against indexed field values

**Alternatives:**
- No query, only list: Forces client-side filtering, inefficient for large collections
- Full query language: Over-engineering, applications can build on top
- Scan changes on every query: Too slow for large objects

## Risks / Trade-offs

**Risks:**
- **Learning curve**: DAG model less familiar than CRUD
  - Mitigation: Document common patterns, provide examples
- **Performance**: Change replay slower than direct read
  - Mitigation: Snapshot hints, client-side caching

**Trade-offs:**
- Flexibility vs simplicity: More powerful but requires more app code
  - Chosen: Flexibility, with bridge library for simple cases
- Thin API vs rich API: Could provide more convenience methods
  - Chosen: Thin API, avoid second-guessing app needs

**Note:** No existing users, so no migration concerns.

## Open Questions

1. **UpdateObjectMetadata operation?** Should we allow updating object metadata without appending a change? Current design: no, keep metadata immutable after creation. Could add if needed.

2. **GetChange by ID?** Should we support fetching a single change by ID? Useful for event handling. Current: included in GetChanges response. Could add dedicated RPC.

3. **Change ordering guarantees?** Should `orderIndex` be globally ordered or per-object? Current: per-object, deterministic within object's DAG.

4. **Pruning/compaction?** Should we expose pruning to remove old changes before a snapshot? Deferred: not needed for MVP, can add later.

# Design: Minimal AnyStore Integration

## Context

This change introduces the first functional integration with AnyStore, moving the plugin from pure scaffolding to working storage capabilities. The implementation must validate the full architectural stack (TypeScript → Rust → gRPC → Go → AnyStore) while remaining minimal enough to be completed in Phase 1.

**Constraints:**
- Desktop-only for now (mobile will use same gRPC API later via gomobile)
- Single AnyStore database instance per application
- No authentication or multi-user support yet
- No sync capabilities (local storage only)
- Keep API surface small (3 operations total)

**Stakeholders:**
- Plugin users who need local document storage
- Future mobile platform implementers (API design must be gomobile-compatible)
- AnySync integration developers (foundation for sync features)

## Goals / Non-Goals

**Goals:**
- Prove end-to-end desktop sidecar architecture works correctly
- Establish patterns for adding more storage operations
- Provide usable local storage for simple use cases
- Validate AnyStore integration approach
- Keep implementation under 500 LOC across all layers

**Non-Goals:**
- Full MongoDB-compatible query API (just basic CRUD)
- Index management (use AnyStore defaults)
- Transaction support (single-document operations only)
- Sync, conflict resolution, or multi-device support
- Mobile platform implementation (desktop validation first)
- Schema validation or type enforcement

## Decisions

### Decision 1: Storage Wrapper Layer

**What:** Create `internal/storage/anystore.go` wrapper around AnyStore rather than using it directly in gRPC handlers.

**Why:**
- Abstracts AnyStore-specific types from gRPC layer
- Enforces gomobile-compatible type boundaries early
- Simplifies testing with mock storage
- Allows future storage backend swapping if needed

**Alternatives considered:**
- Direct AnyStore usage in gRPC handlers: Harder to test, couples gRPC to AnyStore types
- Generic storage interface: Over-engineering for single implementation

### Decision 2: Document Format (JSON Strings)

**What:** Store documents as JSON strings, not structured protobuf messages.

**Why:**
- AnyStore is schema-less and works with `anyenc.Value` (JSON-like)
- Avoids defining rigid protobuf schemas for documents
- Allows application flexibility in document structure
- Matches MongoDB-style document databases
- Simplifies gomobile compatibility (strings are primitive types)

**Alternatives considered:**
- Protobuf messages: Would require pre-defined schemas, limiting flexibility
- Raw bytes: Less developer-friendly, no JSON validation

### Decision 3: Minimal Operation Set (Put/Get/List)

**What:** Implement only 3 operations:
- `Put(collection, id, json)` - Upsert document
- `Get(collection, id)` - Retrieve document
- `List(collection)` - Get all IDs

**Why:**
- Sufficient to demonstrate full stack integration
- Easy to test and validate
- Provides foundation for more operations
- Keeps scope minimal and achievable

**Operations deliberately excluded (future work):**
- Query/filter operations
- Delete operations
- Bulk operations
- Index management

### Decision 4: Collection-Based Organization

**What:** Use collection names as organizational units (like MongoDB collections or SQL tables).

**Why:**
- Matches AnyStore's API (`db.Collection(ctx, name)`)
- Familiar pattern for developers
- Allows logical grouping of related documents
- Supports future schema-per-collection features

**Alternatives considered:**
- Flat key-value store: Less organized, harder to iterate
- Pre-defined collections: Too rigid for Phase 1

### Decision 5: Database Lifecycle

**What:**
- Initialize AnyStore on sidecar startup
- Create database file in platform-specific app data directory
- Close database on sidecar shutdown
- No manual database creation/deletion APIs

**Why:**
- Simplifies application code (automatic setup)
- Matches embedded database pattern
- Avoids database lifecycle management complexity
- Uses Tauri's path APIs for correct platform locations

**Alternatives considered:**
- Manual database management: Added complexity without clear benefit
- In-memory database: Defeats persistence purpose

### Decision 6: Error Handling Strategy

**What:**
- Return specific gRPC error codes: `INVALID_ARGUMENT`, `NOT_FOUND`, `INTERNAL`
- Wrap AnyStore errors with context (collection name, document ID)
- Propagate errors through all layers without silent failures

**Why:**
- Enables meaningful error messages in UI
- Supports debugging and troubleshooting
- Maintains error context across language boundaries
- Aligns with gRPC best practices

### Decision 7: Protobuf Message Design

**What:**
```protobuf
message PutRequest {
  string collection = 1;
  string id = 2;
  string document_json = 3;
}

message GetRequest {
  string collection = 1;
  string id = 2;
}

message ListRequest {
  string collection = 1;
}
```

**Why:**
- Simple, gomobile-compatible types (all strings)
- Clear semantics for each operation
- Extensible (can add fields without breaking changes)

**Alternatives considered:**
- Batch operations: Added complexity for unclear benefit
- Structured document fields: Defeats schema-less purpose

## Risks / Trade-offs

### Risk: AnyStore Dependency Maintenance

**Risk:** AnyStore is relatively new (GitHub shows ~35 stars) and may have breaking changes.

**Mitigation:**
- Pin specific AnyStore version in `go.mod`
- Wrapper layer isolates impact of AnyStore API changes
- Monitor AnyStore repository for major releases
- Establish update testing procedures

### Risk: Performance with Large Collections

**Risk:** `List` operation could be slow/memory-intensive for large collections.

**Mitigation:**
- Document performance characteristics in API docs
- Add pagination support in future iteration if needed
- Recommend using queries instead of full list (future feature)

### Risk: JSON Validation

**Risk:** Invalid JSON in `Put` requests could cause errors at storage layer.

**Mitigation:**
- Validate JSON in Go backend before storing
- Return clear error messages for malformed JSON
- Add JSON validation examples to documentation

### Risk: gomobile Compatibility Unknown

**Risk:** While designed for gomobile compatibility, not yet tested on mobile platforms.

**Mitigation:**
- Keep all protobuf types gomobile-compatible (strings, primitives)
- Plan mobile validation in Phase 2
- Document any discovered mobile limitations

## Migration Plan

### Phase 1: Desktop Implementation (This Change)

1. Add AnyStore dependency to Go backend
2. Implement storage wrapper and gRPC service
3. Add Rust command handlers (desktop-only)
4. Create TypeScript API bindings
5. Update example app with storage demo
6. Validate end-to-end on macOS, Linux, Windows

### Phase 2: Mobile Validation (Future)

1. Test same gRPC API with gomobile bindings
2. Verify JSON string handling on iOS/Android
3. Adjust wrapper if mobile compatibility issues found

### Phase 3: Storage API Expansion (Future)

1. Add Query operation with MongoDB-style filters
2. Add Delete operation
3. Add bulk Put/Delete operations
4. Add index management APIs

### Rollback Plan

If critical issues discovered:
1. AnyStore integration is isolated in Go backend
2. Can revert by removing storage commands from Rust plugin
3. No schema migrations needed (schema-less design)
4. Database file can be deleted without affecting app functionality

## Open Questions

1. **Database file location:** Use Tauri's app data dir, or allow configuration?
   - **Answer:** Start with app data dir, add configuration if requested

2. **Collection naming restrictions:** Should we validate collection names?
   - **Answer:** Pass through to AnyStore, document any errors

3. **Concurrent access:** How to handle multiple processes accessing same database?
   - **Answer:** Out of scope for Phase 1 (single sidecar instance only)

4. **Document size limits:** Should we enforce maximum document size?
   - **Answer:** Let AnyStore handle limits, document behavior

5. **Error message detail:** How much AnyStore error detail should we expose?
   - **Answer:** Expose full context for debugging, sanitize in future if needed

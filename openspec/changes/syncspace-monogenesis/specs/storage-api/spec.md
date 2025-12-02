# Storage API Specification

## REMOVED Requirements

This capability is being REPLACED by the new `syncspace-api` capability.

### Requirement: Storage Service Definition

~~The Go backend SHALL provide a gRPC StorageService with four CRUD operations for document storage and retrieval.~~

**Reason:** Replaced by SyncSpace document operations with opaque data model and Any-Sync integration.

### Requirement: AnyStore Integration Layer

~~The Go backend SHALL provide a storage wrapper in `internal/storage/anystore.go` that abstracts AnyStore-specific types.~~

**Reason:** Direct Any-Store usage replaced by Any-Sync higher-level APIs (ObjectTree).

### Requirement: Protobuf Message Definitions

~~The storage service SHALL define gomobile-compatible protobuf messages using only primitive types.~~

**Reason:** Replaced by unified syncspace.proto defining SyncSpace API.

### Requirement: JSON Document Validation

~~The storage service SHALL validate JSON documents before storing them in AnyStore.~~

**Reason:** Plugin now uses opaque bytes, applications handle their own validation.

### Requirement: Collection-Based Organization

~~The storage service SHALL organize documents into named collections matching AnyStore's collection API.~~

**Reason:** Collection concept retained but now part of SyncSpace document API with opaque data.

### Requirement: Error Context Propagation

~~The storage service SHALL include collection name and document ID in error messages for debugging.~~

**Reason:** Error handling pattern retained but now applies to all SyncSpace operations, not just storage.

## Migration Notes

The `storage-api` capability is being **completely replaced** by the new `syncspace-api` capability. The new API:

- Uses opaque `bytes data` instead of JSON strings
- Integrates with Any-Sync's ObjectTree instead of direct Any-Store
- Adds space management operations
- Adds sync control operations
- Adds event streaming
- Maintains collection-based organization
- Maintains error context propagation patterns

Applications using the old storage API must:
1. Migrate to the SyncSpace document API
2. Handle their own data serialization (e.g., JSON.stringify/parse)
3. Create domain service layers encapsulating their data models
4. Update to use space management if needed

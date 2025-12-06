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

## ADDED Requirements

### Requirement: Deprecated in Favor of SyncSpace API

The storage API SHALL be considered deprecated in favor of the new SyncSpace API.

#### Scenario: Applications must migrate to SyncSpace API

- **GIVEN** an application using the old storage API
- **WHEN** upgrading to the new plugin version
- **THEN** the application must migrate to use the SyncSpace document operations
- **AND** the new `syncspace-api` provides opaque bytes data and Any-Sync integration
- **AND** the new API adds space management, sync control, and event streaming

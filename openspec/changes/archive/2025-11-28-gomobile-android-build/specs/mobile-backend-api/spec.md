# Spec Delta: Mobile Backend API

## ADDED Requirements

### Requirement: gomobile-Compatible Storage API
The Go backend SHALL provide a mobile-specific entrypoint with gomobile-compatible function signatures for all storage operations.

#### Scenario: Initialize storage on mobile
- **GIVEN** a mobile app using the plugin
- **WHEN** the app initializes the Go backend with `InitStorage(dbPath string)`
- **THEN** the backend opens an AnyStore database at the specified path
- **AND** returns nil on success or an error on failure
- **AND** the database connection is ready for storage operations

#### Scenario: Store document via mobile API
- **GIVEN** storage initialized successfully
- **WHEN** calling `StoragePut(collection, id, documentJson string) error`
- **THEN** the document is stored in the specified collection with the given ID
- **AND** returns nil on success or descriptive error on failure
- **AND** the operation is idempotent (subsequent calls update the document)

#### Scenario: Retrieve document via mobile API
- **GIVEN** a document exists in storage
- **WHEN** calling `StorageGet(collection, id string) (string, error)`
- **THEN** returns the document as JSON string and nil error
- **AND** when document doesn't exist, returns empty string and nil error (not an error condition)
- **AND** when collection doesn't exist, returns empty string and nil error

#### Scenario: Delete document via mobile API
- **GIVEN** a document exists in storage
- **WHEN** calling `StorageDelete(collection, id string) (bool, error)`
- **THEN** returns true and nil error when document existed and was deleted
- **AND** returns false and nil error when document didn't exist (idempotent)
- **AND** returns error only for actual storage failures

#### Scenario: List documents via mobile API
- **GIVEN** a collection with multiple documents
- **WHEN** calling `StorageList(collection string) (string, error)`
- **THEN** returns a JSON array string of document IDs: `["id1","id2","id3"]`
- **AND** returns empty array `[]` when collection is empty or doesn't exist
- **AND** returns error only for actual storage failures

### Requirement: Type Compatibility
All exported mobile functions SHALL use only gomobile-compatible types.

#### Scenario: Function signature validation
- **GIVEN** the mobile package is defined
- **WHEN** building with `gomobile bind -target=android`
- **THEN** all exported functions compile without type errors
- **AND** no complex types (maps, channels, interfaces, pointers) in signatures
- **AND** only primitives (string, bool, int, float64) and []byte in signatures
- **AND** standard Go error type used for error returns

### Requirement: Shared Backend Code
The mobile entrypoint SHALL reuse >95% of the existing storage implementation.

#### Scenario: Code reuse validation
- **GIVEN** the mobile and desktop implementations
- **WHEN** analyzing the codebase
- **THEN** both use `internal/storage/anystore.go` for core logic
- **AND** both use the same AnyStore database engine
- **AND** only API boundary differs (gRPC server vs direct function exports)
- **AND** no duplicate storage logic exists

### Requirement: State Management
The mobile backend SHALL manage database connection lifecycle internally.

#### Scenario: Database connection lifecycle
- **GIVEN** mobile app starts and calls `InitStorage`
- **WHEN** the Go backend initializes
- **THEN** a single database connection is created and stored internally
- **AND** subsequent storage operations reuse the same connection
- **AND** connection remains open until process termination
- **AND** no explicit close method required (handled by process cleanup)

## MODIFIED Requirements

None. This is new mobile-specific functionality that doesn't modify existing desktop behavior.

## REMOVED Requirements

None.

## Dependencies

- **Internal:** Requires `internal/storage/anystore.go` (already implemented)
- **External:** Requires `github.com/anyproto/any-store` (already in go.mod)
- **Related Specs:** `storage-api/spec.md` (defines core storage semantics)

## Notes

**Cross-Platform Consistency:** The mobile API is designed to provide the same functionality as the desktop gRPC API, just with a different transport mechanism. The core storage behavior must remain identical across platforms.

**Error Handling Philosophy:** Following Go conventions, errors are returned as the last return value. gomobile automatically converts Go errors to exceptions in Java/Kotlin, which are then caught and propagated through the Tauri plugin error handling.

**Performance Considerations:** Direct function calls avoid the gRPC overhead present on desktop, making mobile operations potentially faster despite running on less powerful hardware.

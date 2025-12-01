# SyncSpace API Specification

## ADDED Requirements

### Requirement: Unified Protobuf Schema

The plugin SHALL define all operations in a single `syncspace.proto` file that serves as the single source of truth for the SyncSpace API.

#### Scenario: Schema includes lifecycle operations

- **GIVEN** the SyncSpace API protobuf schema
- **WHEN** the schema is reviewed
- **THEN** it defines Init and Shutdown RPC methods

#### Scenario: Schema includes space operations

- **GIVEN** the SyncSpace API protobuf schema
- **WHEN** the schema is reviewed
- **THEN** it defines SpaceCreate, SpaceJoin, SpaceLeave, SpaceList, and SpaceDelete RPC methods

#### Scenario: Schema includes document operations

- **GIVEN** the SyncSpace API protobuf schema
- **WHEN** the schema is reviewed
- **THEN** it defines DocCreate, DocGet, DocUpdate, DocDelete, DocList, and DocQuery RPC methods

#### Scenario: Schema includes sync control operations

- **GIVEN** the SyncSpace API protobuf schema
- **WHEN** the schema is reviewed
- **THEN** it defines SyncStart, SyncPause, and SyncStatus RPC methods

#### Scenario: Schema includes event streaming

- **GIVEN** the SyncSpace API protobuf schema
- **WHEN** the schema is reviewed
- **THEN** it defines Subscribe RPC method with server streaming response

### Requirement: Opaque Document Data Model

Document operations SHALL use `bytes data` for opaque application payloads, making the plugin data-model agnostic.

#### Scenario: DocCreate accepts opaque bytes

- **GIVEN** a DocCreateRequest message definition
- **WHEN** the message structure is reviewed
- **THEN** it contains a `bytes data` field for the document payload

#### Scenario: DocCreate includes metadata for indexing

- **GIVEN** a DocCreateRequest message definition
- **WHEN** the message structure is reviewed
- **THEN** it contains a `map<string, string> metadata` field for indexable fields

#### Scenario: DocCreate includes collection for grouping

- **GIVEN** a DocCreateRequest message definition
- **WHEN** the message structure is reviewed
- **THEN** it contains a `string collection` field for logical grouping

#### Scenario: DocGet returns opaque bytes

- **GIVEN** a DocGetResponse message definition
- **WHEN** the message structure is reviewed
- **THEN** it contains a `bytes data` field containing the original document payload

#### Scenario: Application serializes its own data format

- **GIVEN** an application with domain-specific data models
- **WHEN** the application creates a document
- **THEN** it serializes its data to bytes (e.g., JSON, protobuf, msgpack) before calling DocCreate

### Requirement: Space Management

The SyncSpace API SHALL provide operations for creating and managing isolated synchronization spaces.

#### Scenario: SpaceCreate establishes new space

- **GIVEN** a SpaceCreateRequest with space configuration
- **WHEN** SpaceCreate is called
- **THEN** a new space is created via Any-Sync SpaceService and space ID is returned

#### Scenario: SpaceJoin adds device to existing space

- **GIVEN** a SpaceJoinRequest with space ID and optional invite credentials
- **WHEN** SpaceJoin is called
- **THEN** the device joins the space via Any-Sync SpaceService

#### Scenario: SpaceLeave removes device from space

- **GIVEN** a SpaceLeaveRequest with space ID
- **WHEN** SpaceLeave is called
- **THEN** the device leaves the space but space data persists locally

#### Scenario: SpaceList enumerates all spaces

- **GIVEN** multiple spaces have been created or joined
- **WHEN** SpaceList is called
- **THEN** all space IDs are returned

#### Scenario: SpaceDelete removes space and data

- **GIVEN** a SpaceDeleteRequest with space ID
- **WHEN** SpaceDelete is called
- **THEN** the space and all associated data are deleted locally

### Requirement: Document CRUD Operations

The SyncSpace API SHALL provide create, read, update, delete, list, and query operations for documents within spaces.

#### Scenario: DocCreate adds new document to space

- **GIVEN** a DocCreateRequest with space ID, collection, and data
- **WHEN** DocCreate is called
- **THEN** a new ObjectTree is created via Any-Sync and document ID is returned

#### Scenario: DocCreate generates ID if not provided

- **GIVEN** a DocCreateRequest without an ID field
- **WHEN** DocCreate is called
- **THEN** a unique document ID is generated automatically

#### Scenario: DocGet retrieves document by ID

- **GIVEN** a DocGetRequest with space ID, collection, and document ID
- **WHEN** DocGet is called
- **THEN** the document data is retrieved from Any-Sync ObjectTree

#### Scenario: DocUpdate modifies existing document

- **GIVEN** a DocUpdateRequest with space ID, collection, document ID, and new data
- **WHEN** DocUpdate is called
- **THEN** a new change is added to the document's ObjectTree

#### Scenario: DocDelete removes document

- **GIVEN** a DocDeleteRequest with space ID, collection, and document ID
- **WHEN** DocDelete is called
- **THEN** the document is marked for deletion via Any-Sync

#### Scenario: DocList returns all document IDs in collection

- **GIVEN** a DocListRequest with space ID and collection
- **WHEN** DocList is called
- **THEN** all document IDs in the collection are returned

#### Scenario: DocQuery filters documents by metadata

- **GIVEN** a DocQueryRequest with space ID, collection, and metadata filters
- **WHEN** DocQuery is called
- **THEN** only documents matching the metadata criteria are returned

### Requirement: Sync Control

The SyncSpace API SHALL provide operations to control synchronization behavior.

#### Scenario: SyncStart initiates synchronization

- **GIVEN** a space with SyncStart called
- **WHEN** changes are made to documents
- **THEN** Any-Sync HeadSync and ObjectSync mechanisms exchange changes with peers

#### Scenario: SyncPause suspends synchronization

- **GIVEN** a space with active sync
- **WHEN** SyncPause is called
- **THEN** synchronization stops until SyncStart is called again

#### Scenario: SyncStatus reports current state

- **GIVEN** a space with sync either running or paused
- **WHEN** SyncStatus is called
- **THEN** the current sync state and statistics are returned

### Requirement: Event Streaming

The SyncSpace API SHALL provide server-streaming events for document changes and sync status updates.

#### Scenario: Subscribe streams document change events

- **GIVEN** a Subscribe call for a space
- **WHEN** a document is created, updated, or deleted (locally or remotely)
- **THEN** an Event message is streamed to the subscriber

#### Scenario: Subscribe streams sync status events

- **GIVEN** a Subscribe call for a space
- **WHEN** sync state changes (started, paused, syncing, error)
- **THEN** an Event message is streamed to the subscriber

#### Scenario: Events bridge from Any-Sync internal events

- **GIVEN** Any-Sync generates internal events (ObjectTree changes, HeadSync updates)
- **WHEN** these events occur
- **THEN** they are transformed and streamed to plugin event subscribers

### Requirement: Gomobile Compatibility

All protobuf message types SHALL use only gomobile-compatible types (primitives, strings, bytes, maps of primitives).

#### Scenario: No complex nested structures in messages

- **GIVEN** all protobuf message definitions in syncspace.proto
- **WHEN** messages are reviewed for gomobile compatibility
- **THEN** no message contains slices of messages or complex nested structures incompatible with gomobile

#### Scenario: Maps use string keys and primitive values

- **GIVEN** metadata field in DocCreateRequest
- **WHEN** the field type is reviewed
- **THEN** it is defined as `map<string, string>` (gomobile-compatible)

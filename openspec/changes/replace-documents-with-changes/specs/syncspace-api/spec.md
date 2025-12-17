# syncspace-api Spec Delta

## MODIFIED Requirements

### Requirement: Unified Protobuf Schema

The plugin SHALL define all operations in a single `syncspace.proto` file that serves as the single source of truth for the SyncSpace API.

#### Scenario: Schema includes lifecycle operations

- **GIVEN** the SyncSpace API protobuf schema
- **WHEN** the schema is reviewed
- **THEN** it defines Init and Shutdown RPC methods

#### Scenario: Schema includes space operations

- **GIVEN** the SyncSpace API protobuf schema
- **WHEN** the schema is reviewed
- **THEN** it defines CreateSpace, JoinSpace, LeaveSpace, ListSpaces, and DeleteSpace RPC methods

#### Scenario: Schema includes object lifecycle operations

- **GIVEN** the SyncSpace API protobuf schema
- **WHEN** the schema is reviewed
- **THEN** it defines CreateObject, DeleteObject, ListObjects, and QueryObjects RPC methods

#### Scenario: Schema includes change operations

- **GIVEN** the SyncSpace API protobuf schema
- **WHEN** the schema is reviewed
- **THEN** it defines AppendChange, GetChanges, and GetHeads RPC methods

#### Scenario: Schema includes sync control operations

- **GIVEN** the SyncSpace API protobuf schema
- **WHEN** the schema is reviewed
- **THEN** it defines StartSync, PauseSync, and GetSyncStatus RPC methods

#### Scenario: Schema includes event streaming

- **GIVEN** the SyncSpace API protobuf schema
- **WHEN** the schema is reviewed
- **THEN** it defines Subscribe RPC method with server streaming response

### Requirement: Event Streaming

The SyncSpace API SHALL provide server-streaming events for change operations and sync status updates.

#### Scenario: Subscribe streams change received events

- **GIVEN** a Subscribe call for a space
- **WHEN** a change is appended to an object (locally or via sync)
- **THEN** a ChangeReceivedEvent is streamed to the subscriber with objectId, changeId, isLocal flag, and dataType

#### Scenario: Subscribe streams heads changed events

- **GIVEN** a Subscribe call for a space
- **WHEN** an object's heads change (new change appended or sync received)
- **THEN** a HeadsChangedEvent is streamed to the subscriber with objectId, oldHeads, newHeads, and hasMultipleHeads flag

#### Scenario: Subscribe streams sync status events

- **GIVEN** a Subscribe call for a space
- **WHEN** sync state changes (started, paused, syncing, error)
- **THEN** a SyncStatusChangedEvent is streamed to the subscriber

#### Scenario: Events bridge from Any-Sync internal events

- **GIVEN** Any-Sync generates internal events (ObjectTree changes, HeadSync updates)
- **WHEN** these events occur
- **THEN** they are transformed and streamed to plugin event subscribers

## REMOVED Requirements

### Requirement: Opaque Document Data Model

**Reason:** Replaced by Object and Change data models. The opaqueness principle is retained but applied to Change payloads instead of Document payloads.

### Requirement: Document CRUD Operations

**Reason:** Document semantics (create/read/update/delete whole entities) conflict with ObjectTree's change-based model. Replaced by Object lifecycle and Change operations.

## ADDED Requirements

### Requirement: Object Lifecycle Operations

The SyncSpace API SHALL provide operations for creating, deleting, and listing objects within spaces. An object represents an ObjectTree instance that stores a DAG of changes.

#### Scenario: CreateObject establishes new object with initial change

- **GIVEN** a CreateObjectRequest with spaceId, objectType, and initialData
- **WHEN** CreateObject is called
- **THEN** a new ObjectTree is created with an initial change containing the data
- **AND** the objectId and rootChangeId are returned

#### Scenario: CreateObject generates objectId if not provided

- **GIVEN** a CreateObjectRequest without an objectId field
- **WHEN** CreateObject is called
- **THEN** a unique objectId is generated automatically

#### Scenario: CreateObject stores immutable objectType

- **GIVEN** a CreateObjectRequest with objectType
- **WHEN** CreateObject is called
- **THEN** the objectType is stored in the local index for filtering
- **AND** the objectType cannot be changed after creation

#### Scenario: DeleteObject removes object and all changes

- **GIVEN** a DeleteObjectRequest with spaceId and objectId
- **WHEN** DeleteObject is called
- **THEN** the object's ObjectTree and all changes are marked for deletion
- **AND** the response indicates whether the object existed

#### Scenario: ListObjects returns objects in space

- **GIVEN** a ListObjectsRequest with spaceId
- **WHEN** ListObjects is called
- **THEN** ObjectInfo entries are returned with objectId, objectType, metadata, timestamps, changeCount, and headIds

#### Scenario: ListObjects filters by objectType

- **GIVEN** a ListObjectsRequest with objectType filter
- **WHEN** ListObjects is called
- **THEN** only objects matching the objectType are returned

#### Scenario: ListObjects supports pagination

- **GIVEN** a ListObjectsRequest with limit and cursor
- **WHEN** ListObjects is called
- **THEN** at most limit objects are returned with a nextCursor for continuation

#### Scenario: QueryObjects filters objects by indexed metadata

- **GIVEN** a QueryObjectsRequest with spaceId and filters
- **WHEN** QueryObjects is called
- **THEN** only objects matching all filter criteria in the derived index are returned

#### Scenario: QueryObjects supports multiple filter operators

- **GIVEN** a QueryObjectsRequest with filters using different operators
- **WHEN** QueryObjects is called
- **THEN** filters are applied using the specified operators (eq, ne, contains, startsWith)

#### Scenario: QueryObjects combines objectType and metadata filters

- **GIVEN** a QueryObjectsRequest with both objectType and metadata filters
- **WHEN** QueryObjects is called
- **THEN** only objects matching the objectType AND all metadata filters are returned

#### Scenario: QueryObjects supports pagination

- **GIVEN** a QueryObjectsRequest with limit and cursor
- **WHEN** QueryObjects is called
- **THEN** at most limit objects are returned with a nextCursor for continuation

#### Scenario: QueryObjects reflects synced metadata

- **GIVEN** a metadata change received from sync with dataType "metadata"
- **WHEN** the local index is updated and QueryObjects is called
- **THEN** the query results reflect the synced metadata values

### Requirement: Change Operations

The SyncSpace API SHALL provide operations for appending changes to objects and retrieving change history. Changes are the fundamental unit of storage and synchronization.

#### Scenario: AppendChange adds change to object's DAG

- **GIVEN** an AppendChangeRequest with spaceId, objectId, and data
- **WHEN** AppendChange is called
- **THEN** a new change is added to the object's ObjectTree
- **AND** the changeId and current headIds are returned

#### Scenario: AppendChange accepts opaque bytes payload

- **GIVEN** an AppendChangeRequest with bytes data
- **WHEN** AppendChange is called
- **THEN** the data is stored as-is without interpretation
- **AND** applications can use any format (JSON, protobuf, CRDT updates, etc.)

#### Scenario: AppendChange accepts dataType hint

- **GIVEN** an AppendChangeRequest with dataType string
- **WHEN** AppendChange is called
- **THEN** the dataType is stored as metadata for application use

#### Scenario: AppendChange accepts isSnapshot hint

- **GIVEN** an AppendChangeRequest with isSnapshot=true
- **WHEN** AppendChange is called
- **THEN** the change is marked as a snapshot for optimization purposes

#### Scenario: AppendChange appends to current heads by default

- **GIVEN** an AppendChangeRequest without parentIds
- **WHEN** AppendChange is called
- **THEN** the change is appended with current heads as parents

#### Scenario: AppendChange accepts explicit parentIds for merges

- **GIVEN** an AppendChangeRequest with parentIds containing multiple changeIds
- **WHEN** AppendChange is called
- **THEN** the change is created with the specified parents (merge commit)

#### Scenario: GetChanges retrieves change history

- **GIVEN** a GetChangesRequest with spaceId and objectId
- **WHEN** GetChanges is called
- **THEN** Change entries are returned with changeId, data, timestamp, authorIdentity, parentIds, isSnapshot, dataType, and orderIndex

#### Scenario: GetChanges filters by sinceChangeId

- **GIVEN** a GetChangesRequest with sinceChangeId
- **WHEN** GetChanges is called
- **THEN** only changes after the specified change are returned

#### Scenario: GetChanges filters by fromHeads

- **GIVEN** a GetChangesRequest with fromHeads containing changeIds
- **WHEN** GetChanges is called
- **THEN** only changes reachable from the specified heads to current heads are returned

#### Scenario: GetChanges supports metadata-only mode

- **GIVEN** a GetChangesRequest with includeData=false
- **WHEN** GetChanges is called
- **THEN** Change entries are returned without the data field (for sync negotiation)

#### Scenario: GetChanges supports pagination

- **GIVEN** a GetChangesRequest with limit
- **WHEN** GetChanges is called
- **THEN** at most limit changes are returned with hasMore flag and nextCursor

#### Scenario: GetHeads returns current head changeIds

- **GIVEN** a GetHeadsRequest with spaceId and objectId
- **WHEN** GetHeads is called
- **THEN** the current headIds are returned

#### Scenario: GetHeads indicates multiple heads conflict

- **GIVEN** an object with concurrent changes from multiple devices
- **WHEN** GetHeads is called
- **THEN** hasMultipleHeads is true indicating unmerged concurrent edits

### Requirement: Opaque Change Data Model

Change operations SHALL use `bytes data` for opaque application payloads, making the plugin data-model agnostic while exposing the change DAG structure.

#### Scenario: AppendChange accepts opaque bytes

- **GIVEN** an AppendChangeRequest message definition
- **WHEN** the message structure is reviewed
- **THEN** it contains a `bytes data` field for the change payload

#### Scenario: Change includes dataType hint for applications

- **GIVEN** an AppendChangeRequest message definition
- **WHEN** the message structure is reviewed
- **THEN** it contains a `string dataType` field for application-defined type hints

#### Scenario: Application serializes its own change format

- **GIVEN** an application with domain-specific change formats
- **WHEN** the application appends a change
- **THEN** it serializes its change to bytes (delta, operation, snapshot, CRDT update, etc.)

#### Scenario: Change includes system metadata

- **GIVEN** a Change message in GetChangesResponse
- **WHEN** the message structure is reviewed
- **THEN** it contains timestamp, authorIdentity, parentIds, isSnapshot, and orderIndex fields managed by the system

### Requirement: Derived Metadata Index

The SyncSpace API SHALL maintain a local metadata index derived from ObjectTree changes, enabling efficient queries while ensuring metadata syncs across devices.

#### Scenario: Mutable metadata stored as changes

- **GIVEN** an application wants to update object metadata (title, tags, etc.)
- **WHEN** the application appends a change with dataType "metadata"
- **THEN** the metadata is stored in ObjectTree and syncs to other devices

#### Scenario: Local index rebuilt on startup

- **GIVEN** the plugin initializes for a space
- **WHEN** objects are loaded
- **THEN** the local metadata index is rebuilt by scanning changes with dataType "metadata"

#### Scenario: Local index updated on change received

- **GIVEN** a ChangeReceivedEvent arrives with dataType "metadata"
- **WHEN** the event is processed
- **THEN** the local metadata index is updated to reflect the new metadata values

#### Scenario: ListObjects uses derived index

- **GIVEN** ListObjects is called for a space
- **WHEN** the results are computed
- **THEN** metadata fields are read from the derived local index (not by scanning all changes)

#### Scenario: Immutable objectType in index only

- **GIVEN** an object with objectType set at creation
- **WHEN** the objectType is stored
- **THEN** it is stored in the local index only (not in ObjectTree changes) since it never changes

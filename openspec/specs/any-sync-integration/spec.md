# any-sync-integration Specification

## Purpose
Defines the integration layer between the SyncSpace API and Any-Sync's core components (SpaceService, ObjectTree, and synchronization mechanisms). Ensures proper usage of Any-Sync's higher-level abstractions for space management and document storage with built-in synchronization support.
## Requirements
### Requirement: SpaceService Integration

The Go backend SHALL use Any-Sync's SpaceService for all space management operations.

#### Scenario: SpaceCreate uses SpaceService

- **GIVEN** a SpaceCreate handler is invoked
- **WHEN** the handler processes the request
- **THEN** it calls SpaceService.CreateSpace() to create the space

#### Scenario: SpaceJoin uses SpaceService

- **GIVEN** a SpaceJoin handler is invoked
- **WHEN** the handler processes the request
- **THEN** it calls SpaceService.JoinSpace() with space ID and credentials

#### Scenario: SpaceLeave uses SpaceService

- **GIVEN** a SpaceLeave handler is invoked
- **WHEN** the handler processes the request
- **THEN** it calls SpaceService.LeaveSpace() to remove the device

### Requirement: ObjectTree Integration

The Go backend SHALL use Any-Sync's ObjectTree for all document storage operations.

#### Scenario: DocCreate creates ObjectTree

- **GIVEN** a DocCreate handler is invoked
- **WHEN** the handler processes the request
- **THEN** it creates a new ObjectTree with an initial change containing the document data

#### Scenario: DocUpdate adds change to ObjectTree

- **GIVEN** a DocUpdate handler is invoked with existing document ID
- **WHEN** the handler processes the request
- **THEN** it adds a new change to the document's ObjectTree

#### Scenario: DocGet reads from ObjectTree

- **GIVEN** a DocGet handler is invoked
- **WHEN** the handler retrieves the document
- **THEN** it reads the latest state from the document's ObjectTree

#### Scenario: DocDelete marks for deletion in ObjectTree

- **GIVEN** a DocDelete handler is invoked
- **WHEN** the handler processes the request
- **THEN** it marks the document for deletion using Any-Sync's deletion mechanism

### Requirement: HeadStorage Integration

The Go backend SHALL update HeadStorage when documents change.

#### Scenario: DocCreate updates head

- **GIVEN** a new document is created
- **WHEN** the ObjectTree is created
- **THEN** HeadStorage is updated with the new head

#### Scenario: DocUpdate updates head

- **GIVEN** a document is updated
- **WHEN** the change is added to ObjectTree
- **THEN** HeadStorage is updated with the new head

### Requirement: Sync Mechanism Integration

The Go backend SHALL integrate Any-Sync's HeadSync and ObjectSync for synchronization.

#### Scenario: SyncStart initiates HeadSync

- **GIVEN** SyncStart is called for a space
- **WHEN** the handler activates sync
- **THEN** HeadSync periodic discovery is started

#### Scenario: HeadSync discovers remote changes

- **GIVEN** HeadSync is running
- **WHEN** a peer has changes
- **THEN** HeadSync identifies divergent heads

#### Scenario: ObjectSync fetches missing changes

- **GIVEN** HeadSync identified divergent heads
- **WHEN** the sync mechanism processes the divergence
- **THEN** ObjectSync exchanges missing changes with the peer

#### Scenario: SyncPause stops sync mechanisms

- **GIVEN** SyncPause is called for a space
- **WHEN** the handler deactivates sync
- **THEN** HeadSync and ObjectSync are paused

### Requirement: Event Bridging

The Go backend SHALL bridge Any-Sync internal events to plugin events.

#### Scenario: ObjectTree change triggers event

- **GIVEN** an ObjectTree change occurs (local or remote)
- **WHEN** Any-Sync generates the internal event
- **THEN** it is transformed to a plugin Event message and streamed to subscribers

#### Scenario: Sync status change triggers event

- **GIVEN** sync state changes (started, syncing, paused, error)
- **WHEN** Any-Sync's sync mechanism updates state
- **THEN** it is transformed to a plugin Event message and streamed to subscribers

#### Scenario: Event handler callback is invoked

- **GIVEN** SetEventHandler registered a callback
- **WHEN** a plugin Event is generated
- **THEN** the callback is invoked with the serialized Event bytes

### Requirement: DiffContainer Integration

The Go backend SHALL use DiffContainer for change tracking within spaces.

#### Scenario: Document changes added to DiffContainer

- **GIVEN** a document is created or updated
- **WHEN** the ObjectTree change is committed
- **THEN** the change is added to the space's DiffContainer

#### Scenario: DiffContainer used for sync

- **GIVEN** HeadSync identified changes to exchange
- **WHEN** ObjectSync retrieves changes
- **THEN** DiffContainer provides the change history

### Requirement: Error Handling from Any-Sync

The Go backend SHALL properly handle and propagate errors from Any-Sync APIs.

#### Scenario: Any-Sync error is wrapped with context

- **GIVEN** an Any-Sync API call returns an error
- **WHEN** the handler processes the error
- **THEN** it is wrapped with operation context (space ID, document ID, etc.)

#### Scenario: Any-Sync error is converted to appropriate status

- **GIVEN** an Any-Sync error occurs
- **WHEN** the error is returned from a handler
- **THEN** it is converted to an appropriate error type for the client

### Requirement: No Direct Any-Store Usage

The Go backend SHALL NOT use Any-Store APIs directly, only through Any-Sync abstractions.

#### Scenario: Document storage uses ObjectTree not Any-Store

- **GIVEN** a document storage operation
- **WHEN** the implementation is reviewed
- **THEN** it uses Any-Sync's ObjectTree, not direct Any-Store calls

#### Scenario: No Any-Store imports in handlers

- **GIVEN** handler function implementations
- **WHEN** import statements are reviewed
- **THEN** no direct Any-Store package imports are present


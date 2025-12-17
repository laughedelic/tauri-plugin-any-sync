# any-sync-integration Spec Delta

## MODIFIED Requirements

### Requirement: ObjectTree Integration

The Go backend SHALL use Any-Sync's ObjectTree for all object storage and change operations, exposing the DAG structure to applications.

#### Scenario: CreateObject creates ObjectTree with initial change

- **GIVEN** a CreateObject handler is invoked
- **WHEN** the handler processes the request
- **THEN** it creates a new ObjectTree with an initial change containing the provided data
- **AND** the ObjectTree ID becomes the objectId
- **AND** the initial change's CID becomes the rootChangeId

#### Scenario: AppendChange adds change to ObjectTree

- **GIVEN** an AppendChange handler is invoked with existing objectId
- **WHEN** the handler processes the request
- **THEN** it calls AddContent on the object's ObjectTree
- **AND** the change CID is returned as changeId

#### Scenario: AppendChange with explicit parents creates merge

- **GIVEN** an AppendChange handler is invoked with parentIds
- **WHEN** the handler processes the request
- **THEN** it creates a change with the specified parent CIDs
- **AND** this enables merge commits that resolve multiple heads

#### Scenario: GetChanges reads from ObjectTree DAG

- **GIVEN** a GetChanges handler is invoked
- **WHEN** the handler retrieves changes
- **THEN** it traverses the ObjectTree DAG from specified starting point
- **AND** returns changes in deterministic order

#### Scenario: GetHeads reads current ObjectTree heads

- **GIVEN** a GetHeads handler is invoked
- **WHEN** the handler retrieves heads
- **THEN** it returns the current head CIDs from the ObjectTree
- **AND** multiple heads indicate unmerged concurrent changes

#### Scenario: DeleteObject marks ObjectTree for deletion

- **GIVEN** a DeleteObject handler is invoked
- **WHEN** the handler processes the request
- **THEN** it marks the ObjectTree for deletion using Any-Sync's deletion mechanism

### Requirement: HeadStorage Integration

The Go backend SHALL update HeadStorage when changes are appended to objects.

#### Scenario: CreateObject updates head

- **GIVEN** a new object is created
- **WHEN** the ObjectTree is created with initial change
- **THEN** HeadStorage is updated with the initial head

#### Scenario: AppendChange updates head

- **GIVEN** a change is appended to an object
- **WHEN** the change is added to ObjectTree
- **THEN** HeadStorage is updated with the new head(s)

#### Scenario: Multiple heads tracked in HeadStorage

- **GIVEN** concurrent changes arrive from sync
- **WHEN** the changes are added to ObjectTree
- **THEN** HeadStorage tracks all current heads (may be multiple)

### Requirement: Event Bridging

The Go backend SHALL bridge Any-Sync internal events to plugin events with change-level granularity.

#### Scenario: ObjectTree change triggers ChangeReceivedEvent

- **GIVEN** a change is added to an ObjectTree (local or via sync)
- **WHEN** Any-Sync generates the internal event
- **THEN** it is transformed to a ChangeReceivedEvent with objectId, changeId, isLocal, and dataType

#### Scenario: Head transition triggers HeadsChangedEvent

- **GIVEN** an object's heads change
- **WHEN** the transition occurs
- **THEN** it is transformed to a HeadsChangedEvent with objectId, oldHeads, newHeads, and hasMultipleHeads

#### Scenario: Sync status change triggers event

- **GIVEN** sync state changes (started, syncing, paused, error)
- **WHEN** Any-Sync's sync mechanism updates state
- **THEN** it is transformed to a SyncStatusChangedEvent and streamed to subscribers

#### Scenario: Event handler callback is invoked

- **GIVEN** SetEventHandler registered a callback
- **WHEN** a plugin Event is generated
- **THEN** the callback is invoked with the serialized Event bytes

### Requirement: DiffContainer Integration

The Go backend SHALL use DiffContainer for change tracking within spaces, enabling incremental sync.

#### Scenario: Changes added to DiffContainer

- **GIVEN** a change is appended to an object
- **WHEN** the ObjectTree change is committed
- **THEN** the change is added to the space's DiffContainer

#### Scenario: DiffContainer used for sync

- **GIVEN** HeadSync identified changes to exchange
- **WHEN** ObjectSync retrieves changes
- **THEN** DiffContainer provides the change history for the object

#### Scenario: DiffContainer tracks change parentage

- **GIVEN** changes with explicit parentIds are created
- **WHEN** the changes are added to DiffContainer
- **THEN** the DAG structure (parent relationships) is preserved

## REMOVED Requirements

### Requirement: SpaceService Integration (DocCreate scenario)

**Reason:** The specific DocCreate, DocUpdate, DocGet, DocDelete scenarios are replaced by object and change operations. SpaceService is still used for space management.

**Migration:** See MODIFIED ObjectTree Integration requirement for new scenarios.

**Note:** The SpaceService requirement itself is NOT removed - only the document-specific scenarios within it should be considered replaced by the ObjectTree Integration updates.

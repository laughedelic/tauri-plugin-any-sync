# Tasks: Replace Document API with Change-Based API

## 1. Protobuf Schema Update

- [ ] 1.1 Remove document messages from `syncspace.proto`:
  - `CreateDocumentRequest/Response`
  - `GetDocumentRequest/Response`
  - `UpdateDocumentRequest/Response`
  - `DeleteDocumentRequest/Response`
  - `ListDocumentsRequest/Response`
  - `QueryDocumentsRequest/Response`
  - `Document`, `DocumentInfo`
  - `DocumentCreatedEvent`, `DocumentUpdatedEvent`, `DocumentDeletedEvent`

- [ ] 1.2 Add object lifecycle messages:
  - `CreateObjectRequest` (spaceId, objectId?, objectType, initialData) - no mutable metadata
  - `CreateObjectResponse` (objectId, rootChangeId)
  - `DeleteObjectRequest` (spaceId, objectId)
  - `DeleteObjectResponse` (existed)
  - `ListObjectsRequest` (spaceId, objectType?, limit?, cursor?)
  - `ListObjectsResponse` (objects[], nextCursor)
  - `QueryObjectsRequest` (spaceId, objectType?, filters[], limit?, cursor?)
  - `QueryObjectsResponse` (objects[], nextCursor)
  - `QueryFilter` (field, operator, value) - reuse existing message
  - `ObjectInfo` (objectId, objectType, indexedMetadata, createdAt, updatedAt, changeCount, headIds[])

- [ ] 1.3 Add change operation messages:
  - `AppendChangeRequest` (spaceId, objectId, data, dataType?, isSnapshot?, parentIds[]?)
  - `AppendChangeResponse` (changeId, headIds[])
  - `GetChangesRequest` (spaceId, objectId, sinceChangeId?, fromHeads[]?, limit?, includeData?)
  - `GetChangesResponse` (changes[], headIds[], hasMore, nextCursor)
  - `Change` (changeId, data, timestamp, authorIdentity, parentIds[], isSnapshot, dataType, orderIndex)
  - `GetHeadsRequest` (spaceId, objectId)
  - `GetHeadsResponse` (headIds[], hasMultipleHeads)

- [ ] 1.4 Add change-aware event messages:
  - `ChangeReceivedEvent` (objectId, changeId, isLocal, dataType)
  - `HeadsChangedEvent` (objectId, oldHeads[], newHeads[], hasMultipleHeads)

- [ ] 1.5 Update `SyncSpaceService` RPC definitions:
  - Remove: CreateDocument, GetDocument, UpdateDocument, DeleteDocument, ListDocuments, QueryDocuments
  - Add: CreateObject, DeleteObject, ListObjects, QueryObjects, AppendChange, GetChanges, GetHeads

- [ ] 1.6 Run `buf generate` and verify TypeScript client regenerates correctly

## 2. Go Backend - ObjectManager Refactor

- [ ] 2.1 Create `plugin-go-backend/shared/anysync/objects.go`:
  - Refactor from `documents.go`
  - `ObjectManager` struct with space/object tracking
  - `CreateObject()` - creates ObjectTree with initial change
  - `DeleteObject()` - deletes ObjectTree
  - `ListObjects()` - returns object metadata
  - `QueryObjects()` - filters objects by metadata
  - `GetObjectInfo()` - returns single object info

- [ ] 2.2 Create `plugin-go-backend/shared/anysync/changes.go`:
  - `AppendChange()` - adds change to ObjectTree, supports explicit parents
  - `GetChanges()` - retrieves changes from ObjectTree DAG
  - `GetHeads()` - returns current head CIDs

- [ ] 2.3 Implement derived metadata index:
  - Rename metadata files from `documents/{spaceId}.json` to `objects/{spaceId}.json`
  - Update `ObjectIndex` struct to store:
    - `objectType` (immutable, set at creation)
    - `indexedMetadata` (derived from changes with dataType="metadata")
    - `headIds` (for quick conflict detection)
    - `createdAt`, `updatedAt`, `changeCount`
  - Implement `RebuildIndex()` - scan all objects' metadata changes on startup
  - Implement `UpdateIndexFromChange()` - update index when metadata change received
  - Index is derived from ObjectTree changes, not a separate source of truth

- [ ] 2.4 Delete `plugin-go-backend/shared/anysync/documents.go`

## 3. Go Backend - Handlers

- [ ] 3.1 Create `plugin-go-backend/shared/handlers/objects.go`:
  - `CreateObject()` handler
  - `DeleteObject()` handler
  - `ListObjects()` handler
  - `QueryObjects()` handler

- [ ] 3.2 Create `plugin-go-backend/shared/handlers/changes.go`:
  - `AppendChange()` handler
  - `GetChanges()` handler
  - `GetHeads()` handler

- [ ] 3.3 Update `plugin-go-backend/shared/handlers/registry.go`:
  - Remove: CreateDocument, GetDocument, UpdateDocument, DeleteDocument, ListDocuments, QueryDocuments
  - Add: CreateObject, DeleteObject, ListObjects, QueryObjects, AppendChange, GetChanges, GetHeads

- [ ] 3.4 Delete `plugin-go-backend/shared/handlers/documents.go`

## 4. Go Backend - Events

- [ ] 4.1 Update `plugin-go-backend/shared/anysync/events.go`:
  - Remove: `EventDocumentCreated`, `EventDocumentUpdated`, `EventDocumentDeleted`
  - Add: `EventChangeReceived`, `EventHeadsChanged`
  - Update event payload structures

- [ ] 4.2 Update event emission in ObjectManager/ChangeManager:
  - `CreateObject` emits `EventChangeReceived` (for initial change)
  - `AppendChange` emits `EventChangeReceived`
  - Head transitions emit `EventHeadsChanged`

## 5. Go Backend - Tests

- [ ] 5.1 Create `plugin-go-backend/shared/anysync/objects_test.go`:
  - Test CreateObject creates ObjectTree with initial change
  - Test DeleteObject removes object
  - Test ListObjects returns correct objects
  - Test QueryObjects filters by indexed metadata (eq, ne, contains, startsWith)
  - Test index rebuilt from metadata changes on startup
  - Test index updated when metadata change appended
  - Test index updated when metadata change received from sync event

- [ ] 5.2 Create `plugin-go-backend/shared/anysync/changes_test.go`:
  - Test AppendChange adds to ObjectTree
  - Test AppendChange with explicit parents creates merge
  - Test GetChanges retrieves history
  - Test GetChanges with sinceChangeId filters
  - Test GetChanges with fromHeads returns path
  - Test GetHeads returns current heads
  - Test multiple heads detection

- [ ] 5.3 Create `plugin-go-backend/shared/handlers/objects_test.go`:
  - Handler-level tests for object operations (CreateObject, DeleteObject, ListObjects, QueryObjects)

- [ ] 5.4 Create `plugin-go-backend/shared/handlers/changes_test.go`:
  - Handler-level tests for change operations

- [ ] 5.5 Delete old test files:
  - `plugin-go-backend/shared/anysync/documents_test.go`
  - `plugin-go-backend/shared/handlers/documents_test.go`

- [ ] 5.6 Update `plugin-go-backend/shared/handlers/integration_test.go`:
  - Replace document operations with object/change operations

## 6. Example App Update

- [ ] 6.1 Create `example-app/src/services/notes-v2.ts`:
  - Implement notes using change-based API
  - Define `NoteChange` types (init, setText, etc.)
  - Implement state reconstruction from changes
  - Handle multiple heads (show conflict indicator)

- [ ] 6.2 Update `example-app/src/store/useNotesStore.ts`:
  - Replace document operations with change operations
  - Add change history tracking
  - Add conflict state handling

- [ ] 6.3 Update `example-app/src/components/`:
  - Add conflict indicator UI when hasMultipleHeads
  - Optional: Add change history view

- [ ] 6.4 Delete or archive old notes service:
  - `example-app/src/services/notes.ts` (can keep as reference)

## 7. Documentation

- [ ] 7.1 Update `plugin-js-api/README.md`:
  - Document new object/change API
  - Add usage examples
  - Document common patterns (snapshots, deltas, CRDTs)

- [ ] 7.2 Update `example-app/README.md`:
  - Explain change-based architecture
  - Document state reconstruction pattern

## 8. Validation

- [ ] 8.1 Run all Go tests: `task go:test`
- [ ] 8.2 Build all platforms: `task build-all`
- [ ] 8.3 Run example app and verify:
  - Notes can be created
  - Notes can be edited
  - Changes are persisted
  - Multiple heads detected (if simulated)
- [ ] 8.4 Run TypeScript type checking: `task js:build`

## Dependencies

- Tasks 1.x must complete before 2.x-4.x (proto defines types)
- Tasks 2.x and 3.x can be parallelized
- Task 4.x depends on 2.x and 3.x
- Tasks 5.x depend on 2.x and 3.x
- Task 6.x depends on 1.x (needs generated client)
- Task 7.x can be done in parallel with implementation
- Task 8.x must be done last

## Estimated Scope

| Section | New Lines | Deleted Lines |
|---------|-----------|---------------|
| Proto   | ~120      | ~100          |
| Go ObjectManager | ~300 | - |
| Go ChangeManager | ~200 | - |
| Go Handlers | ~200    | ~250          |
| Go Tests | ~400       | ~350          |
| Go Events | ~50       | ~30           |
| Example App | ~200   | ~150          |
| **Total** | ~1470    | ~880          |

Net change: ~590 new lines (refactoring, not pure addition)

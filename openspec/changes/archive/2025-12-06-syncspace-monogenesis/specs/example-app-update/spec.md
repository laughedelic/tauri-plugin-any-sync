# Example App Update Specification

## REMOVED Requirements

### Requirement: Storage Demo Component

~~The example app SHALL provide an interactive UI component for demonstrating storage operations.~~

**Reason:** Replaced by domain service layer pattern and SyncSpace API demonstration.

### Requirement: Storage API Demonstration

~~The example app SHALL demonstrate practical usage of the storage API with example data.~~

**Reason:** Replaced by domain service pattern and generic document API demonstration.

## MODIFIED Requirements

### Requirement: Plugin Integration

The example app SHALL successfully import and initialize the any-sync plugin.

**Changes:**
- Now initializes SyncSpace API instead of direct storage API
- Demonstrates space creation and join workflow

#### Scenario: Plugin initializes SyncSpace

- **GIVEN** the example app
- **WHEN** it starts
- **THEN** it successfully imports and initializes the any-sync plugin with SyncSpace API

### Requirement: Example App Frontend

The existing Svelte frontend SHALL include components to demonstrate plugin functionality.

**Changes:**
- Shows domain service pattern with NotesService
- Demonstrates space management, document CRUD, sync control, and events

#### Scenario: Frontend uses domain service layer

- **GIVEN** the example app UI
- **WHEN** components are loaded
- **THEN** they use NotesService instead of calling storage API directly

## ADDED Requirements

### Requirement: Domain Service Layer

The example app SHALL implement a domain service layer on top of the generic SyncSpace API.

#### Scenario: NotesService encapsulates note operations

- **GIVEN** the example app implements a NotesService
- **WHEN** the service code is reviewed
- **THEN** it provides note-specific methods (createNote, getNote, updateNote, deleteNote, listNotes)

#### Scenario: NotesService uses SyncSpace document API

- **GIVEN** the NotesService implementation
- **WHEN** any note operation is performed
- **THEN** it uses the plugin's generic DocCreate, DocGet, DocUpdate, DocDelete, or DocList operations

#### Scenario: NotesService serializes domain model

- **GIVEN** the NotesService creates a note
- **WHEN** the note data is prepared
- **THEN** it serializes the Note interface/type to bytes before calling DocCreate

#### Scenario: NotesService deserializes domain model

- **GIVEN** the NotesService retrieves a note
- **WHEN** DocGet returns bytes
- **THEN** it deserializes the bytes back to the Note interface/type

### Requirement: Opaque Data Demonstration

The example app SHALL demonstrate how applications handle their own data serialization.

#### Scenario: Example uses JSON serialization

- **GIVEN** the NotesService implementation
- **WHEN** serialization code is reviewed
- **THEN** it uses JSON.stringify for serialization and JSON.parse for deserialization

#### Scenario: Example shows metadata usage

- **GIVEN** the NotesService creates a note
- **WHEN** DocCreate is called
- **THEN** it includes note title in metadata for indexing/search capabilities

### Requirement: Sync Control Demonstration

The example app SHALL demonstrate sync control operations.

#### Scenario: Example provides sync start/pause controls

- **GIVEN** the example app UI
- **WHEN** the user interacts with sync controls
- **THEN** the app calls SyncStart or SyncPause operations

#### Scenario: Example displays sync status

- **GIVEN** the example app UI
- **WHEN** sync is active or paused
- **THEN** the app displays current sync status by calling SyncStatus

### Requirement: Event Subscription Demonstration

The example app SHALL demonstrate event subscription and handling.

#### Scenario: Example subscribes to document events

- **GIVEN** the example app initialization
- **WHEN** the app sets up event handling
- **THEN** it subscribes to document change events via the Subscribe operation

#### Scenario: Example displays real-time updates

- **GIVEN** the example app is subscribed to events
- **WHEN** a document change event is received (local or remote)
- **THEN** the UI updates to reflect the change

#### Scenario: Example subscribes to sync events

- **GIVEN** the example app initialization
- **WHEN** the app sets up event handling
- **THEN** it subscribes to sync status change events

### Requirement: Pattern Documentation

The example app SHALL serve as documentation for the recommended usage pattern.

#### Scenario: Example code is well-commented

- **GIVEN** the NotesService and UI code
- **WHEN** the code is reviewed
- **THEN** it contains comments explaining the domain service pattern

#### Scenario: Example shows complete flow

- **GIVEN** the example app
- **WHEN** a complete user workflow is executed (create space, create note, update note, sync)
- **THEN** all steps work correctly demonstrating the full SyncSpace API

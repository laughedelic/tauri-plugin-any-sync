# Example App Update Specification

## REMOVED Requirements

### Requirement: Direct Storage API Usage

~~The example app SHALL use the storage API functions directly from the plugin.~~

**Reason:** Replaced by domain service layer pattern demonstrating proper usage of generic document API.

## MODIFIED Requirements

### Requirement: API Usage Demonstration

The example app SHALL demonstrate how to use the plugin's API.

**Changes:**
- Demonstrates domain service pattern instead of direct API calls
- Shows how to serialize/deserialize application data models
- Demonstrates space management in addition to document operations

#### Scenario: Example demonstrates space management

- **GIVEN** the example app UI
- **WHEN** the app is running
- **THEN** it provides UI for creating, listing, and joining spaces

#### Scenario: Example demonstrates document operations

- **GIVEN** the example app UI
- **WHEN** the app is running
- **THEN** it provides UI for creating, reading, updating, deleting, and listing documents

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

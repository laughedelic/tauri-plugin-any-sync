# Example App Update Specification

## Purpose
Provides a working example application demonstrating plugin integration, communication with the Go backend, and proper Tauri sidecar configuration for desktop platforms.
## Requirements
### Requirement: Plugin Integration

The example app SHALL successfully import and initialize the any-sync plugin.

**Changes:**
- Now initializes SyncSpace API instead of direct storage API
- Demonstrates space creation and join workflow

#### Scenario: Plugin initializes SyncSpace

- **GIVEN** the example app
- **WHEN** it starts
- **THEN** it successfully imports and initializes the any-sync plugin with SyncSpace API

### Requirement: UI for Ping Test
The example app SHALL provide a button or interface to trigger the ping command.
#### Scenario:
Given users need to test the plugin communication
When the example app loads
Then it should provide a button or interface to trigger the ping command

### Requirement: Response Display
The example app SHALL display the ping response from the Go backend in the UI.
#### Scenario:
Given the ping command returns a response from the Go backend
When the response is received
Then the example app should display the response in the UI

### Requirement: Error Handling Display
The example app SHALL display error messages appropriately when plugin communication fails.
#### Scenario:
Given the plugin communication might fail
When an error occurs
Then the example app should display the error message appropriately

### Requirement: Plugin Configuration
The example app SHALL include the any-sync plugin in the tauri.conf.json capabilities and externalBin configuration for desktop platforms.
#### Scenario:
Given the example app needs to use the plugin
When configuring the Tauri app
Then it should include the any-sync plugin in the tauri.conf.json capabilities

### Requirement: Build Integration
The example app SHALL successfully compile with the plugin dependency.
#### Scenario:
Given the example app needs to be buildable
When running the build process
Then it should successfully compile with the plugin dependency

### Requirement: Example App Frontend

The existing Svelte frontend SHALL include components to demonstrate plugin functionality.

**Changes:**
- Shows domain service pattern with NotesService
- Demonstrates space management, document CRUD, sync control, and events

#### Scenario: Frontend uses domain service layer

- **GIVEN** the example app UI
- **WHEN** components are loaded
- **THEN** they use NotesService instead of calling storage API directly

### Requirement: Tauri Configuration
The existing tauri.conf.json SHALL properly configure the any-sync plugin permissions and capabilities.
#### Scenario:
Given the existing tauri.conf.json
When adding the plugin
Then it should properly configure the any-sync plugin permissions and capabilities

### Requirement: Example App Documentation
The example app SHALL include component-specific AGENTS.md documentation for testing and development.
#### Scenario:
Given developers need to test and work with the example app
When they open the examples directory
Then they should find clear instructions for running, testing, and debugging plugin integration

### Requirement: Proper Sidecar Integration
The example app SHALL demonstrate Tauri's standard sidecar pattern using shell plugin for desktop platforms.
#### Scenario:
Given: desktop platforms require externalBin configuration
When: example app configures plugin
Then: it should use Tauri shell plugin sidecar APIs instead of manual process management

### Requirement: Error Display

The example app SHALL display storage errors in a user-friendly format for debugging.

#### Scenario: Success feedback

- **GIVEN** a successful storage operation
- **WHEN** the operation completes
- **THEN** a success message is displayed with operation details

#### Scenario: Error feedback

- **GIVEN** a failed storage operation (e.g., invalid JSON)
- **WHEN** the operation completes
- **THEN** the error message is displayed in a visible error panel

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


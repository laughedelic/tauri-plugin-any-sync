# Example App Updates

## ADDED Requirements

### Requirement: Storage Demo Component

The example app SHALL provide an interactive UI component for demonstrating storage operations.

#### Scenario: Put document UI

- **GIVEN** the storage demo component
- **WHEN** the user enters collection, ID, and JSON
- **THEN** a "Store Document" button calls storage.put() and displays the result

#### Scenario: Get document UI

- **GIVEN** the storage demo component
- **WHEN** the user enters collection and ID
- **THEN** a "Get Document" button calls storage.get() and displays the JSON

#### Scenario: List collection UI

- **GIVEN** the storage demo component
- **WHEN** the user enters a collection name
- **THEN** a "List Documents" button calls storage.list() and displays all IDs

### Requirement: Storage API Demonstration

The example app SHALL demonstrate practical usage of the storage API with example data.

#### Scenario: Pre-filled examples

- **GIVEN** the storage demo component
- **WHEN** the component is loaded
- **THEN** example values are pre-filled (e.g., collection="todos", id="1", json="{\"title\":\"Test\"}")

#### Scenario: Multiple collections demonstrated

- **GIVEN** the storage demo UI
- **WHEN** the user interacts with examples
- **THEN** multiple collection names are suggested (e.g., "todos", "notes", "settings")

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

### Requirement: Storage State Visualization

The example app SHALL visualize the current state of stored documents.

#### Scenario: Recently stored documents

- **GIVEN** the user has stored documents
- **WHEN** the storage demo is displayed
- **THEN** a list of recent operations is shown with timestamps

#### Scenario: Retrieved document display

- **GIVEN** a Get operation returns a document
- **WHEN** the result is displayed
- **THEN** the JSON is formatted and syntax-highlighted for readability

## MODIFIED Requirements

### Requirement: Example App Integration

The example app SHALL demonstrate all available plugin features through interactive UI components.

#### Scenario: Plugin features are demonstrated

- **GIVEN** the example app
- **WHEN** the app is loaded
- **THEN** both ping functionality and storage operations are accessible from the main UI

#### Scenario: Feature sections are organized

- **GIVEN** the main App component
- **WHEN** the UI is rendered
- **THEN** distinct sections are shown for health check (ping) and storage demos

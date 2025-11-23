# Desktop Integration Updates

## ADDED Requirements

### Requirement: Storage Command Handlers

The Rust plugin SHALL provide Tauri commands for storage operations that communicate with the Go sidecar via gRPC.

#### Scenario: Put command stores document

- **GIVEN** a TypeScript call to storage.put(collection, id, json)
- **WHEN** the Rust command handler is invoked
- **THEN** a gRPC Put request is sent to the sidecar and the result is returned

#### Scenario: Get command retrieves document

- **GIVEN** a TypeScript call to storage.get(collection, id)
- **WHEN** the Rust command handler is invoked
- **THEN** a gRPC Get request is sent to the sidecar and the document JSON is returned

#### Scenario: List command retrieves IDs

- **GIVEN** a TypeScript call to storage.list(collection)
- **WHEN** the Rust command handler is invoked
- **THEN** a gRPC List request is sent to the sidecar and the ID list is returned

### Requirement: Storage Error Handling

The Rust plugin SHALL convert gRPC storage errors to Tauri-compatible error types with meaningful messages.

#### Scenario: NOT_FOUND error is converted

- **GIVEN** a Get request for a non-existent document
- **WHEN** the gRPC call returns NOT_FOUND
- **THEN** the Rust error includes "Document not found" with collection and ID

#### Scenario: INVALID_ARGUMENT error is converted

- **GIVEN** a Put request with invalid JSON
- **WHEN** the gRPC call returns INVALID_ARGUMENT
- **THEN** the Rust error includes "Invalid JSON" with parsing details

#### Scenario: Connection errors are handled

- **GIVEN** the sidecar is not running
- **WHEN** a storage command is invoked
- **THEN** a connection error is returned with troubleshooting guidance

### Requirement: Desktop-Only Storage Commands

The storage commands SHALL be conditionally compiled for desktop platforms only using `#[cfg(desktop)]`.

#### Scenario: Storage commands available on desktop

- **GIVEN** a desktop build (macOS, Linux, Windows)
- **WHEN** the plugin is compiled
- **THEN** storage command handlers are included

#### Scenario: Storage commands excluded on mobile

- **GIVEN** a mobile build (iOS, Android)
- **WHEN** the plugin is compiled
- **THEN** storage command handlers are excluded

### Requirement: gRPC Client Integration

The desktop module SHALL create gRPC client connections to the storage service running in the sidecar.

#### Scenario: Storage client is initialized

- **GIVEN** the desktop module initialization
- **WHEN** the sidecar is started
- **THEN** a gRPC client is created for the StorageService

#### Scenario: Storage client reuses connection

- **GIVEN** multiple storage operations
- **WHEN** commands are invoked
- **THEN** the same gRPC connection is reused for all operations

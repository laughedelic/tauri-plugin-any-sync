# Desktop Integration Specification

## REMOVED Requirements

### Requirement: Storage Command Handlers

~~The Rust plugin SHALL provide Tauri commands for storage operations that communicate with the Go sidecar via gRPC.~~

**Reason:** Replaced by single command handler in single-dispatch pattern.

### Requirement: Storage Error Handling

~~The Rust plugin SHALL convert gRPC storage errors to Tauri-compatible error types with meaningful messages.~~

**Reason:** Error handling is now unified in the single command dispatcher.

### Requirement: gRPC Client Integration

~~The desktop module SHALL create gRPC client connections to the storage service running in the sidecar.~~

**Reason:** IPC abstraction now unified through single-dispatch pattern.

## MODIFIED Requirements

### Requirement: Sidecar Process Spawning

The plugin SHALL spawn the Go backend as a separate process with proper lifecycle management.

**Changes:**
- Sidecar now dispatches all operations through single Command interface
- No changes to process management itself

#### Scenario: Single command dispatcher at sidecar

- **GIVEN** the Go sidecar is spawned
- **WHEN** it initializes
- **THEN** it implements a single Command(cmd string, data []byte) interface instead of per-operation functions

### Requirement: gRPC Client Connection

The plugin SHALL establish a gRPC client connection to communicate with the Go backend.

**Changes:**
- Connection is now used for single Command method instead of per-operation methods
- All operations flow through Command(cmd, data) interface

#### Scenario: Single command gRPC method

- **GIVEN** the Rust plugin establishes gRPC connection
- **WHEN** operations are invoked
- **THEN** all commands use a single Command(cmd, data) gRPC method

## ADDED Requirements

### Requirement: Backend Trait Simplification

The Rust plugin SHALL define a simple backend trait with minimal methods.

#### Scenario: Backend trait has three methods

- **GIVEN** the AnySyncBackend trait definition
- **WHEN** the trait is reviewed
- **THEN** it defines exactly three methods: command, set_event_handler, and shutdown

#### Scenario: Command method signature

- **GIVEN** the AnySyncBackend trait
- **WHEN** the command method is reviewed
- **THEN** it has signature `fn command(&self, cmd: &str, data: &[u8]) -> Result<Vec<u8>, Error>`

#### Scenario: Desktop implementation of backend trait

- **GIVEN** the desktop backend service
- **WHEN** it implements AnySyncBackend
- **THEN** the command method forwards to the sidecar

### Requirement: Simplified Permission System

The Rust plugin SHALL use a single permission for the command handler.

#### Scenario: Single default permission exists

- **GIVEN** the permissions directory
- **WHEN** permission files are reviewed
- **THEN** only one permission file exists: default.toml

#### Scenario: Default permission allows command

- **GIVEN** the default.toml permission file
- **WHEN** the permission is reviewed
- **THEN** it grants access to the "command" Tauri command

#### Scenario: No per-operation permissions

- **GIVEN** the permissions directory
- **WHEN** permission files are reviewed
- **THEN** there are no per-operation permission files (e.g., no storage_put.toml, storage_get.toml)

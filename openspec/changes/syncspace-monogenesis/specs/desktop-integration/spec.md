# Desktop Integration Specification

## REMOVED Requirements

### Requirement: Per-Operation Command Handlers

~~The Rust plugin SHALL define separate Tauri commands for each storage operation.~~

**Reason:** Replaced by single command handler in single-dispatch pattern.

### Requirement: Per-Operation gRPC Client Calls

~~The desktop backend implementation SHALL make specific gRPC calls for each operation.~~

**Reason:** Replaced by single Command call that dispatcher routes internally.

## MODIFIED Requirements

### Requirement: Tauri Command Registration

The Rust plugin SHALL register Tauri commands that are callable from TypeScript.

**Changes:**
- Instead of multiple commands (ping, storage_put, storage_get, etc.), only one command is registered: `command`
- Single command accepts command name and data bytes, returns response bytes

#### Scenario: Single command handler is registered

- **GIVEN** the Rust plugin is initialized
- **WHEN** Tauri commands are registered
- **THEN** exactly one command named "command" is registered

#### Scenario: Command handler accepts command name and data

- **GIVEN** the TypeScript API invokes a command
- **WHEN** the command reaches the Rust handler
- **THEN** it receives cmd: String and data: Vec<u8> parameters

#### Scenario: Command handler returns response bytes

- **GIVEN** the Rust command handler processes a command
- **WHEN** the backend returns a response
- **THEN** the handler returns Vec<u8> to TypeScript

### Requirement: Desktop Backend Service

The Rust plugin SHALL use a desktop service implementation for sidecar communication.

**Changes:**
- Desktop service exposes single Command method instead of per-operation methods
- Service may use gRPC or simplified IPC (decision deferred)

#### Scenario: Desktop service calls sidecar Command

- **GIVEN** the Rust command handler invokes desktop service
- **WHEN** the service processes the call
- **THEN** it forwards cmd and data to the sidecar's Command interface

#### Scenario: Desktop service maintains single connection

- **GIVEN** the desktop service is initialized
- **WHEN** multiple commands are invoked
- **THEN** they all use the same IPC connection to the sidecar

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

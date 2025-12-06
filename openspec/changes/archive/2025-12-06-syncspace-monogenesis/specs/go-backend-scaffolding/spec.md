# Go Backend Scaffolding Specification

## REMOVED Requirements

### Requirement: AnyStore Integration

~~The Go backend SHALL integrate AnyStore for document storage capabilities.~~

**Reason:** Replaced by Any-Sync integration layer that uses SpaceService and ObjectTree.

### Requirement: Storage Module Organization

~~The Go backend SHALL organize storage code in `internal/storage/` to isolate AnyStore integration.~~

**Reason:** Replaced by Any-Sync integration layer.

### Requirement: gRPC Server Registration

~~The Go backend SHALL register the StorageService with the gRPC server during initialization.~~

**Reason:** Replaced by unified single-dispatch SyncSpace API.

## MODIFIED Requirements

### Requirement: Basic Go Backend Structure

The project SHALL provide a Go backend with proper package structure separating API and internal code.

**Changes:**
- Adds `shared/dispatcher/` for command routing
- Adds `shared/handlers/` for operation handlers
- Adds `shared/anysync/` for Any-Sync integration
- Removes `internal/storage/` (replaced by Any-Sync integration)

#### Scenario: Go backend package structure

- **GIVEN** the Go backend project
- **WHEN** the directory structure is reviewed
- **THEN** it includes dispatcher, handlers, and anysync packages for unified API

### Requirement: gRPC Ping Service

The Go backend SHALL provide a gRPC ping service for testing communication between frontend and backend.

**Changes:**
- Ping remains as a single operation in the dispatcher pattern
- No structural changes to ping implementation

#### Scenario: Ping service via dispatcher

- **GIVEN** a ping command via the SyncSpace API
- **WHEN** the command is dispatched
- **THEN** a pong response is returned

## ADDED Requirements

### Requirement: Dispatcher Package

The Go backend SHALL provide a dispatcher package for command routing.

#### Scenario: Dispatcher routes commands to handlers

- **GIVEN** a command name and serialized request data
- **WHEN** Dispatcher.Dispatch is called
- **THEN** the appropriate handler function is invoked

#### Scenario: Dispatcher registers handlers at initialization

- **GIVEN** the dispatcher is created
- **WHEN** handlers are registered
- **THEN** each command name maps to exactly one handler function

### Requirement: Handler Package

The Go backend SHALL organize operation handlers in a handlers package.

#### Scenario: Handlers implement consistent signature

- **GIVEN** any operation handler
- **WHEN** the handler signature is reviewed
- **THEN** it accepts `[]byte` and returns `([]byte, error)`

#### Scenario: Handlers are registered with dispatcher

- **GIVEN** all operation handlers are implemented
- **WHEN** the dispatcher is initialized
- **THEN** all handlers are registered with their command names

### Requirement: Any-Sync Integration Package

The Go backend SHALL provide an Any-Sync integration package wrapping SpaceService and ObjectTree.

#### Scenario: Integration package exposes space operations

- **GIVEN** the Any-Sync integration package
- **WHEN** the package API is reviewed
- **THEN** it provides functions for creating, joining, leaving, listing, and deleting spaces

#### Scenario: Integration package exposes document operations

- **GIVEN** the Any-Sync integration package
- **WHEN** the package API is reviewed
- **THEN** it provides functions for creating, reading, updating, deleting, listing, and querying documents

#### Scenario: Integration package handles Any-Sync lifecycle

- **GIVEN** the plugin initializes
- **WHEN** the Any-Sync integration is initialized
- **THEN** SpaceService and related components are properly initialized

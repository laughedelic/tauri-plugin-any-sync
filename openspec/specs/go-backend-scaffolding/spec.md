# Go Backend Scaffolding Specification

## Purpose
Provides a Go backend with gRPC services for health checks and plugin communication, including proper package structure, Protocol Buffer definitions, and cross-platform build support.
## Requirements
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

### Requirement: gRPC Health Check Service
The Go backend SHALL provide a gRPC health check service that responds to health status requests.
#### Scenario:
Given the desktop plugin needs to verify the Go backend is running
When the plugin calls the health check endpoint
Then the Go backend should respond with a successful health status

### Requirement: gRPC Ping Service

The Go backend SHALL provide a gRPC ping service for testing communication between frontend and backend.

**Changes:**
- Ping remains as a single operation in the dispatcher pattern
- No structural changes to ping implementation

#### Scenario: Ping service via dispatcher

- **GIVEN** a ping command via the SyncSpace API
- **WHEN** the command is dispatched
- **THEN** a pong response is returned

### Requirement: Protocol Buffer Definitions
The project SHALL define Protocol Buffer service and message definitions for type-safe communication.
#### Scenario:
Given the Rust plugin and Go backend need a type-safe communication contract
When the gRPC services are defined
Then they should use Protocol Buffers with clear message definitions for health checks and ping operations

### Requirement: Go Server Configuration
The Go backend SHALL support configurable server settings for binding address and logging.
#### Scenario:
Given the Go backend needs to run as a sidecar process
When the server starts
Then it should bind to localhost on a configurable port with proper logging

### Requirement: Basic Error Handling
The gRPC services SHALL implement proper error handling with appropriate status codes.
#### Scenario:
Given the gRPC service may encounter errors during request processing
When an error occurs
Then the service should return appropriate gRPC status codes with meaningful error messages

### Requirement: Build Configuration
The Go backend SHALL support cross-compilation to multiple target platforms.
#### Scenario:
Given the Go backend needs to be compiled for multiple platforms
When the build process runs
Then it should produce executables for all target platforms using standard Go toolchain

### Requirement: Go Backend Documentation
The Go backend SHALL include component-specific AGENTS.md documentation for development workflows.
#### Scenario:
Given developers need to work with the Go backend code
When they open the plugin-go-backend directory
Then they should find clear instructions for building, testing, and gRPC development

### Requirement: Project Structure
The existing project structure SHALL accommodate the Go backend directory without conflicts.
#### Scenario:
Given the existing Tauri plugin structure
When adding the Go backend
Then the `plugin-go-backend/` directory should integrate cleanly with the existing project layout

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


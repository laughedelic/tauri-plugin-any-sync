# Go Backend Scaffolding Specification

## Purpose
Provides a Go backend with gRPC services for health checks and plugin communication, including proper package structure, Protocol Buffer definitions, and cross-platform build support.
## Requirements
### Requirement: Basic Go Backend Structure
The project SHALL provide a Go backend with proper package structure separating API and internal code.
#### Scenario:
Given the project needs a Go backend for AnySync integration
When the developer sets up the project structure
Then the Go backend should have a proper package structure with clear separation between API and internal code

### Requirement: gRPC Health Check Service
The Go backend SHALL provide a gRPC health check service that responds to health status requests.
#### Scenario:
Given the desktop plugin needs to verify the Go backend is running
When the plugin calls the health check endpoint
Then the Go backend should respond with a successful health status

### Requirement: gRPC Ping Service
The Go backend SHALL provide a gRPC ping service for testing communication between frontend and backend.
#### Scenario:
Given the TypeScript frontend needs to test communication with the Go backend
When the frontend invokes a ping command through the plugin
Then the Go backend should receive the ping request and return a pong response

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
When they open the go-backend directory
Then they should find clear instructions for building, testing, and gRPC development

### Requirement: Project Structure
The existing project structure SHALL accommodate the Go backend directory without conflicts.
#### Scenario:
Given the existing Tauri plugin structure
When adding the Go backend
Then the `go-backend/` directory should integrate cleanly with the existing project layout

### Requirement: AnyStore Integration

The Go backend SHALL integrate AnyStore for document storage capabilities.

#### Scenario: AnyStore is integrated

- **GIVEN** the Go backend implementation
- **WHEN** storage operations are performed
- **THEN** AnyStore library is used for document persistence

### Requirement: Storage Module Organization

The Go backend SHALL organize storage code in `internal/storage/` to isolate AnyStore integration.

#### Scenario: Storage wrapper is in internal/storage

- **GIVEN** the Go backend project structure
- **WHEN** storage code is located
- **THEN** the wrapper implementation is in `internal/storage/anystore.go`

#### Scenario: Storage types are internal

- **GIVEN** the storage wrapper
- **WHEN** exported types are examined
- **THEN** AnyStore-specific types are not exposed outside internal/storage

### Requirement: gRPC Server Registration

The Go backend SHALL register the StorageService with the gRPC server during initialization.

#### Scenario: Storage service is registered

- **GIVEN** the main gRPC server
- **WHEN** services are registered
- **THEN** StorageService is included alongside HealthService

#### Scenario: Storage service uses shared database instance

- **GIVEN** the storage service initialization
- **WHEN** the service is created
- **THEN** it receives a reference to the shared AnyStore database instance


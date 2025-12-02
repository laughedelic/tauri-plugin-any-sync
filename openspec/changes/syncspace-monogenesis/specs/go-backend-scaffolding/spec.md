# Go Backend Scaffolding Specification

## REMOVED Requirements

### Requirement: Per-Operation Mobile Exports

~~The mobile package SHALL export individual functions for each storage operation.~~

**Reason:** Replaced by single Command function in single-dispatch pattern.

### Requirement: Storage Wrapper Abstraction

~~The Go backend SHALL provide a storage wrapper in `internal/storage/anystore.go` that abstracts AnyStore-specific types.~~

**Reason:** Replaced by Any-Sync integration layer that uses SpaceService and ObjectTree.

## MODIFIED Requirements

### Requirement: Mobile Package Structure

The mobile package SHALL provide gomobile-compatible exports for Android and iOS integration.

**Changes:**
- Exports are reduced from N per-operation functions to exactly 4 functions: Init, Command, SetEventHandler, Shutdown
- Functions use dispatcher pattern instead of direct operation handlers

#### Scenario: Mobile package exports Init function

- **GIVEN** the mobile package is compiled with gomobile
- **WHEN** the exported API is inspected
- **THEN** an Init function with signature `Init(dataPath string) error` is available

#### Scenario: Mobile package exports Command function

- **GIVEN** the mobile package is compiled with gomobile
- **WHEN** the exported API is inspected
- **THEN** a Command function with signature `Command(cmd string, data []byte) ([]byte, error)` is available

#### Scenario: Mobile package exports SetEventHandler function

- **GIVEN** the mobile package is compiled with gomobile
- **WHEN** the exported API is inspected
- **THEN** a SetEventHandler function with signature `SetEventHandler(handler func([]byte))` is available

#### Scenario: Mobile package exports Shutdown function

- **GIVEN** the mobile package is compiled with gomobile
- **WHEN** the exported API is inspected
- **THEN** a Shutdown function with signature `Shutdown() error` is available

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

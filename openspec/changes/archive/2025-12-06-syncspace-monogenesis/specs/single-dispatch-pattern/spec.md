# Single Dispatch Pattern Specification

## ADDED Requirements

### Requirement: Single Command Function

The Go backend SHALL expose a single `Command(cmd string, data []byte) ([]byte, error)` function instead of per-operation exports.

#### Scenario: Command function accepts command name

- **GIVEN** a SyncSpace operation needs to be invoked
- **WHEN** the Command function is called
- **THEN** the first parameter is a string identifying the operation (e.g., "DocCreate", "SpaceList")

#### Scenario: Command function accepts serialized request

- **GIVEN** a SyncSpace operation with request parameters
- **WHEN** the Command function is called
- **THEN** the second parameter is the protobuf-serialized request message as bytes

#### Scenario: Command function returns serialized response

- **GIVEN** a Command function call completes successfully
- **WHEN** the response is returned
- **THEN** it is the protobuf-serialized response message as bytes

#### Scenario: Command function returns error for failures

- **GIVEN** a Command function call encounters an error
- **WHEN** the function returns
- **THEN** the error parameter contains the error details

### Requirement: Four Mobile Exports

The Go mobile backend SHALL export exactly four functions via gomobile.

#### Scenario: Init function initializes plugin

- **GIVEN** the mobile platform (Android or iOS)
- **WHEN** the plugin is initialized
- **THEN** `Init(dataPath string) error` function is called with the data directory path

#### Scenario: Command function dispatches operations

- **GIVEN** the plugin is initialized
- **WHEN** any SyncSpace operation needs to be invoked
- **THEN** `Command(cmd string, data []byte) ([]byte, error)` function is called

#### Scenario: SetEventHandler registers callback

- **GIVEN** the application wants to receive events
- **WHEN** event subscription is set up
- **THEN** `SetEventHandler(handler func([]byte))` function is called with the callback

#### Scenario: Shutdown function cleans up

- **GIVEN** the application is closing
- **WHEN** cleanup is needed
- **THEN** `Shutdown() error` function is called

#### Scenario: No other functions are exported

- **GIVEN** the mobile package is compiled with gomobile
- **WHEN** the exported API is inspected
- **THEN** only Init, Command, SetEventHandler, and Shutdown functions are exported

### Requirement: Command Dispatcher

The Go backend SHALL implement a dispatcher that routes command strings to handler functions.

#### Scenario: Dispatcher maintains handler registry

- **GIVEN** the dispatcher is initialized
- **WHEN** handlers are registered
- **THEN** each command name maps to exactly one handler function

#### Scenario: Dispatcher routes to correct handler

- **GIVEN** a Command call with command name "DocCreate"
- **WHEN** the dispatcher processes the command
- **THEN** the DocCreate handler function is invoked

#### Scenario: Dispatcher returns error for unknown command

- **GIVEN** a Command call with unrecognized command name
- **WHEN** the dispatcher processes the command
- **THEN** an error is returned indicating unknown command

#### Scenario: Dispatcher handles malformed request data

- **GIVEN** a Command call with data that fails protobuf deserialization
- **WHEN** the dispatcher processes the command
- **THEN** an error is returned indicating malformed request

### Requirement: Handler Function Pattern

Each SyncSpace operation SHALL be implemented as a handler function with consistent signature.

#### Scenario: Handler receives raw bytes

- **GIVEN** a handler function for an operation
- **WHEN** the handler is invoked by the dispatcher
- **THEN** it receives the serialized protobuf request as bytes

#### Scenario: Handler deserializes request

- **GIVEN** a handler receives request bytes
- **WHEN** the handler processes the request
- **THEN** it deserializes bytes to the appropriate protobuf request type

#### Scenario: Handler executes business logic

- **GIVEN** a handler has deserialized the request
- **WHEN** the handler processes the operation
- **THEN** it invokes Any-Sync APIs to perform the operation

#### Scenario: Handler serializes response

- **GIVEN** a handler has completed the operation
- **WHEN** the handler prepares the response
- **THEN** it serializes the protobuf response type to bytes

#### Scenario: Handler returns bytes or error

- **GIVEN** a handler has processed the operation
- **WHEN** the handler returns
- **THEN** it returns either serialized response bytes or an error

### Requirement: Desktop Unified Interface

The desktop backend SHALL expose the same Command interface as mobile.

#### Scenario: Desktop sidecar uses dispatcher

- **GIVEN** the desktop sidecar process is running
- **WHEN** the Rust plugin sends a command
- **THEN** the command is processed by the same dispatcher as mobile

#### Scenario: Desktop maintains interface consistency

- **GIVEN** both desktop and mobile implementations
- **WHEN** the command interfaces are compared
- **THEN** they both use Command(cmd, data) → (bytes, error) pattern

### Requirement: Single Rust Command Handler

The Rust plugin SHALL expose a single Tauri command that forwards to the backend.

#### Scenario: Rust command handler is defined

- **GIVEN** the Rust plugin implementation
- **WHEN** Tauri commands are registered
- **THEN** exactly one command named "command" is registered

#### Scenario: Rust command forwards to backend

- **GIVEN** the Rust command handler receives a call
- **WHEN** the handler processes the call
- **THEN** it forwards cmd and data to the backend's Command function

#### Scenario: Rust command returns raw bytes

- **GIVEN** the backend returns response bytes
- **WHEN** the Rust command handler receives the response
- **THEN** it returns the bytes directly to TypeScript without modification

### Requirement: Minimal Native Shims

iOS and Android native plugins SHALL contain minimal passthrough code only.

#### Scenario: iOS shim calls Go C export

- **GIVEN** the iOS Swift plugin implementation
- **WHEN** the command method is called
- **THEN** it directly calls the Go C-exported Command function via FFI

#### Scenario: Android shim calls Go JNI

- **GIVEN** the Android Kotlin plugin implementation
- **WHEN** the command method is called
- **THEN** it directly calls the Go JNI Command function

#### Scenario: Native shims contain no business logic

- **GIVEN** iOS and Android plugin implementations
- **WHEN** the code is reviewed
- **THEN** they contain only plugin initialization, FFI bridging, and no business logic

#### Scenario: Native shims under 50 lines each

- **GIVEN** iOS and Android plugin implementations
- **WHEN** lines of code are counted
- **THEN** each implementation is under 50 lines of code

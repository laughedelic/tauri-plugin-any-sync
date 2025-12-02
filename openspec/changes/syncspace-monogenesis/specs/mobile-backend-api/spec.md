# Mobile Backend API Specification

## REMOVED Requirements

### Requirement: Individual Storage Operation Exports

~~The mobile package SHALL export separate functions for Put, Get, Delete, and List operations.~~

**Reason:** Replaced by single Command function.

### Requirement: JSON Serialization in Mobile Layer

~~The mobile package SHALL handle JSON marshaling/unmarshaling for storage list results.~~

**Reason:** All serialization now handled via protobuf, mobile layer only passes bytes.

## MODIFIED Requirements

### Requirement: Gomobile-Compatible Exports

The mobile package SHALL export functions that are compatible with gomobile's type restrictions.

**Changes:**
- Exports reduced to 4 functions total
- All functions use gomobile-compatible types (string, []byte, error, func([]byte))

#### Scenario: Init function is gomobile-compatible

- **GIVEN** the Init function signature
- **WHEN** gomobile compatibility is checked
- **THEN** it uses only string (dataPath) and error return type

#### Scenario: Command function is gomobile-compatible

- **GIVEN** the Command function signature
- **WHEN** gomobile compatibility is checked
- **THEN** it uses only string (cmd), []byte (data), []byte (response), and error

#### Scenario: SetEventHandler function is gomobile-compatible

- **GIVEN** the SetEventHandler function signature
- **WHEN** gomobile compatibility is checked
- **THEN** it uses only func([]byte) callback type

## ADDED Requirements

### Requirement: Four-Function Mobile API

The mobile package SHALL export exactly four functions for the entire plugin API.

#### Scenario: Init initializes plugin with data path

- **GIVEN** the mobile platform needs to initialize the plugin
- **WHEN** Init(dataPath string) is called
- **THEN** the plugin initializes Any-Sync with the provided data path

#### Scenario: Command dispatches operations

- **GIVEN** the mobile platform needs to invoke any operation
- **WHEN** Command(cmd string, data []byte) is called
- **THEN** the command is dispatched to the appropriate handler and response bytes are returned

#### Scenario: SetEventHandler registers callback

- **GIVEN** the mobile platform needs to receive events
- **WHEN** SetEventHandler(handler func([]byte)) is called
- **THEN** the handler callback is registered for future event notifications

#### Scenario: Shutdown cleans up resources

- **GIVEN** the mobile platform is shutting down the plugin
- **WHEN** Shutdown() is called
- **THEN** all resources are cleaned up and Any-Sync is properly shut down

### Requirement: Event Handler Callback

The mobile package SHALL support asynchronous event callbacks via SetEventHandler.

#### Scenario: Event handler is called for plugin events

- **GIVEN** SetEventHandler has registered a callback
- **WHEN** a plugin event occurs (document change, sync status)
- **THEN** the callback is invoked with serialized event bytes

#### Scenario: Event handler can be updated

- **GIVEN** SetEventHandler has been called once
- **WHEN** SetEventHandler is called again with a new handler
- **THEN** the new handler replaces the previous one

### Requirement: Thread-Safe Command Execution

The mobile package SHALL ensure thread-safe execution of Command calls.

#### Scenario: Concurrent Command calls are handled safely

- **GIVEN** multiple threads on mobile platform
- **WHEN** concurrent Command calls are made
- **THEN** they are handled safely without data races

#### Scenario: Dispatcher state is protected

- **GIVEN** the dispatcher maintains internal state
- **WHEN** concurrent operations access the state
- **THEN** appropriate synchronization prevents corruption

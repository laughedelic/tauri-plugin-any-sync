# Mobile Backend API Specification

## REMOVED Requirements

### Requirement: gomobile-Compatible Storage API

~~The Go backend SHALL provide a mobile-specific entrypoint with gomobile-compatible function signatures for all storage operations.~~

**Reason:** Replaced by generic four-function Command API via single-dispatch pattern.

### Requirement: Shared Backend Code

~~The mobile entrypoint SHALL reuse >95% of the existing storage implementation.~~

**Reason:** Replaced by unified dispatcher pattern and Any-Sync integration.

### Requirement: State Management

~~The mobile backend SHALL manage database connection lifecycle internally.~~

**Reason:** Replaced by init-dispatch-shutdown pattern.

## MODIFIED Requirements

### Requirement: Type Compatibility

All exported mobile functions SHALL use only gomobile-compatible types.

**Changes:**
- Exports reduced to 4 functions total: Init, Command, SetEventHandler, Shutdown
- All functions use gomobile-compatible types (string, []byte, error, func([]byte))

#### Scenario: All mobile functions use gomobile-compatible signatures

- **GIVEN** the mobile package exported functions
- **WHEN** gomobile compatibility is checked
- **THEN** all 4 functions (Init, Command, SetEventHandler, Shutdown) use only compatible types

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

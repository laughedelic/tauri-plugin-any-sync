# Spec Delta: Android Plugin Integration

## ADDED Requirements

### Requirement: gomobile Library Loading
The Android plugin SHALL load the gomobile-generated native library at initialization.

#### Scenario: Native library initialization
- **GIVEN** an Android app using the plugin
- **WHEN** the plugin class is first loaded
- **THEN** the plugin loads the `gojni` native library via `System.loadLibrary("gojni")`
- **AND** loading happens in a static initializer block before any instance methods
- **AND** loading failure throws UnsatisfiedLinkError with descriptive message
- **AND** the error is logged to Android logcat for debugging

### Requirement: Storage Command Handlers
The Android plugin SHALL implement Tauri command handlers for all storage operations.

#### Scenario: Handle storage get command
- **GIVEN** the plugin is initialized
- **WHEN** a `storageGet` command is invoked from TypeScript
- **THEN** the plugin parses `StorageGetArgs` (collection, id)
- **AND** calls `Mobile.storageGet(collection, id)` via JNI
- **AND** constructs JSObject response with `documentJson` and `found` fields
- **AND** resolves the invoke with the response object

#### Scenario: Handle storage put command
- **GIVEN** the plugin is initialized
- **WHEN** a `storagePut` command is invoked from TypeScript
- **THEN** the plugin parses `StoragePutArgs` (collection, id, documentJson)
- **AND** calls `Mobile.storagePut(collection, id, documentJson)` via JNI
- **AND** constructs JSObject response with `success: true`
- **AND** resolves the invoke with the response object

#### Scenario: Handle storage delete command
- **GIVEN** the plugin is initialized
- **WHEN** a `storageDelete` command is invoked from TypeScript
- **THEN** the plugin parses `StorageDeleteArgs` (collection, id)
- **AND** calls `Mobile.storageDelete(collection, id)` via JNI
- **AND** constructs JSObject response with `existed` boolean
- **AND** resolves the invoke with the response object

#### Scenario: Handle storage list command
- **GIVEN** the plugin is initialized
- **WHEN** a `storageList` command is invoked from TypeScript
- **THEN** the plugin parses `StorageListArgs` (collection)
- **AND** calls `Mobile.storageList(collection)` via JNI
- **AND** constructs JSObject response with `ids` array
- **AND** resolves the invoke with the response object

### Requirement: Error Propagation
The Android plugin SHALL catch and properly propagate errors from the Go backend.

#### Scenario: Go backend error handling
- **GIVEN** a storage operation that fails in Go
- **WHEN** the Go function throws an exception (converted from Go error)
- **THEN** the Kotlin plugin catches the exception in try-catch block
- **AND** extracts the error message from exception
- **AND** calls `invoke.reject("STORAGE_ERROR", message)`
- **AND** the error propagates to Rust and then TypeScript with proper error type

#### Scenario: Invalid argument handling
- **GIVEN** invalid arguments passed from TypeScript
- **WHEN** the plugin fails to parse arguments
- **THEN** the plugin catches the parse exception
- **AND** rejects the invoke with descriptive error message
- **AND** error includes information about which argument was invalid

### Requirement: Database Path Management
The Android plugin SHALL configure the database path to use Android-appropriate storage location.

#### Scenario: Database initialization path
- **GIVEN** the plugin is constructed with Activity context
- **WHEN** the plugin initializes storage
- **THEN** it calculates dbPath as `activity.filesDir.absolutePath + "/anystore.db"`
- **AND** this path is within app's private internal storage
- **AND** calls `Mobile.initStorage(dbPath)` during plugin initialization
- **AND** logs initialization success or failure

#### Scenario: Storage persistence
- **GIVEN** storage initialized in app's filesDir
- **WHEN** the app is closed and reopened
- **THEN** the same database file is reused
- **AND** previously stored documents are still accessible
- **AND** no data loss occurs

### Requirement: Rust Mobile Module Integration
The Rust mobile module SHALL dispatch storage commands to the Android plugin.

#### Scenario: Rust to Kotlin command dispatch
- **GIVEN** the Rust plugin is built for Android target
- **WHEN** a storage command is called (e.g., `storage_get`)
- **THEN** the Rust mobile module uses `PluginHandle.run_mobile_plugin()`
- **AND** passes the command name ("storageGet") and arguments
- **AND** receives the response as JSObject
- **AND** deserializes to appropriate Rust type
- **AND** returns Result to the command caller

### Requirement: Command Registration
The Android plugin SHALL register all storage commands with Tauri's command system.

#### Scenario: Plugin command registration
- **GIVEN** the ExamplePlugin class with @TauriPlugin annotation
- **WHEN** the Tauri runtime initializes the plugin
- **THEN** all methods annotated with @Command are registered
- **AND** commands are accessible from TypeScript via their method names
- **AND** argument parsing is handled by Tauri framework
- **AND** response serialization handled by Tauri framework

## MODIFIED Requirements

None. The Android plugin implementation is additive and doesn't modify existing requirements.

## REMOVED Requirements

None.

## Dependencies

- **Internal:** Requires mobile backend API spec (mobile-backend-api/spec.md)
- **External:** 
  - Requires Tauri v2.0+ Android plugin framework
  - Requires gomobile-generated .aar artifact
- **Related Specs:** 
  - `storage-api/spec.md` (defines storage command behavior)
  - `desktop-integration/spec.md` (desktop command behavior for comparison)

## Notes

**Platform Parity:** The Android plugin implementation mirrors the desktop command structure as much as possible. The only difference is the transport layer (JNI vs gRPC), not the command semantics.

**Threading Considerations:** gomobile-generated code is generally thread-safe, but the plugin should ensure that concurrent storage operations are handled correctly. The Go backend's internal AnyStore connection handles concurrency.

**Logging Strategy:** Use Android's `android.util.Log` with tag "AnySync" for all plugin logging. This makes debugging easier via `adb logcat | grep AnySync`.

**Gradle Configuration:** The plugin's `build.gradle.kts` must include the .aar dependency. This is handled by the consumer's build.rs copying the .aar to the appropriate location.

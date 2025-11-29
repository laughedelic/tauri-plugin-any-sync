# Integration Testing Specification

## ADDED Requirements

### Requirement: Test Infrastructure Using Tauri IPC

The plugin SHALL provide test infrastructure that uses Tauri's official IPC testing utilities to invoke commands through the actual application invocation path.

#### Scenario: Tests invoke commands through IPC mechanism

- **WHEN** a test needs to call a plugin command
- **THEN** test uses `tauri::test::get_ipc_response()` or `assert_ipc_response()`
- **AND** command is invoked as `"plugin:any-sync|<command_name>"`
- **AND** request payload is serialized through `InvokeBody::Json`
- **AND** response is deserialized from IPC response
- **AND** the full command invocation path is tested (App → Plugin → Backend)

#### Scenario: Test helper creates app with webview

- **WHEN** tests call the test helper function
- **THEN** a Tauri app with `MockRuntime` is created
- **AND** the app includes the any-sync plugin
- **AND** a webview is available for IPC communication
- **AND** the app can spawn platform-specific backend (sidecar or FFI)

### Requirement: Cross-Platform Testing Support

The plugin SHALL support integration testing on desktop and Android platforms, with iOS testing infrastructure documented for future activation when iOS plugin support is implemented.

#### Scenario: Desktop tests use sidecar backend

- **WHEN** tests run on desktop platform (Linux, macOS, Windows)
- **THEN** Go backend is built as sidecar executable using `task backend:build`
- **AND** sidecar process spawns automatically on first command
- **AND** communication occurs via gRPC over localhost
- **AND** tests verify sidecar-specific behavior (process management, gRPC)

#### Scenario: Android tests use FFI bindings

- **WHEN** tests run on Android platform
- **THEN** Go backend is built as native library using `task go:mobile:build-android`
- **AND** gomobile .aar with JNI bindings is used
- **AND** tests verify FFI-specific behavior (library loading, direct calls)
- **AND** Kotlin plugin wrapper is validated

#### Scenario: iOS testing infrastructure is documented but not active

- **WHEN** iOS testing documentation is reviewed
- **THEN** CI job for iOS is documented but commented out
- **AND** documentation explains iOS plugin doesn't exist yet
- **AND** infrastructure is ready to activate when iOS support is added
- **AND** gomobile .xcframework build process is documented

#### Scenario: Shared tests run on all platforms

- **WHEN** test behavior is platform-independent
- **THEN** test implementation is shared (no `#[cfg]` conditionals)
- **AND** test runs on both desktop and mobile
- **AND** same assertions apply regardless of platform

### Requirement: Health Check Testing

The plugin SHALL verify backend connectivity and basic communication through ping tests on all platforms.

#### Scenario: Ping with message succeeds

- **WHEN** test invokes ping command with a message value
- **THEN** backend starts automatically (sidecar or FFI library)
- **AND** connection is established (gRPC or FFI)
- **AND** response echoes the message back
- **AND** the operation succeeds without errors

#### Scenario: Ping with empty message succeeds

- **WHEN** test invokes ping command with None/null message
- **THEN** the operation succeeds without errors
- **AND** response is returned (may be empty or default)

### Requirement: Storage Put and Get Testing

The plugin SHALL verify document creation, retrieval, and update operations through storage tests on all platforms.

#### Scenario: Put and get document succeeds

- **WHEN** test invokes storage_put with JSON document
- **AND** then invokes storage_get for same collection and ID
- **THEN** put operation returns success response
- **AND** get operation returns found=true
- **AND** retrieved JSON matches original document
- **AND** all JSON types are preserved (string, number, bool, null, array, object)

#### Scenario: Get nonexistent document returns not found

- **WHEN** test invokes storage_get for document that doesn't exist
- **THEN** operation succeeds without throwing errors
- **AND** response has found=false
- **AND** document_json is null/None

#### Scenario: Update existing document overwrites

- **WHEN** test invokes storage_put with initial document data
- **AND** then invokes storage_put for same collection/ID with updated data
- **AND** then invokes storage_get for that document
- **THEN** retrieved document contains updated data
- **AND** no duplicate entries are created
- **AND** latest write wins

### Requirement: Storage List Testing

The plugin SHALL verify document listing operations for both populated and empty collections on all platforms.

#### Scenario: List returns all documents in collection

- **WHEN** test invokes storage_put for 5 documents in a collection
- **AND** then invokes storage_list for that collection
- **THEN** response contains exactly 5 document IDs
- **AND** all expected IDs are present in response
- **AND** only documents from that collection are returned

#### Scenario: List empty collection returns empty result

- **WHEN** test invokes storage_list for collection that was never created
- **THEN** operation succeeds without errors
- **AND** response contains zero document IDs
- **AND** no error is thrown

### Requirement: Storage Delete Testing

The plugin SHALL verify document deletion operations for both existing and nonexistent documents on all platforms.

#### Scenario: Delete existing document succeeds

- **WHEN** test invokes storage_put to create a document
- **AND** then invokes storage_delete for that document
- **AND** then invokes storage_get for that document
- **THEN** delete operation returns existed=true
- **AND** subsequent get returns found=false
- **AND** document is no longer retrievable

#### Scenario: Delete nonexistent document is idempotent

- **WHEN** test invokes storage_delete for document that doesn't exist
- **THEN** operation succeeds without errors
- **AND** response has existed=false
- **AND** no error is thrown

### Requirement: Complex Scenario Testing

The plugin SHALL verify correct behavior in complex usage scenarios including collection isolation and complex JSON documents on all platforms.

#### Scenario: Collections are isolated namespaces

- **WHEN** test invokes storage_put with same document ID in three different collections
- **AND** each document has different content
- **AND** then invokes storage_get for each collection
- **THEN** each retrieval returns correct content for that collection
- **AND** no cross-contamination occurs between collections
- **AND** collection acts as isolated namespace

#### Scenario: Complex JSON documents are preserved

- **WHEN** test invokes storage_put with document containing nested objects (3+ levels), arrays with mixed types, Unicode characters (emoji, CJK), special characters (quotes, backslashes, newlines), and all JSON types
- **AND** then invokes storage_get for that document
- **THEN** all fields are preserved exactly as stored
- **AND** character encoding is correct (no mojibake)
- **AND** data integrity is maintained through serialization layers

### Requirement: CI Integration - Desktop

The plugin SHALL run desktop integration tests automatically in CI on every push and pull request.

#### Scenario: Desktop CI builds backend before tests

- **WHEN** desktop CI workflow runs
- **THEN** Go backend is built for target platform using `task backend:build`
- **AND** `ANY_SYNC_GO_BINARIES_DIR` environment variable points to built binaries
- **AND** all system dependencies are installed (webkit2gtk, protoc)
- **AND** tests can locate and spawn sidecar process

#### Scenario: Desktop CI runs tests with correct configuration

- **WHEN** desktop CI executes integration tests
- **THEN** tests run using `task app:test`
- **AND** tests execute with `--test-threads=1` to prevent database conflicts
- **AND** `RUST_LOG=debug` is set for detailed logging
- **AND** tests run on appropriate runner (ubuntu-latest for Linux, macos-latest for macOS)

### Requirement: CI Integration - Mobile

The plugin SHALL run Android integration tests automatically in CI, and document iOS testing infrastructure for future activation.

#### Scenario: Android CI sets up emulator and SDK

- **WHEN** Android CI workflow runs
- **THEN** Android SDK is installed using `android-actions/setup-android`
- **AND** JDK 17 is configured
- **AND** Android emulator is started with API level 33
- **AND** gomobile .aar library is built using `task go:mobile:build-android`
- **AND** emulator is ready before tests run

#### Scenario: Android tests run in emulator

- **WHEN** Android CI executes integration tests
- **THEN** tests run inside Android emulator using `reactivecircus/android-emulator-runner`
- **AND** cargo features select Android-specific code paths
- **AND** tests verify JNI bindings and Kotlin plugin wrapper
- **AND** test results are reported even if emulator is slow to start

#### Scenario: iOS CI job is documented for future use

- **WHEN** iOS CI workflow configuration is reviewed
- **THEN** test-integration-ios job is present but commented out
- **AND** job documents macOS runner requirement (macos-latest)
- **AND** job documents Xcode selection and configuration
- **AND** job documents iOS simulator setup
- **AND** job documents gomobile .xcframework build (`task go:mobile:build-ios`)
- **AND** comment explains job will be enabled when iOS plugin is implemented

### Requirement: CI Path Correctness

The plugin SHALL use current project structure paths in CI workflows without relying on outdated references.

#### Scenario: CI uses current project paths

- **WHEN** CI workflow references project files or directories
- **THEN** paths use current structure (`plugin-go-backend/`, `example-app/`)
- **AND** no references to old paths (`go-backend/`, `examples/tauri-app/`)
- **AND** Task is used for all builds instead of shell scripts
- **AND** no non-existent scripts or paths are referenced

### Requirement: Test Implementation Best Practices

The plugin SHALL enforce best practices for test implementation to ensure reliability and debuggability.

#### Scenario: Tests use unique collection names

- **WHEN** each test needs a collection for storage operations
- **THEN** test uses a unique collection name to avoid cross-test contamination
- **AND** collection name includes test function name or UUID
- **AND** leftover data from previous tests doesn't cause failures

#### Scenario: Assertions provide context

- **WHEN** an assertion fails in CI
- **THEN** error message includes detailed context (expected vs actual values)
- **AND** error message identifies which command or operation failed
- **AND** error source is identifiable without needing to access full logs

#### Scenario: IPC responses are properly deserialized

- **WHEN** test receives IPC response from plugin command
- **THEN** response is deserialized to expected response type
- **AND** deserialization errors are caught and reported clearly
- **AND** test asserts on response fields, not raw JSON strings

#### Scenario: JSON comparison uses parsed values

- **WHEN** test compares JSON documents from storage operations
- **THEN** JSON is parsed to structured data before comparison
- **AND** field order differences don't cause false failures
- **AND** semantic equality is verified, not string equality

### Requirement: Documentation Coverage

The plugin SHALL document integration testing approach, platform-specific setup, and CI workflows.

#### Scenario: AGENTS.md documents testing approach

- **WHEN** developers consult AGENTS.md
- **THEN** integration tests section explains what is tested and on which platforms
- **AND** section shows how to run tests (`task app:test-integration`)
- **AND** section explains IPC testing approach using Tauri utilities
- **AND** section explains `--test-threads=1` requirement
- **AND** section covers both desktop and mobile testing

#### Scenario: Test README documents platform setup

- **WHEN** developers read example-app/src-tauri/tests/README.md
- **THEN** README explains how to set up desktop and Android for local testing
- **AND** README documents Android emulator setup and requirements
- **AND** README explains platform-specific test execution for active platforms
- **AND** README includes troubleshooting guide for Android emulator issues
- **AND** README documents iOS setup guide for future use (clearly marked as not yet active)

#### Scenario: Documentation covers CI workflows

- **WHEN** developers need to understand CI testing
- **THEN** documentation explains active CI jobs (desktop, Android)
- **AND** documentation explains documented-but-inactive iOS job
- **AND** documentation shows how to debug CI failures for active platforms
- **AND** documentation explains build dependencies for each platform
- **AND** documentation clarifies iOS job will be enabled when iOS plugin is implemented
- **AND** paths reference current project structure

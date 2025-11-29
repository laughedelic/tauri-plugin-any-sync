# Integration Tests

This directory contains comprehensive integration tests for the tauri-plugin-any-sync.

## Overview

These integration tests verify end-to-end functionality of the plugin with the Go backend, **without requiring a GUI**. They use `tauri::test::MockRuntime` to create a headless app instance that simulates the webview and allows testing all plugin commands through the actual IPC layer.

## Test Coverage

The integration tests cover all plugin commands across 10 test cases:

### Basic Commands (2 tests)
- `test_ping_command`: Basic health check and message echoing
- `test_ping_command_empty_message`: Handling empty/None messages

### Storage Operations (8 tests)
- `test_storage_put_and_get`: Creating documents and retrieving with JSON integrity verification
- `test_storage_get_nonexistent`: Graceful handling of missing documents
- `test_storage_update_existing_document`: Upsert behavior (update existing documents)
- `test_storage_list`: Listing all documents in a collection
- `test_storage_list_empty`: Empty collection handling
- `test_storage_delete`: Deleting existing documents
- `test_storage_delete_nonexistent`: Idempotent delete (deleting nonexistent documents)
- `test_multiple_collections`: Collection isolation (same ID in different collections)

### What's Tested
- ✅ All 5 plugin commands (ping, storage_put, storage_get, storage_delete, storage_list)
- ✅ IPC layer (commands invoked via `get_ipc_response()`)
- ✅ Desktop sidecar process management (automatic startup)
- ✅ gRPC communication (desktop)
- ✅ Error handling and edge cases
- ✅ JSON serialization/deserialization
- ✅ Complex nested JSON documents with Unicode and special characters
- ✅ Collection isolation

## Running the Tests

### Desktop Tests (macOS, Linux, Windows)

#### Prerequisites

1. **Go backend binary**: Built automatically by the task command, or build manually:
   ```bash
   # From project root
   task go:build  # or task backend:desktop:build
   ```

2. **Environment variable** to use local binaries (optional, set automatically by build.rs):
   ```bash
   export ANY_SYNC_GO_BINARIES_DIR=$(pwd)/binaries
   ```

#### Run Desktop Tests

```bash
# From project root (recommended - handles all setup)
task app:test-integration

# Or manually from example-app/src-tauri
cargo test --test integration --features integration-test -- --test-threads=1

# With detailed logging
RUST_LOG=debug task app:test-integration

# Run a specific test
cargo test --test integration --features integration-test test_ping_command -- --test-threads=1
```

### Mobile Tests (Android)

**Android integration tests run in CI only** due to the complexity of emulator setup. Tests execute automatically on every push/PR via GitHub Actions.

#### CI Testing (Recommended)

Android tests run in the `test-integration-android` job in `.github/workflows/test.yml`:
- Uses `reactivecircus/android-emulator-runner` to spawn an Android emulator
- Runs API level 33 (x86_64 architecture)
- Builds the Android .aar library automatically
- Executes the same integration tests as desktop, but through Android FFI layer

#### Local Development

For local Android development and debugging, use:
```bash
task dev:android  # Run the example app in Android emulator
```

This provides the same integration testing capability through the UI without the complexity of running headless tests on the emulator.

#### Advanced: Local Android Test Execution (Not Recommended)

If you need to run Android tests locally, you would need:

1. **Android SDK and emulator** (already set up if using `task dev:android`)
2. **gomobile for building .aar**:
   ```bash
   go install golang.org/x/mobile/cmd/gomobile@latest
   gomobile init
   ```
3. **Rust Android targets**:
   ```bash
   rustup target add aarch64-linux-android
   ```
4. **cargo-dinghy or similar tool** to execute tests on the emulator
5. **Running emulator** before test execution

This is significantly more complex than desktop testing and is not recommended for routine development. Use CI for Android test verification.

### Why `--test-threads=1`?

Tests are run sequentially to avoid conflicts:
- Each test creates its own app instance
- All instances share the same Go backend sidecar process
- The sidecar uses a single database file
- Running tests in parallel could cause race conditions in the database

## How It Works

### Test Infrastructure

1. **App Setup**: Each test calls `create_test_app()` which:
   - Uses the same `create_app_builder()` function as the production app
   - Creates a `MockRuntime` instead of spawning a real window
   - Creates a webview window for IPC testing
   - Returns `(app, webview, invoke_key)` for test use

2. **Command Execution**: Tests invoke commands through the **actual IPC layer** using `tauri::test::get_ipc_response()`:
   ```rust
   let res = get_ipc_response(
       &webview,
       tauri::webview::InvokeRequest {
           cmd: "plugin:any-sync|ping".into(),
           callback: tauri::ipc::CallbackFn(0),
           error: tauri::ipc::CallbackFn(1),
           body: json!({
               "payload": {
                   "value": "test message"
               }
           }).into(),
           headers: Default::default(),
           url: "tauri://localhost".parse().unwrap(),
           invoke_key: invoke_key.clone(),
       },
   );
   ```

   This approach:
   - ✅ Tests the actual invocation path (same as JavaScript frontend would use)
   - ✅ Verifies IPC serialization/deserialization
   - ✅ Uses official Tauri testing utilities
   - ✅ No test-only code pollution in plugin source

3. **Backend Communication**:
   - **Desktop**: First command automatically spawns the Go sidecar process (gRPC)
   - **Mobile**: Commands invoke the embedded Go library via JNI/FFI (Android .aar)
   - All subsequent commands reuse the same backend instance

4. **Verification**: Tests assert on:
   - Success/failure of operations
   - Response data correctness (parsed from IPC response)
   - Side effects (e.g., document retrieval after put)

## CI Integration

### Desktop Testing (Active)

The `test-integration-desktop` job in `.github/workflows/test.yml` runs desktop tests automatically:

- ✅ Runs on every push and pull request
- ✅ Tests on Ubuntu (fastest CI runner)
- ✅ Automatically builds the Go backend before running tests
- ✅ Installs all required system dependencies (webkit2gtk, protoc, Task)
- ✅ Fails the build if any test fails
- ✅ Uses `--test-threads=1` to prevent database conflicts

### Android Testing (Active)

The `test-integration-android` job in `.github/workflows/test.yml` runs Android tests automatically:

- ✅ Runs on every push and pull request
- ✅ Uses Ubuntu with Android emulator (`reactivecircus/android-emulator-runner`)
- ✅ API level 33, x86_64 architecture, Google APIs
- ✅ Automatically builds the Android .aar library with gomobile
- ✅ Installs required dependencies (Java 17, Android SDK, KVM for hardware acceleration)
- ✅ Runs the same test suite through the Android FFI layer
- ✅ Verifies platform-specific integration (JNI bindings, Kotlin plugin wrapper)

### iOS Testing (Documented - Not Yet Active)

The `test-integration-ios` job structure is documented in `.github/workflows/test.yml` (commented out) but not active:

- 📝 Documented infrastructure ready for when iOS plugin is implemented
- 📝 Will use macOS runner with iOS simulator
- 📝 Will build .xcframework with gomobile
- 📝 Will run same test suite through iOS FFI layer

**Note**: The iOS job is commented out because iOS plugin implementation doesn't exist yet. Uncomment and activate when iOS support is added.

## Troubleshooting

### Test Failures

If tests fail, check:

1. **Go backend built correctly**: `ls -la ../../../binaries/`
2. **Environment variable set**: `echo $ANY_SYNC_GO_BINARIES_DIR`
3. **Database conflicts**: Tests should use `--test-threads=1`
4. **Logs**: Run with `RUST_LOG=debug` to see detailed logs

### Build Errors

If the test binary won't build:

1. **Check Rust version**: `rustc --version` (need 1.77.2+)
2. **Check dependencies**: `cargo tree | grep tauri-plugin-any-sync`
3. **Clean and rebuild**: `cargo clean && cargo build --tests`

### Sidecar Issues

If the Go sidecar doesn't start:

1. **Check binary exists**: `ls -la $ANY_SYNC_GO_BINARIES_DIR/`
2. **Check binary is executable**: `chmod +x $ANY_SYNC_GO_BINARIES_DIR/*`
3. **Test manually**: `$ANY_SYNC_GO_BINARIES_DIR/any-sync-* --help`

## Writing New Tests

When adding new plugin functionality:

1. **Add test for the new command** in `integration.rs`
2. **Follow the IPC testing pattern**:
   ```rust
   #[tokio::test]
   async fn test_your_new_command() {
       let (_app, webview, invoke_key) = create_test_app();

       let res = get_ipc_response(
           &webview,
           tauri::webview::InvokeRequest {
               cmd: "plugin:any-sync|your_command".into(),
               callback: tauri::ipc::CallbackFn(0),
               error: tauri::ipc::CallbackFn(1),
               body: json!({
                   "payload": {
                       "field": "value"
                   }
               }).into(),
               headers: Default::default(),
               url: "tauri://localhost".parse().unwrap(),
               invoke_key: invoke_key.clone(),
           },
       );

       assert!(res.is_ok(), "Command failed: {:?}", res);
       let response = res.unwrap().deserialize::<serde_json::Value>().unwrap();
       // Additional assertions on response...
   }
   ```

3. **Test edge cases**: Empty inputs, nonexistent data, error conditions
4. **Run tests locally** before committing: `task app:test-integration`
5. **Update this README** if adding new test categories
6. **Keep tests platform-agnostic** unless testing platform-specific behavior (use `#[cfg(desktop)]` / `#[cfg(mobile)]` when needed)

## Architecture Notes

### Test Infrastructure
- **MockRuntime**: Tauri's test runtime that doesn't require a window manager or display server
- **IPC Testing**: Uses `tauri::test::get_ipc_response()` to invoke commands through the actual IPC layer
- **Real Backend**: Tests use the actual compiled Go binaries, not mocks
- **Real Communication**: Full communication stack is tested (gRPC for desktop, JNI/FFI for mobile)
- **Real Database**: Uses SQLite, stored in system temp directory during tests
- **No Frontend**: No JavaScript/HTML/CSS needed, tests invoke commands directly via IPC

### Platform Coverage

| Platform | Backend Type | Communication | Test Status | Test Location |
|----------|-------------|---------------|-------------|---------------|
| macOS    | Sidecar     | gRPC          | ✅ Active   | Local + CI    |
| Linux    | Sidecar     | gRPC          | ✅ Active   | Local + CI    |
| Windows  | Sidecar     | gRPC          | ✅ Active   | Local + CI    |
| Android  | Embedded    | JNI (.aar)    | ✅ Active   | CI only       |
| iOS      | Embedded    | FFI (.xcframework) | 📝 Documented | Planned       |

**Legend:**
- ✅ Active: Tests run automatically in CI
- 📝 Documented: Infrastructure documented for future activation
- **Local + CI**: Can run tests locally and in CI
- **CI only**: Tests run in CI with Android emulator (local testing complex)

### Why This Approach?

This testing strategy provides confidence that the entire plugin stack works correctly:

```
Test Code
   ↓
IPC Layer (get_ipc_response)
   ↓
Tauri Command Handler
   ↓
Plugin Rust Core
   ↓
Platform Layer (Desktop Service / Mobile Service)
   ↓
Backend (gRPC Sidecar / gomobile Library)
   ↓
Storage Implementation (AnyStore)
   ↓
SQLite Database
```

**Benefits:**
- Tests the actual code path that production apps use
- Catches IPC serialization bugs
- Verifies platform-specific integration works
- No test-only code in plugin source
- Same tests work across all platforms (unified interface)

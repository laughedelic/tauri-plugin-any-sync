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

#### Prerequisites

1. **Android NDK**: Required for cross-compilation
   ```bash
   # Install via Android Studio SDK Manager or:
   # macOS: brew install android-ndk
   # Linux: Follow Android developer guide
   ```

2. **Rust Android target**:
   ```bash
   rustup target add aarch64-linux-android
   rustup target add armv7-linux-androideabi
   rustup target add i686-linux-android
   rustup target add x86_64-linux-android
   ```

3. **gomobile**: For building the Android .aar
   ```bash
   go install golang.org/x/mobile/cmd/gomobile@latest
   gomobile init
   ```

4. **Build the Android library**:
   ```bash
   task go:mobile:build  # or task backend:mobile:build
   ```

5. **Android emulator or device**: For running tests
   ```bash
   # Create emulator via Android Studio or:
   avdmanager create avd -n test_avd -k "system-images;android-33;google_apis;x86_64"
   
   # Start emulator
   emulator -avd test_avd -no-window -no-audio -no-boot-anim
   ```

#### Run Android Tests

```bash
# Build for Android and run tests on emulator/device
cargo test --test integration --target aarch64-linux-android --features integration-test -- --test-threads=1
```

**Note**: Android testing requires a fully configured Android development environment. The tests use the same code as desktop but execute through the Android FFI layer (.aar) instead of the gRPC sidecar.

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

### Mobile Testing (Planned)

Mobile CI jobs are documented but not yet active (waiting for full mobile platform support):

- **Android**: `test-integration-android` job will use Ubuntu + Android emulator
- **iOS**: `test-integration-ios` job will use macOS + iOS simulator (documented for when iOS support is added)

When enabled, mobile tests will:
- Build the appropriate mobile library (.aar for Android, .xcframework for iOS)
- Start an emulator/simulator
- Run the same test suite through the mobile FFI layer
- Verify platform-specific integration works correctly

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

| Platform | Backend Type | Communication | Test Status |
|----------|-------------|---------------|-------------|
| macOS    | Sidecar     | gRPC          | ✅ Active   |
| Linux    | Sidecar     | gRPC          | ✅ Active   |
| Windows  | Sidecar     | gRPC          | ✅ Active   |
| Android  | Embedded    | JNI (.aar)    | 📝 Ready    |
| iOS      | Embedded    | FFI (.xcframework) | 📝 Documented |

**Legend:**
- ✅ Active: Tests work locally and in CI
- 📝 Ready: Infrastructure ready, needs proper environment setup
- 📝 Documented: Infrastructure documented for future activation

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

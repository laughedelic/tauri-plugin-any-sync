# Integration Tests

This directory contains comprehensive integration tests for the tauri-plugin-any-sync.

## Overview

These integration tests verify end-to-end functionality of the plugin with the Go backend, **without requiring a GUI**. They use `tauri::test::MockRuntime` to create a headless app instance that simulates the webview and allows testing all plugin commands.

## Test Coverage

The integration tests cover:

- **Ping Command**: Basic health check and message echoing
- **Storage Put**: Creating and updating documents
- **Storage Get**: Retrieving documents, handling nonexistent documents
- **Storage Delete**: Deleting documents, handling nonexistent documents
- **Storage List**: Listing documents in collections, empty collections
- **Complex Scenarios**:
  - Multiple collections
  - Complex nested JSON documents
  - Unicode and special characters
  - Document updates
  - Concurrent operations

## Running the Tests

### Prerequisites

1. **Build the Go backend** for your platform:
   ```bash
   # From project root
   ./build-go-backend.sh
   ```

2. **Set environment variable** to use local binaries:
   ```bash
   export ANY_SYNC_GO_BINARIES_DIR=$(pwd)/binaries
   ```

### Run Tests

```bash
# From this directory (examples/tauri-app/src-tauri)
cargo test --test integration -- --test-threads=1

# With detailed logging
RUST_LOG=debug cargo test --test integration -- --test-threads=1 --nocapture

# Run a specific test
cargo test --test integration test_ping_command -- --test-threads=1
```

### Why `--test-threads=1`?

Tests are run sequentially to avoid conflicts:
- Each test creates its own app instance
- All instances share the same Go backend sidecar process
- The sidecar uses a single database file
- Running tests in parallel could cause race conditions in the database

## How It Works

1. **App Setup**: Each test calls `create_test_app()` which:
   - Uses the same `create_app_builder()` function as the production app
   - Creates a `MockRuntime` instead of spawning a real window
   - Initializes the plugin with all its dependencies

2. **Command Execution**: Tests call plugin methods directly via `app.any_sync()`:
   ```rust
   let result = app.any_sync().ping(payload).await;
   ```

3. **Backend Communication**:
   - First command automatically spawns the Go sidecar process
   - Sidecar runs as a real process, communicating via gRPC
   - All subsequent commands reuse the same sidecar instance

4. **Verification**: Tests assert on:
   - Success/failure of operations
   - Response data correctness
   - Side effects (e.g., document retrieval after put)

## CI Integration

The `test-integration` job in `.github/workflows/test.yml` runs these tests automatically:

- Runs on every push and pull request
- Tests on both Ubuntu and macOS
- Automatically builds the Go backend before running tests
- Fails the build if any test fails

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
2. **Follow the pattern**:
   ```rust
   #[tokio::test]
   async fn test_your_new_command() {
       let app = create_test_app();

       let payload = YourRequest { /* ... */ };
       let result = app.any_sync().your_command(payload).await;

       assert!(result.is_ok(), "Command failed: {:?}", result.err());
       // Additional assertions...
   }
   ```

3. **Test edge cases**: Empty inputs, nonexistent data, error conditions
4. **Run tests locally** before committing
5. **Update this README** if adding new test categories

## Architecture Notes

- **MockRuntime**: Tauri's test runtime that doesn't require a window manager
- **Real sidecar**: Tests use the actual Go binary, not a mock
- **Real gRPC**: Full communication stack is tested
- **Real database**: Uses SQLite, stored in system temp directory during tests
- **No frontend**: No JavaScript/HTML/CSS, tests call Rust directly

This approach provides confidence that the entire plugin stack works correctly,
from command invocation through gRPC to the Go backend and back.

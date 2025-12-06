# Rust Plugin Development Guide

## Architecture

**Single-dispatch pattern**: All operations route through one command handler using binary protobuf transport.

**Transport**:
- **Desktop**: Rust → gRPC client → Go sidecar
- **Mobile**: Rust → FFI calls → gomobile library
- Both use `tauri::ipc::Request`/`Response` for binary protobuf (no JSON serialization)

**Adding operations**: Requires only protobuf definition + Go handler (no Rust changes needed).

The plugin provides a thin passthrough layer - all business logic lives in Go handlers.

## Testing

```bash
# Run all tests
task rust:test

# Run specific test (pass args after --)
task rust:test -- test_name

# With logging
RUST_LOG=debug task rust:test
```

## Build Configuration

The `build.rs` script uses environment variables:
- `ANY_SYNC_GO_BINARIES_DIR` - Local binaries directory (development)
- `OUT_DIR` - Cargo build output directory

Platform-specific features determine which binaries are linked or downloaded.

See [root AGENTS.md](../AGENTS.md) for full development workflow.

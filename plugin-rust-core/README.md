# Rust Plugin Core

Tauri plugin implementation for Any-Sync. Provides unified API across desktop and mobile platforms.

## Structure

```
plugin-rust-core/
├── src/
│   ├── lib.rs          # Plugin registration
│   ├── commands.rs     # Tauri command handlers
│   ├── desktop.rs      # Desktop sidecar (gRPC)
│   ├── mobile.rs       # Mobile FFI bindings
│   ├── models.rs       # Data transfer types
│   ├── error.rs        # Error types
│   └── proto/          # Generated protobuf code
├── tests/              # Integration tests
├── permissions/        # Tauri permission system
├── android/            # Android native plugin
├── ios/                # iOS native plugin
└── build.rs            # Build script
```

## Development

```bash
# Build plugin
task rust:build

# Run tests
cd plugin-rust-core && cargo test

# Check and format
cargo clippy
cargo fmt
```

## Build Script

`build.rs` handles:
1. **Protobuf generation**: Compiles `.proto` files from `plugin-go-backend/desktop/proto/`
2. **Binary management**: Links Go binaries from `binaries/` (local dev) or downloads them from GitHub (production)

## Features

Platform-specific features control which binaries are downloaded:
- `all` - All platforms
- `desktop` - Desktop platforms (macOS, Linux, Windows)
- `mobile` - Mobile platforms (Android, iOS)
- Specific platforms: `macos`, `linux`, `windows`, `android`

See [root AGENTS.md](../AGENTS.md) for development workflow.

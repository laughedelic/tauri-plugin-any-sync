<!-- OPENSPEC:START -->
# OpenSpec Instructions

These instructions are for AI assistants working in this project.

Always open `@/openspec/AGENTS.md` when the request:
- Mentions planning or proposals (words like proposal, spec, change, plan)
- Introduces new capabilities, breaking changes, architecture shifts, or big performance/security work
- Sounds ambiguous and you need the authoritative spec before coding

Use `@/openspec/AGENTS.md` to learn:
- How to create and apply change proposals
- Spec format and conventions
- Project structure and guidelines

Keep this managed block so 'openspec update' can refresh the instructions.

<!-- OPENSPEC:END -->

## Project Structure

```
tauri-plugin-any-sync/
├── plugin-rust-core/      # Rust plugin core
│   ├── src/               # Plugin source code
│   ├── tests/             # Integration tests
│   ├── permissions/       # Tauri permission system
│   ├── android/           # Android native plugin
│   ├── ios/               # iOS native plugin
│   └── build.rs           # Build script (protobuf, binaries)
├── plugin-go-backend/     # Go backend implementation
│   ├── desktop/           # Desktop gRPC server + protobuf
│   ├── mobile/            # Mobile FFI bindings (gomobile)
│   └── shared/            # Shared storage logic
├── plugin-js-api/         # TypeScript API
│   └── src/index.ts       # Promise-based frontend API
├── example-app/           # Example Tauri application
└── binaries/              # Compiled Go binaries (desktop + mobile)
```

See component-specific documentation:
- [plugin-rust-core/AGENTS.md](plugin-rust-core/AGENTS.md)
- [plugin-go-backend/AGENTS.md](plugin-go-backend/AGENTS.md)
- [plugin-js-api/README.md](plugin-js-api/README.md)
- [example-app/README.md](example-app/README.md)

## Development Workflow

All build commands use `task` from the project root:

```bash
task build        # Build all components
task build-all    # Build for all platforms
task clean        # Clean build artifacts
task --list       # Show all available tasks
```

Component-specific tasks:
```bash
task go:build          # Go backend (current platform)
task go:build-all      # Go backend (all platforms)
task rust:build        # Rust plugin
task js:build          # TypeScript API
task app:dev           # Run example app
```

Use `task --list` to see all available tasks and their descriptions.

### Adding New API Operations

Follow these steps to add a new operation (e.g., `storageDelete`):

#### 1. Protocol Definition
- **File**: `plugin-go-backend/desktop/proto/{service}.proto`
- **Actions**: Add RPC method and messages

#### 2. Go Implementation
- **Files**: `plugin-go-backend/desktop/api/server/{service}.go`
- **Actions**: Implement gRPC handler and tests

#### 3. Rust Plugin
- **Files**: `plugin-rust-core/src/`, `plugin-rust-core/permissions/`
- Follow detailed instructions in `plugin-rust-core/AGENTS.md`:

#### 5. TypeScript API
- **File**: `plugin-js-api/src/index.ts`
- **Actions**: Add typed function with JSDoc

#### 6. Example App
- **File**: `example-app/src/App.svelte`
- **Actions**: Add UI and handler

#### 7. Build
```bash
task build
```

#### Common Pitfalls
- **Permission files**: All 3 must be updated and plugin rebuilt
- **Sidecar binary**: Example app uses old binary until Go backend rebuilt
- **Type alignment**: Ensure proto ↔ Rust ↔ TypeScript types match

## Build System Integration

### Binary Distribution Architecture

The plugin uses an automated binary distribution system with two distinct modes:

**Consumer/CI Mode (Production)**:
- Plugin downloads pre-compiled Go binaries from GitHub Releases
- Binaries are verified using SHA256 checksums
- Consumer's `build.rs` copies binaries to `src-tauri/binaries/`
- Enabled via Cargo features (e.g., `features = ["all"]` or `["macos"]`)

**Local Development Mode**:
- Set `ANY_SYNC_GO_BINARIES_DIR` environment variable to local binaries path
- Plugin copies binaries from local directory instead of downloading
- Allows developers to test Go backend changes immediately
- No network dependency for development workflows

### Build Flow

**Plugin Build** (`build.rs`):
1. Check for `ANY_SYNC_GO_BINARIES_DIR` environment variable
2. **If set** (development mode):
   - Copy binaries from local path to `OUT_DIR/binaries/`
   - Emit warning message
3. **If not set** (consumer/CI mode):
   - Determine enabled features (e.g., `macos`, `windows`)
   - Download binaries from GitHub Releases for plugin version
   - Download and parse `checksums.txt` from release assets
   - Verify SHA256 checksums for each binary (desktop + mobile)
   - Store verified binaries in `OUT_DIR/binaries/`
   - Fail build with clear error if download or verification fails
4. **For Android**: Symlink (Unix) or copy (Windows) `any-sync-android.aar` from binaries to `android/libs/`
5. Emit `cargo:binaries_dir=<path>` for consumer propagation (both modes)

**Note**: In development mode, symlinks are used instead of copying to save disk space and improve build times. On Windows, files are copied as a fallback since symlinks require admin privileges.

**Android .aar Placement**: The plugin's build.rs automatically manages the .aar placement by creating a symlink/copy to `android/libs/any-sync-android.aar`. This allows the plugin's `android/build.gradle.kts` to reference `implementation(files("libs/any-sync-android.aar"))` relative to its own directory, which works in both development (local path) and production (published crate) scenarios.

**Consumer Build** (`build.rs` in consuming app):
1. Read `DEP_TAURI_PLUGIN_ANY_SYNC_BINARIES_DIR` from plugin
2. Symlink binaries to `src-tauri/binaries/` (copy on Windows)
3. Configure `externalBin` in `tauri.conf.json`
4. Add `.taurignore` to prevent rebuild loops

### Cargo Configuration

**Features** (select which platforms to download):
- `all`
  - `desktop`
    - `macos`
      - `x86_64-apple-darwin`
      - `aarch64-apple-darwin`
    - `linux`
      - `x86_64-unknown-linux-gnu`
      - `aarch64-unknown-linux-gnu`
    - `windows`
      - `x86_64-pc-windows-msvc`
  - `mobile`
    - `android`

**Links** (`links = "tauri-plugin-any-sync"`):
- Enables metadata propagation via environment variables
- Allows consumer `build.rs` to receive `DEP_TAURI_PLUGIN_ANY_SYNC_BINARIES_DIR`

## Communication Flow

### Desktop (gRPC Sidecar)
```
TypeScript → Rust Commands → Desktop Service → gRPC Client → Go Sidecar → AnyStore
```
- Go backend runs as separate process (bundled binary)
- IPC via gRPC over localhost

### Mobile (gomobile binaries Embedded)
```
TypeScript → Rust Commands → Mobile Service → Kotlin/Swift Plugin → JNI/FFI → Go Library → AnyStore
```
- Go backend compiled as native library (.aar/.xcframework)
- Direct in-process function calls via gomobile

**Shared:** Same TypeScript API, same Go storage layer (>95% code reuse)

## Integration Tests

Integration tests verify end-to-end functionality of the plugin with the Go backend, without requiring a GUI. They use `tauri::test` to create a headless app instance.

**Location**: `example-app/src-tauri/tests/integration.rs`

**What is tested**:
- **Process Management**: Automatic sidecar startup when commands are invoked
- **gRPC Communication**: All commands (ping, storage_put, storage_get, storage_delete, storage_list)
- **Error Handling**: Proper error propagation across all layers
- **Data Integrity**: Complex JSON documents, multiple collections, updates, deletes
- **Edge Cases**: Empty collections, nonexistent documents, concurrent operations

**Running integration tests**:

```bash
# Run integration tests from the example app
cd examples/tauri-app/src-tauri
cargo test --test integration -- --test-threads=1

# With detailed logging
RUST_LOG=debug cargo test --test integration -- --test-threads=1 --nocapture
```

**Important notes**:
- Tests run with `--test-threads=1` to avoid database conflicts (each test uses the same sidecar instance)
- The Go backend binary must be built before running tests
- Tests are platform-specific (each platform needs its own binary)
- The `create_app_builder()` function in `lib.rs` ensures tests use the same configuration as production

**CI Integration**:
- The `test-integration` job in `.github/workflows/test.yml` runs these tests on every push/PR
- Runs on Ubuntu (Linux-only for faster CI execution)
- Automatically builds the Go backend before running tests
- Uses `ANY_SYNC_GO_BINARIES_DIR` to point to local binaries

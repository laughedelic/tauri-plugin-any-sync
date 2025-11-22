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

## Architecture Components

### Go Backend (`go-backend/`)
- **Purpose**: gRPC server providing health check and ping services
- **Technology**: Go with gRPC and Protocol Buffers
- **Entry Point**: `cmd/server/main.go`
- **Build Script**: `build-go-backend.sh` (cross-platform compilation)
- **Testing**: Unit tests in `api/server/health_test.go`

### Rust Plugin (`src/`)
- **Purpose**: Tauri plugin that manages Go sidecar process and provides TypeScript API
- **Key Files**:
  - `desktop.rs` - Sidecar process management and gRPC client
  - `commands.rs` - Tauri command handlers
  - `models.rs` - Data transfer types
  - `proto/` - Generated gRPC client code
- **Dependencies**: tokio, tonic, prost, tempfile, uuid

### TypeScript API (`guest-js/`)
- **Purpose**: Frontend API with Promise-based interface
- **Key Function**: `ping(message?: string): Promise<string | null>`
- **Error Handling**: Structured error propagation from Go backend

### Example App (`examples/tauri-app/`)
- **Purpose**: Demonstration of plugin functionality
- **UI**: Svelte frontend with ping test button
- **Configuration**: Plugin capabilities in `tauri.conf.json`

## Development Workflow

### 1. Go Backend Development
```bash
cd go-backend
# Edit proto files

# Generate code
protoc --go_out=. --go-grpc_out=. api/proto/health.proto

# Run tests
go test ./... -v

# Build
./build-go-backend.sh --cross
```

### 2. Rust Plugin Development
```bash
# Build plugin (includes Go backend compilation)
cargo build

# Test
cargo test

# Format code
cargo fmt

# Check code
cargo clippy
```

### 3. TypeScript API Development
```bash
# Edit API

# Build types
bun run build

# Test with example app
cd examples/tauri-app && bun run tauri dev
```

### 4. End-to-End Testing
```bash
# Build everything
./build-go-backend.sh
cargo build

# Test example app
cd examples/tauri-app
bun run tauri dev

# Verify ping functionality
# Click ping button in UI
# Check browser console for response
# Verify Go backend logs
```

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
   - Download matching binaries from GitHub Releases for plugin version
   - Download and parse `checksums.txt` from release assets
   - Verify SHA256 checksums for each binary
   - Store verified binaries in `OUT_DIR/binaries/`
   - Fail build with clear error if download or verification fails
4. Emit `cargo:binaries_dir=<path>` for consumer propagation (both modes)

**Consumer Build** (`build.rs` in consuming app):
1. Read `DEP_ANY_SYNC_GO_BINARIES_DIR` environment variable from plugin
2. Copy binaries to `src-tauri/binaries/` directory
3. Configure `externalBin` in `tauri.conf.json` to bundle binaries

### Cargo Configuration

**Features** (select which platforms to download):
- Individual targets: `x86_64-apple-darwin`, `aarch64-apple-darwin`, `x86_64-unknown-linux-gnu`, `aarch64-unknown-linux-gnu`, `x86_64-pc-windows-msvc`
- Platform groups: `macos`, `linux`, `windows`
- All: `all`

**Links** (`links = "any_sync_go"`):
- Enables metadata propagation via environment variables
- Allows consumer `build.rs` to receive `DEP_ANY_SYNC_GO_BINARIES_DIR`

### Development Workflow

1. **Edit Go backend code**:
   ```bash
   # Edit files in go-backend/
   vim go-backend/cmd/server/main.go
   ```

2. **Build binaries locally**:
   ```bash
   ./build-go-backend.sh
   # or for all platforms:
   ./build-go-backend.sh --cross
   ```

3. **Set environment variable**:
   ```bash
   # One-off
   export ANY_SYNC_GO_BINARIES_DIR=./binaries
   
   # Or persistent in .cargo/config.toml:
   [env]
   ANY_SYNC_GO_BINARIES_DIR = { value = "/path/to/binaries", force = true }
   ```

4. **Build plugin**:
   ```bash
   cargo build
   # Plugin will use local binaries instead of downloading
   ```

### Release Workflow

1. **Bump version** in `Cargo.toml`
2. **Create Git tag**: `git tag v0.2.0`
3. **Push tag**: `git push origin v0.2.0`
4. **GitHub Actions triggers** `.github/workflows/release.yml`:
   - Builds Go binaries for all platforms (`./build-go-backend.sh --cross`)
   - Generates SHA256 checksums
   - Creates GitHub Release
   - Uploads binaries + `checksums.txt`
5. **Consumers update** dependency version
6. **Plugin downloads** matching binaries on next `cargo build`

### Automated Build Process

1. **Rust Build**: `cargo build` triggers `build.rs`
2. **Plugin build.rs**: Downloads or uses local binaries (depending on env var)
3. **Protobuf**: Both Go and Rust code generated from same `.proto` file
4. **Binaries**: Output to `OUT_DIR/binaries/` (not committed)
5. **Consumer build.rs**: Copies to `src-tauri/binaries/` via metadata propagation

### Cross-Platform Support

**Available Platforms**:
- **macOS**: `server-x86_64-apple-darwin` (Intel), `server-aarch64-apple-darwin` (Apple Silicon)
- **Linux**: `server-x86_64-unknown-linux-gnu` (x64), `server-aarch64-unknown-linux-gnu` (ARM64)
- **Windows**: `server-x86_64-pc-windows-msvc` (x64)

## Communication Flow

```
TypeScript UI → Tauri Commands → Rust Plugin → gRPC Client → Go Backend → gRPC Server → Response → UI
```

### Data Flow
1. UI calls `ping("test message")` in TypeScript
2. Tauri invokes Rust `ping` command
3. Rust spawns Go sidecar if not running
4. Rust sends gRPC `PingRequest` to Go backend
5. Go backend processes and returns `PingResponse`
6. Rust converts response and returns to TypeScript
7. UI receives Promise with echoed message

## Tooling Requirements

### Required Tools
- **Go**: 1.21+ (for backend)
- **Rust**: 1.77+ (for plugin)
- **protoc**: Protocol Buffer compiler
- **Node.js**: For TypeScript compilation and example app
- **Tauri CLI**: For app development

### Development Dependencies
```bash
# Go tools
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest

# Rust tools (installed via cargo)
cargo install cargo-watch
cargo install cargo-audit
```

## Testing Strategy

### Unit Tests
- **Go**: `go test ./...` - Tests gRPC services
- **Rust**: `cargo test` - Tests plugin logic
- **Coverage**: Aim for >80% code coverage

### Integration Tests
- **Process Management**: Sidecar startup/shutdown
- **gRPC Communication**: End-to-end message passing
- **Error Handling**: Proper error propagation
- **Resource Cleanup**: Memory and process cleanup

### Manual Testing Checklist
- [ ] Go backend starts and listens on random port
- [ ] Rust plugin spawns sidecar process
- [ ] gRPC connection established successfully
- [ ] Ping request flows through all layers
- [ ] Response returns to UI correctly
- [ ] Sidecar process shuts down gracefully
- [ ] Error handling works across boundaries

## Performance Considerations

### Phase 0 Optimizations
- **Startup Time**: <2 seconds for sidecar spawn
- **Request Latency**: <50ms for simple ping
- **Memory Usage**: <10MB for idle sidecar
- **Binary Size**: <15MB per platform binary

### Monitoring
- **Health Checks**: Every 30 seconds
- **Connection Pooling**: Handled by gRPC library
- **Resource Limits**: Configurable via environment variables

## Security Architecture

### Phase 0 Security
- **Process Isolation**: Sidecar process separates concerns
- **Localhost Only**: Server binds to localhost by default
- **Input Validation**: All gRPC inputs validated
- **No Authentication**: Basic functionality only (Phase 1+)

### Future Security (Phase 1+)
- **Mutual TLS**: Encrypted gRPC communication
- **Authentication**: User authentication for backend
- **Authorization**: Permission-based access control
- **Audit Logging**: Security event logging

## Troubleshooting

### Common Issues
1. **Port Conflicts**: Use random port allocation (port 0)
2. **Build Failures**: Check Go toolchain and protoc installation
3. **gRPC Timeouts**: Increase timeout values in configuration
4. **Process Leaks**: Verify graceful shutdown implementation
5. **Cross-Compilation**: Ensure target-specific toolchains

### Debug Commands
```bash
# Enable verbose logging
export ANY_SYNC_LOG_LEVEL=debug

# Check Go backend logs
./binaries/server --port 8080

# Check Rust plugin logs
RUST_LOG=debug cargo run

# Test gRPC directly
grpcurl -plaintext localhost:8080 anysync.HealthService/Ping
```

## Phase 0 Success Criteria

✅ **Completed**:
- Go backend compiles and runs as standalone server
- Desktop sidecar process spawns and communicates via gRPC
- TypeScript `ping` command round-trips through all layers
- Example app successfully calls plugin and displays response
- Build process produces all necessary artifacts
- Basic error handling works across all boundaries

🔄 **Ready for Phase 1**:
- AnySync/AnyStore integration
- Mobile gomobile binding structure
- Advanced gRPC streaming
- Production-ready error handling
- Comprehensive testing

# Go Backend: Modular Architecture

The Go backend is structured as **three independent modules** sharing minimal core code, enabling clean separation between desktop (gRPC server) and mobile (gomobile bindings) implementations.

## Directory Structure

```
go-backend/
├── go.mod, go.sum, go.work          # Workspace coordination
├── shared/                           # Core module: anysync-backend
│   ├── go.mod
│   └── storage/                      # Shared storage API
├── desktop/                          # Desktop module: anysync-backend/desktop
│   ├── go.mod
│   ├── main.go                       # gRPC server entry point
│   ├── api/proto/                    # Protobuf definitions
│   ├── api/server/                   # gRPC service implementations
│   ├── config/                       # Server configuration
│   └── health/                       # Health check service
└── mobile/                           # Mobile module: anysync-backend/mobile
    ├── go.mod
    ├── main.go                       # Gomobile-exported API
    ├── storage.go                    # Mobile-friendly storage wrapper
    └── tools.go                      # Keeps golang.org/x/mobile in go.mod
```

## Module Breakdown

### Shared Module (`anysync-backend`)
- **Location**: `./shared/`
- **Dependencies**: None (zero external deps)
- **Exports**: `storage/` — Core storage abstractions and types
- **Why separate**: Provides clean core that can be embedded in both desktop and mobile without pulling platform-specific dependencies

### Desktop Module (`anysync-backend/desktop`)
- **Location**: `./desktop/`
- **Dependencies**: 
  - `anysync-backend` (shared module via `replace anysync-backend => ../shared`)
  - `google.golang.org/grpc` (gRPC framework)
  - `google.golang.org/protobuf` (Protobuf runtime)
- **Purpose**: gRPC server for desktop sidecar communication
- **Includes**: Server config, health checks, proto definitions — all gRPC-specific
- **Builds**: Cross-platform binaries (macOS x86_64/ARM64, Linux x86_64/ARM64, Windows)

### Mobile Module (`anysync-backend/mobile`)
- **Location**: `./mobile/`
- **Dependencies**:
  - `anysync-backend` (shared module via `replace anysync-backend => ../shared`)
  - `golang.org/x/mobile` (gomobile tools)
- **Purpose**: Android/iOS native bindings via gomobile
- **Includes**: Exported functions for JNI/FFI calls, mobile-friendly wrapper
- **Builds**: Android `.aar` (all architectures); iOS `.xcframework` can be added similarly

## Key Design Decisions

### Why Three Modules?

1. **Desktop doesn't need gomobile** — Keeps gRPC server lean, no unnecessary mobile build overhead
2. **Mobile doesn't need gRPC** — In-process function calls, not over network
3. **Shared provides clean core** — Both platforms import same storage API, avoiding duplication

### Why Not Everything in `internal/`?

In a single-module monorepo, internal packages would enforce boundaries. However, three separate modules are more idiomatic for:
- **Independent versioning** if modules are published separately
- **Clear public contracts** — Each module declares what it depends on
- **Better IDE support** — Module boundaries are explicit to tooling

### The `v0.0.0 + replace` Pattern

Both desktop and mobile declare:
```go
require anysync-backend v0.0.0
replace anysync-backend => ../shared
```

This is **idiomatic Go** for monorepos:
- `v0.0.0` is required by Go module syntax (dummy version)
- `replace` tells Go to use local directory instead of registry
- Makes changes atomic — all modules always use same shared code version

## Dependency Isolation

```
shared (zero deps)
  ↑
  ├─→ desktop (adds: grpc, protobuf)
  └─→ mobile (adds: golang.org/x/mobile)
```

Each module only pulls what it needs:
- Shared stays minimal and publishable
- Desktop gets full gRPC stack for server
- Mobile gets gomobile for native bindings
- No circular dependencies possible

## Building

```bash
# Desktop binaries (current platform)
./build-go-backend.sh

# Desktop binaries (all platforms) — requires cross-compilation toolchains
./build-go-backend.sh --cross

# Android .aar
./build-go-mobile.sh

# Individual module builds
cd shared && go build ./...
cd desktop && go build ./...
cd mobile && go build ./...
```

## Development Workflow

### Using go.work

The `go.work` file enables seamless development across all three modules:

```
go 1.25
use (
    ./shared
    ./desktop
    ./mobile
)
```

Benefits:
- Run `go test ./...` at project root to test all modules
- IDE treats it as unified project with proper cross-module navigation
- `go mod tidy` keeps all modules synchronized
- Still maintains proper module boundaries (can't import internal packages across modules)

### Dependency Management

Always use `go get` and `go mod tidy`:
```bash
# Add a dependency (updated by Go tools automatically)
go get github.com/anyproto/any-store@latest

# Clean and sync all modules
go mod tidy -C shared && go mod tidy -C desktop && go mod tidy -C mobile
```

### Running Tests

```bash
# Test all modules
cd /path/to/go-backend && go test ./...

# Test specific module
cd desktop && go test ./...
cd mobile && go test ./...

# With coverage
go test ./... -cover
```

## Environment Variables

- `ANY_SYNC_HOST` — Server bind address (default: localhost)
- `ANY_SYNC_PORT` — Server port (default: 0 for random)
- `ANY_SYNC_LOG_LEVEL` — Logging level (default: info)
- `ANY_SYNC_LOG_FORMAT` — Log format (default: json)
- `ANY_SYNC_HEALTH_CHECK_INTERVAL` — Health check interval seconds (default: 30)
- `ANY_SYNC_PORT_FILE` — File to write port number for parent process

## Adding a New Platform

To add iOS, WebAssembly, or other platform:

1. Create `platform/go.mod` and `platform/main.go`
2. Declare: `require anysync-backend v0.0.0` + `replace anysync-backend => ../shared`
3. Import `anysync-backend/storage` from shared
4. Implement platform-specific bindings
5. Add to `go.work`

Same pattern, independent module boundaries, zero code duplication.

## Troubleshooting

**Import error: "module anysync-backend@latest found but does not contain package"**
- Check `replace` directive in go.mod points to correct path (`../shared`)
- Run `go mod tidy` in that module

**gomobile build fails with missing dependency**
- Run `go mod tidy -C mobile`
- Ensure `golang.org/x/mobile` is in mobile/go.mod requires

**Tests fail in IDE but pass in terminal**
- IDE may not be using go.work — restart it or manually configure workspace support

**Port already in use**
- Use `--port 0` for random port allocation (default)
- Check if previous server instance is still running: `lsof -i :8080`

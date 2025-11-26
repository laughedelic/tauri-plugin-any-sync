# Go Backend: Modular Architecture

The Go backend is structured as **three independent modules** sharing minimal core code, enabling clean separation between desktop (gRPC server) and mobile (gomobile bindings) implementations.

## Directory Structure

```
go-backend/
├── go.work                           # Workspace coordination
├── Taskfile.yml                      # Build task definitions
├── shared/                           # Core module: anysync-backend
│   ├── go.mod
│   └── storage/                      # Shared storage API
├── desktop/                          # Desktop module: anysync-backend/desktop
│   ├── go.mod
│   ├── main.go                       # gRPC server entry point
│   ├── proto/                        # Protobuf definitions + generated code
│   │   ├── doc.go                    # go:generate directive
│   │   ├── health.proto
│   │   ├── storage.proto
│   │   ├── *.pb.go                   # Generated message code
│   │   └── *_grpc.pb.go              # Generated service code
│   ├── api/server/                   # gRPC service implementations
│   ├── config/                       # Server configuration
│   └── health/                       # Health check service
└── mobile/                           # Mobile module: anysync-backend/mobile
    ├── go.mod
    ├── main.go                       # Gomobile-exported API
    ├── storage.go                    # Mobile-friendly storage wrapper
    └── tools.go                      # Keeps golang.org/x/mobile in go.mod
```

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

Build tasks are orchestrated through **Taskfile**, which provides a unified interface for Go builds and handles tool checking:

```bash
# From project root OR go-backend directory:
task backend:build          # Desktop binaries (current platform)
task backend:mobile         # Android .aar
task backend:test           # Run tests

# See all available tasks
task --list
```

**What Taskfile does:**
- Checks that required tools exist (`go`, `protoc`, `gomobile`, etc.)
- Invokes build scripts in `scripts/` with proper paths
- Provides consistent interface across different platforms

**Individual script execution** (if needed):
- Desktop: `go-backend/scripts/build-desktop.sh [--cross]`
- Mobile: `go-backend/scripts/build-mobile.sh`

Both scripts automatically trigger protobuf code generation via `go generate`.

### Protobuf Code Generation

Proto files in `desktop/proto/` are automatically generated via `go:generate` directives in `doc.go`:

```bash
# Automatic (via build scripts or go generate)
go generate ./proto

# Manual (equivalent command)
protoc --go_out=. --go-grpc_out=. \
  --go_opt=paths=source_relative --go-grpc_opt=paths=source_relative \
  proto/health.proto proto/storage.proto
```

The `paths=source_relative` option ensures `.pb.go` files land next to `.proto` sources, avoiding nested directory structures.

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

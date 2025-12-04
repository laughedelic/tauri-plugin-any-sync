# Go Backend Development Guide

## Quick Start

```bash
# Build desktop binaries (current platform)
task go:desktop:build

# Build Android .aar
task go:mobile:build

# Build for all platforms
task go:build-all
```

All tasks use Taskfile from project root.

## Module Architecture

```
plugin-go-backend/
├── shared/                    # Shared logic
│   ├── dispatcher/            # Command routing
│   ├── handlers/              # Operation implementations
│   └── anysync/               # Any-Sync integration
├── desktop/                   # gRPC server (TransportService)
│   ├── proto/transport/v1/    # Transport protocol (4 methods)
│   └── main.go                # Server + dispatcher integration
└── mobile/                    # gomobile FFI bindings
    └── main.go                # 4-function API (Init, Command, Subscribe, Shutdown)
```

**Transport Layer**:
- Desktop: gRPC `TransportService` with `Command(cmd, protobuf_bytes)` RPC
- Mobile: Direct Go exports via gomobile FFI
- Both use same dispatcher routing to `shared/handlers/`

**Module dependencies:**
- `shared`: Common interface with Any-Sync/Any-Store dependencies
- `desktop`: Adds grpc + protobuf
- `mobile`: Adds golang.org/x/mobile

## Dependencies

### Managing Dependencies

**Always use native Go tools to manage dependencies:**

```bash
# Add a new dependency (automatically updates go.mod and go.sum)
go get github.com/anyproto/any-store@latest

# Remove unused dependencies and download missing ones
go mod tidy

# Verify dependencies
go mod verify
```

**Never manually edit go.mod or go.sum** - let Go tooling handle version resolution and checksums.

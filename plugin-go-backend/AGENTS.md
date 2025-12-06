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

## Handler Implementation Pattern

Handlers in `shared/handlers/` receive and return binary protobuf:

1. **Unmarshal request** from protobuf bytes
2. **Interact with Any-Sync** components (SpaceManager, ObjectTree, EventManager)
3. **Marshal response** to protobuf bytes

Example:
```go
func CreateDocument(req []byte) ([]byte, error) {
    var request pb.CreateDocumentRequest
    if err := proto.Unmarshal(req, &request); err != nil {
        return nil, err
    }

    // Use Any-Sync ObjectTree for document storage
    documentId, err := deps.ObjectTree.CreateDocument(...)
    if err != nil {
        return nil, err
    }

    response := &pb.CreateDocumentResponse{DocumentId: documentId}
    return proto.Marshal(response)
}
```

**Register in `handlers/registry.go`**:
```go
func init() {
    dispatcher.Register("CreateDocument", CreateDocument)
}
```

**Local-first vs Network**:
- Current handlers operate on local Any-Sync structures
- Network sync (coordinator, peers, JoinSpace/LeaveSpace) deferred to future work
- All data uses cryptographic keys and sync-ready structures

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

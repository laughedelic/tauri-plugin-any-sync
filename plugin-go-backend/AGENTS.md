# Go Backend Development Guide

## Quick Start

```bash
# Install build dependencies
task backend:install-deps

# Build desktop binaries (current platform)
task backend:build

# Build for all platforms
task backend:build-all

# Build Android .aar
task backend:mobile

# Run tests
task backend:test

# See all available tasks
task --list
```

All build tasks are orchestrated through Taskfile. See `Taskfile.yml` in `go-backend/` and project root for details.

## Module Architecture

The backend is three independent Go modules sharing core storage code:

```
go-backend/
├── shared/                    # anysync-backend (storage API)
│   └── storage/
├── desktop/                   # anysync-backend/desktop (gRPC server)
│   ├── main.go
│   ├── proto/                 # Protocol Buffer definitions + generated code
│   │   ├── doc.go            # go:generate directive for protoc
│   │   ├── health.proto
│   │   ├── storage.proto
│   │   ├── *.pb.go           # Generated gRPC server code
│   │   └── *_grpc.pb.go      # Generated gRPC client/server code
│   ├── api/server/           # gRPC service implementations
│   ├── config/
│   └── health/
└── mobile/                    # anysync-backend/mobile (gomobile bindings)
    ├── main.go
    └── storage.go
```

**Module dependencies:**
- `shared`: Zero external dependencies (just any-store)
- `desktop`: Adds grpc + protobuf
- `mobile`: Adds golang.org/x/mobile

**Development:** Use `go.work` for seamless workspace development across modules. Rebuilt modules use `replace` directives pointing to local paths.

## Proto Generation with go:generate

Proto files are in `desktop/proto/` and automatically generated via `go:generate`:

```bash
# Auto-generate via go:generate (used by build scripts)
go generate ./proto

# Manual protoc invocation (equivalent)
protoc --go_out=. --go-grpc_out=. \
  --go_opt=paths=source_relative --go-grpc_opt=paths=source_relative \
  proto/health.proto proto/storage.proto
```

The `paths=source_relative` option ensures generated `.pb.go` files land next to their `.proto` sources instead of creating nested directories.

## Development Workflow

### 1. Protocol Buffer Development

Protocol Buffer definitions are in `api/proto/health.proto`.

```bash
# Generate Go code from protobuf
cd go-backend
protoc --go_out=. --go-grpc_out=. api/proto/health.proto

# Generate Rust code from protobuf
protoc --rust_out=src/proto --rust-grpc_out=src/proto api/proto/health.proto
```

### 2. gRPC Service Implementation

Services are implemented in `api/server/health.go`:

```go
// HealthServer implements the HealthService gRPC service
type HealthServer struct {
    pb.UnimplementedHealthServiceServer
    serverID string
    startTime time.Time
}

// NewHealthServer creates a new HealthServer instance
func NewHealthServer() *HealthServer {
    return &HealthServer{
        serverID: fmt.Sprintf("any-sync-%d", time.Now().Unix()),
        startTime: time.Now(),
    }
}

// Check implements the health check RPC
func (s *HealthServer) Check(ctx context.Context, req *pb.HealthCheckRequest) (*pb.HealthCheckResponse, error) {
    return &pb.HealthCheckResponse{
        Status:  pb.HealthCheckResponse_SERVING,
        Message: fmt.Sprintf("Server %s is running", s.serverID),
    }, nil
}

// Ping implements the ping RPC for round-trip testing
func (s *HealthServer) Ping(ctx context.Context, req *pb.PingRequest) (*pb.PingResponse, error) {
    if req == nil {
        return nil, status.Error(codes.InvalidArgument, "ping request cannot be nil")
    }

    response := &pb.PingResponse{
        Message:          fmt.Sprintf("Echo: %s", req.GetMessage()),
        RequestTimestamp: req.GetTimestamp(),
        ResponseTimestamp: time.Now().Unix(),
        ServerId:         s.serverID,
    }

    return response, nil
}
```

### 3. Configuration Management

Configuration is handled in `internal/config/config.go`:

```go
type Config struct {
    Host                string
    Port                int
    LogLevel           string
    LogFormat          string
    HealthCheckInterval int
}

// NewConfig creates a new Config instance with defaults and environment overrides
func NewConfig() *Config {
    config := &Config{
        Host:                "localhost",
        Port:                0, // 0 means random port
        LogLevel:            "info",
        LogFormat:          "json",
        HealthCheckInterval: 30,
    }
    
    // Override with environment variables if present
    if host := os.Getenv("ANY_SYNC_HOST"); host != "" {
        config.Host = host
    }
    
    if portStr := os.Getenv("ANY_SYNC_PORT"); portStr != "" {
        if port, err := strconv.Atoi(portStr); err == nil {
            config.Port = port
        }
    }
    
    return config
}
```

### 4. Server Entry Point

The main server is in `cmd/server/main.go`:

```go
func main() {
    flag.Parse()
    
    // Load configuration
    cfg := config.NewConfig()
    
    // Override with command line flags
    if *port != 0 {
        cfg.Port = *port
    }
    if *host != "localhost" {
        cfg.Host = *host
    }

    // Create gRPC server
    grpcServer := grpc.NewServer()
    healthServer := server.NewHealthServer()
    pb.RegisterHealthServiceServer(grpcServer, healthServer)

    // Create listener
    lis, err := net.Listen("tcp", cfg.GetAddress())
    if err != nil {
        log.Fatalf("Failed to listen: %v", err)
    }

    // Start serving
    log.Printf("Server listening on %s", cfg.GetAddress())
    
    // Write port to file for parent process communication
    if portFile := os.Getenv("ANY_SYNC_PORT_FILE"); portFile != "" {
        actualPort := lis.Addr().(*net.TCPAddr).Port
        if err := os.WriteFile(portFile, []byte(fmt.Sprintf("%d", actualPort)), 0644); err != nil {
            log.Printf("Warning: Failed to write port file: %v", err)
        }
    }

    if err := grpcServer.Serve(lis); err != nil {
        log.Fatalf("Failed to serve: %v", err)
    }
}
```

## Build System

### Cross-Platform Build Script

The `build-go-backend.sh` script handles cross-compilation:

```bash
# Build for current platform only
./build-go-backend.sh

# Build for all platforms
./build-go-backend.sh --cross

# Supported platforms:
# - darwin-amd64
# - darwin-arm64  
# - linux-amd64
# - linux-arm64
# - windows-amd64
```

### Build Integration with Rust

The `build.rs` file integrates Go compilation with the Rust build process:

- Verifies Go toolchain availability
- Runs Go build script during Rust compilation
- Places binaries in `binaries/` directory
- Emits cargo metadata for plugin integration

## Testing

### Unit Tests

Run unit tests for gRPC services:

```bash
cd go-backend
go test ./api/server -v
go test ./internal/health -v
go test ./internal/config -v
```

### Integration Testing

Test server startup and communication:

```bash
# Start server
./binaries/any-sync-aarch64-apple-darwin --port 8080 &

# Test with grpcurl (requires grpcurl installation)
grpcurl -plaintext -d '{"message":"test"} localhost:8080 anysync.HealthService/Ping

# Or use Go test client
go run test_client.go
```

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

### Core Dependencies

- `google.golang.org/grpc` - gRPC framework
- `google.golang.org/protobuf` - Protocol Buffer support
- `github.com/google/uuid` - UUID generation

### Development Dependencies

- `google.golang.org/grpc/codes` - gRPC status codes
- `google.golang.org/grpc/status` - gRPC error handling
- `google.golang.org/protobuf/types/known/timestamppb` - Timestamp support

## Environment Variables

- `ANY_SYNC_HOST` - Server bind address (default: localhost)
- `ANY_SYNC_PORT` - Server port (default: 0 for random)
- `ANY_SYNC_LOG_LEVEL` - Logging level (default: info)
- `ANY_SYNC_LOG_FORMAT` - Log format (default: json)
- `ANY_SYNC_HEALTH_CHECK_INTERVAL` - Health check interval in seconds (default: 30)
- `ANY_SYNC_PORT_FILE` - File to write port number for parent process

## Debugging

### Common Issues

1. **Port already in use**: Server will exit with error. Use `--port 0` for random port allocation.
2. **Protobuf generation errors**: Ensure `protoc` and Go plugins are in PATH.
3. **Cross-compilation failures**: Check target-specific Go installation and CGO requirements.

### Debug Logging

Enable debug logging:

```bash
export ANY_SYNC_LOG_LEVEL=debug
./binaries/any-sync-aarch64-apple-darwin
```

### Health Check Endpoint

Test server health:

```bash
# Health check (empty request)
grpcurl -plaintext localhost:8080 anysync.HealthService/Check

# Ping with message
grpcurl -plaintext -d '{"message":"hello","timestamp":1234567890}' localhost:8080 anysync.HealthService/Ping
```

## Performance Considerations

- Server uses lightweight gRPC with minimal overhead
- Connection pooling is handled by gRPC library
- Health checks run on configurable intervals
- Graceful shutdown ensures clean resource cleanup

## Security Notes

- Server binds to localhost by default for security
- No authentication in Phase 0 (will be added in later phases)
- Process isolation through sidecar architecture
- Input validation on all gRPC endpoints
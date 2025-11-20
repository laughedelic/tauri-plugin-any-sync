# Example App Testing Guide

This guide covers testing the any-sync plugin with the example Tauri application.

## Quick Start

```bash
cd examples/tauri-app
bun install
bun run tauri dev
```

## Testing Workflow

### 1. Plugin Integration Testing

The example app demonstrates the complete any-sync plugin integration:

```svelte
<script>
  import { ping } from 'tauri-plugin-any-sync-api'
  
  function _ping() {
    ping("Pong!").then(updateResponse).catch(updateResponse)
  }
</script>

<button on:click="{_ping}">Ping</button>
<div>{@html response}</div>
```

### 2. End-to-End Testing

Test the complete communication flow:

1. **UI Layer**: Svelte component calls `ping()` function
2. **TypeScript Layer**: API function invokes Tauri command
3. **Rust Layer**: Plugin spawns Go sidecar and routes gRPC call
4. **Go Backend**: gRPC server processes ping and returns response
5. **Return Path**: Response flows back through all layers to UI

### 3. Expected Behavior

When you click the "Ping" button:

1. **First Click**: 
   - Sidecar process spawns (takes ~1-2 seconds)
   - "Starting server..." message appears
   - Ping request sent to Go backend
   - Response: "Echo: Pong!" with timestamp

2. **Subsequent Clicks**:
   - Sidecar process already running
   - Immediate response (<100ms)
   - Server ID remains consistent

### 4. Error Scenarios

Test error handling:

```javascript
// Test with null/undefined
ping().then(response => console.log(response))

// Test with empty string
ping("").then(response => console.log(response))

// Test with long message
ping("A".repeat(1000)).then(response => console.log(response))
```

## Development Commands

### Running the Example

```bash
# Development mode
bun run tauri dev

# Production build
bun run tauri build

# Test specific functionality
bun run tauri dev -- --help
```

### Debug Mode

Enable verbose logging:

```bash
# Go backend debug logging
export ANY_SYNC_LOG_LEVEL=debug
bun run tauri dev

# Rust plugin debug logging
RUST_LOG=debug bun run tauri dev
```

## Testing Checklist

### Basic Functionality
- [ ] App starts successfully
- [ ] Plugin loads without errors
- [ ] Ping button responds to clicks
- [ ] Response displays in UI
- [ ] Console shows no errors

### Process Management
- [ ] Go sidecar process starts on first ping
- [ ] Process uses random port allocation
- [ ] Port file created and cleaned up
- [ ] Process shuts down gracefully on app exit

### Error Handling
- [ ] Network errors display in UI
- [ ] Process spawn errors handled gracefully
- [ ] gRPC timeout errors handled
- [ ] Invalid input validation works

### Performance
- [ ] Initial ping <3 seconds
- [ ] Subsequent pings <200ms
- [ ] Memory usage stable
- [ ] No memory leaks detected

## Advanced Testing

### Load Testing
```bash
# Test multiple concurrent pings
for i in {1..10}; do
  curl -X POST http://localhost:1420/api/ping \
    -H "Content-Type: application/json" \
    -d '{"message":"test'$i'}' &
done
```

### Stress Testing
```javascript
// Rapid fire test
setInterval(() => {
  ping("stress test").catch(console.error);
}, 100);
```

### Error Injection
```bash
# Test with unavailable backend
killall server
# Then try ping - should handle connection error

# Test with malformed responses
# Modify Go backend to return errors
```

## Troubleshooting

### Common Issues

1. **Plugin Not Found**
   ```
   Error: Plugin any-sync not found
   ```
   **Solution**: Check `tauri.conf.json` permissions and plugin installation

2. **Sidecar Won't Start**
   ```
   Error: Failed to start server
   ```
   **Solution**: 
   - Check Go installation: `go version`
   - Verify binary exists: `ls ../binaries/`
   - Check permissions: `chmod +x ../binaries/server`

3. **gRPC Connection Failed**
   ```
   Error: Connection refused
   ```
   **Solution**:
   - Check if server is running: `ps aux | grep server`
   - Verify port allocation: `netstat -an | grep LISTEN`
   - Check firewall settings

4. **Build Failures**
   ```
   Error: Go toolchain not found
   ```
   **Solution**:
   ```bash
   # Install Go
   brew install go
   
   # Install protoc
   brew install protobuf
   
   # Set PATH
   export PATH=$PATH:$(go env GOPATH)/bin
   ```

### Debug Information

Enable comprehensive logging:

```bash
# All logs
export RUST_LOG=debug
export ANY_SYNC_LOG_LEVEL=debug
bun run tauri dev

# Go backend logs only
export ANY_SYNC_LOG_LEVEL=debug
./binaries/server --port 8080

# Check sidecar process
ps aux | grep "[s]erver"

# Network connections
lsof -i :8080
```

## Performance Metrics

### Expected Performance
- **Startup Time**: <2 seconds for cold start
- **Ping Latency**: <50ms for warm calls
- **Memory Usage**: <20MB for idle sidecar
- **CPU Usage**: <5% for idle sidecar

### Monitoring Performance
```bash
# Memory usage
ps aux | grep server | awk '{print $6}'

# CPU usage
top -p $(pgrep server)

# Network latency
ping -c 4 localhost
```

## Integration Testing

### Automated Tests
```bash
# Run all tests
bun test

# E2E tests
bun run test:e2e

# Performance tests
bun run test:performance
```

### Manual Testing Guide
1. **Functionality**: Test all plugin features manually
2. **Compatibility**: Test on different OS versions
3. **Edge Cases**: Test with unusual inputs
4. **Recovery**: Test crash recovery scenarios
5. **Documentation**: Verify all examples work

## Success Criteria

✅ **Phase 0 Complete**:
- All basic functionality works
- End-to-end communication verified
- Error handling implemented
- Performance within acceptable limits
- Documentation updated

🔄 **Ready for Phase 1**:
- AnySync/AnyStore integration
- Mobile platform support
- Advanced gRPC features
- Production deployment
# Mobile Package - gomobile Bindings

gomobile FFI bindings for embedding the Go backend in mobile apps (Android/iOS).

## 4-Function API

The mobile layer exports 4 functions that dispatch to shared handlers via binary protobuf:

```go
func Init(dataDir string) error
func Command(cmdName string, protobufBytes []byte) ([]byte, error)
func SetEventHandler(handler func([]byte))
func Shutdown() error
```

### Usage Pattern

**Kotlin (Android)**:
```kotlin
import anysync.Mobile

// Initialize
Mobile.init("/data/data/com.app/anysync")

// Execute command with protobuf bytes
val request = InitRequest.newBuilder()
    .setDataDir("/data/...")
    .build()
val responseBytes = Mobile.command("Init", request.toByteArray())
val response = InitResponse.parseFrom(responseBytes)

// Set event handler
Mobile.setEventHandler { bytes ->
    val event = Event.parseFrom(bytes)
    // Handle event
}

// Shutdown
Mobile.shutdown()
```

**Swift (iOS)**:
```swift
import AnySync

// Initialize
try! AnysyncInit("/Library/Application Support/anysync")

// Execute command
let request = InitRequest.with { $0.dataDir = "..." }
let requestData = try! request.serializedData()
let responseData = try! AnysyncCommand("Init", requestData)
let response = try! InitResponse(serializedData: responseData)

// Shutdown
try! AnysyncShutdown()
```

## Binary Dispatch Pattern

All operations route through `Command(cmdName, protobufBytes)`:

1. Mobile app encodes protobuf request
2. Calls `Command("OperationName", bytes)`
3. Go dispatcher routes to handler in `shared/handlers/`
4. Handler decodes request, executes logic, encodes response
5. Returns protobuf response bytes

Same handlers used by desktop gRPC layer (100% code reuse).

## Building

### Android (.aar)
```bash
task go:mobile:build
# Outputs: binaries/any-sync-android.aar
```

### iOS (.xcframework)
```bash
gomobile bind -target=ios -o binaries/AnySync.xcframework ./mobile
```

## Architecture

**Mobile flow**:
```
TypeScript → Rust → Kotlin/Swift → gomobile FFI → Go Command() → Dispatcher → Handlers → Any-Sync
```

**vs Desktop flow**:
```
TypeScript → Rust → gRPC Client → Go gRPC Server → Dispatcher → Handlers → Any-Sync
```

Both platforms share:
- `shared/dispatcher/` - Command routing
- `shared/handlers/` - Operation logic
- `shared/anysync/` - Any-Sync integration

See [root README](../../README.md) for architecture overview.

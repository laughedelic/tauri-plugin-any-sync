# iOS Plugin Development Guide

This guide covers development and integration of the iOS plugin for the any-sync Tauri plugin.

## Quick Start

```bash
# Build iOS plugin
cd ios
swift build

# Run tests
swift test

# Build Xcode project
swift package generate-xcodeproj
```

## Architecture Overview

The iOS plugin follows Tauri's Swift plugin architecture:

```
ios/
├── Sources/
│   └── ExamplePlugin.swift    # Main plugin class with Tauri commands
├── Tests/
│   └── PluginTests/
│       └── PluginTests.swift   # Unit tests
├── Package.swift              # Swift Package Manager configuration
└── README.md                  # iOS-specific documentation
```

### Key Components

- **Plugin Class** (`ExamplePlugin.swift`): Tauri plugin interface with command handlers
- **Argument Classes**: Codable structs for command parameters
- **Package Config** (`Package.swift`): Dependencies and build settings
- **Tests**: Unit tests for plugin functionality

## Development Workflow

### 1. Plugin Command Implementation

Commands are implemented in `ExamplePlugin.swift`:

```swift
class ExamplePlugin: Plugin {
  @objc public func ping(_ invoke: Invoke) throws {
    let args = try invoke.parseArgs(PingArgs.self)
    
    // Process the ping request
    let response = processPing(args.value ?? "")
    
    invoke.resolve(["value": response])
  }
  
  private func processPing(_ message: String) -> String {
    // Implementation logic here
    return "Echo: \(message)"
  }
}
```

### 2. Command Arguments

Define argument classes conforming to `Decodable`:

```swift
class PingArgs: Decodable {
  let value: String?
}

class ComplexArgs: Decodable {
  let message: String
  let timestamp: Int64
  let options: [String: Any]?
}
```

### 3. Plugin Registration

Register the plugin with the Tauri runtime:

```swift
@_cdecl("init_plugin_any_sync")
func initPlugin() -> Plugin {
  return ExamplePlugin()
}
```

## gomobile Integration (Phase 1+)

### Planned Architecture

For Phase 1+, the iOS plugin will integrate with Go backend via gomobile:

```swift
import Foundation

class GoMobileBridge {
    private let goModule: AnyMobile
    
    init() {
        // Load gomobile framework
        guard let frameworkPath = Bundle.main.path(forResource: "AnyMobile", ofType: "framework"),
              let bundle = Bundle(path: frameworkPath) else {
            fatalError("Failed to load AnyMobile framework")
        }
        
        bundle.load()
        
        // Initialize Go module
        self.goModule = AnyMobile()
    }
    
    func ping(_ message: String) -> String {
        return goModule.ping(message)
    }
}
```

### gomobile Build Process

```bash
# Generate iOS framework from Go code
cd go-backend
gomobile bind -target=ios -o ../ios/AnyMobile.framework

# Build with Swift Package Manager
cd ios
swift build
```

## Build System

### Swift Package Manager

The `Package.swift` handles:

- **Tauri Plugin Dependencies**: Core Tauri iOS plugin framework
- **Swift Configuration**: Language version and compiler options
- **Platform Support**: iOS version requirements
- **External Dependencies**: Third-party libraries

### Build Commands

```bash
# Debug build
swift build

# Release build
swift build -c release

# Run tests
swift test

# Generate Xcode project
swift package generate-xcodeproj

# Build for specific platform
swift build --target x86_64-apple-ios-simulator
```

## Testing

### Unit Tests

Test plugin functionality in `Tests/PluginTests/`:

```swift
import XCTest
@testable import ExamplePlugin

final class PluginTests: XCTestCase {
    func testPingCommand() throws {
        let plugin = ExamplePlugin()
        let invoke = MockInvoke()
        
        try plugin.ping(invoke)
        
        XCTAssertEqual(invoke.result["value"] as? String, "Echo: test")
    }
}
```

### Mock Objects

Create mock objects for testing:

```swift
class MockInvoke: Invoke {
    var result: [String: Any] = [:]
    var error: Error?
    
    func resolve(_ result: [String: Any]) {
        self.result = result
    }
    
    func reject(_ error: Error) {
        self.error = error
    }
}
```

## Dependencies

### Core Dependencies

- `SwiftRs`: Swift runtime for Tauri plugins
- `Tauri`: Core Tauri iOS plugin framework
- `Foundation`: iOS standard library

### Development Dependencies

- `XCTest`: Unit testing framework
- `SwiftLint`: Code style enforcement

## Configuration

### Package Configuration

Configure build settings in `Package.swift`:

```swift
let package = Package(
    name: "ExamplePlugin",
    platforms: [
        .iOS(.v13)
    ],
    products: [
        .library(name: "ExamplePlugin", targets: ["ExamplePlugin"])
    ],
    dependencies: [
        .package(url: "https://github.com/tauri-apps/tauri-ios-plugin", from: "1.0.0")
    ],
    targets: [
        .target(name: "ExamplePlugin", dependencies: ["Tauri"]),
        .testTarget(name: "PluginTests", dependencies: ["ExamplePlugin"])
    ]
)
```

### Info.plist

Add necessary permissions and configurations:

```xml
<key>NSAppTransportSecurity</key>
<dict>
    <key>NSAllowsArbitraryLoads</key>
    <true/>
</dict>
```

## Debugging

### NSLog Debugging

Use NSLog for debugging:

```swift
import Foundation

class ExamplePlugin: Plugin {
    @objc public func ping(_ invoke: Invoke) throws {
        NSLog("AnySync: Processing ping command")
        
        let args = try invoke.parseArgs(PingArgs.self)
        NSLog("AnySync: Received message: \(args.value ?? "nil")")
        
        // Process request...
        
        invoke.resolve(["value": response])
    }
}
```

### Debug Commands

```bash
# View device logs
xcrun simctl spawn booted log stream --predicate 'subsystem == "com.plugin.any-sync"'

# Install on simulator
xcrun simctl install booted .build/debug/ExamplePlugin.framework

# Run with debugger
lldb -s debug_commands.txt
```

## Performance Considerations

### Memory Management

- Use ARC properly to avoid memory leaks
- Clean up resources in plugin lifecycle
- Avoid retain cycles with closures

### Threading

- Run heavy operations on background queues
- Update UI on main thread only
- Use proper synchronization for shared resources

### Network Operations

- Use URLSession for network requests
- Implement proper timeout configurations
- Handle network state changes

## Security Notes

### Input Validation

- Validate all command arguments
- Sanitize inputs before processing
- Implement proper error handling

### Code Signing

- Properly sign frameworks and binaries
- Handle certificate validation
- Use secure communication channels

## Troubleshooting

### Common Issues

1. **Build Failures**
   ```
   error: no such module 'Tauri'
   ```
   **Solution**: Check Swift Package Manager dependencies and repository access

2. **Runtime Errors**
   ```
   unrecognized selector sent to instance
   ```
   **Solution**: Verify plugin methods are properly exposed with @objc

3. **gomobile Integration**
   ```
   dyld: Library not loaded: @rpath/AnyMobile.framework
   ```
   **Solution**: Ensure gomobile framework is properly embedded and signed

### Debug Commands

```bash
# Check Swift dependencies
swift package show-dependencies

# Verify plugin registration
nm .build/debug/ExamplePlugin.framework/ExamplePlugin | grep init_plugin

# Test gomobile integration
otool -L AnyMobile.framework/AnyMobile
```

## Phase 1+ Planning

### gomobile Integration Steps

1. **Go Backend Preparation**
   - Implement gomobile-compatible Go API
   - Add mobile-specific build targets
   - Generate iOS framework

2. **iOS Plugin Updates**
   - Load gomobile framework
   - Implement Swift bridge functions
   - Add error handling for native calls

3. **Testing and Validation**
   - Unit tests for Go bridge
   - Integration tests for end-to-end flow
   - Performance testing of native calls

### Expected Architecture

```
TypeScript UI → Tauri Commands → iOS Plugin → gomobile Bridge → Go Backend
```

## Success Criteria

✅ **Phase 0 Complete**:
- Basic iOS plugin structure established
- Tauri command framework working
- Build system configured
- Unit tests implemented

🔄 **Ready for Phase 1**:
- gomobile integration complete
- End-to-end communication with Go backend
- Performance optimization
- Production deployment ready
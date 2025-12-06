# iOS Plugin

## Structure

```
ios/
├── Sources/ExamplePlugin.swift  # Single command() method calling Go via gomobile
└── Tests/PluginTests/           # Basic instantiation tests
```

## Implementation

Minimal passthrough to Go backend:

```swift
import Mobile  // Gomobile-generated framework

class AnySyncPlugin: Plugin {
  @objc public func command(_ invoke: Invoke) throws {
    let args = try invoke.parseArgs(CommandArgs.self)
    let response = try MobileCommand(args.cmd, Data(args.data))
    invoke.resolve(["data": [UInt8](response ?? Data())])
  }
}
```

## Building gomobile Framework

Generate `Mobile.xcframework` from Go backend:

```bash
cd plugin-go-backend/mobile
gomobile bind -target=ios -o ../../binaries/any-sync-ios.xcframework .
```

Framework provides:
- `MobileInit()` - Initialize backend
- `MobileCommand(cmd, data)` - Execute command
- `MobileShutdown()` - Cleanup

## Notes

- **Package.swift**: Tauri dependency path is app-specific (resolved at app build time)
- **Testing**: Unit tests verify instantiation only; full tests are integration-level
- **No business logic**: This layer is pure passthrough

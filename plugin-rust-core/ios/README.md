# iOS Native Plugin

Minimal Swift shim for Tauri plugin integration.

## Architecture

**Single-dispatch pattern**: All commands route through one function that calls gomobile FFI.

```swift
@objc public func command(_ invoke: Invoke) throws {
    let cmd = invoke.getString("cmd")
    let data = invoke.getData("data")

    let response = try AnysyncCommand(cmd, data)  // gomobile FFI call
    invoke.resolve(["data": response])
}
```

The native layer is a thin passthrough - all logic lives in Go backend via FFI.

See [plugin-go-backend/mobile](../../plugin-go-backend/mobile/README.md) for FFI details.

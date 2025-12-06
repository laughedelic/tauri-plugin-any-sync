# Android Native Plugin

Minimal Kotlin shim for Tauri plugin integration.

## Architecture

**Single-dispatch pattern**: All commands route through one function that calls gomobile FFI.

```kotlin
@Command
fun command(invoke: Invoke) {
    val cmd = invoke.getString("cmd")
    val data = invoke.getByteArray("data")

    val response = Mobile.command(cmd, data)  // gomobile FFI call
    invoke.resolve(JSObject().put("data", response))
}
```

The native layer is a thin passthrough - all logic lives in Go backend via FFI.

See [plugin-go-backend/mobile](../../plugin-go-backend/mobile/README.md) for FFI details.

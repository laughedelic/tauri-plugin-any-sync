# Android Plugin

## Structure

```
android/
├── libs/any-sync-android.aar    # Gomobile-generated library (symlinked from binaries/)
└── src/main/java/AnySyncPlugin.kt  # Single command() method calling Go via JNI
```

## Implementation

Minimal passthrough to Go backend:

```kotlin
import mobile.Mobile  // Gomobile-generated JNI bindings

class AnySyncPlugin(activity: Activity): Plugin(activity) {
  init { System.loadLibrary("gojni") }
  
  @Command
  fun command(invoke: Invoke) {
    val args = invoke.parseArgs(CommandArgs::class.java)
    val response = Mobile.command(args.cmd, args.data)
    invoke.resolve(JSObject().put("data", response))
  }
}
```

## Building gomobile AAR

Generate `any-sync-android.aar` from Go backend:

```bash
cd plugin-go-backend/mobile
gomobile bind -target=android -androidapi=21 -o ../../binaries/any-sync-android.aar .
```

AAR provides:
- `Mobile.init()` - Initialize backend
- `Mobile.command(cmd, data)` - Execute command  
- `Mobile.shutdown()` - Cleanup

## Plugin Build

The plugin's `build.rs` symlinks the .aar to `android/libs/` automatically.

`build.gradle.kts` includes it:
```kotlin
implementation(files("libs/any-sync-android.aar"))
```

## Debugging

```bash
adb logcat | grep AnySync
```

Key log messages:
- `"Successfully loaded gojni library"` - JNI OK
- `"Mobile backend initialized"` - Init OK
- `"command: cmd=..., data.size=..."` - Command received

## Notes

- **No business logic**: This layer is pure passthrough
- **Response structure**: Must return `{"data": ByteArray}` matching Rust's `CommandResponse`
- **Testing**: Unit tests verify args parsing; full tests are integration-level

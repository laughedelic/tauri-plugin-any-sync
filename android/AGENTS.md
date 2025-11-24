# Android Plugin Development Guide

This guide covers development and integration of the Android plugin for the any-sync Tauri plugin.

## Quick Start

```bash
# Build Android plugin
cd android
./gradlew build

# Run tests
./gradlew test

# Build with AAR output
./gradlew assembleRelease
```

## Architecture Overview

The Android plugin integrates the Go backend via gomobile, using JNI to call Go functions directly.

```
android/
├── libs/
│   └── any-sync-android.aar    # Go mobile library (gomobile build)
├── src/main/java/
│   └── ExamplePlugin.kt        # Main plugin with storage commands + JNI calls
├── build.gradle.kts            # Gradle build configuration (includes .aar)
├── proguard-rules.pro          # ProGuard configuration
└── settings.gradle.kts         # Gradle settings
```

### Key Components

- **Plugin Class** (`ExamplePlugin.kt`): Tauri plugin interface with storage command handlers
- **Go Mobile Library** (`libs/any-sync-android.aar`): Native Go backend compiled with gomobile
- **JNI Integration**: Direct function calls from Kotlin to Go via `mobile.Mobile` class
- **Build Config** (`build.gradle.kts`): Dependencies including .aar library

### Communication Flow

```
TypeScript API → Tauri Command → Kotlin Plugin → JNI → Go Mobile → AnyStore
```

Unlike desktop (which uses gRPC sidecar), Android embeds the Go backend as a native library:
- **Desktop**: Process IPC via gRPC (separate sidecar process)
- **Android**: In-process JNI calls (embedded library)

## Development Workflow

### 1. Plugin Command Implementation

Commands are implemented in `ExamplePlugin.kt`:

```kotlin
@TauriPlugin
class ExamplePlugin(private val activity: Activity): Plugin(activity) {
    private val implementation = Example()

    @Command
    fun ping(invoke: Invoke) {
        val args = invoke.parseArgs(PingArgs::class.java)
        
        val ret = JSObject()
        ret.put("value", implementation.pong(args.value ?: "default value"))
        invoke.resolve(ret)
    }
}
```

### 2. Command Arguments

Define argument classes with `@InvokeArg`:

```kotlin
@InvokeArg
class PingArgs {
    var value: String? = null
}
```

### 3. Implementation Logic

Keep business logic separate from plugin framework:

```kotlin
class Example {
    fun pong(value: String): String {
        Log.i("Pong", value)
        return value
    }
}
```

## gomobile Integration (Phase 1+)

### Planned Architecture

For Phase 1+, the Android plugin will integrate with Go backend via gomobile:

```kotlin
class GoMobileBridge {
    private external fun nativePing(message: String): String
    
    init {
        System.loadLibrary("anymobile")
    }
    
    fun ping(message: String): String {
        return try {
            nativePing(message)
        } catch (e: Exception) {
            "Error: ${e.message}"
        }
    }
}
```

### gomobile Build Process

```bash
# Generate Android AAR from Go code
cd go-backend
gomobile bind -target=android -o ../android/libs/anymobile.aar

# Build with Android library
cd android
./gradlew build
```

## Build System

### Gradle Configuration

The `build.gradle.kts` handles:

- **Tauri Plugin Dependencies**: Core Tauri Android plugin framework
- **Kotlin Configuration**: Language version and compiler options
- **Android SDK**: Target and minimum SDK versions
- **Build Types**: Debug and release configurations

### Build Commands

```bash
# Debug build
./gradlew assembleDebug

# Release build
./gradlew assembleRelease

# Run tests
./gradlew test

# Run instrumented tests
./gradlew connectedAndroidTest

# Clean build
./gradlew clean
```

## Testing

### Unit Tests

Test implementation logic in `src/test/`:

```kotlin
@Test
fun testPong() {
    val example = Example()
    val result = example.pong("test")
    assertEquals("test", result)
}
```

### Integration Tests

Test plugin commands in `src/androidTest/`:

```kotlin
@Test
fun testPingCommand() {
    val plugin = ExamplePlugin(activity)
    val invoke = MockInvoke()
    plugin.ping(invoke)
    assertEquals("test", invoke.result)
}
```

## Dependencies

### Core Dependencies

- `app.tauri:plugin`: Tauri plugin framework
- `org.jetbrains.kotlin:kotlin-stdlib`: Kotlin standard library
- `androidx.core:core-ktx`: Android KTX extensions

### Development Dependencies

- `junit:junit`: Unit testing framework
- `androidx.test.ext:junit`: Android test extensions
- `androidx.test.espresso`: UI testing framework

## Configuration

### Android Manifest

Key permissions and configurations in `AndroidManifest.xml`:

```xml
<uses-permission android:name="android.permission.INTERNET" />
<uses-permission android:name="android.permission.ACCESS_NETWORK_STATE" />
```

### Build Variants

Configure debug and release variants:

```kotlin
android {
    buildTypes {
        debug {
            isDebuggable = true
            applicationIdSuffix = ".debug"
        }
        release {
            isMinifyEnabled = true
            proguardFiles(getDefaultProguardFile("proguard-android.txt"))
        }
    }
}
```

## Debugging

### Logcat Debugging

Use Android Log for debugging:

```kotlin
import android.util.Log

class Example {
    fun pong(value: String): String {
        Log.d("AnySync", "Processing pong: $value")
        return value
    }
}
```

### Debug Commands

```bash
# View logs
adb logcat | grep AnySync

# Install debug APK
adb install app/build/outputs/apk/debug/app-debug.apk

# Run with debugger
./gradlew installDebug
adb shell am start -n com.plugin.any-sync/.MainActivity
```

## Performance Considerations

### Memory Management

- Avoid memory leaks in long-running operations
- Use weak references for Activity contexts
- Clean up resources in plugin lifecycle methods

### Threading

- Run heavy operations on background threads
- Use coroutines for async operations
- Update UI on main thread only

### Network Operations

- Use proper timeout configurations
- Implement retry logic for network failures
- Handle network state changes

## Security Notes

### Input Validation

- Validate all command arguments
- Sanitize inputs before processing
- Implement proper error handling

### Permissions

- Request minimum necessary permissions
- Explain permission usage to users
- Handle permission denials gracefully

## Troubleshooting

### Common Issues

1. **Build Failures**
   ```
   Could not find com.tauri:plugin
   ```
   **Solution**: Check Tauri plugin dependency version and repository configuration

2. **Runtime Errors**
   ```
   ClassNotFoundException: ExamplePlugin
   ```
   **Solution**: Verify plugin is properly registered in Tauri configuration

3. **gomobile Integration**
   ```
   UnsatisfiedLinkError: nativePing
   ```
   **Solution**: Ensure gomobile library is properly built and loaded

### Debug Commands

```bash
# Check Gradle dependencies
./gradlew dependencies

# Verify plugin registration
adb shell dumpsys package com.plugin.any-sync

# Test gomobile integration
adb shell am start -n com.plugin.any-sync/.MainActivity -e action test_ping
```

## Phase 1+ Planning

### gomobile Integration Steps

1. **Go Backend Preparation**
   - Implement gomobile-compatible Go API
   - Add mobile-specific build targets
   - Generate Android AAR library

2. **Android Plugin Updates**
   - Load gomobile library
   - Implement JNI bridge functions
   - Add error handling for native calls

3. **Testing and Validation**
   - Unit tests for Go bridge
   - Integration tests for end-to-end flow
   - Performance testing of native calls

### Expected Architecture

```
TypeScript UI → Tauri Commands → Android Plugin → gomobile Bridge → Go Backend
```

## Success Criteria

✅ **Phase 0 Complete**:
- Basic Android plugin structure established
- Tauri command framework working
- Build system configured
- Unit tests implemented

🔄 **Ready for Phase 1**:
- gomobile integration complete
- End-to-end communication with Go backend
- Performance optimization
- Production deployment ready
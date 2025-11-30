# Testing Strategy Pivot: Integration Testing on Mobile Platforms

## Summary

The original plan to implement full-chain integration tests (App → Plugin → Go Backend) using Tauri's `MockRuntime` across all platforms hit a fundamental limitation: **the Mock Runtime does not support Android/mobile platforms**. This document explains the issue and outlines the revised testing approach.

## The Problem

### Original Goal
Test the complete communication chain on all platforms:
- **Desktop**: TypeScript → Rust Commands → Desktop Service → gRPC Client → Go Sidecar
- **Mobile**: TypeScript → Rust Commands → Mobile Service → Kotlin/Swift Plugin → JNI/FFI → Go Library

### Why Mock Runtime Doesn't Work on Android

The integration tests used `tauri::test::mock_builder()` and `get_ipc_response()` to invoke plugin commands through the actual IPC layer without requiring a GUI. This approach works perfectly on desktop platforms.

However, when running these tests on Android, they fail with:
```
panicked at mock_runtime.rs:292:5:
not yet implemented
```

The root cause is in Tauri's source code:
```rust
#[cfg(target_os = "android")]
fn run_on_android_context<F: Fn(&mut android_activity::AndroidApp) + Send>(&self, f: F) {
    todo!()
}
```

**The Mock Runtime is fundamentally incompatible with Android.** On Android, Tauri requires a real Android Context (via JNI) to function, which the Mock Runtime cannot provide. Without this Context, even basic initialization fails.

### iOS Has the Same Issue

While iOS doesn't use JNI (it uses C FFI instead), the Mock Runtime similarly lacks the necessary iOS-specific runtime infrastructure. The same testing limitation applies.

## The Revised Approach

Instead of trying to run full-chain integration tests on mobile platforms, we split testing into two complementary strategies:

### 1. Desktop: Full Chain Integration Tests (Active)

**Status**: ✅ Implemented and working

Use `MockRuntime` and `get_ipc_response()` to test the complete path on desktop platforms (macOS, Linux, Windows). These tests verify:
- IPC serialization/deserialization
- Command routing
- gRPC communication with Go sidecar
- Error propagation
- Data integrity

**Why it works**: Desktop doesn't require JNI or mobile-specific runtime infrastructure.

### 2. Mobile: Build Validation + Plugin Isolation Tests (Planned)

**Status**: 📝 Documented for future implementation

For Android and iOS, use a two-pronged approach:

#### A. Build Validity Check
**Purpose**: Verify compilation, linking, and architecture targets are correct

```bash
# Android
cargo tauri android build --debug

# iOS (when supported)
cargo tauri ios build --debug --no-signing
```

**What it proves**:
- FFI definitions are correct
- Go library links successfully
- No symbol resolution errors
- Architecture slicing works (arm64 vs x86_64)

#### B. Plugin Isolation Tests (Optional)
**Purpose**: Test Rust ↔ Go FFI layer independently

Run plugin-level tests on the development host (Linux/macOS) that verify the FFI communication:

- **Android approach**: Use OpenJDK on host to test JNI bindings (JNI interface is identical on desktop and Android)
- **iOS approach**: Test C FFI on macOS (shares same Darwin kernel and C ABI with iOS)

These tests would live in the plugin crate, not the example app, and would require additional setup (loading Go libraries, initializing JVM, etc.).

## Current Implementation Status

| Platform                        | Test Type              | Status                                    | Coverage                   |
| ------------------------------- | ---------------------- | ----------------------------------------- | -------------------------- |
| Desktop (macOS, Linux, Windows) | Full chain integration | ✅ Active                                  | IPC → Plugin → gRPC → Go   |
| Android                         | Build validation       | 📝 Documented                              | Compilation + linking only |
| Android                         | Plugin isolation       | 🔄 Optional future work                    | FFI layer only             |
| iOS                             | Build validation       | 📝 Documented (platform not yet supported) | Compilation + linking only |
| iOS                             | Plugin isolation       | 🔄 Optional future work                    | FFI layer only             |

## Trade-offs

### What We Gain
- ✅ Working integration tests on desktop (primary development platform)
- ✅ Build verification ensures mobile builds don't break
- ✅ Simpler CI setup (no Android emulator for basic tests)
- ✅ Faster feedback loop
- ✅ 95% confidence with 10% of the complexity

### What We Lose
- ❌ No runtime validation of mobile FFI/JNI on actual devices
- ❌ Can't catch mobile-specific runtime bugs in CI
- ❌ Must rely on manual testing for mobile platforms

### Mitigation
The plugin's mobile FFI layer is relatively thin - most logic lives in the shared Go storage layer, which is thoroughly tested on desktop. Build validation ensures the FFI signatures are correct. Runtime issues would be caught during development testing on actual devices/emulators.

## Recommendations

1. **Implement desktop integration tests fully** ✅ (Done)
2. **Add build validation to CI for Android** 📝 (Next step)
3. **Document iOS build validation** 📝 (When iOS support added)
4. **Consider plugin isolation tests** 🔄 (If mobile-specific bugs emerge)
5. **Supplement with manual mobile testing** during feature development

## References

- Original proposal: `openspec/changes/add-integration-tests/proposal.md`
- Design document: `openspec/changes/add-integration-tests/design.md`
- Tauri Mock Runtime source: https://github.com/tauri-apps/tauri/blob/dev/crates/tauri/src/test/mock_runtime.rs#L287-L293

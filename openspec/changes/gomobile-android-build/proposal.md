# Change: Add Android Support via gomobile

## Why

The plugin currently works on desktop platforms (Windows, macOS, Linux) using the sidecar pattern. To achieve true cross-platform capability, we need to integrate the same Go backend with Android using gomobile as planned in Phase 2 of the architecture.

This validates that AnyStore (which uses pure-Go `modernc.org/sqlite` with zero CGo) can be built with gomobile and that the same storage API works identically across platforms.

## What Changes

- **Add gomobile mobile entrypoint** (`go-backend/cmd/mobile/`) exposing storage functions
- **Build .aar artifacts** via `gomobile bind -target=android`
- **Distribute .aar files** following same pattern as desktop binaries (GitHub Releases, checksums, auto-download)
- **Implement Android plugin** in Kotlin to call Go functions via JNI
- **Initialize Android support** in example app (`tauri android init`)
- **Update build system** to handle .aar downloads and integration

## Impact

- **Affected specs**:
  - `mobile-backend-api` (NEW) - gomobile-compatible Go API
  - `android-plugin-integration` (NEW) - Kotlin plugin implementation
  - `binaries-distribution` (MODIFIED) - Extend to support mobile artifacts

- **Affected code**:
  - `go-backend/cmd/mobile/` (NEW) - Mobile entrypoint
  - `android/src/main/java/ExamplePlugin.kt` (MODIFIED) - Storage commands
  - `build.rs` (MODIFIED) - .aar download logic
  - `.github/workflows/release.yml` (MODIFIED) - Build .aar files
  - `examples/tauri-app/` (MODIFIED) - Android initialization

- **Not affected**:
  - TypeScript API (unchanged, works on both platforms)
  - Desktop sidecar pattern (unchanged)
  - Core storage logic (>95% code reuse)
  - iOS (deferred to Phase 3)

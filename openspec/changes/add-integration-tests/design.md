# Design: Integration Testing Strategy

## Context

The tauri-plugin-any-sync uses different architectures for desktop and mobile:
- **Desktop**: Go sidecar process, gRPC communication
- **Mobile**: gomobile FFI bindings (.aar for Android, .xcframework for iOS)

Both platforms share 95%+ of the Go backend code (storage layer) and expose the same Tauri commands to the application. From the app's perspective, the interface is identical - only the target platform differs. Integration tests must verify this unified interface works correctly on all platforms.

**Test Scope:** `App's Rust backend → (Tauri command) → Plugin's Rust core → Platform implementation → Go backend`

**Out of Scope:**
- TypeScript API testing (would require different infrastructure, can be deferred)
- GUI/WebDriver testing (not needed for backend integration)
- Frontend testing (not the goal)

## Goals / Non-Goals

**Goals:**
- Test desktop and Android platforms (iOS setup included but not validated - platform support doesn't exist yet)
- Use actual Tauri command invocation path (not bypassing with direct access)
- Tests run automatically in CI for desktop and Android
- Verify end-to-end functionality across all plugin layers
- Catch platform-specific regressions (sidecar vs FFI)
- Test all 5 commands and error handling scenarios
- Provide iOS testing infrastructure ready for future iOS platform support

**Non-Goals:**
- TypeScript API testing in isolation (different test infrastructure needed)
- GUI/WebDriver testing (backend-only focus)
- Performance/load testing (functional correctness only)

## Decisions

### Decision 1: Test Desktop and Android Platforms (iOS Setup for Future)

**Choice:** Implement integration tests for desktop and Android, with iOS infrastructure documented but not active (iOS platform support doesn't exist yet).

**Rationale:**
- Plugin provides unified interface - must verify it works on available platforms
- Desktop uses gRPC sidecar, Android uses FFI - different code paths need testing
- Platform-specific bugs won't be caught without testing both
- iOS plugin implementation doesn't exist yet, but we can prepare the testing infrastructure
- When iOS support is added later, tests will be ready to validate it immediately

**Implementation:**
- Use conditional compilation: `#[cfg(desktop)]` / `#[cfg(mobile)]` for platform-specific tests
- Shared test functions where behavior is identical
- Active CI jobs: Ubuntu for desktop, Ubuntu+emulator for Android
- Documented but inactive: macOS for iOS (to be enabled when iOS plugin is implemented)

**Alternatives considered:**
- **Desktop-only testing:** Wouldn't catch Android-specific FFI issues, gomobile binding bugs, or platform differences
- **Wait for iOS before adding mobile tests:** Would delay Android testing unnecessarily
- **Separate test files per platform:** More duplication, harder to maintain

### Decision 2: Use Tauri's IPC Testing Utilities

**Choice:** Use `tauri::test::get_ipc_response()` and `assert_ipc_response()` to invoke plugin commands.

**Rationale:**
- **Tests the actual invocation path:** Commands are invoked the same way the app invokes them
- **Official Tauri approach:** Uses documented testing utilities, not custom workarounds
- **No plugin source pollution:** No need for test-only traits or extensions in plugin code
- **Tests IPC serialization:** Verifies request/response serialization works correctly
- **Maintainable:** Uses Tauri's stable testing API

**Code pattern:**
```rust
use tauri::test::{get_ipc_response, mock_context, noop_assets, MockRuntime, INVOKE_KEY};
use tauri::ipc::{InvokeRequest, InvokeBody};

let app = create_app_builder::<MockRuntime>()
    .build(mock_context(noop_assets()))
    .expect("failed to build test app");

let webview = app.webview_windows().values().next().expect("no webview");

let request = InvokeRequest {
    cmd: "plugin:any-sync|ping".into(),
    callback: 0.into(),
    error: 1.into(),
    body: InvokeBody::Json(serde_json::json!({
        "value": "test message"
    })),
    headers: Default::default(),
    invoke_key: INVOKE_KEY,
};

let response = get_ipc_response(webview, request);
// Assert on response
```

**Alternatives considered:**
- **Custom AnySyncExt trait:** Would pollute plugin source with test-only code, bypasses actual invocation path, not an established Tauri pattern
- **Direct service access:** Bypasses command layer entirely, doesn't test IPC serialization
- **Mock IPC from JavaScript:** Requires different test infrastructure, doesn't test Rust→Plugin path

### Decision 3: Platform-Specific CI Jobs

**Choice:** Add separate CI jobs for desktop and Android testing; document iOS job for future activation.

**Rationale:**
- Desktop and Android require different CI environments
- Desktop: Standard Ubuntu runner with webkit2gtk
- Android: Ubuntu runner with Android SDK and emulator
- iOS: Will need macOS runner with Xcode and iOS simulator (when platform support exists)

**CI Structure:**
```yaml
test-integration-desktop:
  runs-on: ubuntu-latest
  # Build desktop sidecar, run desktop tests

test-integration-android:
  runs-on: ubuntu-latest
  # Build gomobile .aar, set up Android SDK/emulator, run Android tests

# iOS job documented but commented out - to be enabled when iOS plugin implemented
# test-integration-ios:
#   runs-on: macos-latest
#   # Build gomobile .xcframework, set up iOS simulator, run iOS tests
```

**Alternatives considered:**
- **Single job for all platforms:** Not feasible - can't run Android emulator and iOS simulator together, different build requirements
- **Manual testing only:** Wouldn't catch regressions, slow feedback loop
- **Wait to document iOS until later:** Better to have infrastructure ready so it can be enabled immediately when iOS support lands

### Decision 4: Sequential Test Execution

**Choice:** Run tests with `--test-threads=1` to prevent database conflicts.

**Rationale:**
- Both desktop and mobile tests may share database state (depending on implementation)
- SQLite doesn't handle parallel access well without WAL mode
- Sequential execution is fast enough (~1-2 seconds per test, <30 seconds total)
- Simpler than managing per-test database isolation

**Alternatives considered:**
- **Separate database per test:** Would require modifying backend to accept different DB paths, adds complexity
- **Parallel execution with locking:** Could use file locks but adds brittleness

### Decision 5: Fix CI Paths (No Compatibility Shims)

**Choice:** Update CI workflow to use current paths and Task-based builds.

**Rationale:**
- Project structure has changed, old paths are incorrect
- Task-based builds are more maintainable than shell scripts
- No reason to maintain backwards compatibility in CI workflow
- Aligns with current project conventions

**Alternatives considered:**
- **Create symlinks for old paths:** Technical debt, confusing for contributors
- **Create `build-go-backend.sh` script:** Duplicates Task functionality

## Risks / Trade-offs

**Risk 1: Mobile CI complexity and cost**
- Android emulator startup is slow (~2-3 minutes)
- iOS requires macOS runner (more expensive than Ubuntu)
- **Mitigation:** Run mobile tests only on main branch / PR approval, not every commit

**Risk 2: Tests may be flaky in CI**
- Emulators can be unstable
- Sidecar/FFI connection timing issues
- **Mitigation:** Use `RUST_LOG=debug` for logs, add retry logic for initial connections, ensure adequate timeouts

**Risk 3: Database conflicts if `--test-threads=1` is forgotten**
- **Mitigation:** Enforce in Task command, document clearly, add assertion in test setup

**Trade-off: IPC testing adds serialization complexity**
- **Pro:** Tests actual invocation path, catches serialization bugs, official Tauri approach
- **Con:** Slightly more verbose test code (need to build InvokeRequest)
- **Decision:** Worth it for testing the real path and avoiding plugin source pollution

## Mobile Testing Implementation

### Android Testing (Active)

**Approach:** Use Android emulator on Ubuntu runner

```yaml
test-integration-android:
  runs-on: ubuntu-latest
  steps:
    - name: Set up JDK
      uses: actions/setup-java@v4
      with:
        java-version: '17'

    - name: Setup Android SDK
      uses: android-actions/setup-android@v3

    - name: Build gomobile library
      run: task go:mobile:build-android

    - name: Run Android instrumented tests
      uses: reactivecircus/android-emulator-runner@v2
      with:
        api-level: 33
        target: google_apis
        arch: x86_64
        script: cargo test --test integration --features android
```

**What it tests:**
- Kotlin plugin wrapper
- JNI bindings to Go
- gomobile .aar integration
- Full round-trip on Android

### iOS Testing (Infrastructure Only - Not Active)

**Status:** iOS plugin implementation doesn't exist yet. This CI job will be documented but commented out, ready to be activated when iOS support is added.

**Planned approach:** Use iOS simulator on macOS runner

```yaml
# Commented out - enable when iOS plugin is implemented
# test-integration-ios:
#   runs-on: macos-latest
#   steps:
#     - name: Select Xcode
#       run: sudo xcode-select -s /Applications/Xcode_15.0.app
#
#     - name: Build gomobile framework
#       run: task go:mobile:build-ios
#
#     - name: Run iOS tests
#       run: cargo test --test integration --features ios
```

**What it will test (when enabled):**
- Swift plugin wrapper
- FFI bindings to Go
- gomobile .xcframework integration
- Full round-trip on iOS

**Rationale for documenting now:**
- When iOS support is added, testing infrastructure is immediately available
- Consistent testing approach across all mobile platforms
- No need to redesign testing strategy later

## Migration Plan

No migration needed - this is net-new functionality. Existing empty test stubs will be implemented.

**Rollout:**
1. Implement desktop tests first (faster CI, simpler setup)
2. Verify desktop CI passes reliably
3. Add Android tests and CI job
4. Verify Android CI passes reliably
5. Document iOS CI job (commented out) for future activation
6. Make desktop and Android tests required checks for PRs
7. **Future:** When iOS plugin is implemented, uncomment iOS CI job and make it required

**Rollback:**
- Can disable specific platform checks if tests prove flaky
- Test code is isolated to `tests/` directory, no production impact

## Open Questions

None - approach uses official Tauri testing utilities and established patterns.
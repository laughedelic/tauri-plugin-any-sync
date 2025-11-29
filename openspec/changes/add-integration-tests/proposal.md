# Change: Add Integration Testing for Desktop and Mobile Platforms

## Why

The plugin currently has no working integration tests despite having test scaffolding in place. All 11 test functions in `example-app/src-tauri/tests/integration.rs` are empty TODOs. The CI workflow has outdated paths and references non-existent scripts.

Without integration tests, we cannot verify that the plugin correctly integrates with applications across the full command invocation path (App's Rust backend → Plugin's Rust core → Go backend) or catch regressions in the platform-specific implementations (desktop sidecar vs mobile FFI bindings). The plugin provides a unified interface regardless of platform, so tests should verify this works correctly on both desktop and mobile.

## What Changes

- Implement 11 integration test functions covering all 5 plugin commands (ping, storage_put, storage_get, storage_delete, storage_list)
- Use Tauri's official `get_ipc_response()` / `assert_ipc_response()` utilities to test the actual command invocation path
- Support desktop and Android platform testing (iOS setup included but not validated - iOS platform support doesn't exist yet)
- Fix CI workflow paths (`go-backend/` → `plugin-go-backend/`, `examples/tauri-app/` → `example-app/`)
- Replace non-existent `build-go-backend.sh` script reference with `task backend:build` / `task go:mobile:build-android`
- Add CI jobs for desktop and Android testing; document iOS setup for future use
- Update documentation to reflect testing strategy for current platforms and future iOS support

## Impact

- **Affected specs**: integration-testing (new capability)
- **Affected code**:
  - `example-app/src-tauri/tests/integration.rs` - Implement all test functions (~500-600 lines)
  - `.github/workflows/test.yml` - Fix paths, add desktop and Android CI jobs, document iOS job (not active)
  - `CLAUDE.md` - Update integration tests section
  - `example-app/src-tauri/tests/README.md` - Document desktop and Android testing; include iOS setup guide for future
- **Testing impact**: Provides comprehensive end-to-end test coverage for desktop and Android platforms; iOS testing infrastructure ready for when iOS support is added
- **CI impact**: Integration tests will run on every push/PR for desktop and Android, catching regressions early
- **Breaking changes**: None - this is net-new test infrastructure

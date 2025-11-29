# Implementation Tasks

## 1. Foundation - Desktop Testing

- [x] 1.1 Fix outdated paths in `.github/workflows/test.yml` (go-backend → plugin-go-backend, examples → example-app)
- [x] 1.2 Replace `build-go-backend.sh` reference with `task backend:build` in desktop CI
- [x] 1.3 Add Task installation step to CI workflow
- [x] 1.4 Update test helper in `example-app/src-tauri/tests/integration.rs` to create webview for IPC testing
- [x] 1.5 Add `integration-test` feature flag and build.rs support for automatic binary linking

## 2. Desktop Basic Tests

- [x] 2.1 Implement `test_ping_command` using `get_ipc_response()` - verify sidecar starts and gRPC works
- [x] 2.2 Implement `test_ping_command_empty_message` - verify handling of None/empty messages
- [x] 2.3 Run tests locally with `task app:test-integration` to verify IPC infrastructure works
- [x] 2.4 Update task configuration to use `--features integration-test` for binary linking
- [ ] 2.5 Verify desktop tests pass in CI

## 3. Desktop Storage Tests

- [x] 3.1 Implement `test_storage_put_and_get` - create document, retrieve, verify JSON integrity
- [x] 3.2 Implement `test_storage_get_nonexistent` - verify graceful handling of missing documents
- [x] 3.3 Implement `test_storage_update_existing_document` - verify upsert behavior
- [x] 3.4 Implement `test_storage_list` - put multiple docs, list collection, verify all IDs
- [x] 3.5 Implement `test_storage_list_empty` - list non-existent collection, verify empty result
- [x] 3.6 Implement `test_storage_delete` - delete existing doc, verify it's gone
- [x] 3.7 Implement `test_storage_delete_nonexistent` - verify idempotent delete

## 4. Desktop Complex Scenarios

- [x] 4.1 Implement `test_multiple_collections` - verify collection isolation
- [x] 4.2 Implement `test_complex_json_document` - verify nested objects, arrays, Unicode, special chars (covered in storage_put_and_get)
- [x] 4.3 Verify all desktop tests pass locally (10 tests passing in 1.20s)

## 5. Mobile Testing Infrastructure - Android

- [ ] 5.1 Add `test-integration-android` job to `.github/workflows/test.yml`
- [ ] 5.2 Configure Android SDK and emulator setup in CI
- [ ] 5.3 Add gomobile Android library build step (`task go:mobile:build-android`)
- [ ] 5.4 Set up Android-specific test execution with cargo features

## 6. iOS Testing Infrastructure (Documentation Only)

- [ ] 6.1 Document `test-integration-ios` job in `.github/workflows/test.yml` (commented out)
- [ ] 6.2 Document macOS runner setup with Xcode and iOS simulator requirements
- [ ] 6.3 Document gomobile iOS framework build step (`task go:mobile:build-ios`) for future use
- [ ] 6.4 Add note that iOS job will be enabled when iOS plugin implementation is added

## 7. Android Test Implementation

- [ ] 7.1 Add `#[cfg(mobile)]` variants of tests that need Android-specific behavior
- [ ] 7.2 Verify Android tests can invoke plugin commands through IPC
- [ ] 7.3 Run Android tests locally with emulator
- [ ] 7.4 Verify Android tests pass in CI
- [ ] 7.5 Make Android CI job a required check

## 8. Documentation

- [ ] 8.1 Update CLAUDE.md Integration Tests section with desktop workflows
- [x] 8.2 Document IPC testing approach (using `get_ipc_response()`, proper payload wrapping, etc.)
- [x] 8.3 Update `example-app/src-tauri/tests/README.md` with comprehensive instructions
- [x] 8.4 Document binary linking approach via build.rs and integration-test feature
- [x] 8.5 Document Android testing setup and requirements
- [x] 8.6 Add troubleshooting guide for common test issues
- [x] 8.7 Document platform coverage table (desktop active, mobile ready/documented)

## 9. Validation

- [x] 9.1 Run full desktop test suite with `task app:test-integration` (10 tests passing in 1.20s)
- [x] 9.2 Verify tests run consistently with --test-threads=1
- [ ] 9.3 Verify desktop CI job passes
- [x] 9.4 Verify test execution time is reasonable (all 10 tests in 1.20s)
- [ ] 9.5 Run `openspec validate add-integration-tests --strict`

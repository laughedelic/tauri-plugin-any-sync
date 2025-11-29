# Implementation Tasks

## 1. Foundation - Desktop Testing

- [ ] 1.1 Fix outdated paths in `.github/workflows/test.yml` (go-backend → plugin-go-backend, examples → example-app)
- [ ] 1.2 Replace `build-go-backend.sh` reference with `task backend:build` in desktop CI
- [ ] 1.3 Add Task installation step to CI workflow
- [ ] 1.4 Update test helper in `example-app/src-tauri/tests/integration.rs` to create webview for IPC testing

## 2. Desktop Basic Tests

- [ ] 2.1 Implement `test_ping_command` using `get_ipc_response()` - verify sidecar starts and gRPC works
- [ ] 2.2 Implement `test_ping_command_empty_message` - verify handling of None/empty messages
- [ ] 2.3 Run tests locally with `task app:test` to verify IPC infrastructure works
- [ ] 2.4 Verify desktop tests pass in CI

## 3. Desktop Storage Tests

- [ ] 3.1 Implement `test_storage_put_and_get` - create document, retrieve, verify JSON integrity
- [ ] 3.2 Implement `test_storage_get_nonexistent` - verify graceful handling of missing documents
- [ ] 3.3 Implement `test_storage_update_existing_document` - verify upsert behavior
- [ ] 3.4 Implement `test_storage_list` - put multiple docs, list collection, verify all IDs
- [ ] 3.5 Implement `test_storage_list_empty` - list non-existent collection, verify empty result
- [ ] 3.6 Implement `test_storage_delete` - delete existing doc, verify it's gone
- [ ] 3.7 Implement `test_storage_delete_nonexistent` - verify idempotent delete

## 4. Desktop Complex Scenarios

- [ ] 4.1 Implement `test_multiple_collections` - verify collection isolation
- [ ] 4.2 Implement `test_complex_json_document` - verify nested objects, arrays, Unicode, special chars
- [ ] 4.3 Verify all desktop tests pass locally and in CI

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

- [ ] 8.1 Update CLAUDE.md Integration Tests section with desktop and Android workflows
- [ ] 8.2 Document IPC testing approach (using `get_ipc_response()`)
- [ ] 8.3 Update `example-app/src-tauri/tests/README.md` with desktop and Android instructions
- [ ] 8.4 Document iOS setup guide for future use (clearly marked as not yet active)
- [ ] 8.5 Document CI jobs for desktop and Android; include commented iOS job
- [ ] 8.6 Add troubleshooting guide for Android emulator issues

## 9. Validation

- [ ] 9.1 Run full desktop test suite with `RUST_LOG=debug task app:test`
- [ ] 9.2 Run Android tests with emulator locally
- [ ] 9.3 Verify desktop and Android CI jobs pass
- [ ] 9.4 Verify CI execution time is reasonable (<8 minutes total for desktop + Android)
- [ ] 9.5 Confirm iOS documentation is clear about future activation
- [ ] 9.6 Run `openspec validate add-integration-tests --strict`

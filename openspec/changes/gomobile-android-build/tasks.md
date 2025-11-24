# Tasks: gomobile Android Build Integration

## Overview

Implementation tasks for integrating Go backend with Android via gomobile, organized to deliver incremental user-visible progress with validation at each step.

## Task List

### Phase 1: Go Mobile Backend Foundation

#### Task 1.1: Create mobile package structure
**Estimated effort:** 1 hour  
**Dependencies:** None  
**Deliverable:** Go package skeleton for mobile entrypoint

**Steps:**
1. Create `go-backend/cmd/mobile/` directory
2. Create `storage.go` with package declaration
3. Add `go.mod` replace directive if needed
4. Create basic README documenting gomobile API

**Validation:**
- [ ] Directory structure matches design
- [ ] Package compiles with `go build`

---

#### Task 1.2: Implement gomobile-compatible storage functions
**Estimated effort:** 3 hours  
**Dependencies:** Task 1.1  
**Deliverable:** Exported Go functions callable from Android

**Steps:**
1. Implement `InitStorage(dbPath string) error`
2. Implement `StoragePut(collection, id, documentJson string) error`
3. Implement `StorageGet(collection, id string) (string, error)`
4. Implement `StorageDelete(collection, id string) (bool, error)`
5. Implement `StorageList(collection string) (string, error)` - returns JSON array
6. Add internal state management for DB instance
7. Add proper error handling and logging

**Validation:**
- [ ] All functions compile without errors
- [ ] No complex types in signatures (only string, bool, error)
- [ ] Internal tests pass (if added)

---

#### Task 1.3: Test gomobile build locally
**Estimated effort:** 2 hours  
**Dependencies:** Task 1.2  
**Deliverable:** Working .aar artifact

**Steps:**
1. Install gomobile: `go install golang.org/x/mobile/cmd/gomobile@latest`
2. Run `gomobile init`
3. Build .aar: `gomobile bind -target=android -o test.aar ./cmd/mobile`
4. Inspect .aar contents (unzip and check .so files, Java classes)
5. Verify all Android ABIs present (arm64-v8a, armeabi-v7a, x86, x86_64)
6. Document any issues or warnings

**Validation:**
- [ ] .aar file builds successfully
- [ ] Contains libgojni.so for all ABIs
- [ ] Generated Java classes match expected API
- [ ] File size reasonable (<25MB)

---

### Phase 2: Android Plugin Integration

#### Task 2.1: Update Kotlin plugin structure
**Estimated effort:** 1 hour  
**Dependencies:** None (can parallel with Phase 1)  
**Deliverable:** Updated plugin with storage command handlers

**Steps:**
1. Open `android/src/main/java/ExamplePlugin.kt`
2. Add argument classes: `StorageGetArgs`, `StoragePutArgs`, `StorageDeleteArgs`, `StorageListArgs`
3. Add command method stubs: `@Command fun storageGet(invoke: Invoke)`
4. Add companion object with library loading: `System.loadLibrary("gojni")`
5. Add initialization logic in constructor

**Validation:**
- [ ] Kotlin code compiles
- [ ] Plugin structure follows Tauri conventions
- [ ] All 4 storage commands defined

---

#### Task 2.2: Implement JNI calls to Go backend
**Estimated effort:** 3 hours  
**Dependencies:** Task 1.3, Task 2.1  
**Deliverable:** Working JNI integration

**Steps:**
1. Place .aar in `android/libs/` for testing
2. Update `android/build.gradle.kts` to include .aar dependency
3. Import generated Mobile class: `import mobile.Mobile`
4. Implement `storageGet` calling `Mobile.storageGet()`
5. Implement `storagePut` calling `Mobile.storagePut()`
6. Implement `storageDelete` calling `Mobile.storageDelete()`
7. Implement `storageList` calling `Mobile.storageList()`
8. Add try-catch blocks and error propagation
9. Add Android logging for debugging

**Validation:**
- [ ] Code compiles with .aar dependency
- [ ] No compilation errors or warnings
- [ ] Error handling implemented for all commands

---

#### Task 2.3: Update Rust mobile module
**Estimated effort:** 2 hours  
**Dependencies:** Task 2.2  
**Deliverable:** Rust plugin dispatches to Kotlin

**Steps:**
1. Open `src/mobile.rs`
2. Update `AnySync::storage_get` to call Android plugin
3. Add similar methods for put, delete, list
4. Ensure proper serialization/deserialization
5. Map errors from Kotlin to Rust error types
6. Remove or update placeholder "ping" command if needed

**Validation:**
- [ ] Rust code compiles for Android target
- [ ] Mobile module mirrors desktop API
- [ ] Error types properly mapped

---

### Phase 3: Build System Integration

#### Task 3.1: Create gomobile build script
**Estimated effort:** 2 hours  
**Dependencies:** Task 1.3  
**Deliverable:** Automated .aar build script

**Steps:**
1. Create `build-go-mobile.sh` in repository root
2. Add functions: `check_gomobile()`, `build_android_aar()`
3. Build .aar with proper naming: `anysync-mobile.aar`
4. Generate SHA256 checksum
5. Add option for architecture-specific builds
6. Test script on clean machine (or clean environment)

**Validation:**
- [ ] Script runs without errors
- [ ] Produces .aar in `binaries/` directory
- [ ] Checksum file generated
- [ ] Documentation added to README

---

#### Task 3.2: Update CI workflow for Android builds
**Estimated effort:** 2 hours  
**Dependencies:** Task 3.1  
**Deliverable:** CI builds .aar artifacts

**Steps:**
1. Open `.github/workflows/release.yml`
2. Add Android build job (similar to desktop builds)
3. Install gomobile in CI environment
4. Run `build-go-mobile.sh --cross` (if multi-arch)
5. Upload .aar and checksums to GitHub Release
6. Test workflow on a test tag

**Validation:**
- [ ] CI job completes successfully
- [ ] .aar uploaded to release assets
- [ ] Checksum file present
- [ ] Download link accessible

---

#### Task 3.3: Implement .aar download in plugin build.rs
**Estimated effort:** 2 hours  
**Dependencies:** Task 3.2  
**Deliverable:** Plugin downloads .aar during build

**Steps:**
1. Open `build.rs`
2. Add Android-specific build logic (with `#[cfg(target_os = "android")]`)
3. Implement `download_android_aar()` function (similar to binary download)
4. Download .aar from GitHub releases
5. Verify checksum
6. Place in appropriate target directory
7. Emit cargo metadata for consumer build.rs
8. Test with example app

**Validation:**
- [ ] Plugin builds successfully for Android
- [ ] .aar downloaded automatically
- [ ] Checksum verification works
- [ ] Metadata emitted correctly

---

### Phase 4: Example App Android Support

#### Task 4.1: Initialize Android support in example app
**Estimated effort:** 1 hour  
**Dependencies:** None (can parallel)  
**Deliverable:** Android project structure

**Steps:**
1. `cd examples/tauri-app`
2. Run `tauri android init`
3. Accept defaults or configure as needed
4. Commit generated files (except large binaries)
5. Document any manual configuration needed

**Validation:**
- [ ] `gen/android/` directory created
- [ ] Android project structure valid
- [ ] Gradle builds successfully

---

#### Task 4.2: Update example app build configuration
**Estimated effort:** 2 hours  
**Dependencies:** Task 3.3, Task 4.1  
**Deliverable:** Example app builds for Android

**Steps:**
1. Open `examples/tauri-app/src-tauri/build.rs`
2. Add Android-specific logic to copy .aar
3. Read `DEP_TAURI_PLUGIN_ANY_SYNC_AAR_PATH` env var
4. Copy .aar to `gen/android/libs/`
5. Update `gen/android/app/build.gradle` if needed
6. Test build: `tauri android build`

**Validation:**
- [ ] Build completes successfully
- [ ] .aar present in Android project
- [ ] APK generated
- [ ] APK size reasonable

---

#### Task 4.3: Test on Android emulator
**Estimated effort:** 3 hours  
**Dependencies:** Task 4.2, Task 2.3  
**Deliverable:** Working app on Android

**Steps:**
1. Create Android Virtual Device (AVD) if needed (API 24+, arm64 or x86_64)
2. Start emulator
3. Run `tauri android dev`
4. Test storage operations:
   - Put a document
   - Get the document
   - List documents
   - Delete document
5. Check logcat for errors
6. Test error cases (invalid JSON, missing docs)
7. Document any issues found

**Validation:**
- [ ] App launches on emulator
- [ ] UI renders correctly
- [ ] All storage operations work
- [ ] Errors handled gracefully
- [ ] Data persists across app restarts

---

### Phase 5: Documentation and Validation

#### Task 5.1: Document Android build process
**Estimated effort:** 2 hours  
**Dependencies:** All previous tasks  
**Deliverable:** Complete documentation

**Steps:**
1. Update `README.md` with Android build instructions
2. Update `android/AGENTS.md` with gomobile details
3. Document required tools (Android SDK, gomobile)
4. Add troubleshooting section
5. Document known limitations
6. Add example output/screenshots

**Validation:**
- [ ] Documentation complete and accurate
- [ ] Another developer can follow instructions
- [ ] Troubleshooting section helpful

---

#### Task 5.2: Cross-platform validation
**Estimated effort:** 2 hours  
**Dependencies:** Task 4.3  
**Deliverable:** Verified cross-platform consistency

**Steps:**
1. Run example app on desktop (macOS/Linux/Windows)
2. Run example app on Android emulator
3. Perform same operations on both platforms
4. Verify API behavior identical
5. Test error handling consistency
6. Document any platform differences found

**Validation:**
- [ ] Same TypeScript code works on both
- [ ] Data format compatible
- [ ] Error messages similar
- [ ] Performance acceptable on Android

---

#### Task 5.3: Update project specs
**Estimated effort:** 1 hour  
**Dependencies:** All previous tasks  
**Deliverable:** Spec deltas applied

**Steps:**
1. Review all spec deltas in this change
2. Verify all requirements met
3. Update any specs that need refinement
4. Prepare for archiving change

**Validation:**
- [ ] All requirements implemented
- [ ] All scenarios pass
- [ ] Specs accurate and complete

---

## Task Summary

**Total Estimated Effort:** ~27 hours (~4 days)

**Parallelizable Work:**
- Phase 1 (Go) and Phase 2.1 (Kotlin structure) can run in parallel
- Phase 4.1 (Android init) can start early

**Critical Path:**
- Phase 1 → Phase 2 → Phase 3 → Phase 4 (linear dependency)

**Risk Areas:**
- Task 1.3: First gomobile build might reveal issues
- Task 2.2: JNI integration sometimes has subtle issues
- Task 4.3: Emulator testing can uncover platform-specific bugs

**Incremental Validation:**
Each phase delivers working artifacts that can be tested independently before proceeding to the next phase.

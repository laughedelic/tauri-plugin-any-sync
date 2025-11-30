# Cherry-Pick Guide: Useful Changes from Integration Testing Branch

This document identifies changes from the `feat/add-integration-testing` branch that are valuable regardless of the specific testing implementation approach.

## ✅ Changes Worth Keeping

### 1. **Android Build Configuration** (`.cargo/config.toml`)
**Status**: ✅ Keep entirely

**Why**: Essential for any Android development. Configures NDK linkers for all Android architectures.

**File**: `.cargo/config.toml` (new file)
```toml
# Android NDK linker configuration
[target.x86_64-linux-android]
linker = "x86_64-linux-android21-clang"
# ... etc for all Android targets
```

**How to cherry-pick**: Copy the entire file
```bash
git checkout feat/add-integration-testing -- .cargo/config.toml
```

---

### 2. **Plugin Build Script Improvements** (`plugin-rust-core/build.rs`)
**Status**: ✅ Keep the protobuf generation fix

**Why**: Correctly skips protobuf generation for mobile targets (which use FFI, not gRPC). This prevents build failures on Android/iOS.

**Change**:
```rust
// Generate protobuf code for desktop targets only (mobile uses FFI, not gRPC)
let target = std::env::var("TARGET").unwrap_or_default();
let is_mobile_target = target.contains("android") || target.contains("ios");

if !is_mobile_target {
    if let Err(e) = generate_protobuf() {
        eprintln!("Error: Failed to generate protobuf code: {}", e);
        std::process::exit(1);
    }
}
```

**How to cherry-pick**: Use interactive staging
```bash
git show feat/add-integration-testing:plugin-rust-core/build.rs > /tmp/build.rs
# Manually apply just the protobuf generation section
```

---

### 3. **Task Configuration Updates**
**Status**: ✅ Keep selectively

#### Root `Taskfile.yml`
- **Keep**: New `test` task that orchestrates all tests
- **Consider**: The commented-out `output: prefixed` (personal preference)

```yaml
test:
  desc: Run all tests
  deps:
    - backend:test
    - core:test
    - app:test-integration  # Can be removed if not implementing tests
```

#### `example-app/Taskfile.yml`
- **Keep**: `init:android` task (useful for Android development)
- **Skip**: `test-integration` tasks (specific to MockRuntime approach)

**How to cherry-pick**:
```bash
# For root Taskfile
git diff main feat/add-integration-testing -- Taskfile.yml > /tmp/taskfile.patch
# Manually apply just the `test` task

# For example-app Taskfile
git checkout feat/add-integration-testing -- example-app/Taskfile.yml
# Then remove the test-integration tasks if not needed
```

---

### 4. **Documentation Updates** (`AGENTS.md`)
**Status**: ✅ Keep structure, update content

**Why**: The integration testing section provides valuable context about testing strategy, even if implementation changes.

**What to keep**:
- The "Integration Tests" section header
- Location information
- Platform support table (update status as needed)
- Running instructions (modify for new approach)

**What to update**:
- Change status from "Active" to "Planned" or remove mobile entries
- Update test coverage description if implementation differs

**How to cherry-pick**:
```bash
git diff main feat/add-integration-testing -- AGENTS.md > /tmp/agents.patch
# Manually extract and adapt the Integration Tests section
```

---

### 5. **Mobile Backend Task Improvements** (`plugin-go-backend/mobile/Taskfile.yml`)
**Status**: ⚠️ Review changes

**Changes**: Minor formatting or build improvements

**How to check**:
```bash
git diff main feat/add-integration-testing -- plugin-go-backend/mobile/Taskfile.yml
```

If there are useful build optimizations, keep them. If just whitespace, skip.

---

## ❌ Changes to Skip (Testing-Specific)

### 1. **Complete Integration Test Implementation**
- `example-app/src-tauri/tests/integration.rs` (711 lines)
- `example-app/src-tauri/tests/README.md` (319 lines)

**Why skip**: These are specific to the MockRuntime approach which doesn't work on Android. The new approach will require different test structure.

---

### 2. **Test-Specific Cargo Configuration**
- `example-app/src-tauri/Cargo.toml` - `integration-test` feature
- `example-app/src-tauri/build.rs` - Test binary symlinking logic

**Why skip**: Designed for the current test approach. New approach may need different setup.

---

### 3. **CI Workflow for Integration Tests**
- `.github/workflows/test.yml` - `test-integration-desktop` and `test-integration-android` jobs

**Why skip**: These jobs are specific to the MockRuntime test implementation. Keep the structural improvements (formatting, comments) but remove the integration test jobs.

**Exception**: If you implement build validation only, you might want the Android build job structure as a starting point.

---

### 4. **OpenSpec Documentation**
- `openspec/changes/add-integration-tests/` (entire directory)

**Why skip**: This documents the specific implementation approach being deferred. Keep `TESTING_PIVOT.md` as a reference for why the approach changed.

---

## 📋 Recommended Cherry-Pick Strategy

### Option A: Interactive Rebase (Most Control)
```bash
# Create a new branch from main
git checkout main
git checkout -b cherry-picked-android-improvements

# Interactive rebase to pick specific commits
git cherry-pick --no-commit feat/add-integration-testing
git reset HEAD  # Unstage everything
git add -p      # Interactively stage only desired changes

# Or use individual files
git checkout feat/add-integration-testing -- .cargo/config.toml
# ... manually edit plugin-rust-core/build.rs, etc.

git commit -m "feat: add Android build configuration and mobile target improvements"
```

### Option B: Manual File-by-File (Most Precise)
```bash
git checkout main
git checkout -b android-build-improvements

# 1. Android config
git checkout feat/add-integration-testing -- .cargo/config.toml

# 2. Plugin build script - manual edit needed
git show feat/add-integration-testing:plugin-rust-core/build.rs > /tmp/build.rs
# Open both files, manually copy just the mobile target check

# 3. Documentation - manual edit
git show feat/add-integration-testing:AGENTS.md > /tmp/agents.md
# Extract integration testing section, modify status

git add .cargo/config.toml plugin-rust-core/build.rs AGENTS.md
git commit -m "feat: add Android build configuration

- Add .cargo/config.toml with NDK linker settings
- Skip protobuf generation for mobile targets in plugin build
- Document integration testing strategy (deferred implementation)"
```

### Option C: Squash Merge Specific Files (Simplest)
```bash
git checkout main
git checkout -b essential-android-improvements

# Cherry-pick just the essential files
git checkout feat/add-integration-testing -- .cargo/config.toml

# For build.rs, use difftool to selectively merge
git difftool main feat/add-integration-testing -- plugin-rust-core/build.rs

git commit -m "feat: essential Android build improvements"
```

---

## 🎯 Minimal Recommended Cherry-Pick

If you want just the essentials with minimal effort:

```bash
git checkout main
git checkout -b android-essentials

# Only these two files are essential and don't need modification
git checkout feat/add-integration-testing -- .cargo/config.toml

# For build.rs, apply the mobile target check manually (see above)
# It's about 8 lines of code

git commit -m "feat: add Android build essentials

- Configure NDK linkers for Android targets
- Skip protobuf generation on mobile platforms"
```

---

## 📊 Summary Table

| File/Directory                             | Action                      | Effort | Value  |
| ------------------------------------------ | --------------------------- | ------ | ------ |
| `.cargo/config.toml`                       | ✅ Copy entire file          | Low    | High   |
| `plugin-rust-core/build.rs`                | ⚠️ Manual merge ~8 lines     | Medium | High   |
| `Taskfile.yml`                             | ⚠️ Copy `test` task          | Low    | Medium |
| `example-app/Taskfile.yml`                 | ⚠️ Copy `init:android`       | Low    | Medium |
| `AGENTS.md`                                | ⚠️ Adapt integration section | Medium | Medium |
| `.github/workflows/test.yml`               | ⚠️ Keep formatting only      | Low    | Low    |
| `example-app/src-tauri/tests/*`            | ❌ Skip                      | -      | -      |
| `openspec/changes/add-integration-tests/*` | ❌ Skip                      | -      | -      |
| `TESTING_PIVOT.md`                         | ⚠️ Keep as reference         | None   | Medium |

**Legend**:
- ✅ = Keep entirely
- ⚠️ = Keep partially / needs review
- ❌ = Skip entirely

---

## 🔄 Next Steps After Cherry-Picking

1. **Test Android build** after cherry-picking:
   ```bash
   task go:mobile:build-android
   cargo build --target aarch64-linux-android
   ```

2. **Update documentation** to reflect deferred testing implementation

3. **Consider implementing** the simplified testing approach from `TESTING_PIVOT.md`:
   - Build validation CI job for Android
   - Optional: Plugin isolation tests on host machine

4. **Close or update** the integration testing PR with explanation of the pivot

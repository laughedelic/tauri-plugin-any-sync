# Implementation Tasks

## 1. GitHub Release Workflow

- [ ] 1.1 Create `.github/workflows/release.yml` workflow
- [ ] 1.2 Configure workflow to trigger on version tags (v*)
- [ ] 1.3 Add build job using `build-go-backend.sh --cross` for all platforms
- [ ] 1.4 Generate SHA256 checksums for each binary artifact
- [ ] 1.5 Create checksums.txt file with format: `<hash>  <filename>`
- [ ] 1.6 Upload binaries and checksums.txt to GitHub release assets
- [ ] 1.7 Test workflow on a test tag/release
- [ ] 1.8 Verify all binaries and checksums.txt are uploaded correctly

## 2. Plugin Build Script Enhancement

- [ ] 2.1 Add `reqwest` (blocking feature) to build-dependencies in `Cargo.toml`
- [ ] 2.2 Add `sha2` crate for checksum verification to build-dependencies
- [ ] 2.3 Add environment variable handling in `build.rs`
  - [ ] 2.3.1 Check for ANY_SYNC_GO_BINARIES_DIR environment variable
  - [ ] 2.3.2 Emit `cargo:rerun-if-env-changed=ANY_SYNC_GO_BINARIES_DIR`
  - [ ] 2.3.3 Validate local path exists if environment variable is set
- [ ] 2.4 Implement `copy_local_binaries()` function for local development mode
  - [ ] 2.4.1 Copy binaries from local path to OUT_DIR
  - [ ] 2.4.2 Emit warning about using local binaries
  - [ ] 2.4.3 Handle invalid paths with clear error messages
- [ ] 2.5 Implement `download_binaries()` function for consumer/CI mode
  - [ ] 2.5.1 Determine version from Cargo.toml (CARGO_PKG_VERSION)
  - [ ] 2.5.2 Construct GitHub Release URL for the version tag
  - [ ] 2.5.3 Detect enabled features (via cfg! macros) to determine which binaries to download
  - [ ] 2.5.4 Download platform binaries based on enabled features from GitHub release assets
  - [ ] 2.5.5 Download checksums.txt from release assets
  - [ ] 2.5.6 Verify SHA256 checksums for each downloaded binary
  - [ ] 2.5.7 Store verified binaries in `OUT_DIR/binaries/`
  - [ ] 2.5.8 Fail build with clear error messages on download/checksum failures (no fallback)
- [ ] 2.6 Emit `cargo:binaries_dir=<dir>` for consumer propagation (both modes)

## 3. Cargo Configuration

- [ ] 3.1 Add `links = "any_sync_go"` to `[package]` section in `Cargo.toml`
- [ ] 3.2 Define feature flags in `Cargo.toml` matching go-backend build targets:
  - [ ] 3.2.1 Individual target features:
    - `x86_64-apple-darwin` (macOS Intel)
    - `aarch64-apple-darwin` (macOS Apple Silicon)
    - `x86_64-unknown-linux-gnu` (Linux x64)
    - `aarch64-unknown-linux-gnu` (Linux ARM64)
    - `x86_64-pc-windows-msvc` (Windows x64)
  - [ ] 3.2.2 Platform group features:
    - `macos` = ["x86_64-apple-darwin", "aarch64-apple-darwin"]
    - `linux` = ["x86_64-unknown-linux-gnu", "aarch64-unknown-linux-gnu"]
    - `windows` = ["x86_64-pc-windows-msvc"]
  - [ ] 3.2.3 `all` = ["macos", "linux", "windows"] (all platforms)
- [ ] 3.3 Set no default features (users must explicitly choose or features will be empty)
- [ ] 3.4 Update `include` field to include downloaded binaries in package

## 4. Documentation Updates

- [ ] 4.1 Update README.md installation section:
  - [ ] 4.1.1 Add consumer build script example with DEP_ANY_SYNC_GO_BINARIES_DIR
  - [ ] 4.1.2 Document feature selection (individual targets, platform groups, all)
  - [ ] 4.1.3 Show example Cargo.toml dependency with features
  - [ ] 4.1.4 Document ANY_SYNC_GO_BINARIES_DIR environment variable for local development
  - [ ] 4.1.5 Add example .cargo/config.toml configuration for persistent local dev setup
  - [ ] 4.1.6 Explain two-mode architecture (consumer/CI vs local development)
  - [ ] 4.1.7 Keep existing manual download instructions as legacy alternative
  - [ ] 4.1.8 Add troubleshooting section:
    - Download failures
    - Checksum verification failures
    - Invalid local binaries path
    - Feature selection guidance
- [ ] 4.2 Update `AGENTS.md` with binary distribution architecture notes
  - [ ] 4.2.1 Document two modes: consumer/CI mode vs local development mode
  - [ ] 4.2.2 Explain environment variable override pattern
  - [ ] 4.2.3 Update "Build System Integration" section with new download approach
  - [ ] 4.2.4 Document feature selection strategy for different use cases
- [ ] 4.3 Add inline code comments in build.rs explaining:
  - [ ] 4.3.1 The Cargo links mechanism and metadata propagation
  - [ ] 4.3.2 Environment variable handling logic
  - [ ] 4.3.3 Feature detection for platform-specific downloads
  - [ ] 4.3.4 Checksum verification process

## 5. Testing and Validation

- [ ] 5.1 Test consumer/CI mode (without ANY_SYNC_GO_BINARIES_DIR)
  - [ ] 5.1.1 Verify binaries download from GitHub releases for specified version
  - [ ] 5.1.2 Test checksum verification rejects corrupted/modified downloads
  - [ ] 5.1.3 Test download failure produces helpful error message (no fallback)
  - [ ] 5.1.4 Test missing checksums.txt produces clear error
  - [ ] 5.1.5 Verify Cargo caching works (second build reuses downloads)
- [ ] 5.2 Test local development mode (with ANY_SYNC_GO_BINARIES_DIR)
  - [ ] 5.2.1 Set environment variable pointing to ./binaries
  - [ ] 5.2.2 Verify local binaries are copied to OUT_DIR
  - [ ] 5.2.3 Verify warning message is emitted
  - [ ] 5.2.4 Test invalid path produces clear error
- [ ] 5.3 Test consumer build script in example app
  - [ ] 5.3.1 Verify DEP_ANY_SYNC_GO_BINARIES_DIR propagation
  - [ ] 5.3.2 Test binary copying to src-tauri/binaries/
- [ ] 5.4 Test with different feature combinations
  - [ ] 5.4.1 No features (should skip downloads)
  - [ ] 5.4.2 `all` feature (downloads all platforms)
  - [ ] 5.4.3 Platform groups (`macos`, `linux`, `windows`)
  - [ ] 5.4.4 Individual target features (e.g., `x86_64-apple-darwin` only)
  - [ ] 5.4.5 Mixed features (e.g., `macos` + `windows`)
- [ ] 5.5 Test cross-compilation scenarios
- [ ] 5.6 Test .cargo/config.toml persistent configuration

## 6. Migration Support

- [ ] 6.1 Add deprecation notice to manual download instructions
- [ ] 6.2 Ensure backward compatibility with existing manual setup
- [ ] 6.3 Create migration guide for existing users

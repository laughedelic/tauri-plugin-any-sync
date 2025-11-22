# Implementation Tasks

## 1. GitHub Release Workflow

- [x] 1.1 Create `.github/workflows/release.yml` workflow
- [x] 1.2 Configure workflow to trigger on version tags (v*)
- [x] 1.3 Add build job using `build-go-backend.sh --cross` for all platforms
- [x] 1.4 Generate SHA256 checksums for each binary artifact
- [x] 1.5 Create checksums.txt file with format: `<hash>  <filename>`
- [x] 1.6 Upload binaries and checksums.txt to GitHub release assets
- [x] 1.7 Test workflow on a test tag/release
- [x] 1.8 Verify all binaries and checksums.txt are uploaded correctly

## 2. Plugin Build Script Enhancement

- [x] 2.1 Add `reqwest` (blocking feature) to build-dependencies in `Cargo.toml`
- [x] 2.2 Add `sha2` crate for checksum verification to build-dependencies
- [x] 2.3 Add environment variable handling in `build.rs`
  - [x] 2.3.1 Check for ANY_SYNC_GO_BINARIES_DIR environment variable
  - [x] 2.3.2 Emit `cargo:rerun-if-env-changed=ANY_SYNC_GO_BINARIES_DIR`
  - [x] 2.3.3 Validate local path exists if environment variable is set
- [x] 2.4 Implement `copy_local_binaries()` function for local development mode
  - [x] 2.4.1 Copy binaries from local path to OUT_DIR
  - [x] 2.4.2 Emit warning about using local binaries
  - [x] 2.4.3 Handle invalid paths with clear error messages
- [x] 2.5 Implement `download_binaries()` function for consumer/CI mode
  - [x] 2.5.1 Determine version from Cargo.toml (CARGO_PKG_VERSION)
  - [x] 2.5.2 Construct GitHub Release URL for the version tag
  - [x] 2.5.3 Detect enabled features (via cfg! macros) to determine which binaries to download
  - [x] 2.5.4 Download platform binaries based on enabled features from GitHub release assets
  - [x] 2.5.5 Download checksums.txt from release assets
  - [x] 2.5.6 Verify SHA256 checksums for each downloaded binary
  - [x] 2.5.7 Store verified binaries in `OUT_DIR/binaries/`
  - [x] 2.5.8 Fail build with clear error messages on download/checksum failures (no fallback)
- [x] 2.6 Emit `cargo:binaries_dir=<dir>` for consumer propagation (both modes)

## 3. Cargo Configuration

- [x] 3.1 Add `links = "any_sync_go"` to `[package]` section in `Cargo.toml`
- [x] 3.2 Define feature flags in `Cargo.toml` matching go-backend build targets:
  - [x] 3.2.1 Individual target features:
    - `x86_64-apple-darwin` (macOS Intel)
    - `aarch64-apple-darwin` (macOS Apple Silicon)
    - `x86_64-unknown-linux-gnu` (Linux x64)
    - `aarch64-unknown-linux-gnu` (Linux ARM64)
    - `x86_64-pc-windows-msvc` (Windows x64)
  - [x] 3.2.2 Platform group features:
    - `macos` = ["x86_64-apple-darwin", "aarch64-apple-darwin"]
    - `linux` = ["x86_64-unknown-linux-gnu", "aarch64-unknown-linux-gnu"]
    - `windows` = ["x86_64-pc-windows-msvc"]
  - [x] 3.2.3 `all` = ["macos", "linux", "windows"] (all platforms)
- [x] 3.3 Set no default features (users must explicitly choose or features will be empty)
- [x] 3.4 Update `include` field to include downloaded binaries in package

## 4. Documentation Updates

- [x] 4.1 Update README.md installation section:
  - [x] 4.1.1 Add consumer build script example with DEP_ANY_SYNC_GO_BINARIES_DIR
  - [x] 4.1.2 Document feature selection (individual targets, platform groups, all)
  - [x] 4.1.3 Show example Cargo.toml dependency with features
  - [x] 4.1.4 Document ANY_SYNC_GO_BINARIES_DIR environment variable for local development
  - [x] 4.1.5 Add example .cargo/config.toml configuration for persistent local dev setup
  - [x] 4.1.6 Explain two-mode architecture (consumer/CI vs local development)
  - [x] 4.1.7 Keep existing manual download instructions as legacy alternative
  - [x] 4.1.8 Add troubleshooting section:
    - Download failures
    - Checksum verification failures
    - Invalid local binaries path
    - Feature selection guidance
- [x] 4.2 Update `AGENTS.md` with binary distribution architecture notes
  - [x] 4.2.1 Document two modes: consumer/CI mode vs local development mode
  - [x] 4.2.2 Explain environment variable override pattern
  - [x] 4.2.3 Update "Build System Integration" section with new download approach
  - [x] 4.2.4 Document feature selection strategy for different use cases
- [x] 4.3 Add inline code comments in build.rs explaining:
  - [x] 4.3.1 The Cargo links mechanism and metadata propagation
  - [x] 4.3.2 Environment variable handling logic
  - [x] 4.3.3 Feature detection for platform-specific downloads
  - [x] 4.3.4 Checksum verification process

## 5. Testing and Validation

- [x] 5.1 Test consumer/CI mode (without ANY_SYNC_GO_BINARIES_DIR)
  - [x] 5.1.1 Verify binaries download from GitHub releases for specified version
  - [x] 5.1.2 Test checksum verification rejects corrupted/modified downloads
  - [x] 5.1.3 Test download failure produces helpful error message (no fallback)
  - [x] 5.1.4 Test missing checksums.txt produces clear error
  - [x] 5.1.5 Verify Cargo caching works (second build reuses downloads)
- [x] 5.2 Test local development mode (with ANY_SYNC_GO_BINARIES_DIR)
  - [x] 5.2.1 Set environment variable pointing to ./binaries
  - [x] 5.2.2 Verify local binaries are copied to OUT_DIR
  - [x] 5.2.3 Verify warning message is emitted
  - [x] 5.2.4 Test invalid path produces clear error
- [x] 5.3 Test consumer build script in example app
  - [x] 5.3.1 Verify DEP_ANY_SYNC_GO_BINARIES_DIR propagation
  - [x] 5.3.2 Test binary copying to src-tauri/binaries/
- [x] 5.4 Test with different feature combinations
  - [x] 5.4.1 No features (should skip downloads)
  - [x] 5.4.2 `all` feature (downloads all platforms)
  - [x] 5.4.3 Platform groups (`macos`, `linux`, `windows`)
  - [x] 5.4.4 Individual target features (e.g., `x86_64-apple-darwin` only)
  - [x] 5.4.5 Mixed features (e.g., `macos` + `windows`)
- [x] 5.5 Test cross-compilation scenarios
- [x] 5.6 Test .cargo/config.toml persistent configuration

## 6. Migration Support

- [x] 6.1 Add deprecation notice to manual download instructions
- [x] 6.2 Ensure backward compatibility with existing manual setup
- [x] 6.3 Create migration guide for existing users

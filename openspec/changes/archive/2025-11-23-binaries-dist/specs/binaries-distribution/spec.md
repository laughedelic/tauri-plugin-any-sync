# Binaries Distribution Specification

## ADDED Requirements

### Requirement: GitHub Release Workflow
The plugin SHALL publish pre-compiled Go backend binaries to GitHub Releases on version tags.
#### Scenario: Release creation
- **WHEN** a version tag (v*) is pushed to the repository
- **THEN** a GitHub Actions workflow builds all platform binaries using build-go-backend.sh
- **AND** generates SHA256 checksums for each binary
- **AND** creates a GitHub Release with binaries and checksums as assets

### Requirement: Binary Download with Verification
The plugin's build script SHALL download platform-specific binaries from GitHub Releases with checksum verification when not using local binaries.
#### Scenario: Successful binary download
- **WHEN** the plugin is compiled as a dependency
- **AND** ANY_SYNC_GO_BINARIES_DIR environment variable is not set
- **THEN** build.rs downloads the required platform binaries from the matching GitHub Release version
- **AND** verifies SHA256 checksums before using the binaries
- **AND** stores binaries in OUT_DIR for propagation to consumers

#### Scenario: Download failure handling
- **WHEN** binary download fails due to network issues or rate limiting
- **AND** ANY_SYNC_GO_BINARIES_DIR environment variable is not set
- **THEN** build.rs fails the build with a clear error message
- **AND** suggests setting ANY_SYNC_GO_BINARIES_DIR for offline development
- **AND** provides troubleshooting guidance for network issues

### Requirement: Cargo Links Metadata Propagation
The plugin SHALL use Cargo's links mechanism to broadcast binary paths to consuming applications.
#### Scenario: Metadata broadcasting
- **WHEN** the plugin's build script completes successfully
- **THEN** it emits cargo:binaries_dir=<path> to standard output
- **AND** Cargo propagates this as DEP_TAURI_PLUGIN_ANY_SYNC_BINARIES_DIR environment variable to consumer build scripts

### Requirement: Platform-Selective Downloads via Features
The plugin SHALL provide Cargo features for selective platform binary downloads.
#### Scenario: Feature-based platform selection
- **WHEN** a consumer specifies features in their Cargo.toml dependency declaration
- **THEN** the plugin downloads only the binaries for enabled platforms
- **AND** supports features: all, macos, linux, windows, and platform-specific targets

#### Scenario: No default features
- **WHEN** no features are explicitly specified by the consumer
- **THEN** no platform binaries are downloaded (no default features)
- **AND** users must explicitly select which platforms they need

### Requirement: Checksum Verification Security
The plugin SHALL verify the integrity of downloaded binaries using SHA256 checksums.
#### Scenario: Valid checksum verification
- **WHEN** binaries are downloaded from GitHub Releases
- **THEN** build.rs downloads the checksums.txt file from the same release
- **AND** verifies each binary's SHA256 hash matches the published checksum
- **AND** proceeds with build only if all checksums match

#### Scenario: Checksum mismatch handling
- **WHEN** a downloaded binary's checksum does not match the published value
- **THEN** build.rs rejects the binary and fails the build
- **AND** provides a clear error message indicating the checksum mismatch
- **AND** suggests potential MITM attack or corrupted download

### Requirement: Consumer Build Script Integration
The plugin SHALL provide documentation and examples for consumer build scripts to copy binaries to their application.
#### Scenario: Consumer build script pattern
- **WHEN** a consumer application builds with the plugin as a dependency
- **THEN** their build.rs can read DEP_TAURI_PLUGIN_ANY_SYNC_BINARIES_DIR
- **AND** copy the appropriate platform binaries to their src-tauri/binaries/ directory
- **AND** rename binaries following Tauri's sidecar naming convention (binary-<target-triple>)

### Requirement: Local Development Mode
The plugin SHALL support local development using environment variable override to bypass GitHub downloads.
#### Scenario: Local binaries override
- **WHEN** ANY_SYNC_GO_BINARIES_DIR environment variable is set
- **THEN** build.rs uses binaries from the specified local directory
- **AND** copies them to OUT_DIR instead of downloading from GitHub
- **AND** emits a warning indicating local binaries are being used
- **AND** validates that the specified directory exists and contains binaries

#### Scenario: Invalid local binaries path
- **WHEN** ANY_SYNC_GO_BINARIES_DIR is set but path doesn't exist
- **THEN** build.rs fails with a clear error message
- **AND** indicates the invalid path
- **AND** suggests verifying the path or unsetting the environment variable

#### Scenario: Developer workflow
- **WHEN** a developer modifies Go backend code
- **THEN** they can set ANY_SYNC_GO_BINARIES_DIR=./binaries (or via .cargo/config.toml)
- **AND** run cargo build to immediately test changes
- **AND** binaries are copied to OUT_DIR maintaining the same downstream flow as downloaded binaries

### Requirement: Build Caching
The plugin SHALL leverage Cargo's incremental build system for caching.
#### Scenario: Cargo-based caching
- **WHEN** binaries have been previously downloaded for a specific version
- **THEN** Cargo's incremental build system avoids re-running build.rs
- **AND** binaries stored in OUT_DIR are reused automatically
- **AND** reduces build time and bandwidth usage on subsequent builds
**Note:** Explicit cache checking in build.rs is not implemented; we rely on Cargo's built-in incremental build mechanism which skips build.rs re-execution when dependencies haven't changed.

### Requirement: Cross-Compilation Support
The plugin SHALL support cross-compilation scenarios where target platform differs from host.
#### Scenario: Cross-compilation binary selection
- **WHEN** building for a different target platform than the host (cross-compilation)
- **THEN** build.rs downloads binaries for the TARGET platform
- **AND** ensures consumer build scripts receive the correct platform binaries
- **AND** supports common cross-compilation targets (macOS arm64 from x64, Linux from macOS, etc.)

### Requirement: Versioned Binary Distribution
The plugin SHALL tie binary versions to plugin crate versions for reproducible builds.
#### Scenario: Version-locked binary downloads
- **WHEN** a consumer depends on a specific plugin version
- **THEN** build.rs downloads binaries from the matching GitHub Release tag
- **AND** ensures consistent binary versions across builds
- **AND** prevents version mismatch between plugin code and binaries

### Requirement: Error Reporting and Diagnostics
The plugin SHALL provide clear error messages for binary distribution failures.
#### Scenario: Download error diagnostics
- **WHEN** binary download fails
- **THEN** build.rs outputs a detailed error message including:
  - The GitHub Release URL being accessed
  - The specific binary that failed to download
  - Network error details or HTTP status codes
  - Suggested remediation steps (check network, GitHub status, use local build)

#### Scenario: Build progress feedback
- **WHEN** downloading binaries during build
- **THEN** build.rs outputs basic progress messages indicating:
  - Which binaries are being downloaded
**Note:** Detailed download progress bars and verbose checksum verification logging are not currently implemented. Basic logging (which binary is downloading) is considered sufficient for now.

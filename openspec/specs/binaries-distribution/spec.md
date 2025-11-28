# binaries-distribution Specification

## Purpose
TBD - created by archiving change binaries-dist. Update Purpose after archive.
## Requirements
### Requirement: GitHub Release Workflow
The plugin SHALL publish pre-compiled Go backend binaries **for all supported platforms** to GitHub Releases on version tags.

#### Scenario: Release creation **with mobile artifacts**
- **WHEN** a version tag (v*) is pushed to the repository
- **THEN** a GitHub Actions workflow builds all **desktop** platform binaries using build-go-backend.sh
- **AND** **builds Android .aar using gomobile bind**
- **AND** generates SHA256 checksums for **all artifacts (desktop + mobile)**
- **AND** creates a GitHub Release with **all** binaries and checksums as assets

### Requirement: Binary Download with Verification
The plugin's build script SHALL download platform-specific binaries **including mobile platforms** from GitHub Releases with checksum verification when not using local binaries.

#### Scenario: **Android binary download**
- **WHEN** the plugin is compiled **for Android target**
- **AND** ANY_SYNC_GO_BINARIES_DIR environment variable is not set
- **THEN** build.rs downloads **`any-sync-android.aar`** from the matching GitHub Release version
- **AND** verifies SHA256 checksums before using the **artifact**
- **AND** stores **artifact** in OUT_DIR for propagation to consumers

### Requirement: Cargo Links Metadata Propagation
The plugin SHALL use Cargo's links mechanism to broadcast binary paths to consuming applications.

#### Scenario: Metadata broadcasting
- **WHEN** the plugin's build script completes successfully
- **THEN** it emits cargo:**binaries_dir**=<path> to standard output
- **AND** Cargo propagates this as DEP_TAURI_PLUGIN_ANY_SYNC_**BINARIES_DIR** environment variable to consumer build scripts
- **AND** this single metadata value covers both desktop binaries and Android .aar (all in same directory)

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
The plugin SHALL provide documentation and examples for consumer build scripts to link or copy binaries to their application.
#### Scenario: Consumer build script pattern
- **WHEN** a consumer application builds with the plugin as a dependency
- **THEN** their build.rs can read DEP_TAURI_PLUGIN_ANY_SYNC_BINARIES_DIR
- **AND** create a symlink (Unix) or copy files (Windows) to their src-tauri/binaries/ directory
- **AND** binaries follow Tauri's sidecar naming convention (binary-<target-triple>)
- **AND** .taurignore prevents rebuild loops from file watcher

### Requirement: Local Development Mode
The plugin SHALL support local development **for all platforms** using environment variable override to bypass GitHub downloads.

#### Scenario: Local binaries override **for all platforms**
- **WHEN** ANY_SYNC_GO_BINARIES_DIR environment variable is set
- **THEN** build.rs uses binaries from the specified local directory **for any platform**
- **AND** **selects the appropriate artifact based on target platform** (desktop binary or `any-sync-android.aar`)
- **AND** copies **the correct artifact** to OUT_DIR instead of downloading from GitHub
- **AND** emits a warning indicating local binaries are being used
- **AND** validates that the specified directory exists and contains **required artifacts**

#### Scenario: Developer workflow **with mobile**
- **WHEN** a developer modifies Go backend code **for mobile**
- **THEN** they can set ANY_SYNC_GO_BINARIES_DIR=./binaries (same as desktop)
- **AND** run cargo build to immediately test changes **on Android**
- **AND** **Android artifacts** are copied to OUT_DIR maintaining the same downstream flow as downloaded binaries

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

### Requirement: Mobile Platform Build
The repository SHALL provide build tooling for Android .aar artifacts alongside desktop binaries.

#### Scenario: Build Android AAR
- **GIVEN** the Go mobile backend source code
- **WHEN** running the build script with mobile support
- **THEN** gomobile bind generates `any-sync-android.aar`
- **AND** outputs to `binaries/` directory alongside desktop binaries
- **AND** includes all Android ABIs (arm64-v8a, armeabi-v7a, x86, x86_64)
- **AND** generates SHA256 checksum file
- **AND** total .aar size is <25MB

### Requirement: Platform-Specific Binary Naming
All Go backend artifacts SHALL use consistent naming pattern distinguishing platforms.

#### Scenario: Binary naming convention
- **GIVEN** builds for multiple platforms
- **THEN** desktop binaries follow: `any-sync-<target-triple>` (e.g., `any-sync-x86_64-apple-darwin`)
- **AND** Android binary follows: `any-sync-android.aar`
- **AND** future iOS binary will follow: `any-sync-ios.xcframework`
- **AND** all artifacts stored in same `binaries/` directory

### Requirement: Android .aar Accessibility
The plugin SHALL make the Android .aar accessible to the plugin's Gradle build.

#### Scenario: Plugin self-manages .aar placement
- **WHEN** the plugin's build script completes successfully
- **AND** `any-sync-android.aar` exists in binaries directory
- **THEN** build.rs symlinks (Unix) or copies (Windows) the .aar to `android/libs/any-sync-android.aar`
- **AND** the plugin's `android/build.gradle.kts` references `implementation(files("libs/any-sync-android.aar"))`
- **AND** this works in both development (local path) and production (published crate) scenarios
- **AND** consumer's build.rs requires no Android-specific logic


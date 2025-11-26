# Spec Delta: binaries-distribution

## ADDED Requirements

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

## MODIFIED Requirements

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

### Requirement: Cargo Links Metadata Propagation
The plugin SHALL use Cargo's links mechanism to broadcast binary paths to consuming applications.

#### Scenario: Metadata broadcasting
- **WHEN** the plugin's build script completes successfully
- **THEN** it emits cargo:**binaries_dir**=<path> to standard output
- **AND** Cargo propagates this as DEP_TAURI_PLUGIN_ANY_SYNC_**BINARIES_DIR** environment variable to consumer build scripts
- **AND** this single metadata value covers both desktop binaries and Android .aar (all in same directory)

### Requirement: Android .aar Accessibility
The plugin SHALL make the Android .aar accessible to the plugin's Gradle build.

#### Scenario: Plugin self-manages .aar placement
- **WHEN** the plugin's build script completes successfully
- **AND** `any-sync-android.aar` exists in binaries directory
- **THEN** build.rs symlinks (Unix) or copies (Windows) the .aar to `android/libs/any-sync-android.aar`
- **AND** the plugin's `android/build.gradle.kts` references `implementation(files("libs/any-sync-android.aar"))`
- **AND** this works in both development (local path) and production (published crate) scenarios
- **AND** consumer's build.rs requires no Android-specific logic

## REMOVED Requirements

None.

## Dependencies

- **Internal:** Requires mobile backend API (mobile-backend-api/spec.md)
- **External:** Requires gomobile toolchain (build time)
- **Related Specs:** android-plugin-integration/spec.md (consumes .aar)

## Notes

**Unified Distribution:** Mobile artifacts follow the same distribution pattern as desktop: same GitHub Release, same checksum verification, same local override mechanism, same binaries directory. The only difference is the artifact name and extension.

**Naming Rationale:** `any-sync-android.aar` clearly distinguishes from future `any-sync-ios.xcframework` while maintaining consistency with desktop pattern `any-sync-<platform>`.

**Single Binaries Directory:** All cross-platform artifacts live in `binaries/` directory. Build system selects the correct one based on target platform. No separate mobile-specific environment variables or directories needed.

**Self-Contained Plugin:** The plugin manages its own .aar placement by symlinking it from the binaries directory to `android/libs/`. This design:
- Keeps the plugin self-contained (Gradle references local `libs/` directory)
- Works identically in development and production (published crate)
- Eliminates need for consumer build scripts to handle Android-specific logic
- Leverages Gradle's project structure where plugin's `projectDir` is set in `tauri.settings.gradle`

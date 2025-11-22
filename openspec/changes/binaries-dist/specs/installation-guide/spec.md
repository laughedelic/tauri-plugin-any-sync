# Installation Guide Specification

## MODIFIED Requirements

### Requirement: Desktop Configuration Steps
The installation guide SHALL include step-by-step build script configuration for automated binary setup on desktop platforms.
#### Scenario: Automated desktop setup instructions
- **WHEN** developers are setting up for Windows, macOS, or Linux
- **THEN** they should find instructions to:
  - Add the plugin dependency with desired features to Cargo.toml
  - Create a build.rs file with binary copying logic
  - Configure externalBin in tauri.conf.json
  - Set up required shell permissions in capabilities

#### Scenario: Feature selection guidance
- **WHEN** developers need to choose which platform binaries to download
- **THEN** the guide explains the available features (all-platforms, desktop-only, platform-specific)
- **AND** provides recommendations based on target platforms
- **AND** shows example Cargo.toml dependency configurations

### Requirement: Binary Setup Examples
The installation guide SHALL include concrete build.rs code examples for consumers.
#### Scenario: Copy-paste ready consumer build script
- **WHEN** developers need to configure binary copying
- **THEN** they find a complete build.rs example that:
  - Reads DEP_ANY_SYNC_GO_BINARIES_DIR environment variable
  - Handles missing environment variable gracefully
  - Creates the binaries directory if needed
  - Copies and renames binaries for the target platform
  - Includes proper error handling
  - Works with Tauri's sidecar naming conventions

### Requirement: Troubleshooting Section
The installation guide SHALL include solutions for binary download and automation issues.
#### Scenario: Download failure troubleshooting
- **WHEN** developers encounter binary download failures
- **THEN** the troubleshooting section provides solutions for:
  - Network connectivity issues (retrying the build)
  - Checksum verification failures

#### Scenario: Build script debugging
- **WHEN** consumer build scripts fail to copy binaries correctly
- **THEN** the troubleshooting section explains:
  - How to verify environment variables are set
  - How to check if binaries were downloaded
  - Common path and naming issues
  - Platform-specific considerations

## ADDED Requirements

### Requirement: Alternative Manual Installation
The installation guide SHALL document manual binary download as a legacy alternative approach.
#### Scenario: Manual installation fallback
- **WHEN** developers cannot use automated binary downloads (air-gapped environments, corporate proxies)
- **THEN** the guide provides instructions to:
  - Manually download binaries from GitHub Releases
  - Place them in src-tauri/binaries/ with correct naming
  - Configure externalBin and permissions (same as automated approach)
  - Mark this approach as legacy/discouraged

### Requirement: Migration Guide for Existing Users
The installation guide SHALL provide migration instructions for users currently using manual downloads.
#### Scenario: Migration from manual to automated
- **WHEN** existing users want to adopt automated binary distribution
- **THEN** the guide explains:
  - How to remove manually downloaded binaries
  - How to add the build.rs script
  - How to configure Cargo features
  - That existing tauri.conf.json and capabilities config remains unchanged
  - Benefits of migration (automatic updates, reduced manual steps)

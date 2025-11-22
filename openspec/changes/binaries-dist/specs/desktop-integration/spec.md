# Desktop Integration Specification

## MODIFIED Requirements

### Requirement: Binary Distribution Strategy
The plugin SHALL use automated binary downloads from GitHub Releases via Cargo links for desktop platforms.
#### Scenario: Automated binary provisioning
- **WHEN** users install the plugin in their Tauri application
- **THEN** the plugin's build script automatically downloads platform-specific binaries from GitHub Releases
- **AND** propagates binary locations via Cargo environment variables
- **AND** consumers copy binaries to their src-tauri/binaries/ directory using build scripts

#### Scenario: Feature-based platform selection
- **WHEN** consumers configure plugin dependency features in Cargo.toml
- **THEN** only binaries for selected platforms are downloaded
- **AND** reduces download size and build time for single-platform projects

### Requirement: User Configuration Requirements
The plugin SHALL require consumers to add a build script for binary copying on desktop platforms.
#### Scenario: Consumer build script setup
- **WHEN** developers install the plugin for desktop usage
- **THEN** they add a build.rs file that reads DEP_ANY_SYNC_GO_BINARIES_DIR
- **AND** copies binaries to src-tauri/binaries/ with proper naming
- **AND** configures externalBin in tauri.conf.json (unchanged from previous approach)
- **AND** configures shell permissions in capabilities (unchanged from previous approach)

### Requirement: Installation Documentation
The plugin SHALL provide clear examples of consumer build scripts for automated binary setup.
#### Scenario: Documented consumer setup
- **WHEN** developers read the installation documentation
- **THEN** they find copy-paste ready build.rs examples
- **AND** understand how to configure Cargo features for their target platforms
- **AND** see step-by-step instructions for externalBin and permissions configuration
- **AND** have troubleshooting guidance for download failures

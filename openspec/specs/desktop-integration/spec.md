# Desktop Integration Specification

## Purpose
Manages the Go backend as a sidecar process on desktop platforms with health monitoring, gRPC communication, and graceful lifecycle management.
## Requirements
### Requirement: Sidecar Process Spawning
The plugin SHALL spawn the Go backend as a separate process with proper lifecycle management.
#### Scenario:
Given the Tauri application starts on desktop
When the plugin initializes
Then it should spawn the Go backend as a separate process with proper lifecycle management

### Requirement: Process Health Monitoring
The plugin SHALL perform periodic health checks on the Go backend process and detect failures.
#### Scenario:
Given the Go backend is running as a sidecar process
When the plugin needs to verify the backend is responsive
Then it should perform periodic health checks and detect process failures

### Requirement: gRPC Client Connection
The plugin SHALL establish a gRPC client connection to communicate with the Go backend.
#### Scenario:
Given the Rust plugin needs to communicate with the Go backend
When the plugin establishes communication
Then it should create a gRPC client connection to the backend's localhost port

### Requirement: Port Management
The plugin SHALL allocate unique available ports to avoid conflicts between multiple application instances.
#### Scenario:
Given multiple instances of the application might run simultaneously
When each instance starts its sidecar
Then each should use a unique available port to avoid conflicts

### Requirement: Graceful Shutdown
The plugin SHALL gracefully terminate the Go backend process and clean up resources on application exit.
#### Scenario:
Given the Tauri application is closing
When the plugin teardown occurs
Then it should gracefully terminate the Go backend process and clean up resources

### Requirement: Process Restart Logic
The plugin SHALL attempt to restart the sidecar process with configurable retry limits when health checks fail.
#### Scenario:
Given the Go backend process crashes or becomes unresponsive
When the health check fails
Then the plugin should attempt to restart the sidecar process with configurable retry limits

### Requirement: Error Propagation
The plugin SHALL properly propagate process management errors to the TypeScript layer with meaningful messages.
#### Scenario:
Given the sidecar process fails to start or crashes
When errors occur during process management
Then they should be properly propagated to the TypeScript layer with meaningful error messages

### Requirement: Plugin Initialization
The existing plugin setup SHALL include sidecar process initialization and gRPC client creation.
#### Scenario:
Given: existing plugin setup in lib.rs
When: adding desktop integration
Then: the setup should include sidecar process initialization and gRPC client creation

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
- **THEN** they add a build.rs file that reads DEP_TAURI_PLUGIN_ANY_SYNC_BINARIES_DIR
- **AND** copies binaries to src-tauri/binaries/ with proper naming
- **AND** configures externalBin in tauri.conf.json (unchanged from previous approach)
- **AND** configures shell permissions in capabilities (unchanged from previous approach)

### Requirement: Mobile Zero Configuration
The plugin SHALL require zero additional configuration for mobile platforms using gomobile.
#### Scenario:
Given: developer wants to use the plugin on mobile (iOS/Android)
When: they install the plugin
Then: it should work immediately without any sidecar configuration

### Requirement: Installation Documentation
The plugin SHALL provide clear examples of consumer build scripts for automated binary setup.
#### Scenario: Documented consumer setup
- **WHEN** developers read the installation documentation
- **THEN** they find copy-paste ready build.rs examples
- **AND** understand how to configure Cargo features for their target platforms
- **AND** see step-by-step instructions for externalBin and permissions configuration
- **AND** have troubleshooting guidance for download failures

### Requirement: Binary Discovery Enhancement
The plugin SHALL enhance binary discovery to work with Tauri's sidecar naming conventions.
#### Scenario:
Given: Tauri expects binaries with target-triple suffixes
When: the plugin searches for Go backend binary
Then: it should find the correct platform-specific binary in the expected location

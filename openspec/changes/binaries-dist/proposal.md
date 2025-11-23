# Change: Automate Binary Distribution via Cargo Links and GitHub Releases

## Why

Currently, plugin users must manually download Go backend binaries from GitHub releases and place them in their app's `src-tauri/binaries/` directory. This creates friction during installation and requires manual updates when new binaries are released. The plugin needs an automated mechanism to distribute pre-compiled Go backend binaries to consuming applications using standard Cargo features.

## What Changes

- Add `links = "tauri-plugin-any-sync"` to plugin's `Cargo.toml` to enable metadata propagation
- Modify plugin's `build.rs` with two distinct modes:
  - **Consumer/CI mode**: Download binaries from GitHub releases with checksum verification
  - **Local development mode**: Use `ANY_SYNC_GO_BINARIES_DIR` environment variable to override with local binaries
- Add Cargo features for selective platform binary downloads (`all`, `macos`, `linux`, `windows`, platform-specific features). No features are enabled by default - users must explicitly select platforms to download.
- Emit `cargo:binaries_dir=<path>` environment variable for consuming applications
- Create GitHub Actions workflow to build and publish binaries on release
- Update installation documentation with consumer build script examples and local development setup
- **BREAKING**: Existing desktop-integration spec requirements updated to reference new automated approach

## Impact

- **Affected specs**: 
  - `binaries-distribution` (new capability)
  - `desktop-integration` (modified - updated binary distribution requirements)
  - `installation-guide` (modified - simplified setup instructions)
- **Affected code**: 
  - `Cargo.toml` - add `links` key and feature flags
  - `build.rs` - add binary download logic and path broadcasting
  - `.github/workflows/` - new release workflow
  - `README.md` - updated installation instructions with build script examples
- **User impact**: 
  - Simpler installation (no manual binary downloads)
  - Automatic binary updates via dependency updates
  - Optional cross-platform binary bundling via features
  - Clear separation between consumer usage (downloads) and local development (environment variable override)

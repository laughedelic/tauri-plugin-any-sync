# Binary Distribution Design

## Context

The plugin currently builds Go backend binaries locally via `build-go-backend.sh` during compilation. Users must manually copy these binaries to their application's `src-tauri/binaries/` directory and configure `externalBin` in `tauri.conf.json`. This manual process is error-prone and creates friction during plugin adoption.

Cargo provides a `links` mechanism for build scripts to propagate metadata to downstream crates via environment variables. Combined with GitHub Releases for binary distribution, we can automate the entire binary distribution pipeline.

**Stakeholders:**
- Plugin consumers (Tauri application developers)
- Plugin maintainers
- CI/CD systems

**Constraints:**
- Must work with standard Cargo tooling (no custom build tools)
- Must support cross-compilation scenarios
- Must allow offline development (local builds)
- Must be secure (checksum verification)
- Desktop platforms only (Phase 0 scope)

## Goals / Non-Goals

**Goals:**
- Eliminate manual binary download steps from installation
- Enable automatic binary updates via dependency version bumps
- Provide flexible platform selection via Cargo features
- Maintain development workflow (local builds still work)
- Secure binary distribution with checksum verification

**Non-Goals:**
- Mobile platform binaries (gomobile artifacts - future phase)
- Dynamic binary selection at runtime
- Binary caching across multiple projects (Cargo handles this)
- Binary signing/notarization (future enhancement)
- Alternative distribution channels (npm, homebrew, etc.)

## Decisions

### Decision 1: Use Cargo `links` for Metadata Propagation

**What:** Leverage Cargo's `links` key to broadcast binary paths from plugin to consumer.

**Why:**
- Standard Cargo feature (no custom tooling)
- Automatic environment variable prefixing (`DEP_<LINKS_KEY>_<VARIABLE>`)
- Works with all Cargo build scenarios
- Zero consumer-side plugin dependencies

**Alternatives considered:**
- Custom procedural macro: More complex, adds compile-time dependencies
- Environment variables set by user: Defeats automation purpose
- Workspace-based approach: Doesn't work for external consumers

### Decision 2: GitHub Releases as Binary Distribution Channel

**What:** Publish pre-compiled binaries to GitHub Releases on version tags.

**Why:**
- Already using GitHub for source code hosting
- Free artifact hosting for open-source projects
- Built-in versioning tied to Git tags
- Standard practice in Rust ecosystem
- Simple HTTP downloads (reqwest)

**Alternatives considered:**
- crates.io package data: Limited to package size (<10MB typically), our binaries are larger
- Separate CDN: Additional infrastructure cost and complexity
- Git LFS: Slower downloads, complicates repository

### Decision 3: Feature Flags for Platform Selection

**What:** Cargo features control which platform binaries are downloaded.

**Why:**
- Users building for single platform don't need all binaries
- Reduces download size and build time
- Standard Cargo mechanism for optional functionality
- Supports cross-compilation scenarios

**Feature structure:**

Features correspond to the go-backend build targets and their combinations:

```toml
[features]
x86_64-apple-darwin = []
aarch64-apple-darwin = []
x86_64-unknown-linux-gnu = []
aarch64-unknown-linux-gnu = []
x86_64-pc-windows-msvc = []
macos = ["x86_64-apple-darwin", "aarch64-apple-darwin"]
linux = ["x86_64-unknown-linux-gnu", "aarch64-unknown-linux-gnu"]
windows = ["x86_64-pc-windows-msvc"]
all = ["macos", "linux", "windows"]
```

**Alternatives considered:**
- Target-based detection: Doesn't handle cross-compilation well
- Always download all: Wastes bandwidth for single-platform developers
- Environment variables: Less discoverable, not Cargo-native

### Decision 4: Local Development via Environment Variable Override

**What:** Support local development using `ANY_SYNC_GO_BINARIES_DIR` environment variable to override GitHub downloads.

**Why:**
- Standard pattern used by sys crates (e.g., `OPENSSL_DIR`, `SQLITE3_LIB_DIR`)
- Passes data value (path) not just boolean flag
- Allows developers to build Go binaries with any tooling (IDE, Makefile, direct `go build`)
- Clear separation: local development vs consumer installation (two distinct scenarios, not fallback)
- No accidental commits (env var not in Cargo.toml)
- Works with workspace `.cargo/config.toml` for persistent configuration

**Implementation:**
```rust
// In build.rs
const ENV_VAR_NAME: &str = "ANY_SYNC_GO_BINARIES_DIR";

fn main() {
    println!("cargo:rerun-if-env-changed={}", ENV_VAR_NAME);
    
    let out_dir = PathBuf::from(env::var("OUT_DIR").unwrap());
    
    if let Ok(local_path) = env::var(ENV_VAR_NAME) {
        // LOCAL DEVELOPMENT MODE
        let local_binaries = Path::new(&local_path);
        if !local_binaries.exists() {
            panic!("Local binaries directory not found: {}", local_path);
        }
        // Copy local binaries to OUT_DIR
        copy_binaries(&local_binaries, &out_dir);
        println!("cargo:warning=Using local binaries from: {}", local_path);
    } else {
        // CONSUMER/CI MODE
        download_binaries_from_github(&out_dir)?;
    }
    
    // Emit path for consumers (same for both modes)
    println!("cargo:binaries_dir={}", out_dir.display());
}
```

**Usage patterns:**

*Local development (one-off):*
```bash
export ANY_SYNC_GO_BINARIES_DIR=./binaries
cargo build
```

*Local development (persistent via `.cargo/config.toml`):*
```toml
[env]
ANY_SYNC_GO_BINARIES_DIR = { value = "/absolute/path/to/binaries", force = true }
```

**Why not a Cargo feature?**
- Features are compile-time flags, not data values
- Features can be accidentally committed to Cargo.toml
- Env var allows building Go binaries with any workflow
- Standard Rust ecosystem pattern for overriding bundled dependencies

**Trade-offs:**
- Developers must set env var explicitly for local development
- Requires understanding of environment variable configuration
- Different workflow than automatic fallback (but more explicit and intentional)

### Decision 5: Checksum Verification

**What:** Verify SHA256 checksums of downloaded binaries against published checksums.

**Why:**
- Prevents corrupted downloads
- Mitigates MITM attacks
- Standard security practice
- Minimal performance overhead

**Implementation:**
- Generate SHA256 checksums in GitHub Actions workflow
- Upload checksums as `checksums.txt` in release assets
- Download and verify in build.rs before using binaries

### Decision 6: Consumer Build Script Pattern

**What:** Consumers add a `build.rs` that reads `DEP_TAURI_PLUGIN_ANY_SYNC_BINARIES_DIR` and copies binaries to `src-tauri/binaries/`.

**Why:**
- Simple, explicit, and transparent
- Consumers control where binaries are placed
- Works with Tauri's sidecar naming conventions
- No magic or hidden behavior

**Example consumer build.rs:**
```rust
fn main() {
    if let Ok(binaries_dir) = env::var("DEP_TAURI_PLUGIN_ANY_SYNC_BINARIES_DIR") {
        let manifest_dir = env::var("CARGO_MANIFEST_DIR").unwrap();
        let dest_dir = Path::new(&manifest_dir).join("binaries");
        fs::create_dir_all(&dest_dir).unwrap();

        // Copy binaries for target platform
        let target = env::var("TARGET").unwrap();
        // ... copy logic
    }
}
```

## Architecture

### Build Flow

**Consumer/CI Mode (Production):**
```
Plugin Build:
1. Plugin build.rs runs
2. Check ANY_SYNC_GO_BINARIES_DIR env var (not set)
3. Download binaries from GitHub releases based on features
4. Verify checksums
5. Store in OUT_DIR/binaries/
6. Emit cargo:binaries_dir=<OUT_DIR>/binaries
7. Cargo propagates as DEP_TAURI_PLUGIN_ANY_SYNC_BINARIES_DIR

Consumer Build:
1. Consumer build.rs runs
2. Reads DEP_TAURI_PLUGIN_ANY_SYNC_BINARIES_DIR
3. Copies binaries to src-tauri/binaries/
4. Tauri bundles via externalBin config
```

**Local Development Mode:**
```
Plugin Build:
1. Plugin build.rs runs
2. Check ANY_SYNC_GO_BINARIES_DIR env var (set to ./binaries)
3. Copy binaries from local path to OUT_DIR/binaries/
4. Emit cargo:binaries_dir=<OUT_DIR>/binaries
5. Cargo propagates as DEP_TAURI_PLUGIN_ANY_SYNC_BINARIES_DIR
   (Same downstream flow - consumers don't know the difference)

Developer Workflow:
1. Modify Go backend code
2. Run ./build-go-backend.sh (outputs to ./binaries/)
3. Set ANY_SYNC_GO_BINARIES_DIR=./binaries (or via .cargo/config.toml)
4. Run cargo build
5. Test changes immediately
```

### Release Flow

```
Developer:
1. Bump version in Cargo.toml
2. Create Git tag: git tag v0.2.0
3. Push tag: git push origin v0.2.0

GitHub Actions:
1. Workflow triggers on tag push
2. Runs build-go-backend.sh --cross
3. Generates SHA256 checksums
4. Creates GitHub Release
5. Uploads binaries + checksums

Consumer:
1. Updates dependency: tauri-plugin-any-sync = "0.2.0"
2. Runs cargo build
3. Plugin build.rs downloads v0.2.0 binaries
4. Consumer build.rs copies to src-tauri/binaries/
```

## Risks / Trade-offs

### Risk: Binary Size Impact
**Impact:** Larger dependency download size for consumers.

**Mitigation:**
- Feature flags allow downloading only needed platforms
- No features enabled by default - users must explicitly choose platforms
- Document minimal feature set for single-platform builds
- Consider separate mobile distribution in future

### Risk: Network Dependency
**Impact:** Cannot build offline without local binaries or Go toolchain.

**Mitigation:**
- Cargo caches downloads after first successful build
- Use ANY_SYNC_GO_BINARIES_DIR env var for offline development
- Document environment variable override pattern
- CI systems typically have network access

### Trade-off: Build Complexity
**Impact:** build.rs becomes more complex with download logic and environment variable handling.

**Mitigation:**
- Comprehensive inline documentation
- Clear error messages for both download failures and invalid local paths
- Unit tests for download/verify logic
- Simple environment variable override pattern (standard sys crate practice)

## Migration Plan

### Phase 1: Add Download Capability (Non-Breaking)
1. Implement download logic in build.rs
2. Keep local builds working as-is
3. Add feature flags
4. Create release workflow
5. Test with internal projects

### Phase 2: Update Documentation
1. Add new automated approach to README
2. Keep manual instructions as alternative
3. Add migration guide for existing users
4. Update example app with build.rs example

### Phase 3: Deprecation (Future)
1. Mark manual download instructions as legacy
2. Add deprecation warnings
3. Continue supporting both approaches for 2-3 releases
4. Eventually remove manual instructions (keep in archive)

### Rollback Plan

If automation causes issues:
1. Set ANY_SYNC_GO_BINARIES_DIR to use local binaries (bypasses downloads)
2. Manual download instructions remain in docs as alternative
3. Can revert changes without breaking existing users
4. Environment variable approach provides immediate workaround

## Open Questions

1. **Q:** Should we sign/notarize binaries for macOS/Windows?
   **A:** Defer to future phase. Current unsigned approach works for development, but production apps may need signing.

2. **Q:** Should we provide alternative download sources (mirrors)?
   **A:** Start with GitHub Releases only. Add mirrors if reliability issues arise.

3. **Q:** How to handle pre-release versions (alpha, beta)?
   **A:** Use Git tag naming convention (`v0.2.0-beta.1`). Build.rs can parse semver and download matching release.

4. **Q:** Should we compress binaries before upload?
   **A:** Current binaries are already compiled/stripped. Compression savings would be minimal (~10-15%). Keep uncompressed for simplicity.

5. **Q:** What's the maximum acceptable download size?
   **A:** All desktop platforms (~45MB total). With features, users can reduce to ~15MB for single platform.

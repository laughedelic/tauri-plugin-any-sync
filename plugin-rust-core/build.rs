use std::env;
use std::error::Error;
use std::fs;
use std::path::Path;
use std::path::PathBuf;

fn main() {
    // println!("cargo:warning=ANY_SYNC_GO_BINARIES_DIR={}", env::var("ANY_SYNC_GO_BINARIES_DIR").unwrap_or_default());

    // Generate protobuf code for desktop targets only (mobile uses FFI, not gRPC)
    let target = env::var("TARGET").unwrap_or_default();
    let is_mobile_target = target.contains("android") || target.contains("ios");
    println!("cargo:warning=Building for target: {}", target);

    if !is_mobile_target {
        if let Err(e) = generate_protobuf() {
            eprintln!("Error: Failed to generate protobuf code: {}", e);
            std::process::exit(1);
        }
    }

    // Manage binaries (download or use local)
    if let Err(e) = manage_binaries() {
        eprintln!("Error managing binaries: {}", e);
        std::process::exit(1);
    }

    tauri_plugin::Builder::new(&["command"])
        .android_path("android")
        .ios_path("ios")
        .build();
}

/// Manages binaries: either downloads from GitHub or uses local directory
fn manage_binaries() -> Result<(), Box<dyn Error>> {
    let target = env::var("TARGET").unwrap();
    let out_dir = PathBuf::from(env::var("OUT_DIR")?);
    let binaries_out_dir = out_dir.join("binaries");

    // Check if local development mode is enabled
    if let Ok(local_path) = env::var("ANY_SYNC_GO_BINARIES_DIR") {
        let local_binaries = Path::new(&local_path);
        println!("cargo:rerun-if-changed={}", local_binaries.display());

        // Check if the provided directory exists and has any files
        let binaries_missing = !local_binaries.exists()
            || !local_binaries.is_dir()
            || fs::read_dir(local_binaries)?.next().is_none();

        if binaries_missing {
            return Err(format!(
                "Local binaries directory is missing or empty: {}",
                local_binaries.display()
            )
            .into());
        }

        // In development mode, link local binaries to out_dir
        println!(
            "cargo:warning=Linking local binaries from: {}",
            local_binaries.canonicalize()?.display()
        );
        create_link(&local_binaries, &binaries_out_dir)?;
    } else if env::var_os("CI").is_none() {
        // Download from GitHub (non-CI)
        download_binaries_from_github(&binaries_out_dir)?;
    }

    // For Android: symlink .aar to plugin's android/libs/ directory
    // This allows the plugin's gradle file to reference libs/any-sync-android.aar
    if target.contains("android") {
        let aar_file_name = "any-sync-android.aar";
        let aar_file = binaries_out_dir.join(&aar_file_name);
        if aar_file.exists() {
            let aar_dest = env::current_dir()?
                .join("android")
                .join("libs")
                .join(&aar_file_name);
            create_link(&aar_file, &aar_dest)?;
        }
    }

    // For iOS: point swift-rs linker to the right framework inside the xcframework bundle
    if target.contains("-apple-ios") {
        let framework_name = "AnySync";
        let xcframework_ext = ".xcframework";
        let xcframework_path = binaries_out_dir.join(framework_name.to_string() + xcframework_ext);
        if xcframework_path.exists() {
            // xcframework bundles different architectures in subfolders
            let framework_path = if target == "aarch64-apple-ios" {
                xcframework_path.join("ios-arm64")
            } else {
                xcframework_path.join("ios-arm64_x86_64-simulator")
            };

            // -F flag (Search Path) -> points to the folder CONTAINING the .framework
            println!(
                "cargo:rustc-link-search=framework={}",
                framework_path.display()
            );

            // -framework flag (The Lib) -> assumes "AnySync.framework" exists in the search path
            println!("cargo:rustc-link-lib=framework={}", framework_name);
        }
    }

    // Emit metadata for consumer crates
    println!("cargo:binaries_dir={}", binaries_out_dir.display());

    Ok(())
}

/// Download binaries from GitHub releases (consumer/CI mode)
fn download_binaries_from_github(dest_dir: &PathBuf) -> Result<(), Box<dyn Error>> {
    // Get plugin version from Cargo.toml
    let version = env::var("CARGO_PKG_VERSION")?;

    // Determine which binaries to download based on enabled features
    let binaries_to_download = determine_binaries_to_download()?;

    if binaries_to_download.is_empty() {
        // No features enabled, skip download
        println!("cargo:warning=No platform features enabled, skipping binary downloads");
        fs::create_dir_all(dest_dir)?;
        return Ok(());
    }

    // Create destination directory
    fs::create_dir_all(dest_dir)?;

    // Construct GitHub release URL
    let release_url = format!(
        "https://github.com/laughedelic/tauri-plugin-any-sync/releases/download/v{}/",
        version
    );

    // Download checksums.txt first
    let checksums_content = download_file(&format!("{}checksums.txt", release_url))?;
    let checksums_path = dest_dir.join("checksums.txt");
    fs::write(&checksums_path, &checksums_content)?;

    // Parse checksums
    let checksums_str = String::from_utf8(checksums_content)?;
    let checksums = parse_checksums(&checksums_str)?;

    // Download each binary
    for binary_name in binaries_to_download {
        let url = format!("{}{}", release_url, binary_name);
        println!("Downloading: {}", url);

        let binary_content = download_file(&url)?;

        // Verify checksum
        let expected_checksum = checksums
            .get(&binary_name)
            .ok_or_else(|| format!("Checksum not found for binary: {}", binary_name))?;

        let actual_checksum = compute_sha256(&binary_content);

        if actual_checksum != *expected_checksum {
            return Err(format!(
                "Checksum verification failed for {}: expected {}, got {}",
                binary_name, expected_checksum, actual_checksum
            )
            .into());
        }

        // Handle xcframework zip files specially - extract them
        if binary_name.ends_with(".xcframework.zip") {
            // Extract zip to dest_dir (zip contains xcframework folder at root)
            let cursor = std::io::Cursor::new(&binary_content);
            let mut archive = zip::ZipArchive::new(cursor)?;
            archive.extract(dest_dir)?;

            println!(
                "cargo:warning=Extracted {} to {}",
                binary_name,
                dest_dir.display()
            );
        } else {
            // Write binary to destination
            let dest_path = dest_dir.join(&binary_name);
            fs::write(&dest_path, &binary_content)?;

            // Make binary executable on Unix-like systems
            #[cfg(unix)]
            {
                use std::os::unix::fs::PermissionsExt;
                let permissions = fs::Permissions::from_mode(0o755);
                fs::set_permissions(&dest_path, permissions)?;
            }
        }
    }

    Ok(())
}

/// Determine which binaries to download based on enabled features
fn determine_binaries_to_download() -> Result<Vec<String>, Box<dyn Error>> {
    let mut binaries = Vec::new();

    // Check which platform features are enabled
    if cfg!(feature = "x86_64-apple-darwin") {
        binaries.push("any-sync-x86_64-apple-darwin".to_string());
    }
    if cfg!(feature = "aarch64-apple-darwin") {
        binaries.push("any-sync-aarch64-apple-darwin".to_string());
    }
    if cfg!(feature = "x86_64-unknown-linux-gnu") {
        binaries.push("any-sync-x86_64-unknown-linux-gnu".to_string());
    }
    if cfg!(feature = "aarch64-unknown-linux-gnu") {
        binaries.push("any-sync-aarch64-unknown-linux-gnu".to_string());
    }
    if cfg!(feature = "x86_64-pc-windows-msvc") {
        binaries.push("any-sync-x86_64-pc-windows-msvc.exe".to_string());
    }
    if cfg!(feature = "android") {
        binaries.push("any-sync-android.aar".to_string());
    }
    if cfg!(feature = "ios") {
        binaries.push("any-sync-ios.xcframework.zip".to_string());
    }

    Ok(binaries)
}

/// Download file from URL
fn download_file(url: &str) -> Result<Vec<u8>, Box<dyn Error>> {
    let client = reqwest::blocking::Client::new();
    let response = client.get(url).send()?;

    if !response.status().is_success() {
        return Err(format!(
            "Failed to download from {}: HTTP {}",
            url,
            response.status()
        )
        .into());
    }

    Ok(response.bytes()?.to_vec())
}

/// Parse checksums.txt format: "<hash>  <filename>"
fn parse_checksums(
    content: &str,
) -> Result<std::collections::HashMap<String, String>, Box<dyn Error>> {
    use std::collections::HashMap;

    let mut checksums = HashMap::new();

    for line in content.lines() {
        let line = line.trim();
        if line.is_empty() {
            continue;
        }

        // Split on whitespace: "<hash>  <filename>"
        let parts: Vec<&str> = line.split_whitespace().collect();
        if parts.len() < 2 {
            return Err(format!("Invalid checksum line format: {}", line).into());
        }

        let hash = parts[0];
        let filename = parts[1];

        checksums.insert(filename.to_string(), hash.to_string());
    }

    Ok(checksums)
}

/// Compute SHA256 checksum of bytes
fn compute_sha256(data: &[u8]) -> String {
    use sha2::{Digest, Sha256};

    let mut hasher = Sha256::new();
    hasher.update(data);
    let result = hasher.finalize();

    format!("{:x}", result)
}

fn generate_protobuf() -> Result<(), Box<dyn Error>> {
    // Generate protobuf code from the new unified transport and syncspace APIs
    println!("cargo:rerun-if-changed=../buf/proto/dispatch-transport/transport/v1/transport.proto");
    println!("cargo:rerun-if-changed=../buf/proto/syncspace-api/syncspace/v1/syncspace.proto");

    // Generate protobuf code using tonic for gRPC client
    tonic_build::configure()
        .build_server(false) // We only need the client (server runs in Go)
        .compile_protos(
            &[
                "../buf/proto/dispatch-transport/transport/v1/transport.proto",
                "../buf/proto/syncspace-api/syncspace/v1/syncspace.proto",
            ],
            &[
                "../buf/proto/dispatch-transport",
                "../buf/proto/syncspace-api",
            ],
        )?;

    Ok(())
}

/// Creates a filesystem link using the best unprivileged method for the platform.
///
/// - **Windows**: Uses `Junctions` for directories and `Hard Links` for files.
/// - **Unix**: Uses `Symlinks` for both.
pub fn create_link<P: AsRef<Path>, Q: AsRef<Path>>(src: P, dst: Q) -> std::io::Result<()> {
    let src = src.as_ref().canonicalize().unwrap_or_else(|err| {
        panic!(
            "Failed to canonicalize source path {:?}: {}",
            src.as_ref(),
            err
        )
    });
    let dst = dst.as_ref();

    // Clean Destination: Remove existing link or directory safely
    if let Ok(meta) = fs::symlink_metadata(dst) {
        if meta.is_dir() {
            fs::remove_dir_all(dst)?;
        } else {
            fs::remove_file(dst)?;
        }
    }

    // Create parent directories for destination if they don't exist
    if let Some(parent) = dst.parent() {
        fs::create_dir_all(parent)?;
    }

    // Create Link: Platform-specific implementation
    #[cfg(unix)]
    {
        std::os::unix::fs::symlink(src, dst)
    }

    #[cfg(windows)]
    {
        if src.is_dir() {
            // Directory -> Create Junction (No Admin needed)
            junction::create(src, dst)
        } else {
            // File -> Create Hard Link (No Admin needed, must be same drive)
            fs::hard_link(src, dst)
        }
    }
}

const COMMANDS: &[&str] = &["command"];

fn main() {
    println!(
        "cargo:warning=ANY_SYNC_GO_BINARIES_DIR={}",
        std::env::var("ANY_SYNC_GO_BINARIES_DIR").unwrap_or_default()
    );

    // Generate protobuf code for desktop targets only (mobile uses FFI, not gRPC)
    let target = std::env::var("TARGET").unwrap_or_default();
    let is_mobile_target = target.contains("android") || target.contains("ios");

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

    tauri_plugin::Builder::new(COMMANDS)
        .android_path("android")
        .ios_path("ios")
        .build();
}

/// Manages binaries: either downloads from GitHub or uses local directory
fn manage_binaries() -> Result<(), Box<dyn std::error::Error>> {
    use std::env;
    use std::fs;
    use std::path::PathBuf;

    let out_dir = PathBuf::from(env::var("OUT_DIR")?);
    let binaries_out_dir = out_dir.join("binaries");

    // Check if local development mode is enabled
    if let Ok(local_path) = env::var("ANY_SYNC_GO_BINARIES_DIR") {
        // LOCAL DEVELOPMENT MODE
        let local_binaries = std::path::Path::new(&local_path);
        println!("cargo:rerun-if-changed={}", local_binaries.display());

        // Check if binaries directory exists and has any files
        let binaries_missing = !local_binaries.exists()
            || !local_binaries.is_dir()
            || std::fs::read_dir(local_binaries)?.next().is_none();

        if binaries_missing {
            return Err(format!(
                "Local binaries directory is missing or empty: {}",
                local_binaries.display()
            )
            .into());
        }

        link_local_binaries(&local_path, &binaries_out_dir)?;
    } else if env::var_os("CI").is_none() {
        // CONSUMER: Download from GitHub (non-CI)
        download_binaries_from_github(&binaries_out_dir)?;
    }

    // For Android: symlink .aar to plugin's android/libs/ directory
    // This allows the plugin's gradle file to reference libs/any-sync-android.aar
    let aar_file = binaries_out_dir.join("any-sync-android.aar");
    if aar_file.exists() {
        let android_libs = env::current_dir()?.join("android").join("libs");
        fs::create_dir_all(&android_libs)?;
        let aar_dest = android_libs.join("any-sync-android.aar");

        // Remove existing file/symlink if present
        if aar_dest.exists() || aar_dest.symlink_metadata().is_ok() {
            fs::remove_file(&aar_dest).ok();
        }

        // Create symlink (Unix) or copy (Windows)
        #[cfg(unix)]
        {
            std::os::unix::fs::symlink(&aar_file, &aar_dest)?;
        }
        #[cfg(windows)]
        {
            fs::copy(&aar_file, &aar_dest)?;
        }
    }

    // For iOS: symlink .xcframework to plugin's ios/Frameworks/ directory
    // This allows the Package.swift to reference Frameworks/any-sync-ios.xcframework
    let xcframework_dir = binaries_out_dir.join("any-sync-ios.xcframework");
    if xcframework_dir.exists() {
        let ios_frameworks = env::current_dir()?.join("ios").join("Frameworks");
        println!(
            "cargo:warning=Linking iOS xcframework to {}",
            ios_frameworks.display()
        );
        fs::create_dir_all(&ios_frameworks)?;
        let xcframework_dest = ios_frameworks.join("any-sync-ios.xcframework");

        // Remove existing directory/symlink if present
        if xcframework_dest.symlink_metadata().is_ok() {
            if xcframework_dest.is_dir() && !xcframework_dest.is_symlink() {
                fs::remove_dir_all(&xcframework_dest).ok();
            } else {
                fs::remove_file(&xcframework_dest).ok();
            }
        }

        // Create symlink (Unix) or copy directory (Windows)
        #[cfg(unix)]
        {
            std::os::unix::fs::symlink(&xcframework_dir, &xcframework_dest)?;
            println!(
                "cargo:warning=Created symlink: {} -> {}",
                xcframework_dest.display(),
                xcframework_dir.display()
            );
        }
        #[cfg(windows)]
        {
            copy_dir_recursive(&xcframework_dir, &xcframework_dest)?;
        }
    }

    // Emit metadata for consumer crates
    println!("cargo:binaries_dir={}", binaries_out_dir.display());

    Ok(())
}

/// Link binaries from local directory (development mode)
fn link_local_binaries(
    local_path: &str,
    dest_dir: &std::path::PathBuf,
) -> Result<(), Box<dyn std::error::Error>> {
    use std::fs;

    let local_binaries = std::path::Path::new(local_path);

    // Validate the path exists
    if !local_binaries.exists() {
        return Err(format!(
            "Local binaries directory not found at: {}",
            local_binaries.display()
        )
        .into());
    }

    if !local_binaries.is_dir() {
        return Err(format!(
            "Expected a directory but found a file at: {}",
            local_binaries.display()
        )
        .into());
    }

    println!(
        "cargo:warning=Linking local binaries from: {}",
        local_binaries.canonicalize()?.display()
    );

    // Create destination directory
    fs::create_dir_all(dest_dir)?;

    // Create symlinks to binaries from source to destination
    for entry in fs::read_dir(local_binaries)? {
        let entry = entry?;
        let path = entry.path();
        let file_name = entry.file_name();
        let dest_path = dest_dir.join(&file_name);

        // Remove existing symlink/file/directory if it exists
        if dest_path.symlink_metadata().is_ok() {
            if dest_path.is_dir() && !dest_path.is_symlink() {
                fs::remove_dir_all(&dest_path).ok();
            } else {
                fs::remove_file(&dest_path).ok();
            }
        }

        // Use absolute path for symlink target
        let absolute_source = path.canonicalize()?;

        // Create symlink (Unix) or copy (Windows fallback)
        #[cfg(unix)]
        {
            std::os::unix::fs::symlink(&absolute_source, &dest_path)?;
        }

        #[cfg(windows)]
        {
            // Windows symlinks require admin privileges, so fall back to copying
            if path.is_file() {
                fs::copy(&absolute_source, &dest_path)?;
            } else if path.is_dir() {
                copy_dir_recursive(&absolute_source, &dest_path)?;
            }
        }
    }
    Ok(())
}

/// Download binaries from GitHub releases (consumer/CI mode)
fn download_binaries_from_github(
    dest_dir: &std::path::PathBuf,
) -> Result<(), Box<dyn std::error::Error>> {
    use std::env;
    use std::fs;

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
fn determine_binaries_to_download() -> Result<Vec<String>, Box<dyn std::error::Error>> {
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
fn download_file(url: &str) -> Result<Vec<u8>, Box<dyn std::error::Error>> {
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
) -> Result<std::collections::HashMap<String, String>, Box<dyn std::error::Error>> {
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

/// Copy directory recursively (for Windows fallback)
#[cfg(windows)]
fn copy_dir_recursive(
    src: &std::path::Path,
    dst: &std::path::Path,
) -> Result<(), Box<dyn std::error::Error>> {
    use std::fs;

    fs::create_dir_all(dst)?;
    for entry in fs::read_dir(src)? {
        let entry = entry?;
        let path = entry.path();
        let dest_path = dst.join(entry.file_name());

        if path.is_dir() {
            copy_dir_recursive(&path, &dest_path)?;
        } else {
            fs::copy(&path, &dest_path)?;
        }
    }
    Ok(())
}

fn generate_protobuf() -> Result<(), Box<dyn std::error::Error>> {
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

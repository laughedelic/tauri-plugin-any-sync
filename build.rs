const COMMANDS: &[&str] = &["ping"];

fn main() {
    // Generate protobuf code first
    if let Err(e) = generate_protobuf() {
        eprintln!("Warning: Failed to generate protobuf code: {}", e);
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
    use std::path::PathBuf;

    const ENV_VAR_NAME: &str = "ANY_SYNC_GO_BINARIES_DIR";

    // Register dependency on environment variable changes
    println!("cargo:rerun-if-env-changed={}", ENV_VAR_NAME);

    let out_dir = PathBuf::from(env::var("OUT_DIR")?);
    let binaries_out_dir = out_dir.join("binaries");

    // Check if local development mode is enabled
    if let Ok(local_path) = env::var(ENV_VAR_NAME) {
        // LOCAL DEVELOPMENT MODE
        copy_local_binaries(&local_path, &binaries_out_dir)?;
        println!("cargo:warning=Using local binaries from: {}", local_path);
    } else {
        // CONSUMER/CI MODE: Download from GitHub
        download_binaries_from_github(&binaries_out_dir)?;
    }

    // Emit metadata for consumer crates
    println!("cargo:binaries_dir={}", binaries_out_dir.display());

    Ok(())
}

/// Copy binaries from local directory (development mode)
fn copy_local_binaries(
    local_path: &str,
    dest_dir: &std::path::PathBuf,
) -> Result<(), Box<dyn std::error::Error>> {
    use std::fs;

    let local_binaries = std::path::Path::new(local_path);

    // Validate the path exists
    if !local_binaries.exists() {
        return Err(format!(
            "Local binaries directory not found at: {}",
            local_path
        )
        .into());
    }

    if !local_binaries.is_dir() {
        return Err(format!(
            "Expected a directory but found a file at: {}",
            local_path
        )
        .into());
    }

    // Create destination directory
    fs::create_dir_all(dest_dir)?;

    // Copy all binaries from source to destination
    for entry in fs::read_dir(local_binaries)? {
        let entry = entry?;
        let path = entry.path();

        // Only copy binary files (skip non-files like directories)
        if path.is_file() {
            let file_name = entry.file_name();
            let dest_file = dest_dir.join(&file_name);
            fs::copy(&path, &dest_file)?;
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
        "https://github.com/sst/tauri-plugin-any-sync/releases/download/v{}/",
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
        let expected_checksum = checksums.get(&binary_name).ok_or_else(|| {
            format!("Checksum not found for binary: {}", binary_name)
        })?;

        let actual_checksum = compute_sha256(&binary_content);

        if actual_checksum != *expected_checksum {
            return Err(format!(
                "Checksum verification failed for {}: expected {}, got {}",
                binary_name, expected_checksum, actual_checksum
            )
            .into());
        }

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

    Ok(())
}

/// Determine which binaries to download based on enabled features
fn determine_binaries_to_download() -> Result<Vec<String>, Box<dyn std::error::Error>> {
    let mut binaries = Vec::new();

    // Check which platform features are enabled
    if cfg!(feature = "x86_64-apple-darwin") {
        binaries.push("server-x86_64-apple-darwin".to_string());
    }
    if cfg!(feature = "aarch64-apple-darwin") {
        binaries.push("server-aarch64-apple-darwin".to_string());
    }
    if cfg!(feature = "x86_64-unknown-linux-gnu") {
        binaries.push("server-x86_64-unknown-linux-gnu".to_string());
    }
    if cfg!(feature = "aarch64-unknown-linux-gnu") {
        binaries.push("server-aarch64-unknown-linux-gnu".to_string());
    }
    if cfg!(feature = "x86_64-pc-windows-msvc") {
        binaries.push("server-x86_64-pc-windows-msvc.exe".to_string());
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

fn generate_protobuf() -> Result<(), Box<dyn std::error::Error>> {
    println!("cargo:rerun-if-changed=go-backend/api/proto/health.proto");

    // Generate protobuf code
    tonic_build::configure()
        .build_server(false) // We only need the client
        .out_dir("src/proto")
        .compile_protos(
            &["go-backend/api/proto/health.proto"],
            &["go-backend/api/proto"],
        )?;

    Ok(())
}

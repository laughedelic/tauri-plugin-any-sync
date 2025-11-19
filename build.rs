const COMMANDS: &[&str] = &["ping"];

fn main() {
    // Generate protobuf code first
    if let Err(e) = generate_protobuf() {
        eprintln!("Warning: Failed to generate protobuf code: {}", e);
    }

    // Build Go backend
    if let Err(e) = build_go_backend() {
        eprintln!("Warning: Failed to build Go backend: {}", e);
        eprintln!("The Go backend will need to be built manually with ./build-go-backend.sh");
    }

    tauri_plugin::Builder::new(COMMANDS)
        .android_path("android")
        .ios_path("ios")
        .build();
}

fn build_go_backend() -> Result<(), Box<dyn std::error::Error>> {
    use std::env;
    use std::path::Path;
    use std::process::Command;

    println!("cargo:rerun-if-changed=go-backend/");
    println!("cargo:rerun-if-changed=build-go-backend.sh");

    // Only build Go backend during actual build, not during cargo check
    if env::var("CARGO_CFG_TARGET_OS").is_err() {
        return Ok(());
    }

    let manifest_dir = env::var("CARGO_MANIFEST_DIR")?;
    let project_root = Path::new(&manifest_dir);

    // Check if Go is available
    let go_check = Command::new("go").arg("version").output();
    if go_check.is_err() {
        return Err("Go toolchain not found. Please install Go to build the backend.".into());
    }

    // Check if build script exists
    let build_script = project_root.join("build-go-backend.sh");
    if !build_script.exists() {
        return Err("Go backend build script not found".into());
    }

    println!("Building Go backend...");

    let output = Command::new("bash")
        .arg(build_script.to_str().unwrap())
        .current_dir(project_root)
        .output()?;

    if !output.status.success() {
        let stderr = String::from_utf8_lossy(&output.stderr);
        let stdout = String::from_utf8_lossy(&output.stdout);
        return Err(format!(
            "Go backend build failed:\nstdout: {}\nstderr: {}",
            stdout, stderr
        )
        .into());
    }

    println!("Go backend built successfully");

    // Emit cargo metadata to include binaries in package
    let binaries_dir = project_root.join("binaries");
    if binaries_dir.exists() {
        println!(
            "cargo:rustc-env=ANY_SYNC_BINARIES_DIR={}",
            binaries_dir.display()
        );
    }

    Ok(())
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

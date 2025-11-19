fn main() -> Result<(), Box<dyn std::error::Error>> {
    // Generate protobuf code
    tonic_build::configure()
        .build_server(false) // We only need the client
        .out_dir("src/proto")
        .compile(
            &["go-backend/api/proto/health.proto"],
            &["go-backend/api/proto"],
        )?;
    
    println!("cargo:rerun-if-changed=go-backend/api/proto/health.proto");
    
    Ok(())
}
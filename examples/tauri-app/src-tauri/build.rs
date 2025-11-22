use std::env;
use std::fs;
use std::path::Path;

fn main() {
    // Copy binaries from plugin
    if let Ok(binaries_dir) = env::var("DEP_ANY_SYNC_GO_BINARIES_DIR") {
        let manifest_dir = env::var("CARGO_MANIFEST_DIR").unwrap();
        let dest_dir = Path::new(&manifest_dir).join("binaries");

        // Create destination directory
        fs::create_dir_all(&dest_dir).unwrap();

        // Copy binaries from plugin's output directory
        for entry in fs::read_dir(&binaries_dir).unwrap() {
            let entry = entry.unwrap();
            let path = entry.path();

            // Only copy binary files
            if path.is_file() {
                let file_name = entry.file_name();
                let dest = dest_dir.join(&file_name);
                fs::copy(&path, &dest).unwrap();
            }
        }
    }

    tauri_build::build()
}

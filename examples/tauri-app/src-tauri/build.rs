use std::{env, fs, path::Path};

fn main() {
    // Link binaries directory from plugin
    if let Ok(binaries_dir) = env::var("DEP_TAURI_PLUGIN_ANY_SYNC_BINARIES_DIR") {
        let dest_dir = Path::new(&env::var("CARGO_MANIFEST_DIR").unwrap()).join("binaries");
        // Clean up existing directory/symlink
        let _ = fs::remove_dir_all(&dest_dir).or_else(|_| fs::remove_file(&dest_dir));

        let source = Path::new(&binaries_dir).canonicalize().unwrap();

        #[cfg(unix)]
        std::os::unix::fs::symlink(&source, &dest_dir).unwrap();

        #[cfg(windows)]
        {
            fs::create_dir_all(&dest_dir).unwrap();
            for entry in fs::read_dir(&source).unwrap().flatten() {
                if entry.path().is_file() {
                    fs::copy(&entry.path(), dest_dir.join(entry.file_name())).unwrap();
                }
            }
        }
    }

    tauri_build::build()
}

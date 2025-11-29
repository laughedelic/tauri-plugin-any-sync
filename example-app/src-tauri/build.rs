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

        // Symlink test binary for integration tests
        if std::env::var("CARGO_FEATURE_INTEGRATION_TEST").is_ok() {
            let target = env::var("TARGET").unwrap();
            let src_bin = source.join(format!("any-sync-{}", target));
            if src_bin.exists() {
                let current_exe = env::current_exe().unwrap();
                // target/debug/[build/hash/build-script-build] -> target/debug/deps/
                let deps_dir = current_exe.ancestors().nth(3).unwrap().join("deps");

                let dst_bin = deps_dir.join("any-sync");
                println!(
                    "cargo:warning=Linking test binary from {:?} to {:?}",
                    src_bin, dst_bin
                );
                let _ = fs::remove_file(&dst_bin);
                std::os::unix::fs::symlink(&src_bin, &dst_bin).unwrap();
            }
        }
    }

    tauri_build::build()
}

// Learn more about Tauri commands at https://v2.tauri.app/develop/calling-rust/#commands
#[tauri::command]
fn greet(name: &str) -> String {
    format!("Hello, {}! You've been greeted from Rust!", name)
}

/// Creates and configures the Tauri application builder with all plugins and handlers.
/// This function is used by both the production app and integration tests.
pub fn create_app_builder<R: tauri::Runtime>() -> tauri::Builder<R> {
    let mut builder = tauri::Builder::<R>::new().invoke_handler(tauri::generate_handler![greet]);

    // Only add shell plugin for desktop runtimes
    #[cfg(desktop)]
    {
        builder = builder.plugin(tauri_plugin_shell::init());
    }

    builder = builder.plugin(tauri_plugin_any_sync::init());

    // Disable default macOS menu to avoid main thread requirement in tests
    #[cfg(target_os = "macos")]
    {
        builder = builder.enable_macos_default_menu(false);
    }

    builder
}

#[cfg_attr(mobile, tauri::mobile_entry_point)]
pub fn run() {
    // Initialize logging
    env_logger::Builder::from_env(env_logger::Env::default().default_filter_or("info")).init();

    log::info!("Starting Tauri application");
    eprintln!("Starting Tauri application - this should print to stderr");

    create_app_builder::<tauri::Wry>()
        .run(tauri::generate_context!())
        .expect("error while running tauri application");
}

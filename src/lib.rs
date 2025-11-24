use tauri::{
    plugin::{Builder, TauriPlugin},
    Manager, Runtime,
};
#[cfg(desktop)]
use tauri_plugin_shell::ShellExt;

pub use models::*;

#[cfg(desktop)]
mod desktop;
#[cfg(mobile)]
mod mobile;

mod commands;
mod error;
mod models;
mod proto;
mod service;

pub use error::{Error, Result};

// Note: The AnySyncExt trait and direct AnySync access has been removed.
// All functionality is now accessed through the AnySyncService trait,
// which provides a unified interface for both desktop and mobile platforms.

/// Initializes the plugin.
pub fn init<R: Runtime>() -> TauriPlugin<R> {
    Builder::new("any-sync")
        .invoke_handler(tauri::generate_handler![
            commands::ping,
            commands::storage_put,
            commands::storage_get,
            commands::storage_delete,
            commands::storage_list
        ])
        .setup(|app, api| {
            log::debug!("Initializing any-sync plugin");

            // Initialize shell plugin for sidecar support
            #[cfg(desktop)]
            let _shell = app.shell();

            // Create the service trait object based on platform
            #[cfg(mobile)]
            let service: Box<dyn service::AnySyncService> = {
                Box::new(service::mobile::MobileService::new(app, api)?)
            };
            #[cfg(desktop)]
            let service: Box<dyn service::AnySyncService> = {
                Box::new(service::desktop::DesktopService::new(app, api)?)
            };
            
            // Manage the service for use in commands
            app.manage(service);

            log::debug!("any-sync plugin initialized successfully");
            Ok(())
        })
        .build()
}

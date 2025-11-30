use async_trait::async_trait;
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
#[cfg(desktop)]
mod proto;

pub use error::{Error, Result};

/// Service trait that abstracts platform-specific implementations.
/// Desktop uses async gRPC calls, Mobile uses sync FFI wrapped in spawn_blocking.
#[async_trait]
pub trait AnySyncService: Send + Sync {
    /// Ping the backend service
    async fn ping(&self, payload: PingRequest) -> Result<PingResponse>;

    /// Store a document in a collection
    async fn storage_put(&self, payload: PutRequest) -> Result<PutResponse>;

    /// Retrieve a document from a collection
    async fn storage_get(&self, payload: GetRequest) -> Result<GetResponse>;

    /// Delete a document from a collection
    async fn storage_delete(&self, payload: DeleteRequest) -> Result<DeleteResponse>;

    /// List all document IDs in a collection
    async fn storage_list(&self, payload: ListRequest) -> Result<ListResponse>;
}

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
            let service: Box<dyn AnySyncService> = { Box::new(mobile::AnySync::new(app, api)?) };
            #[cfg(desktop)]
            let service: Box<dyn AnySyncService> = { Box::new(desktop::AnySync::new(app, api)?) };

            // Manage the service for use in commands
            app.manage(service);

            log::debug!("any-sync plugin initialized successfully");
            Ok(())
        })
        .build()
}

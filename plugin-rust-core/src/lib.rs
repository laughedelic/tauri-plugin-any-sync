use async_trait::async_trait;
use tauri::{
    plugin::{Builder, TauriPlugin},
    Manager, Runtime,
};

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

/// Backend trait that abstracts platform-specific implementations.
/// Desktop uses gRPC to communicate with sidecar.
/// Mobile uses FFI to call native library.
#[async_trait]
pub trait AnySyncBackend: Send + Sync {
    /// Execute a command - single entry point for all operations.
    ///
    /// # Arguments
    /// * `cmd` - Command name (e.g., "syncspace.v1.SpaceCreate")
    /// * `data` - Protobuf-encoded request bytes
    ///
    /// # Returns
    /// Protobuf-encoded response bytes or error
    async fn command(&self, cmd: &str, data: &[u8]) -> Result<Vec<u8>>;

    /// Register event handler callback (receives protobuf-encoded events)
    fn set_event_handler(&self, handler: Box<dyn Fn(Vec<u8>) + Send + Sync>);

    /// Shutdown the backend
    async fn shutdown(&self) -> Result<()>;
}

/// Initializes the plugin.
pub fn init<R: Runtime>() -> TauriPlugin<R> {
    Builder::new("any-sync")
        .invoke_handler(tauri::generate_handler![commands::command])
        .setup(|app, _api| {
            // Create platform-specific backend
            #[cfg(desktop)]
            let backend: Box<dyn AnySyncBackend> = Box::new(desktop::DesktopBackend::new(app)?);

            #[cfg(mobile)]
            let backend: Box<dyn AnySyncBackend> = Box::new(mobile::MobileBackend::new(app, _api)?);

            // Register backend as app state
            app.manage(backend);

            Ok(())
        })
        .build()
}

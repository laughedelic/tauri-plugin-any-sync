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
mod proto;



pub use error::{Error, Result};

#[cfg(desktop)]
use desktop::AnySync;
#[cfg(mobile)]
use mobile::AnySync;

/// Extensions to [`tauri::App`], [`tauri::AppHandle`] and [`tauri::Window`] to access the any-sync APIs.
pub trait AnySyncExt<R: Runtime> {
    fn any_sync(&self) -> &AnySync<R>;
}

impl<R: Runtime, T: Manager<R>> crate::AnySyncExt<R> for T {
    fn any_sync(&self) -> &AnySync<R> {
        self.state::<AnySync<R>>().inner()
    }
}

/// Initializes the plugin.
pub fn init<R: Runtime>() -> TauriPlugin<R> {
    Builder::new("any-sync")
        .invoke_handler(tauri::generate_handler![commands::ping])
        .setup(|app, api| {
            #[cfg(mobile)]
            let any_sync = mobile::init(app, api)?;
            #[cfg(desktop)]
            let any_sync = desktop::init(app, api)?;
            app.manage(any_sync);
            Ok(())
        })
        .build()
}

use async_trait::async_trait;
use log::{debug, error, info};
use serde::de::DeserializeOwned;
use tauri::{
    plugin::{PluginApi, PluginHandle},
    AppHandle, Runtime,
};

use crate::{AnySyncBackend, Result};

#[cfg(target_os = "ios")]
tauri::ios_plugin_binding!(init_plugin_any_sync);

/// Mobile backend that calls native FFI (Android Kotlin or iOS Swift).
pub struct MobileBackend<R: Runtime>(PluginHandle<R>);

impl<R: Runtime> MobileBackend<R> {
    /// Initialize the mobile backend by registering platform-specific plugin.
    pub fn new<C: DeserializeOwned>(_app: &AppHandle<R>, api: PluginApi<R, C>) -> Result<Self> {
        #[cfg(target_os = "android")]
        let handle = api.register_android_plugin("com.plugin.anysync", "AnySyncPlugin")?;
        #[cfg(target_os = "ios")]
        let handle = api.register_ios_plugin(init_plugin_any_sync)?;

        info!("Mobile backend initialized");
        Ok(MobileBackend(handle))
    }

    /// Helper to call mobile plugin methods asynchronously via spawn_blocking.
    async fn call_plugin<F, T>(&self, f: F) -> Result<T>
    where
        F: FnOnce(&PluginHandle<R>) -> Result<T> + Send + 'static,
        T: Send + 'static,
    {
        let handle = PluginHandle::clone(&self.0);
        tokio::task::spawn_blocking(move || f(&handle))
            .await
            .map_err(|e| format!("Task join error: {}", e).into())?
    }
}

#[async_trait]
impl<R: Runtime> AnySyncBackend for MobileBackend<R> {
    async fn command(&self, cmd: &str, data: &[u8]) -> Result<Vec<u8>> {
        debug!("Mobile backend: executing command '{}'", cmd);

        let cmd = cmd.to_string();
        let data = data.to_vec();

        self.call_plugin(move |handle| {
            handle
                .run_mobile_plugin("command", (cmd, data))
                .map_err(|e| {
                    error!("Mobile plugin call failed: {}", e);
                    format!("Mobile plugin error: {}", e).into()
                })
        })
        .await
    }

    fn set_event_handler(&self, _handler: Box<dyn Fn(Vec<u8>) + Send + Sync>) {
        // TODO: Implement event handler registration for mobile
        debug!("Event handler set (not yet implemented on mobile)");
    }

    async fn shutdown(&self) -> Result<()> {
        info!("Mobile backend shutdown");
        Ok(())
    }
}

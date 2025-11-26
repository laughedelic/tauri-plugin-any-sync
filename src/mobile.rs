use async_trait::async_trait;
use serde::de::DeserializeOwned;
use tauri::{
    plugin::{PluginApi, PluginHandle},
    AppHandle, Runtime,
};

use crate::{models::*, AnySyncService, Result};

#[cfg(target_os = "ios")]
tauri::ios_plugin_binding!(init_plugin_any_sync);

/// Access to the any-sync APIs on mobile platforms.
/// Wraps platform-specific plugin calls (Kotlin on Android, Swift on iOS).
pub struct AnySync<R: Runtime>(PluginHandle<R>);

impl<R: Runtime> AnySync<R> {
    /// Initialize the mobile service by registering platform-specific plugins.
    pub fn new<C: DeserializeOwned>(
        _app: &AppHandle<R>,
        api: PluginApi<R, C>,
    ) -> crate::Result<AnySync<R>> {
        #[cfg(target_os = "android")]
        let handle = api.register_android_plugin("com.plugin.anysync", "AnySyncPlugin")?;
        #[cfg(target_os = "ios")]
        let handle = api.register_ios_plugin(init_plugin_any_sync)?;
        Ok(AnySync(handle))
    }

    /// Helper to call mobile plugin methods asynchronously via spawn_blocking
    async fn call_plugin<F, T>(&self, f: F) -> Result<T>
    where
        F: FnOnce(&PluginHandle<R>) -> crate::Result<T> + Send + 'static,
        T: Send + 'static,
    {
        let handle = PluginHandle::clone(&self.0);
        tokio::task::spawn_blocking(move || f(&handle))
            .await
            .map_err(|e| crate::Error::Storage(format!("Task join error: {}", e)))?
    }
}

#[async_trait]
impl<R: Runtime> AnySyncService for AnySync<R> {
    async fn ping(&self, _payload: PingRequest) -> Result<PingResponse> {
        // Mobile ping is a no-op, just return success
        Ok(PingResponse {
            value: Some("pong (mobile)".to_string()),
        })
    }

    async fn storage_put(&self, payload: PutRequest) -> Result<PutResponse> {
        self.call_plugin(move |handle| {
            handle
                .run_mobile_plugin("storagePut", payload)
                .map_err(Into::into)
        })
        .await
    }

    async fn storage_get(&self, payload: GetRequest) -> Result<GetResponse> {
        self.call_plugin(move |handle| {
            handle
                .run_mobile_plugin("storageGet", payload)
                .map_err(Into::into)
        })
        .await
    }

    async fn storage_delete(&self, payload: DeleteRequest) -> Result<DeleteResponse> {
        self.call_plugin(move |handle| {
            handle
                .run_mobile_plugin("storageDelete", payload)
                .map_err(Into::into)
        })
        .await
    }

    async fn storage_list(&self, payload: ListRequest) -> Result<ListResponse> {
        self.call_plugin(move |handle| {
            handle
                .run_mobile_plugin("storageList", payload)
                .map_err(Into::into)
        })
        .await
    }
}

use serde::de::DeserializeOwned;
use tauri::{
    plugin::{PluginApi, PluginHandle},
    AppHandle, Runtime,
};

use crate::models::*;

#[cfg(target_os = "ios")]
tauri::ios_plugin_binding!(init_plugin_any_sync);

// initializes the Kotlin or Swift plugin classes
pub fn init<R: Runtime, C: DeserializeOwned>(
    _app: &AppHandle<R>,
    api: PluginApi<R, C>,
) -> crate::Result<AnySync<R>> {
    #[cfg(target_os = "android")]
    let handle = api.register_android_plugin("com.plugin.anysync", "ExamplePlugin")?;
    #[cfg(target_os = "ios")]
    let handle = api.register_ios_plugin(init_plugin_any_sync)?;
    Ok(AnySync(handle))
}

/// Access to the any-sync APIs.
pub struct AnySync<R: Runtime>(PluginHandle<R>);

impl<R: Runtime> AnySync<R> {
    pub fn storage_get(&self, payload: GetRequest) -> crate::Result<GetResponse> {
        self.0
            .run_mobile_plugin("storageGet", payload)
            .map_err(Into::into)
    }

    pub fn storage_put(&self, payload: PutRequest) -> crate::Result<PutResponse> {
        self.0
            .run_mobile_plugin("storagePut", payload)
            .map_err(Into::into)
    }

    pub fn storage_delete(&self, payload: DeleteRequest) -> crate::Result<DeleteResponse> {
        self.0
            .run_mobile_plugin("storageDelete", payload)
            .map_err(Into::into)
    }

    pub fn storage_list(&self, payload: ListRequest) -> crate::Result<ListResponse> {
        self.0
            .run_mobile_plugin("storageList", payload)
            .map_err(Into::into)
    }
}

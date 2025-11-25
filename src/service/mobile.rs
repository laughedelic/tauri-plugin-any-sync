use async_trait::async_trait;
use log::info;
use serde::de::DeserializeOwned;
use std::sync::Arc;
use tauri::{plugin::PluginApi, AppHandle, Runtime};

use crate::mobile::{init as mobile_init, AnySync};
use crate::models::*;
use crate::Result;

/// Mobile service implementation using sync FFI wrapped in spawn_blocking
pub struct MobileService<R: Runtime> {
    any_sync: Arc<AnySync<R>>,
}

impl<R: Runtime> MobileService<R> {
    pub fn new<C: DeserializeOwned>(app: &AppHandle<R>, api: PluginApi<R, C>) -> Result<Self> {
        let any_sync = mobile_init(app, api)?;
        Ok(Self {
            any_sync: Arc::new(any_sync),
        })
    }
}

#[async_trait]
impl<R: Runtime> super::AnySyncService for MobileService<R> {
    async fn call_blocking<F, T>(&self, f: F) -> Result<T>
    where
        F: FnOnce() -> Result<T> + Send + 'static,
        T: Send + 'static,
    {
        tokio::task::spawn_blocking(f)
            .await
            .map_err(|e| crate::Error::Storage(format!("Task join error: {}", e)))?
    }

    // Mobile ping is a no-op, just return success
    async fn ping(&self, payload: PingRequest) -> Result<PingResponse> {
        info!("Mobile service: ping");
        Ok(PingResponse {
            value: Some("pong (mobile)".to_string()),
        })
    }

    async fn storage_put(&self, payload: PutRequest) -> Result<PutResponse> {
        info!("Mobile service: storage_put");
        self.call_blocking(move || self.any_sync.clone().storage_put(payload))
            .await
    }

    async fn storage_get(&self, payload: GetRequest) -> Result<GetResponse> {
        info!("Mobile service: storage_get");
        self.call_blocking(move || self.any_sync.clone().storage_get(payload))
            .await
    }

    async fn storage_delete(&self, payload: DeleteRequest) -> Result<DeleteResponse> {
        info!("Mobile service: storage_delete");
        self.call_blocking(move || self.any_sync.clone().storage_delete(payload))
            .await
    }

    async fn storage_list(&self, payload: ListRequest) -> Result<ListResponse> {
        info!("Mobile service: storage_list");
        self.call_blocking(move || self.any_sync.clone().storage_list(payload))
            .await
    }
}

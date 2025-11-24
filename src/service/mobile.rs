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
    async fn ping(&self, payload: PingRequest) -> Result<PingResponse> {
        info!("Mobile service: ping");
        let any_sync = self.any_sync.clone();
        tokio::task::spawn_blocking(move || {
            // Mobile ping is a no-op, just return success
            Ok(PingResponse {
                value: Some("pong (mobile)".to_string()),
            })
        })
        .await
        .map_err(|e| crate::Error::Storage(format!("Task join error: {}", e)))?
    }

    async fn storage_put(&self, payload: PutRequest) -> Result<PutResponse> {
        info!("Mobile service: storage_put");
        let any_sync = self.any_sync.clone();
        tokio::task::spawn_blocking(move || any_sync.storage_put(payload))
            .await
            .map_err(|e| crate::Error::Storage(format!("Task join error: {}", e)))?
    }

    async fn storage_get(&self, payload: GetRequest) -> Result<GetResponse> {
        info!("Mobile service: storage_get");
        let any_sync = self.any_sync.clone();
        tokio::task::spawn_blocking(move || any_sync.storage_get(payload))
            .await
            .map_err(|e| crate::Error::Storage(format!("Task join error: {}", e)))?
    }

    async fn storage_delete(&self, payload: DeleteRequest) -> Result<DeleteResponse> {
        info!("Mobile service: storage_delete");
        let any_sync = self.any_sync.clone();
        tokio::task::spawn_blocking(move || any_sync.storage_delete(payload))
            .await
            .map_err(|e| crate::Error::Storage(format!("Task join error: {}", e)))?
    }

    async fn storage_list(&self, payload: ListRequest) -> Result<ListResponse> {
        info!("Mobile service: storage_list");
        let any_sync = self.any_sync.clone();
        tokio::task::spawn_blocking(move || any_sync.storage_list(payload))
            .await
            .map_err(|e| crate::Error::Storage(format!("Task join error: {}", e)))?
    }
}

use async_trait::async_trait;
use log::info;
use serde::de::DeserializeOwned;
use std::sync::Arc;
use tauri::{plugin::PluginApi, AppHandle, Runtime};

use crate::desktop::{init as desktop_init, AnySync};
use crate::models::*;
use crate::Result;

/// Desktop service implementation using async gRPC calls to sidecar
pub struct DesktopService<R: Runtime> {
    any_sync: Arc<AnySync<R>>,
}

impl<R: Runtime> DesktopService<R> {
    pub fn new<C: DeserializeOwned>(app: &AppHandle<R>, api: PluginApi<R, C>) -> Result<Self> {
        let any_sync = desktop_init(app, api)?;
        Ok(Self {
            any_sync: Arc::new(any_sync),
        })
    }
}

#[async_trait]
impl<R: Runtime> super::AnySyncService for DesktopService<R> {
    async fn ping(&self, payload: PingRequest) -> Result<PingResponse> {
        info!("Desktop service: ping");
        self.any_sync.ping(payload).await
    }

    async fn storage_put(&self, payload: PutRequest) -> Result<PutResponse> {
        info!("Desktop service: storage_put");
        self.any_sync.storage_put(payload).await
    }

    async fn storage_get(&self, payload: GetRequest) -> Result<GetResponse> {
        info!("Desktop service: storage_get");
        self.any_sync.storage_get(payload).await
    }

    async fn storage_delete(&self, payload: DeleteRequest) -> Result<DeleteResponse> {
        info!("Desktop service: storage_delete");
        self.any_sync.storage_delete(payload).await
    }

    async fn storage_list(&self, payload: ListRequest) -> Result<ListResponse> {
        info!("Desktop service: storage_list");
        self.any_sync.storage_list(payload).await
    }
}

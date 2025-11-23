use log::{debug, error, info};
use tauri::{command, AppHandle, Runtime};

use crate::models::*;
use crate::AnySyncExt;
use crate::Result;

#[command]
pub(crate) async fn ping<R: Runtime>(
    app: AppHandle<R>,
    payload: PingRequest,
) -> Result<PingResponse> {
    info!("Received ping command from frontend");
    debug!("Ping payload: {:?}", payload);

    match app.any_sync().ping(payload).await {
        Ok(response) => {
            info!("Ping command completed successfully");
            debug!("Ping response: {:?}", response);
            Ok(response)
        }
        Err(e) => {
            error!("Ping command failed: {}", e);
            Err(e)
        }
    }
}

#[command]
pub(crate) async fn storage_put<R: Runtime>(
    app: AppHandle<R>,
    payload: PutRequest,
) -> Result<PutResponse> {
    info!("Received storage_put command from frontend");
    debug!("Put payload: collection={}, id={}", payload.collection, payload.id);

    match app.any_sync().storage_put(payload).await {
        Ok(response) => {
            info!("Storage put command completed successfully");
            Ok(response)
        }
        Err(e) => {
            error!("Storage put command failed: {}", e);
            Err(e)
        }
    }
}

#[command]
pub(crate) async fn storage_get<R: Runtime>(
    app: AppHandle<R>,
    payload: GetRequest,
) -> Result<GetResponse> {
    info!("Received storage_get command from frontend");
    debug!("Get payload: collection={}, id={}", payload.collection, payload.id);

    match app.any_sync().storage_get(payload).await {
        Ok(response) => {
            info!("Storage get command completed successfully");
            Ok(response)
        }
        Err(e) => {
            error!("Storage get command failed: {}", e);
            Err(e)
        }
    }
}

#[command]
pub(crate) async fn storage_list<R: Runtime>(
    app: AppHandle<R>,
    payload: ListRequest,
) -> Result<ListResponse> {
    info!("Received storage_list command from frontend");
    debug!("List payload: collection={}", payload.collection);

    match app.any_sync().storage_list(payload).await {
        Ok(response) => {
            info!("Storage list command completed successfully");
            Ok(response)
        }
        Err(e) => {
            error!("Storage list command failed: {}", e);
            Err(e)
        }
    }
}

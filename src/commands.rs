use log::{debug, error, info};
use tauri::{command, State};

use crate::models::*;
use crate::AnySyncService;
use crate::Result;

#[command]
pub(crate) async fn ping(
    service: State<'_, Box<dyn AnySyncService>>,
    payload: PingRequest,
) -> Result<PingResponse> {
    info!("Received ping command from frontend");
    debug!("Ping payload: {:?}", payload);

    match service.ping(payload).await {
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
pub(crate) async fn storage_put(
    service: State<'_, Box<dyn AnySyncService>>,
    payload: PutRequest,
) -> Result<PutResponse> {
    info!("Received storage_put command from frontend");
    debug!(
        "Put payload: collection={}, id={}",
        payload.collection, payload.id
    );

    match service.storage_put(payload).await {
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
pub(crate) async fn storage_get(
    service: State<'_, Box<dyn AnySyncService>>,
    payload: GetRequest,
) -> Result<GetResponse> {
    info!("Received storage_get command from frontend");
    debug!(
        "Get payload: collection={}, id={}",
        payload.collection, payload.id
    );

    match service.storage_get(payload).await {
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
pub(crate) async fn storage_delete(
    service: State<'_, Box<dyn AnySyncService>>,
    payload: DeleteRequest,
) -> Result<DeleteResponse> {
    info!("Received storage_delete command from frontend");
    debug!(
        "Delete payload: collection={}, id={}",
        payload.collection, payload.id
    );

    match service.storage_delete(payload).await {
        Ok(response) => {
            info!("Storage delete command completed successfully");
            Ok(response)
        }
        Err(e) => {
            error!("Storage delete command failed: {}", e);
            Err(e)
        }
    }
}

#[command]
pub(crate) async fn storage_list(
    service: State<'_, Box<dyn AnySyncService>>,
    payload: ListRequest,
) -> Result<ListResponse> {
    info!("Received storage_list command from frontend");
    debug!("List payload: collection={}", payload.collection);

    match service.storage_list(payload).await {
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

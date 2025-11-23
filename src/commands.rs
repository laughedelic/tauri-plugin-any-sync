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

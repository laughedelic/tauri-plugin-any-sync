use tauri::{AppHandle, command, Runtime};

use crate::models::*;
use crate::Result;
use crate::AnySyncExt;

#[command]
pub(crate) async fn ping<R: Runtime>(
    app: AppHandle<R>,
    payload: PingRequest,
) -> Result<PingResponse> {
    app.any_sync().ping(payload)
}

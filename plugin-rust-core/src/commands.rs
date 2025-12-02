use log::{debug, error, info};
use tauri::{command, State};

use crate::AnySyncBackend;
use crate::Result;

/// Single command dispatch handler.
/// Takes a command name and opaque request bytes, forwards to backend,
/// and returns opaque response bytes.
#[command]
pub(crate) async fn command(
    backend: State<'_, Box<dyn AnySyncBackend>>,
    cmd: String,
    data: Vec<u8>,
) -> Result<Vec<u8>> {
    info!("Received command from frontend: {}", cmd);
    debug!("Command data length: {} bytes", data.len());

    match backend.command(&cmd, &data).await {
        Ok(response) => {
            info!("Command '{}' completed successfully", cmd);
            debug!("Response length: {} bytes", response.len());
            Ok(response)
        }
        Err(e) => {
            error!("Command '{}' failed: {}", cmd, e);
            Err(e)
        }
    }
}

use tauri::{command, ipc, State};

use crate::AnySyncBackend;
use crate::Result;

/// Single command dispatch handler.
/// Receives raw protobuf bytes from TypeScript, forwards to backend,
/// and returns raw protobuf bytes back.
///
/// Command name is passed via X-Command header to avoid JSON serialization
/// of the binary payload.
#[command]
pub(crate) async fn command(
    backend: State<'_, Box<dyn AnySyncBackend>>,
    request: ipc::Request<'_>,
) -> Result<ipc::Response> {
    // Extract command name from header
    let cmd = request
        .headers()
        .get("X-Command")
        .ok_or_else(|| crate::Error::Storage("Missing X-Command header".to_string()))?
        .to_str()
        .map_err(|e| crate::Error::Storage(format!("Invalid X-Command header: {e}")))?;

    // Extract raw protobuf bytes from request body
    let ipc::InvokeBody::Raw(request_bytes) = request.body() else {
        return Err(crate::Error::Storage(
            "Request body must be raw bytes".to_string(),
        ));
    };

    // Call backend with protobuf bytes
    let response_bytes = backend
        .command(cmd, request_bytes)
        .await
        .map_err(|e| crate::Error::Storage(format!("Command failed: {e}")))?;

    // Return raw protobuf bytes (no JSON serialization)
    Ok(ipc::Response::new(response_bytes))
}

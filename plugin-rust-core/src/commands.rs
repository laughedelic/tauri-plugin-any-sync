use tauri::{command, ipc, State};

use crate::AnySyncBackend;
use crate::Result;

/// Single command dispatch handler.
/// Receives raw protobuf bytes from TypeScript, forwards to backend,
/// and returns raw protobuf bytes back.
///
/// Platform-specific behavior:
/// - Desktop: Command name in X-Command header, request body is raw bytes (InvokeBody::Raw)
/// - Mobile: Command name in X-Command header, request body is JSON array (InvokeBody::Json)
///           because mobile IPC serializes Uint8Array as JSON array
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

    // Extract protobuf bytes from request body
    // Desktop: InvokeBody::Raw (Uint8Array sent as raw binary)
    // Mobile: InvokeBody::Json (Uint8Array serialized as JSON array)
    let request_bytes = match request.body() {
        #[cfg(not(target_os = "android"))]
        ipc::InvokeBody::Raw(bytes) => bytes.to_vec(),

        // NOTE: this is an issue with Tauri IPC serialization on android and linux specifically (https://github.com/tauri-apps/tauri/issues/10573)
        // TODO: test on linux and adjust
        #[cfg(target_os = "android")]
        ipc::InvokeBody::Json(value) => {
            // On mobile, Uint8Array is serialized as JSON array of numbers
            let array = value
                .as_array()
                .ok_or_else(|| crate::Error::Storage("Request body must be array".to_string()))?;

            array
                .iter()
                .map(|v| {
                    v.as_u64()
                        .and_then(|n| u8::try_from(n).ok())
                        .ok_or_else(|| crate::Error::Storage("Invalid byte in array".to_string()))
                })
                .collect::<Result<Vec<u8>>>()?
        }

        _ => {
            return Err(crate::Error::Storage(format!(
                "Unsupported request body type: {:?}",
                request.body()
            )));
        }
    };

    // Call backend with protobuf bytes
    let response_bytes = backend
        .command(cmd, &request_bytes)
        .await
        .map_err(|e| crate::Error::Storage(format!("Command failed: {e}")))?;

    // Return raw protobuf bytes (no JSON serialization)
    Ok(ipc::Response::new(response_bytes))
}

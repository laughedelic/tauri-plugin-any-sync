use log::{debug, error, info, warn};
use serde::de::DeserializeOwned;
use std::sync::Arc;
use std::time::SystemTime;
use tauri::{plugin::PluginApi, AppHandle, Runtime};
use tauri_plugin_shell::ShellExt;
use tokio::sync::RwLock;
use tokio::time::{sleep, timeout, Duration};
use tonic::transport::{Channel, Endpoint};
use tonic::Request;

use crate::models::*;
use crate::proto::anysync::{
    health_service_client::HealthServiceClient, HealthCheckRequest, PingRequest as GrpcPingRequest,
    PingResponse as GrpcPingResponse,
};
use crate::Result;

pub fn init<R: Runtime, C: DeserializeOwned>(
    app: &AppHandle<R>,
    _api: PluginApi<R, C>,
) -> crate::Result<AnySync<R>> {
    // Initialize sidecar process manager
    let manager = Arc::new(RwLock::new(SidecarManager::new()));

    Ok(AnySync {
        app: app.clone(),
        manager,
    })
}

/// Manages Go backend sidecar process
pub struct SidecarManager {
    child: Option<tauri_plugin_shell::process::CommandChild>,
    port: Option<u16>,
    client: Option<HealthServiceClient<Channel>>,
    is_running: bool,
}

impl SidecarManager {
    pub fn new() -> Self {
        Self {
            child: None,
            port: None,
            client: None,
            is_running: false,
        }
    }

    async fn start<R: Runtime>(&mut self, app: &AppHandle<R>) -> Result<()> {
        if self.is_running {
            info!("Sidecar already running, skipping startup");
            return Ok(());
        }

        info!("Starting Go backend sidecar...");

        // Create a temporary file for port communication
        let port_file = tempfile::NamedTempFile::new()?;
        let port_file_path = port_file.path().to_str().unwrap();
        debug!("Created port file: {}", port_file_path);

        // Start sidecar using Tauri shell plugin
        // Tauri automatically finds any-sync-{target-triple} in plugin's binaries/
        debug!("Creating sidecar command for 'any-sync'");
        let sidecar_command = app.shell().sidecar("any-sync").map_err(|e| {
            error!("Failed to create sidecar command: {}", e);
            std::io::Error::other(format!("Failed to create sidecar command: {}", e))
        })?;

        debug!("Spawning sidecar process...");
        let (_rx, child) = sidecar_command
            .env("ANY_SYNC_PORT_FILE", port_file_path)
            .spawn()
            .map_err(|e| {
                error!("Failed to spawn sidecar: {}", e);
                std::io::Error::other(format!("Failed to spawn sidecar: {}", e))
            })?;
        debug!("Sidecar process spawned successfully");

        // Wait for server to write port
        debug!("Waiting for server to write port to file...");
        let port = self
            .wait_for_port(port_file_path, Duration::from_secs(10))
            .await?;
        debug!("Server listening on port {}", port);
        // Connect to server
        debug!("Connecting to gRPC server at localhost:{}", port);
        let endpoint =
            Endpoint::from_shared(format!("http://localhost:{}", port)).map_err(|e| {
                error!("Invalid endpoint: {}", e);
                std::io::Error::other(format!("Invalid endpoint: {}", e))
            })?;
        let channel = endpoint.connect().await.map_err(|e| {
            error!("Failed to connect to gRPC server: {}", e);
            std::io::Error::other(format!("Failed to connect: {}", e))
        })?;
        let mut client = HealthServiceClient::new(channel);
        debug!("gRPC client created successfully");

        // Test the connection
        debug!("Testing gRPC connection with health check...");
        self.test_connection(&mut client).await?;
        debug!("Health check passed, sidecar is ready");

        self.child = Some(child);
        self.port = Some(port);
        self.client = Some(client);
        self.is_running = true;

        info!("Sidecar startup complete");
        Ok(())
    }

    async fn wait_for_port(&self, port_file: &str, timeout_duration: Duration) -> Result<u16> {
        let start_time = tokio::time::Instant::now();
        debug!(
            "Waiting for port file with timeout of {:?}",
            timeout_duration
        );

        loop {
            if start_time.elapsed() > timeout_duration {
                error!(
                    "Timeout waiting for server to start after {:?}",
                    timeout_duration
                );
                return Err(std::io::Error::new(
                    std::io::ErrorKind::TimedOut,
                    "Timeout waiting for server to start",
                )
                .into());
            }

            if let Ok(content) = tokio::fs::read_to_string(port_file).await {
                debug!("Port file content: {}", content);
                if let Ok(port) = content.trim().parse::<u16>() {
                    info!("Successfully read port {} from file", port);
                    return Ok(port);
                } else {
                    warn!("Failed to parse port from file content: {}", content);
                }
            }

            sleep(Duration::from_millis(100)).await;
        }
    }

    async fn test_connection(&mut self, client: &mut HealthServiceClient<Channel>) -> Result<()> {
        let request = Request::new(HealthCheckRequest {});
        debug!("Sending health check request");

        match timeout(Duration::from_secs(5), client.check(request)).await {
            Ok(Ok(response)) => {
                let response = response.into_inner();
                debug!("Health check response status: {:?}", response.status());
                if response.status()
                    == crate::proto::anysync::health_check_response::ServingStatus::Serving
                {
                    info!("Health check successful - server is serving");
                    Ok(())
                } else {
                    error!("Server is not serving: {:?}", response.status());
                    Err(std::io::Error::other("Server is not serving").into())
                }
            }
            Ok(Err(e)) => {
                error!("Connection test failed: {}", e);
                Err(std::io::Error::other(format!("Connection test failed: {}", e)).into())
            }
            Err(_) => {
                error!("Connection test timed out after 5 seconds");
                Err(
                    std::io::Error::new(std::io::ErrorKind::TimedOut, "Connection test timed out")
                        .into(),
                )
            }
        }
    }

    pub async fn stop(&mut self) -> Result<()> {
        info!("Stopping sidecar process");
        if let Some(child) = self.child.take() {
            match child.kill() {
                Ok(_) => info!("Sidecar process killed successfully"),
                Err(e) => warn!("Failed to kill sidecar process: {}", e),
            }
        }

        self.is_running = false;
        self.port = None;
        self.client = None;

        info!("Sidecar stopped");
        Ok(())
    }

    pub async fn ping<R: Runtime>(
        &mut self,
        app: &AppHandle<R>,
        message: Option<String>,
    ) -> Result<GrpcPingResponse> {
        if !self.is_running {
            info!("Sidecar not running, starting it first");
            self.start(app).await?;
        }

        if let Some(client) = &mut self.client {
            let timestamp = SystemTime::now()
                .duration_since(SystemTime::UNIX_EPOCH)
                .unwrap()
                .as_secs();

            let msg = message.as_deref().unwrap_or("");
            debug!("Sending ping request with message: '{}'", msg);

            let request = Request::new(GrpcPingRequest {
                message: message.unwrap_or_default(),
                timestamp: timestamp as i64,
            });

            match timeout(Duration::from_secs(10), client.ping(request)).await {
                Ok(Ok(response)) => {
                    let response = response.into_inner();
                    info!("Ping successful, received: '{}'", response.message);
                    Ok(response)
                }
                Ok(Err(e)) => {
                    error!("Ping failed: {}", e);
                    Err(std::io::Error::other(format!("Ping failed: {}", e)).into())
                }
                Err(_) => {
                    error!("Ping timed out after 10 seconds");
                    Err(std::io::Error::new(std::io::ErrorKind::TimedOut, "Ping timed out").into())
                }
            }
        } else {
            error!("No gRPC client available");
            Err(
                std::io::Error::new(std::io::ErrorKind::NotConnected, "No gRPC client available")
                    .into(),
            )
        }
    }
}

/// Access to any-sync APIs.
pub struct AnySync<R: Runtime> {
    app: AppHandle<R>,
    manager: Arc<RwLock<SidecarManager>>,
}

impl<R: Runtime> AnySync<R> {
    pub async fn ping(&self, payload: PingRequest) -> crate::Result<PingResponse> {
        let mut manager = self.manager.write().await;
        let response = manager.ping(&self.app, payload.value).await?;

        Ok(PingResponse {
            value: Some(response.message),
        })
    }
}

impl<R: Runtime> Drop for AnySync<R> {
    fn drop(&mut self) {
        let manager = self.manager.clone();
        let rt = tokio::runtime::Runtime::new();
        if let Ok(rt) = rt {
            rt.block_on(async {
                let _ = manager.write().await.stop().await;
            });
        }
    }
}

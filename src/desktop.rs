use serde::de::DeserializeOwned;
use std::path::PathBuf;
use std::sync::Arc;
use std::time::SystemTime;
use tauri::{plugin::PluginApi, AppHandle, Runtime};
use tokio::process::Command;
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
    process: Option<tokio::process::Child>,
    port: Option<u16>,
    client: Option<HealthServiceClient<Channel>>,
    is_running: bool,
}

impl SidecarManager {
    pub fn new() -> Self {
        Self {
            process: None,
            port: None,
            client: None,
            is_running: false,
        }
    }

    async fn start(&mut self) -> Result<()> {
        if self.is_running {
            return Ok(());
        }

        // Find server binary
        let binary_path = self.find_server_binary()?;

        // Create a temporary file for port communication
        let port_file = tempfile::NamedTempFile::new()?;
        let port_file_path = port_file.path().to_str().unwrap();

        // Start server process
        let mut cmd = Command::new(&binary_path);
        cmd.env("ANY_SYNC_PORT_FILE", port_file_path);

        #[cfg(unix)]
        {
            cmd.kill_on_drop(true);
        }

        let child = cmd.spawn().map_err(|e| {
            std::io::Error::new(
                std::io::ErrorKind::Other,
                format!("Failed to start server: {}", e),
            )
        })?;

        // Wait for server to write port
        let port = self
            .wait_for_port(port_file_path, Duration::from_secs(10))
            .await?;

        // Connect to server
        let endpoint =
            Endpoint::from_shared(format!("http://localhost:{}", port)).map_err(|e| {
                std::io::Error::new(
                    std::io::ErrorKind::Other,
                    format!("Invalid endpoint: {}", e),
                )
            })?;
        let channel = endpoint.connect().await.map_err(|e| {
            std::io::Error::new(
                std::io::ErrorKind::Other,
                format!("Failed to connect: {}", e),
            )
        })?;
        let mut client = HealthServiceClient::new(channel);

        // Test the connection
        self.test_connection(&mut client).await?;

        self.process = Some(child);
        self.port = Some(port);
        self.client = Some(client);
        self.is_running = true;

        Ok(())
    }

    fn find_server_binary(&self) -> Result<PathBuf> {
        // Try different binary names based on platform
        let binary_names = if cfg!(target_os = "windows") {
            vec!["server.exe", "server-windows-amd64.exe"]
        } else if cfg!(target_os = "macos") {
            vec!["server", "server-darwin-arm64", "server-darwin-amd64"]
        } else {
            vec!["server", "server-linux-amd64", "server-linux-arm64"]
        };

        let binaries_dir = PathBuf::from("binaries");

        for name in binary_names {
            let path = binaries_dir.join(name);
            if path.exists() {
                return Ok(path);
            }
        }

        Err(std::io::Error::new(
            std::io::ErrorKind::NotFound,
            "Server binary not found. Please run ./build-go-backend.sh first.",
        )
        .into())
    }

    async fn wait_for_port(&self, port_file: &str, timeout_duration: Duration) -> Result<u16> {
        let start_time = tokio::time::Instant::now();

        loop {
            if start_time.elapsed() > timeout_duration {
                return Err(std::io::Error::new(
                    std::io::ErrorKind::TimedOut,
                    "Timeout waiting for server to start",
                )
                .into());
            }

            if let Ok(content) = tokio::fs::read_to_string(port_file).await {
                if let Ok(port) = content.trim().parse::<u16>() {
                    return Ok(port);
                }
            }

            sleep(Duration::from_millis(100)).await;
        }
    }

    async fn test_connection(&mut self, client: &mut HealthServiceClient<Channel>) -> Result<()> {
        let request = Request::new(HealthCheckRequest {});

        match timeout(Duration::from_secs(5), client.check(request)).await {
            Ok(Ok(response)) => {
                let response = response.into_inner();
                if response.status()
                    == crate::proto::anysync::health_check_response::ServingStatus::Serving
                {
                    Ok(())
                } else {
                    Err(
                        std::io::Error::new(std::io::ErrorKind::Other, "Server is not serving")
                            .into(),
                    )
                }
            }
            Ok(Err(e)) => Err(std::io::Error::new(
                std::io::ErrorKind::Other,
                format!("Connection test failed: {}", e),
            )
            .into()),
            Err(_) => Err(std::io::Error::new(
                std::io::ErrorKind::TimedOut,
                "Connection test timed out",
            )
            .into()),
        }
    }

    pub async fn stop(&mut self) -> Result<()> {
        if let Some(mut process) = self.process.take() {
            #[cfg(unix)]
            {
                let _ = process.kill().await;
            }
            #[cfg(not(unix))]
            {
                let _ = process.kill();
            }

            let _ = process.wait().await;
        }

        self.is_running = false;
        self.port = None;
        self.client = None;

        Ok(())
    }

    pub async fn ping(&mut self, message: Option<String>) -> Result<GrpcPingResponse> {
        if !self.is_running {
            self.start().await?;
        }

        if let Some(client) = &mut self.client {
            let timestamp = SystemTime::now()
                .duration_since(SystemTime::UNIX_EPOCH)
                .unwrap()
                .as_secs();

            let request = Request::new(GrpcPingRequest {
                message: message.unwrap_or_default(),
                timestamp: timestamp as i64,
            });

            match timeout(Duration::from_secs(10), client.ping(request)).await {
                Ok(Ok(response)) => Ok(response.into_inner()),
                Ok(Err(e)) => Err(std::io::Error::new(
                    std::io::ErrorKind::Other,
                    format!("Ping failed: {}", e),
                )
                .into()),
                Err(_) => {
                    Err(std::io::Error::new(std::io::ErrorKind::TimedOut, "Ping timed out").into())
                }
            }
        } else {
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
        let response = manager.ping(payload.value).await?;

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

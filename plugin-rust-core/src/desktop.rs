use async_trait::async_trait;
use log::{debug, error, info, warn};
use std::sync::Arc;
use tauri::{AppHandle, Manager, Runtime};
use tauri_plugin_shell::ShellExt;
use tokio::sync::RwLock;
use tokio::time::{sleep, timeout, Duration};
use tonic::transport::{Channel, Endpoint};
use tonic::Request;

use crate::proto::transport::v1::{
    transport_service_client::TransportServiceClient, CommandRequest as GrpcCommandRequest,
    InitRequest,
};
use crate::{AnySyncBackend, Result};

/// Desktop backend that communicates with Go sidecar via gRPC.
pub struct DesktopBackend<R: Runtime> {
    app: AppHandle<R>,
    manager: Arc<RwLock<SidecarManager>>,
}

impl<R: Runtime> DesktopBackend<R> {
    /// Create new desktop backend.
    pub fn new(app: &AppHandle<R>) -> Result<Self> {
        Ok(DesktopBackend {
            app: app.clone(),
            manager: Arc::new(RwLock::new(SidecarManager::new())),
        })
    }

    /// Helper to execute a gRPC operation with timeout and error handling.
    async fn call_grpc<F, T>(&self, op_name: &str, f: F) -> Result<T>
    where
        F: std::future::Future<Output = std::result::Result<tonic::Response<T>, tonic::Status>>,
    {
        match timeout(Duration::from_secs(30), f).await {
            Ok(Ok(response)) => {
                debug!("{} successful", op_name);
                Ok(response.into_inner())
            }
            Ok(Err(e)) => {
                error!("{} failed: {}", op_name, e);
                Err(std::io::Error::other(format!("{}: {}", op_name, e)).into())
            }
            Err(_) => {
                error!("{} timed out", op_name);
                Err(std::io::Error::new(
                    std::io::ErrorKind::TimedOut,
                    format!("{} timed out", op_name),
                )
                .into())
            }
        }
    }
}

/// Manages Go backend sidecar process lifecycle.
struct SidecarManager {
    child: Option<tauri_plugin_shell::process::CommandChild>,
    port: Option<u16>,
    client: Option<TransportServiceClient<Channel>>,
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

    /// Start the sidecar and establish gRPC connection.
    async fn start<R: Runtime>(&mut self, app: &AppHandle<R>) -> Result<()> {
        if self.is_running {
            info!("Sidecar already running, skipping startup");
            return Ok(());
        }

        info!("Starting Go backend sidecar...");

        // Create temp file for port communication
        let port_file = tempfile::NamedTempFile::new()?;
        let port_file_path = port_file.path().to_str().unwrap();
        debug!("Created port file: {}", port_file_path);

        // Get database path (outside src-tauri to avoid file watcher loops)
        let db_path = app
            .path()
            .app_data_dir()
            .map_err(|e| {
                error!("Failed to get app data dir: {}", e);
                std::io::Error::other(format!("Failed to get app data dir: {}", e))
            })?
            .join("any-sync-data");

        // Ensure the data directory exists
        if let Some(parent) = db_path.parent() {
            std::fs::create_dir_all(parent)?;
        }

        let db_path_str = db_path
            .to_str()
            .ok_or_else(|| std::io::Error::other("Invalid database path"))?;
        debug!("Database path: {}", db_path_str);

        // Start sidecar using Tauri shell plugin
        // Tauri automatically finds any-sync-{target-triple} in binaries/
        debug!("Creating sidecar command for 'any-sync'");
        let sidecar_command = app.shell().sidecar("any-sync").map_err(|e| {
            error!("Failed to create sidecar command: {}", e);
            std::io::Error::other(format!("Failed to create sidecar command: {}", e))
        })?;

        debug!("Spawning sidecar process...");
        let (_rx, child) = sidecar_command
            .env("ANY_SYNC_PORT_FILE", port_file_path)
            .env("ANY_SYNC_DB_PATH", db_path_str)
            .spawn()
            .map_err(|e| {
                error!("Failed to spawn sidecar: {}", e);
                std::io::Error::other(format!("Failed to spawn sidecar: {}", e))
            })?;
        debug!("Sidecar process spawned successfully");

        // Wait for port file
        debug!("Waiting for server to write port to file...");
        let port = self
            .wait_for_port(port_file_path, Duration::from_secs(10))
            .await?;
        info!("Server listening on port {}", port);

        // Connect to gRPC server
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

        let client = TransportServiceClient::new(channel);
        debug!("gRPC client created successfully");

        // Test connection with init call
        debug!("Testing gRPC connection with init request...");
        let mut test_client = client.clone();
        let request = Request::new(InitRequest {
            storage_path: db_path_str.to_string(),
            network_id: String::new(),
            config_json: String::new(),
        });
        match timeout(Duration::from_secs(5), test_client.init(request)).await {
            Ok(Ok(_)) => {
                info!("Init call passed, sidecar is ready");
            }
            Ok(Err(e)) => {
                error!("Connection test failed: {}", e);
                return Err(std::io::Error::other(format!("Connection test failed: {}", e)).into());
            }
            Err(_) => {
                error!("Connection test timed out");
                return Err(std::io::Error::new(
                    std::io::ErrorKind::TimedOut,
                    "Connection test timed out",
                )
                .into());
            }
        }

        self.child = Some(child);
        self.port = Some(port);
        self.client = Some(client);
        self.is_running = true;

        info!("Sidecar startup complete");
        Ok(())
    }

    /// Wait for sidecar to write port to file with timeout.
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

    /// Stop the sidecar process.
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

    /// Ensure sidecar is running and return reference to client.
    pub async fn get_client<R: Runtime>(
        &mut self,
        app: &AppHandle<R>,
    ) -> Result<&mut TransportServiceClient<Channel>> {
        if !self.is_running {
            info!("Sidecar not running, starting it first");
            self.start(app).await?;
        }

        self.client.as_mut().ok_or_else(|| {
            std::io::Error::new(
                std::io::ErrorKind::NotConnected,
                "Transport client not available",
            )
            .into()
        })
    }
}

#[async_trait]
impl<R: Runtime> AnySyncBackend for DesktopBackend<R> {
    async fn command(&self, cmd: &str, data: &[u8]) -> Result<Vec<u8>> {
        debug!("Desktop backend: executing command '{}'", cmd);

        let mut manager = self.manager.write().await;
        let client = manager.get_client(&self.app).await?;

        let request = GrpcCommandRequest {
            cmd: cmd.to_string(),
            data: data.to_vec(),
        };

        let response = self
            .call_grpc("Command", client.command(Request::new(request)))
            .await?;

        Ok(response.data)
    }

    fn set_event_handler(&self, _handler: Box<dyn Fn(Vec<u8>) + Send + Sync>) {
        // TODO: Implement event streaming from gRPC Subscribe RPC
        debug!("Event handler set (not yet implemented)");
    }

    async fn shutdown(&self) -> Result<()> {
        info!("Desktop backend shutdown");
        let mut manager = self.manager.write().await;
        manager.stop().await?;
        Ok(())
    }
}

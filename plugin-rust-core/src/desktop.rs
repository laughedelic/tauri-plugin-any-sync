use async_trait::async_trait;
use log::{debug, error, info, warn};
use serde::de::DeserializeOwned;
use std::sync::Arc;
use std::time::SystemTime;
use tauri::{plugin::PluginApi, AppHandle, Manager, Runtime};
use tauri_plugin_shell::ShellExt;
use tokio::sync::RwLock;
use tokio::time::{sleep, timeout, Duration};
use tonic::transport::{Channel, Endpoint};
use tonic::Request;

use crate::models::*;
use crate::proto::anysync::{
    health_service_client::HealthServiceClient, storage_service_client::StorageServiceClient,
    DeleteRequest as GrpcDeleteRequest, GetRequest as GrpcGetRequest, HealthCheckRequest,
    ListRequest as GrpcListRequest, PingRequest as GrpcPingRequest, PutRequest as GrpcPutRequest,
};
use crate::{AnySyncService, Result};

/// Access to any-sync APIs.
pub struct AnySync<R: Runtime> {
    app: AppHandle<R>,
    manager: Arc<RwLock<SidecarManager>>,
}

impl<R: Runtime> AnySync<R> {
    pub fn new<C: DeserializeOwned>(
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

    /// Helper to execute a gRPC operation with timeout and error handling
    async fn call_grpc<F, T>(&self, op_name: &str, f: F) -> Result<T>
    where
        F: std::future::Future<Output = std::result::Result<tonic::Response<T>, tonic::Status>>,
    {
        match timeout(Duration::from_secs(10), f).await {
            Ok(Ok(response)) => {
                info!("{} successful", op_name);
                Ok(response.into_inner())
            }
            Ok(Err(e)) => {
                error!("{} failed: {}", op_name, e);
                Err(std::io::Error::other(format!("{} failed: {}", op_name, e)).into())
            }
            Err(_) => {
                error!("{} timed out after 10 seconds", op_name);
                Err(std::io::Error::new(
                    std::io::ErrorKind::TimedOut,
                    format!("{} timed out", op_name),
                )
                .into())
            }
        }
    }
}

/// Manages Go backend sidecar process
struct SidecarManager {
    child: Option<tauri_plugin_shell::process::CommandChild>,
    port: Option<u16>,
    client: Option<HealthServiceClient<Channel>>,
    storage_client: Option<StorageServiceClient<Channel>>,
    is_running: bool,
}

impl SidecarManager {
    pub fn new() -> Self {
        Self {
            child: None,
            port: None,
            client: None,
            storage_client: None,
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

        // Set database path to user's data directory (outside src-tauri to avoid file watcher loops)
        let db_path = app
            .path()
            .app_data_dir()
            .map_err(|e| {
                error!("Failed to get app data dir: {}", e);
                std::io::Error::other(format!("Failed to get app data dir: {}", e))
            })?
            .join("anystore.db");

        // Ensure the data directory exists
        if let Some(parent) = db_path.parent() {
            std::fs::create_dir_all(parent)?;
        }

        let db_path_str = db_path
            .to_str()
            .ok_or_else(|| std::io::Error::other("Invalid database path"))?;
        debug!("Database path: {}", db_path_str);

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
            .env("ANY_SYNC_DB_PATH", db_path_str)
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
        // Clone the channel to create both HealthServiceClient and StorageServiceClient.
        // Channels are designed to be cheaply cloneable in tonic, allowing multiple service clients
        // to share the same underlying connection. This ensures efficient resource usage.
        let mut client = HealthServiceClient::new(channel.clone());
        let storage_client = StorageServiceClient::new(channel);
        debug!("gRPC clients created successfully");

        // Test the connection
        debug!("Testing gRPC connection with health check...");
        self.test_connection(&mut client).await?;
        debug!("Health check passed, sidecar is ready");

        self.child = Some(child);
        self.port = Some(port);
        self.client = Some(client);
        self.storage_client = Some(storage_client);
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
        self.storage_client = None;

        info!("Sidecar stopped");
        Ok(())
    }

    /// Ensure sidecar is running and return references to both clients.
    /// This is the main interface for the service implementation.
    pub async fn get_clients<R: Runtime>(
        &mut self,
        app: &AppHandle<R>,
    ) -> Result<(
        &mut HealthServiceClient<Channel>,
        &mut StorageServiceClient<Channel>,
    )> {
        if !self.is_running {
            info!("Sidecar not running, starting it first");
            self.start(app).await?;
        }

        let health_client = self.client.as_mut().ok_or_else(|| {
            std::io::Error::new(
                std::io::ErrorKind::NotConnected,
                "Health client not available",
            )
        })?;

        let storage_client = self.storage_client.as_mut().ok_or_else(|| {
            std::io::Error::new(
                std::io::ErrorKind::NotConnected,
                "Storage client not available",
            )
        })?;

        Ok((health_client, storage_client))
    }
}

#[async_trait]
impl<R: Runtime> AnySyncService for AnySync<R> {
    async fn ping(&self, payload: PingRequest) -> Result<PingResponse> {
        let mut manager = self.manager.write().await;
        manager.get_clients(&self.app).await?;
        let health_client = manager.client.as_mut().ok_or_else(|| {
            std::io::Error::new(std::io::ErrorKind::NotConnected, "No health client")
        })?;

        let timestamp = SystemTime::now()
            .duration_since(SystemTime::UNIX_EPOCH)
            .unwrap()
            .as_secs();

        let request = Request::new(GrpcPingRequest {
            message: payload.value.unwrap_or_default(),
            timestamp: timestamp as i64,
        });
        let response = self.call_grpc("Ping", health_client.ping(request)).await?;
        Ok(PingResponse {
            value: Some(response.message),
        })
    }

    async fn storage_put(&self, payload: PutRequest) -> Result<PutResponse> {
        let mut manager = self.manager.write().await;
        manager.get_clients(&self.app).await?;
        let storage_client = manager.storage_client.as_mut().ok_or_else(|| {
            std::io::Error::new(std::io::ErrorKind::NotConnected, "No storage client")
        })?;

        let request = Request::new(GrpcPutRequest {
            collection: payload.collection,
            id: payload.id,
            document_json: payload.document_json,
        });
        let response = self.call_grpc("Put", storage_client.put(request)).await?;
        Ok(PutResponse {
            success: response.success,
        })
    }

    async fn storage_get(&self, payload: GetRequest) -> Result<GetResponse> {
        let mut manager = self.manager.write().await;
        manager.get_clients(&self.app).await?;
        let storage_client = manager.storage_client.as_mut().ok_or_else(|| {
            std::io::Error::new(std::io::ErrorKind::NotConnected, "No storage client")
        })?;

        let request = Request::new(GrpcGetRequest {
            collection: payload.collection,
            id: payload.id,
        });
        let response = self.call_grpc("Get", storage_client.get(request)).await?;
        Ok(GetResponse {
            document_json: if response.found {
                Some(response.document_json)
            } else {
                None
            },
            found: response.found,
        })
    }

    async fn storage_delete(&self, payload: DeleteRequest) -> Result<DeleteResponse> {
        let mut manager = self.manager.write().await;
        manager.get_clients(&self.app).await?;
        let storage_client = manager.storage_client.as_mut().ok_or_else(|| {
            std::io::Error::new(std::io::ErrorKind::NotConnected, "No storage client")
        })?;

        let request = Request::new(GrpcDeleteRequest {
            collection: payload.collection,
            id: payload.id,
        });
        let response = self
            .call_grpc("Delete", storage_client.delete(request))
            .await?;
        Ok(DeleteResponse {
            existed: response.existed,
        })
    }

    async fn storage_list(&self, payload: ListRequest) -> Result<ListResponse> {
        let mut manager = self.manager.write().await;
        manager.get_clients(&self.app).await?;
        let storage_client = manager.storage_client.as_mut().ok_or_else(|| {
            std::io::Error::new(std::io::ErrorKind::NotConnected, "No storage client")
        })?;

        let request = Request::new(GrpcListRequest {
            collection: payload.collection,
        });
        let response = self.call_grpc("List", storage_client.list(request)).await?;
        Ok(ListResponse { ids: response.ids })
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

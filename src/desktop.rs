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
    DeleteRequest as GrpcDeleteRequest, DeleteResponse as GrpcDeleteResponse,
    GetRequest as GrpcGetRequest, GetResponse as GrpcGetResponse, HealthCheckRequest,
    ListRequest as GrpcListRequest, ListResponse as GrpcListResponse,
    PingRequest as GrpcPingRequest, PingResponse as GrpcPingResponse, PutRequest as GrpcPutRequest,
    PutResponse as GrpcPutResponse,
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

    pub async fn storage_put<R: Runtime>(
        &mut self,
        app: &AppHandle<R>,
        collection: String,
        id: String,
        document_json: String,
    ) -> Result<GrpcPutResponse> {
        if !self.is_running {
            info!("Sidecar not running, starting it first");
            self.start(app).await?;
        }

        if let Some(client) = &mut self.storage_client {
            debug!(
                "Sending put request for collection='{}', id='{}'",
                collection, id
            );

            let request = Request::new(GrpcPutRequest {
                collection,
                id,
                document_json,
            });

            match timeout(Duration::from_secs(10), client.put(request)).await {
                Ok(Ok(response)) => {
                    let response = response.into_inner();
                    info!("Put successful: success={}", response.success);
                    Ok(response)
                }
                Ok(Err(e)) => {
                    error!("Put failed: {}", e);
                    Err(crate::Error::Storage(format!(
                        "Put operation failed: {}",
                        e
                    )))
                }
                Err(_) => {
                    error!("Put timed out after 10 seconds");
                    Err(crate::Error::Storage("Put operation timed out".to_string()))
                }
            }
        } else {
            error!("No storage gRPC client available");
            Err(crate::Error::Storage(
                "Storage client not available".to_string(),
            ))
        }
    }

    pub async fn storage_get<R: Runtime>(
        &mut self,
        app: &AppHandle<R>,
        collection: String,
        id: String,
    ) -> Result<GrpcGetResponse> {
        if !self.is_running {
            info!("Sidecar not running, starting it first");
            self.start(app).await?;
        }

        if let Some(client) = &mut self.storage_client {
            debug!(
                "Sending get request for collection='{}', id='{}'",
                collection, id
            );

            let request = Request::new(GrpcGetRequest { collection, id });

            match timeout(Duration::from_secs(10), client.get(request)).await {
                Ok(Ok(response)) => {
                    let response = response.into_inner();
                    info!("Get successful: found={}", response.found);
                    Ok(response)
                }
                Ok(Err(e)) => {
                    error!("Get failed: {}", e);
                    Err(crate::Error::Storage(format!(
                        "Get operation failed: {}",
                        e
                    )))
                }
                Err(_) => {
                    error!("Get timed out after 10 seconds");
                    Err(crate::Error::Storage("Get operation timed out".to_string()))
                }
            }
        } else {
            error!("No storage gRPC client available");
            Err(crate::Error::Storage(
                "Storage client not available".to_string(),
            ))
        }
    }

    pub async fn storage_delete<R: Runtime>(
        &mut self,
        app: &AppHandle<R>,
        collection: String,
        id: String,
    ) -> Result<GrpcDeleteResponse> {
        if !self.is_running {
            info!("Sidecar not running, starting it first");
            self.start(app).await?;
        }

        if let Some(client) = &mut self.storage_client {
            debug!(
                "Sending delete request: collection='{}', id='{}'",
                collection, id
            );

            let request = Request::new(GrpcDeleteRequest { collection, id });

            match timeout(Duration::from_secs(10), client.delete(request)).await {
                Ok(Ok(response)) => {
                    let response = response.into_inner();
                    info!("Delete successful: existed={}", response.existed);
                    Ok(response)
                }
                Ok(Err(e)) => {
                    error!("Delete failed: {}", e);
                    Err(crate::Error::Storage(format!(
                        "Delete operation failed: {}",
                        e
                    )))
                }
                Err(_) => {
                    error!("Delete timed out after 10 seconds");
                    Err(crate::Error::Storage(
                        "Delete operation timed out".to_string(),
                    ))
                }
            }
        } else {
            error!("No storage gRPC client available");
            Err(crate::Error::Storage(
                "Storage client not available".to_string(),
            ))
        }
    }

    pub async fn storage_list<R: Runtime>(
        &mut self,
        app: &AppHandle<R>,
        collection: String,
    ) -> Result<GrpcListResponse> {
        if !self.is_running {
            info!("Sidecar not running, starting it first");
            self.start(app).await?;
        }

        if let Some(client) = &mut self.storage_client {
            debug!("Sending list request for collection='{}'", collection);

            let request = Request::new(GrpcListRequest { collection });

            match timeout(Duration::from_secs(10), client.list(request)).await {
                Ok(Ok(response)) => {
                    let response = response.into_inner();
                    info!("List successful: {} documents found", response.ids.len());
                    Ok(response)
                }
                Ok(Err(e)) => {
                    error!("List failed: {}", e);
                    Err(crate::Error::Storage(format!(
                        "List operation failed: {}",
                        e
                    )))
                }
                Err(_) => {
                    error!("List timed out after 10 seconds");
                    Err(crate::Error::Storage(
                        "List operation timed out".to_string(),
                    ))
                }
            }
        } else {
            error!("No storage gRPC client available");
            Err(crate::Error::Storage(
                "Storage client not available".to_string(),
            ))
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

    pub async fn storage_put(&self, payload: PutRequest) -> crate::Result<PutResponse> {
        let mut manager = self.manager.write().await;
        let response = manager
            .storage_put(
                &self.app,
                payload.collection,
                payload.id,
                payload.document_json,
            )
            .await?;

        Ok(PutResponse {
            success: response.success,
        })
    }

    pub async fn storage_get(&self, payload: GetRequest) -> crate::Result<GetResponse> {
        let mut manager = self.manager.write().await;
        let response = manager
            .storage_get(&self.app, payload.collection, payload.id)
            .await?;

        Ok(GetResponse {
            document_json: if response.found {
                Some(response.document_json)
            } else {
                None
            },
            found: response.found,
        })
    }

    pub async fn storage_delete(&self, payload: DeleteRequest) -> crate::Result<DeleteResponse> {
        let mut manager = self.manager.write().await;
        let response = manager
            .storage_delete(&self.app, payload.collection, payload.id)
            .await?;

        Ok(DeleteResponse {
            existed: response.existed,
        })
    }

    pub async fn storage_list(&self, payload: ListRequest) -> crate::Result<ListResponse> {
        let mut manager = self.manager.write().await;
        let response = manager.storage_list(&self.app, payload.collection).await?;

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

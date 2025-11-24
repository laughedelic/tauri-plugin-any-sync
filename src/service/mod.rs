use async_trait::async_trait;

use crate::models::*;
use crate::Result;

#[cfg(desktop)]
pub mod desktop;
#[cfg(mobile)]
pub mod mobile;

/// Service trait that abstracts platform-specific implementations.
/// Desktop uses async gRPC calls, Mobile uses sync FFI wrapped in spawn_blocking.
#[async_trait]
pub trait AnySyncService: Send + Sync {
    /// Ping the backend service
    async fn ping(&self, payload: PingRequest) -> Result<PingResponse>;

    /// Store a document in a collection
    async fn storage_put(&self, payload: PutRequest) -> Result<PutResponse>;

    /// Retrieve a document from a collection
    async fn storage_get(&self, payload: GetRequest) -> Result<GetResponse>;

    /// Delete a document from a collection
    async fn storage_delete(&self, payload: DeleteRequest) -> Result<DeleteResponse>;

    /// List all document IDs in a collection
    async fn storage_list(&self, payload: ListRequest) -> Result<ListResponse>;
}

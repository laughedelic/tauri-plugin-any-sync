# Change: Add Minimal AnyStore Integration with Desktop Sidecar

## Why

The plugin currently has scaffolding for gRPC communication and sidecar management, but lacks actual storage functionality. This change adds a minimal integration with AnyStore to validate the end-to-end architecture: Go backend → Rust plugin → TypeScript API → Tauri example app.

AnyStore provides a document-oriented database with MongoDB-style queries backed by SQLite, offering local-first storage that aligns perfectly with the plugin's data sovereignty objectives.

## What Changes

- **Add AnyStore dependency** to Go backend and integrate with gRPC server
- **Implement minimal storage API** with 3 operations:
  - `put(collection, id, document)` - Store a document
  - `get(collection, id)` - Retrieve a document by ID
  - `list(collection)` - List all document IDs in a collection
- **Add protobuf definitions** for storage operations (StorageService)
- **Expose storage commands** in Rust plugin (desktop sidecar only)
- **Add TypeScript bindings** for storage operations
- **Update example app** to demonstrate storage usage with interactive UI
- **Document integration** with usage examples and API reference

This is explicitly **NOT** a full storage API implementation. The goal is to:
1. Validate the desktop sidecar architecture end-to-end
2. Prove AnyStore integration works correctly
3. Provide a foundation for future storage features
4. Keep the scope minimal (< 500 lines of new code across all layers)

## Impact

- **Affected specs**:
  - `storage-api` (NEW) - Core storage operations specification
  - `go-backend-scaffolding` (MODIFIED) - Add AnyStore dependency and implementation
  - `desktop-integration` (MODIFIED) - Add storage command handlers
  - `example-app-update` (MODIFIED) - Add storage demonstration UI

- **Affected code**:
  - `go-backend/api/proto/storage.proto` (NEW) - Protobuf service definition
  - `go-backend/api/server/storage.go` (NEW) - gRPC server implementation
  - `go-backend/internal/storage/anystore.go` (NEW) - AnyStore integration layer
  - `go-backend/go.mod` (MODIFIED) - Add AnyStore dependency
  - `src/commands.rs` (MODIFIED) - Add storage command handlers
  - `guest-js/index.ts` (MODIFIED) - Add TypeScript storage API
  - `examples/tauri-app/src/` (MODIFIED) - Add storage demo component

- **Not affected**:
  - Mobile platforms (storage API will be desktop-only for now)
  - Authentication, sync, or conflict resolution (deferred to Phase 4)
  - Advanced AnyStore features (indexes, complex queries, transactions)

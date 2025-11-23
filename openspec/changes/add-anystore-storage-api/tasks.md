# Implementation Tasks

## 1. Go Backend Storage Integration

- [x] 1.1 Add AnyStore dependency to `go-backend/go.mod`
- [x] 1.2 Create `go-backend/internal/storage/anystore.go` with storage wrapper
- [x] 1.3 Define `storage.proto` with Put, Get, List RPC methods
- [x] 1.4 Implement gRPC StorageService in `go-backend/api/server/storage.go`
- [x] 1.5 Wire storage service into main gRPC server
- [x] 1.6 Add unit tests for storage operations

## 2. Rust Plugin Storage Commands

- [x] 2.1 Update protobuf generation to include storage service
- [x] 2.2 Add storage command handlers in `src/commands.rs` (desktop-only)
- [x] 2.3 Add gRPC client calls in `src/desktop.rs` for storage operations
- [x] 2.4 Add error handling for storage-specific errors
- [x] 2.5 Add integration tests for storage commands

## 3. TypeScript API Bindings

- [x] 3.1 Add storage functions to `guest-js/index.ts` (put, get, list)
- [x] 3.2 Add TypeScript type definitions for storage operations
- [x] 3.3 Export storage API in main package entry point
- [x] 3.4 Add JSDoc documentation for storage functions

## 4. Example Application Updates

- [ ] 4.1 Create storage demo component in `examples/tauri-app/src/lib/Storage.svelte`
- [ ] 4.2 Add UI for putting documents (collection, id, JSON data)
- [ ] 4.3 Add UI for getting documents by ID
- [ ] 4.4 Add UI for listing collection contents
- [ ] 4.5 Integrate storage demo into main App.svelte
- [ ] 4.6 Add example data and usage instructions

## 5. Documentation and Validation

- [ ] 5.1 Update README.md with storage API usage examples
- [ ] 5.2 Document storage API in openspec specs
- [ ] 5.3 Add architecture notes about AnyStore integration
- [ ] 5.4 Run end-to-end validation: example app → plugin → sidecar → AnyStore
- [ ] 5.5 Verify storage persistence across app restarts
- [ ] 5.6 Test error scenarios (invalid JSON, missing documents, etc.)

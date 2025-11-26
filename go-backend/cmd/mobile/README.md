# Mobile Package - gomobile Bindings

This package provides gomobile-compatible bindings for the AnySync storage API, allowing the Go backend to be embedded in mobile applications (Android and iOS).

## API Overview

All functions use simple types (string, bool, error) that are compatible with gomobile. Complex data is serialized as JSON strings.

### InitStorage

```go
func InitStorage(dbPath string) error
```

Initializes the storage with the given database path. Must be called before any other storage operations.

**Parameters:**
- `dbPath`: Absolute path to the SQLite database file

**Returns:**
- `error`: Error if initialization fails, nil on success

### StoragePut

```go
func StoragePut(collection, id, documentJson string) error
```

Stores a document in the specified collection.

**Parameters:**
- `collection`: Collection name
- `id`: Document ID
- `documentJson`: Document as JSON string

**Returns:**
- `error`: Error if operation fails, nil on success

### StorageGet

```go
func StorageGet(collection, id string) (string, error)
```

Retrieves a document from the specified collection.

**Parameters:**
- `collection`: Collection name
- `id`: Document ID

**Returns:**
- `string`: Document as JSON string
- `error`: Error if not found or operation fails

### StorageDelete

```go
func StorageDelete(collection, id string) (bool, error)
```

Deletes a document from the specified collection.

**Parameters:**
- `collection`: Collection name
- `id`: Document ID

**Returns:**
- `bool`: True if document was deleted, false if it didn't exist
- `error`: Error if operation fails

### StorageList

```go
func StorageList(collection string) (string, error)
```

Lists all document IDs in the specified collection.

**Parameters:**
- `collection`: Collection name

**Returns:**
- `string`: JSON array of document IDs, e.g., `["id1", "id2", "id3"]`
- `error`: Error if operation fails

## Building

### Android

```bash
gomobile bind -target=android -o any-sync-android.aar ./cmd/mobile
```

This generates an `.aar` file containing:
- `libgojni.so` for all Android ABIs (arm64-v8a, armeabi-v7a, x86, x86_64)
- Generated Java classes in the `mobile` package

### iOS (Future)

```bash
gomobile bind -target=ios -o AnySync.xcframework ./cmd/mobile
```

## Usage

### Android (Kotlin)

```kotlin
import mobile.Mobile

// Initialize storage
Mobile.initStorage("/data/data/com.example.app/databases/anystore.db")

// Store a document
Mobile.storagePut("users", "user1", """{"name": "Alice", "age": 30}""")

// Retrieve a document
val json = Mobile.storageGet("users", "user1")

// List documents
val ids = Mobile.storageList("users")

// Delete a document
val deleted = Mobile.storageDelete("users", "user1")
```

## Architecture

This package is part of the mobile integration layer:

```
TypeScript API → Rust Plugin → Kotlin/Swift Bridge → gomobile JNI → Go Mobile Package → Storage Layer
```

The same storage implementation (`internal/storage`) is used by both:
- Desktop: via gRPC sidecar (`cmd/server`)
- Mobile: via direct function calls (`cmd/mobile`)

This ensures >95% code reuse and consistent behavior across platforms.

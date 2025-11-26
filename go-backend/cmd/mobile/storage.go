// Package mobile provides gomobile-compatible bindings for the AnySync storage API.
// This package exports functions that can be called from Android via JNI.
//
// All functions use simple types (string, bool, error) that are compatible with gomobile.
// Complex data is serialized as JSON strings.
package mobile

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"anysync-backend/internal/storage"
)

var (
	// Global storage instance
	store   *storage.Store
	storeMu sync.RWMutex
)

// InitStorage initializes the storage with the given database path.
// Must be called before any other storage operations.
// Returns an error if initialization fails.
func InitStorage(dbPath string) error {
	storeMu.Lock()
	defer storeMu.Unlock()

	var err error
	store, err = storage.New(dbPath)
	if err != nil {
		return fmt.Errorf("failed to initialize storage: %w", err)
	}

	return nil
}

// StoragePut stores a document in the specified collection.
// documentJson must be a valid JSON string.
// Returns an error if the operation fails.
func StoragePut(collection, id, documentJson string) error {
	storeMu.RLock()
	defer storeMu.RUnlock()

	if store == nil {
		return fmt.Errorf("storage not initialized")
	}

	ctx := context.Background()
	return store.Put(ctx, collection, id, documentJson)
}

// StorageGet retrieves a document from the specified collection.
// Returns the document as a JSON string, or an error if not found.
func StorageGet(collection, id string) (string, error) {
	storeMu.RLock()
	defer storeMu.RUnlock()

	if store == nil {
		return "", fmt.Errorf("storage not initialized")
	}

	ctx := context.Background()
	return store.Get(ctx, collection, id)
}

// StorageDelete deletes a document from the specified collection.
// Returns true if the document was deleted, false if it didn't exist.
// Returns an error if the operation fails.
func StorageDelete(collection, id string) (bool, error) {
	storeMu.RLock()
	defer storeMu.RUnlock()

	if store == nil {
		return false, fmt.Errorf("storage not initialized")
	}

	ctx := context.Background()
	return store.Delete(ctx, collection, id)
}

// StorageList lists all document IDs in the specified collection.
// Returns a JSON array of strings, e.g., ["id1", "id2", "id3"].
// Returns an error if the operation fails.
func StorageList(collection string) (string, error) {
	storeMu.RLock()
	defer storeMu.RUnlock()

	if store == nil {
		return "", fmt.Errorf("storage not initialized")
	}

	ctx := context.Background()
	ids, err := store.List(ctx, collection)
	if err != nil {
		return "", err
	}

	jsonBytes, err := json.Marshal(ids)
	if err != nil {
		return "", fmt.Errorf("failed to marshal IDs: %w", err)
	}

	return string(jsonBytes), nil
}

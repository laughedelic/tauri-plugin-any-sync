package storage

import (
	"context"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
)

// Helper function to create a test store
func createTestStore(t *testing.T) (*Store, func()) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	store, err := New(dbPath)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	cleanup := func() {
		if err := store.Close(); err != nil {
			t.Errorf("Failed to close store: %v", err)
		}
	}
	return store, cleanup
}

func TestNew(t *testing.T) {
	t.Run("CreateStore", func(t *testing.T) {
		tmpDir := t.TempDir()
		dbPath := filepath.Join(tmpDir, "test.db")
		store, err := New(dbPath)
		if err != nil {
			t.Fatalf("Failed to create store: %v", err)
		}
		defer store.Close()

		if store == nil {
			t.Error("Expected store to be non-nil")
		}
		if store.db == nil {
			t.Error("Expected store.db to be non-nil")
		}
	})

	t.Run("CreateStoreInvalidPath", func(t *testing.T) {
		// Try to create store with invalid path
		_, err := New("/invalid/path/that/does/not/exist/test.db")
		if err == nil {
			t.Error("Expected error when creating store with invalid path")
		}
	})
}

func TestPut(t *testing.T) {
	t.Run("PutSimpleDocument", func(t *testing.T) {
		store, cleanup := createTestStore(t)
		defer cleanup()

		ctx := context.Background()
		err := store.Put(ctx, "users", "user1", `{"name":"Alice","age":30}`)
		if err != nil {
			t.Fatalf("Put failed: %v", err)
		}
	})

	t.Run("PutEmptyDocument", func(t *testing.T) {
		store, cleanup := createTestStore(t)
		defer cleanup()

		ctx := context.Background()
		err := store.Put(ctx, "users", "empty", "{}")
		if err != nil {
			t.Fatalf("Put empty document failed: %v", err)
		}

		// Verify it was stored with id field
		result, err := store.Get(ctx, "users", "empty")
		if err != nil {
			t.Fatalf("Get failed: %v", err)
		}
		if result == "" {
			t.Error("Expected document, got empty string")
		}
	})

	t.Run("PutComplexDocument", func(t *testing.T) {
		store, cleanup := createTestStore(t)
		defer cleanup()

		ctx := context.Background()
		complexDoc := `{"name":"Bob","email":"bob@example.com","address":{"city":"NYC","zip":"10001"},"tags":["developer","golang"]}`
		err := store.Put(ctx, "users", "user2", complexDoc)
		if err != nil {
			t.Fatalf("Put complex document failed: %v", err)
		}
	})

	t.Run("PutInvalidJSON", func(t *testing.T) {
		store, cleanup := createTestStore(t)
		defer cleanup()

		ctx := context.Background()
		err := store.Put(ctx, "users", "invalid", "not valid json")
		if err == nil {
			t.Error("Expected error when putting invalid JSON")
		}
		if !strings.Contains(err.Error(), "invalid JSON") {
			t.Errorf("Expected 'invalid JSON' error, got: %v", err)
		}
	})

	t.Run("PutUpdate", func(t *testing.T) {
		store, cleanup := createTestStore(t)
		defer cleanup()

		ctx := context.Background()
		// Put initial document
		err := store.Put(ctx, "users", "user3", `{"name":"Charlie","version":1}`)
		if err != nil {
			t.Fatalf("Initial Put failed: %v", err)
		}

		// Update the document
		err = store.Put(ctx, "users", "user3", `{"name":"Charlie Updated","version":2}`)
		if err != nil {
			t.Fatalf("Update Put failed: %v", err)
		}

		// Verify update
		result, err := store.Get(ctx, "users", "user3")
		if err != nil {
			t.Fatalf("Get failed: %v", err)
		}
		if !strings.Contains(result, "Charlie Updated") {
			t.Error("Document was not updated")
		}
	})
}

func TestGet(t *testing.T) {
	t.Run("GetExistingDocument", func(t *testing.T) {
		store, cleanup := createTestStore(t)
		defer cleanup()

		ctx := context.Background()
		testDoc := `{"name":"Alice","email":"alice@example.com"}`
		err := store.Put(ctx, "users", "user1", testDoc)
		if err != nil {
			t.Fatalf("Put failed: %v", err)
		}

		result, err := store.Get(ctx, "users", "user1")
		if err != nil {
			t.Fatalf("Get failed: %v", err)
		}
		if result == "" {
			t.Error("Expected document, got empty string")
		}

		// Verify it's valid JSON
		var parsed map[string]interface{}
		if err := json.Unmarshal([]byte(result), &parsed); err != nil {
			t.Fatalf("Result is not valid JSON: %v", err)
		}

		// Verify id field is present
		if id, ok := parsed["id"].(string); !ok || id != "user1" {
			t.Errorf("Expected id='user1', got: %v", parsed["id"])
		}

		// Verify original fields exist
		if name, ok := parsed["name"].(string); !ok || name != "Alice" {
			t.Errorf("Expected name='Alice', got: %v", parsed["name"])
		}
	})

	t.Run("GetNonExistentDocument", func(t *testing.T) {
		store, cleanup := createTestStore(t)
		defer cleanup()

		ctx := context.Background()
		result, err := store.Get(ctx, "users", "nonexistent")
		if err != nil {
			t.Fatalf("Get should not error for non-existent document: %v", err)
		}
		if result != "" {
			t.Errorf("Expected empty string for non-existent document, got: %s", result)
		}
	})

	t.Run("GetFromNonExistentCollection", func(t *testing.T) {
		store, cleanup := createTestStore(t)
		defer cleanup()

		ctx := context.Background()
		result, err := store.Get(ctx, "nonexistent_collection", "doc1")
		if err != nil {
			t.Fatalf("Get from non-existent collection failed: %v", err)
		}
		if result != "" {
			t.Errorf("Expected empty string from non-existent collection, got: %s", result)
		}
	})
}

func TestList(t *testing.T) {
	t.Run("ListEmptyCollection", func(t *testing.T) {
		store, cleanup := createTestStore(t)
		defer cleanup()

		ctx := context.Background()
		ids, err := store.List(ctx, "empty_collection")
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}
		if len(ids) != 0 {
			t.Errorf("Expected 0 docs in empty collection, got %d", len(ids))
		}
	})

	t.Run("ListSingleDocument", func(t *testing.T) {
		store, cleanup := createTestStore(t)
		defer cleanup()

		ctx := context.Background()
		err := store.Put(ctx, "items", "item1", `{"x":1}`)
		if err != nil {
			t.Fatalf("Put failed: %v", err)
		}

		ids, err := store.List(ctx, "items")
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}
		if len(ids) != 1 {
			t.Errorf("Expected 1 doc, got %d", len(ids))
		}
		if ids[0] != "item1" {
			t.Errorf("Expected id 'item1', got '%s'", ids[0])
		}
	})

	t.Run("ListMultipleDocuments", func(t *testing.T) {
		store, cleanup := createTestStore(t)
		defer cleanup()

		ctx := context.Background()
		// Put multiple documents
		testDocs := map[string]string{
			"doc1": `{"type":"a"}`,
			"doc2": `{"type":"b"}`,
			"doc3": `{"type":"c"}`,
		}

		for id, doc := range testDocs {
			err := store.Put(ctx, "items", id, doc)
			if err != nil {
				t.Fatalf("Put failed for %s: %v", id, err)
			}
		}

		ids, err := store.List(ctx, "items")
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}
		if len(ids) != len(testDocs) {
			t.Errorf("Expected %d docs, got %d", len(testDocs), len(ids))
		}

		// Verify all IDs are present
		idMap := make(map[string]bool)
		for _, id := range ids {
			idMap[id] = true
		}
		for expectedId := range testDocs {
			if !idMap[expectedId] {
				t.Errorf("Expected ID '%s' not found in results", expectedId)
			}
		}
	})

	t.Run("ListMultipleCollections", func(t *testing.T) {
		store, cleanup := createTestStore(t)
		defer cleanup()

		ctx := context.Background()
		// Put documents in different collections
		store.Put(ctx, "users", "u1", `{"name":"Alice"}`)
		store.Put(ctx, "users", "u2", `{"name":"Bob"}`)
		store.Put(ctx, "items", "i1", `{"x":1}`)

		// List users collection
		userIds, err := store.List(ctx, "users")
		if err != nil {
			t.Fatalf("List users failed: %v", err)
		}
		if len(userIds) != 2 {
			t.Errorf("Expected 2 users, got %d", len(userIds))
		}

		// List items collection
		itemIds, err := store.List(ctx, "items")
		if err != nil {
			t.Fatalf("List items failed: %v", err)
		}
		if len(itemIds) != 1 {
			t.Errorf("Expected 1 item, got %d", len(itemIds))
		}
	})
}

func TestClose(t *testing.T) {
	t.Run("CloseStore", func(t *testing.T) {
		tmpDir := t.TempDir()
		dbPath := filepath.Join(tmpDir, "test.db")
		store, err := New(dbPath)
		if err != nil {
			t.Fatalf("Failed to create store: %v", err)
		}

		err = store.Close()
		if err != nil {
			t.Errorf("Close failed: %v", err)
		}
	})
}

func TestIntegration(t *testing.T) {
	t.Run("CompleteWorkflow", func(t *testing.T) {
		store, cleanup := createTestStore(t)
		defer cleanup()

		ctx := context.Background()

		// 1. Verify empty collection
		ids, _ := store.List(ctx, "tasks")
		if len(ids) != 0 {
			t.Errorf("Expected empty collection initially, got %d items", len(ids))
		}

		// 2. Add multiple documents
		tasks := []struct {
			id   string
			data string
		}{
			{"task1", `{"title":"Write tests","priority":"high"}`},
			{"task2", `{"title":"Review code","priority":"medium"}`},
			{"task3", `{"title":"Deploy app","priority":"low"}`},
		}

		for _, task := range tasks {
			err := store.Put(ctx, "tasks", task.id, task.data)
			if err != nil {
				t.Fatalf("Failed to put task %s: %v", task.id, err)
			}
		}

		// 3. Verify all documents exist
		ids, err := store.List(ctx, "tasks")
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}
		if len(ids) != len(tasks) {
			t.Errorf("Expected %d tasks, got %d", len(tasks), len(ids))
		}

		// 4. Retrieve and verify each document
		for _, task := range tasks {
			result, err := store.Get(ctx, "tasks", task.id)
			if err != nil {
				t.Fatalf("Failed to get task %s: %v", task.id, err)
			}
			if result == "" {
				t.Errorf("Expected document for task %s, got empty", task.id)
			}

			// Verify it's valid JSON with id field
			var parsed map[string]interface{}
			if err := json.Unmarshal([]byte(result), &parsed); err != nil {
				t.Fatalf("Result is not valid JSON for task %s: %v", task.id, err)
			}
			if id, ok := parsed["id"].(string); !ok || id != task.id {
				t.Errorf("Expected id='%s', got: %v", task.id, parsed["id"])
			}
		}

		// 5. Update a document
		err = store.Put(ctx, "tasks", "task1", `{"title":"Write tests - DONE","priority":"high","completed":true}`)
		if err != nil {
			t.Fatalf("Failed to update task1: %v", err)
		}

		updated, err := store.Get(ctx, "tasks", "task1")
		if err != nil {
			t.Fatalf("Failed to get updated task1: %v", err)
		}
		if !strings.Contains(updated, "DONE") || !strings.Contains(updated, "completed") {
			t.Error("Task1 was not properly updated")
		}

		// 6. Verify count hasn't changed after update
		ids, err = store.List(ctx, "tasks")
		if err != nil {
			t.Fatalf("List failed after update: %v", err)
		}
		if len(ids) != len(tasks) {
			t.Errorf("Expected %d tasks after update, got %d", len(tasks), len(ids))
		}
	})
}

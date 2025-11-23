package storage

import (
	"context"
	"fmt"

	anystore "github.com/anyproto/any-store"
	"github.com/anyproto/any-store/anyenc"
)

// Store wraps AnyStore database operations
type Store struct {
	db anystore.DB
}

// New creates a new Store instance with the given database path
func New(dbPath string) (*Store, error) {
	db, err := anystore.Open(context.Background(), dbPath, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	return &Store{db: db}, nil
}

// Close closes the database connection
func (s *Store) Close() error {
	return s.db.Close()
}

// Put stores a document in the specified collection
// The document must be valid JSON
func (s *Store) Put(ctx context.Context, collection, id, documentJSON string) error {
	// Create arena for new values
	a := &anyenc.Arena{}

	// Parse the input JSON document
	doc, err := anyenc.ParseJson(documentJSON)
	if err != nil {
		return fmt.Errorf("invalid JSON document: %w", err)
	}

	// Set the id field (AnyStore requires 'id' field in the document)
	doc.Set("id", a.NewString(id))

	// Get or create collection
	coll, err := s.db.Collection(ctx, collection)
	if err != nil {
		return fmt.Errorf("failed to get collection %q: %w", collection, err)
	}

	// UpsertOne will insert if not exists, or update if exists
	err = coll.UpsertOne(ctx, doc)
	if err != nil {
		return fmt.Errorf("failed to upsert document in collection %q with id %q: %w", collection, id, err)
	}

	return nil
}

// Get retrieves a document from the specified collection by ID
// Returns empty string if document not found
func (s *Store) Get(ctx context.Context, collection, id string) (string, error) {
	coll, err := s.db.Collection(ctx, collection)
	if err != nil {
		return "", fmt.Errorf("failed to get collection %q: %w", collection, err)
	}

	// Find document by ID
	doc, err := coll.FindId(ctx, id)
	if err != nil {
		// Document not found
		return "", nil
	}

	// Convert to JSON string using String() method
	return doc.Value().String(), nil
}

// Delete removes a document from the specified collection by ID
// Returns true if the document existed and was deleted, false if it didn't exist
// This operation is idempotent - deleting a non-existent document returns (false, nil)
func (s *Store) Delete(ctx context.Context, collection, id string) (bool, error) {
	coll, err := s.db.Collection(ctx, collection)
	if err != nil {
		return false, fmt.Errorf("failed to get collection %q: %w", collection, err)
	}

	// Check if document exists before deleting
	_, err = coll.FindId(ctx, id)
	if err != nil {
		// Document not found - this is not an error, just return false
		return false, nil
	}

	// Delete the document by ID
	err = coll.DeleteId(ctx, id)
	if err != nil {
		return false, fmt.Errorf("failed to delete document in collection %q with id %q: %w", collection, id, err)
	}

	return true, nil
}

// List returns all document IDs in the specified collection
func (s *Store) List(ctx context.Context, collection string) ([]string, error) {
	coll, err := s.db.Collection(ctx, collection)
	if err != nil {
		return nil, fmt.Errorf("failed to get collection %q: %w", collection, err)
	}

	// Find all documents (nil filter means all)
	query := coll.Find(nil)
	iter, err := query.Iter(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list documents in collection %q: %w", collection, err)
	}
	defer iter.Close()

	var ids []string
	for iter.Next() {
		doc, err := iter.Doc()
		if err != nil {
			iter.Close()
			return nil, fmt.Errorf("error retrieving document in collection %q: %w", collection, err)
		}
		// Get the id field from the document
		idStr := doc.Value().GetString("id")
		if idStr != "" {
			ids = append(ids, idStr)
		}
	}

	if err := iter.Err(); err != nil {
		return nil, fmt.Errorf("error iterating documents in collection %q: %w", collection, err)
	}

	return ids, nil
}

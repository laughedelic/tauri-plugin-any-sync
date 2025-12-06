package anysync

import (
	"testing"
	"time"

	"github.com/anyproto/any-sync/commonspace/object/accountdata"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDocumentManager(t *testing.T) {
	tempDir := t.TempDir()
	keys, err := accountdata.NewRandom()
	require.NoError(t, err)

	sm, err := NewSpaceManager(tempDir, keys, NewEventManager())
	require.NoError(t, err)
	defer sm.Close()

	dm, err := NewDocumentManager(sm, keys, NewEventManager())
	require.NoError(t, err)
	assert.NotNil(t, dm)
}

func TestNewDocumentManager_NilSpaceManager(t *testing.T) {
	keys, err := accountdata.NewRandom()
	require.NoError(t, err)

	dm, err := NewDocumentManager(nil, keys, NewEventManager())
	assert.Error(t, err)
	assert.Nil(t, dm)
	assert.Contains(t, err.Error(), "space manager required")
}

func TestNewDocumentManager_NilKeys(t *testing.T) {
	tempDir := t.TempDir()
	keys, err := accountdata.NewRandom()
	require.NoError(t, err)

	sm, err := NewSpaceManager(tempDir, keys, NewEventManager())
	require.NoError(t, err)
	defer sm.Close()

	dm, err := NewDocumentManager(sm, nil, NewEventManager())
	assert.Error(t, err)
	assert.Nil(t, dm)
	assert.Contains(t, err.Error(), "account keys required")
}

func TestCreateDocument_Success(t *testing.T) {
	tempDir := t.TempDir()
	keys, err := accountdata.NewRandom()
	require.NoError(t, err)

	sm, err := NewSpaceManager(tempDir, keys, NewEventManager())
	require.NoError(t, err)
	defer sm.Close()

	// Create a space first
	err = sm.CreateSpace("test-space", "Test Space", nil)
	require.NoError(t, err)

	spaces := sm.ListSpaces()
	require.Len(t, spaces, 1)
	spaceID := spaces[0].SpaceID

	// Create document manager
	dm, err := NewDocumentManager(sm, keys, NewEventManager())
	require.NoError(t, err)

	// Create a document
	docData := []byte("Hello, World!")
	docID, err := dm.CreateDocument(spaceID, "Test Document", docData, map[string]string{
		"author": "test",
	})
	require.NoError(t, err)
	assert.NotEmpty(t, docID)
}

func TestCreateDocument_InvalidSpace(t *testing.T) {
	tempDir := t.TempDir()
	keys, err := accountdata.NewRandom()
	require.NoError(t, err)

	sm, err := NewSpaceManager(tempDir, keys, NewEventManager())
	require.NoError(t, err)
	defer sm.Close()

	dm, err := NewDocumentManager(sm, keys, NewEventManager())
	require.NoError(t, err)

	// Try to create document in non-existent space
	docData := []byte("Hello, World!")
	docID, err := dm.CreateDocument("invalid-space-id", "Test Document", docData, nil)
	assert.Error(t, err)
	assert.Empty(t, docID)
}

func TestGetDocument_Success(t *testing.T) {
	tempDir := t.TempDir()
	keys, err := accountdata.NewRandom()
	require.NoError(t, err)

	sm, err := NewSpaceManager(tempDir, keys, NewEventManager())
	require.NoError(t, err)
	defer sm.Close()

	// Create a space
	err = sm.CreateSpace("test-space", "Test Space", nil)
	require.NoError(t, err)

	spaces := sm.ListSpaces()
	spaceID := spaces[0].SpaceID

	dm, err := NewDocumentManager(sm, keys, NewEventManager())
	require.NoError(t, err)

	// Create a document
	docData := []byte("Hello, World!")
	docID, err := dm.CreateDocument(spaceID, "Test Document", docData, map[string]string{
		"author": "test",
	})
	require.NoError(t, err)

	// Retrieve the document
	retrievedData, meta, err := dm.GetDocument(spaceID, docID)
	require.NoError(t, err)
	assert.Equal(t, docData, retrievedData)
	assert.Equal(t, "Test Document", meta.Title)
	assert.Equal(t, "test", meta.Metadata["author"])
	assert.Equal(t, docID, meta.DocumentID)
	assert.Equal(t, spaceID, meta.SpaceID)
}

func TestGetDocument_NotFound(t *testing.T) {
	tempDir := t.TempDir()
	keys, err := accountdata.NewRandom()
	require.NoError(t, err)

	sm, err := NewSpaceManager(tempDir, keys, NewEventManager())
	require.NoError(t, err)
	defer sm.Close()

	// Create a space
	err = sm.CreateSpace("test-space", "Test Space", nil)
	require.NoError(t, err)

	spaces := sm.ListSpaces()
	spaceID := spaces[0].SpaceID

	dm, err := NewDocumentManager(sm, keys, NewEventManager())
	require.NoError(t, err)

	// Try to get non-existent document
	data, meta, err := dm.GetDocument(spaceID, "non-existent-doc-id")
	assert.Error(t, err)
	assert.Nil(t, data)
	assert.Nil(t, meta)
}

func TestUpdateDocument_Success(t *testing.T) {
	tempDir := t.TempDir()
	keys, err := accountdata.NewRandom()
	require.NoError(t, err)

	sm, err := NewSpaceManager(tempDir, keys, NewEventManager())
	require.NoError(t, err)
	defer sm.Close()

	// Create a space
	err = sm.CreateSpace("test-space", "Test Space", nil)
	require.NoError(t, err)

	spaces := sm.ListSpaces()
	spaceID := spaces[0].SpaceID

	dm, err := NewDocumentManager(sm, keys, NewEventManager())
	require.NoError(t, err)

	// Create a document
	docData := []byte("Version 1")
	docID, err := dm.CreateDocument(spaceID, "Test Document", docData, nil)
	require.NoError(t, err)

	// Add a small delay to ensure timestamp difference
	time.Sleep(1 * time.Second)

	// Update the document with new data and metadata
	newData := []byte("Version 2")
	updateMetadata := map[string]string{
		"title":   "Updated Document",
		"updated": "2023-01-01T00:00:00Z",
		"custom":  "value",
	}
	err = dm.UpdateDocument(spaceID, docID, newData, updateMetadata)
	require.NoError(t, err)

	// Retrieve and verify the updated document
	retrievedData, meta, err := dm.GetDocument(spaceID, docID)
	require.NoError(t, err)
	assert.Equal(t, newData, retrievedData)
	assert.Greater(t, meta.UpdatedAt, meta.CreatedAt)

	// Verify metadata was updated
	assert.Equal(t, "Updated Document", meta.Title, "Title should be updated")
	assert.Equal(t, "Updated Document", meta.Metadata["title"], "Metadata title should be updated")
	assert.Equal(t, "2023-01-01T00:00:00Z", meta.Metadata["updated"], "Custom metadata should be preserved")
	assert.Equal(t, "value", meta.Metadata["custom"], "Custom metadata should be preserved")
}

func TestUpdateDocument_NotFound(t *testing.T) {
	tempDir := t.TempDir()
	keys, err := accountdata.NewRandom()
	require.NoError(t, err)

	sm, err := NewSpaceManager(tempDir, keys, NewEventManager())
	require.NoError(t, err)
	defer sm.Close()

	// Create a space
	err = sm.CreateSpace("test-space", "Test Space", nil)
	require.NoError(t, err)

	spaces := sm.ListSpaces()
	spaceID := spaces[0].SpaceID

	dm, err := NewDocumentManager(sm, keys, NewEventManager())
	require.NoError(t, err)

	// Try to update non-existent document
	newData := []byte("Version 2")
	err = dm.UpdateDocument(spaceID, "non-existent-doc-id", newData, nil)
	assert.Error(t, err)
}

func TestDeleteDocument_Success(t *testing.T) {
	tempDir := t.TempDir()
	keys, err := accountdata.NewRandom()
	require.NoError(t, err)

	sm, err := NewSpaceManager(tempDir, keys, NewEventManager())
	require.NoError(t, err)
	defer sm.Close()

	// Create a space
	err = sm.CreateSpace("test-space", "Test Space", nil)
	require.NoError(t, err)

	spaces := sm.ListSpaces()
	spaceID := spaces[0].SpaceID

	dm, err := NewDocumentManager(sm, keys, NewEventManager())
	require.NoError(t, err)

	// Create a document
	docData := []byte("Hello, World!")
	docID, err := dm.CreateDocument(spaceID, "Test Document", docData, nil)
	require.NoError(t, err)

	// Delete the document
	err = dm.DeleteDocument(spaceID, docID)
	require.NoError(t, err)

	// Verify document is gone
	data, meta, err := dm.GetDocument(spaceID, docID)
	assert.Error(t, err)
	assert.Nil(t, data)
	assert.Nil(t, meta)
}

func TestDeleteDocument_NotFound(t *testing.T) {
	tempDir := t.TempDir()
	keys, err := accountdata.NewRandom()
	require.NoError(t, err)

	sm, err := NewSpaceManager(tempDir, keys, NewEventManager())
	require.NoError(t, err)
	defer sm.Close()

	// Create a space
	err = sm.CreateSpace("test-space", "Test Space", nil)
	require.NoError(t, err)

	spaces := sm.ListSpaces()
	spaceID := spaces[0].SpaceID

	dm, err := NewDocumentManager(sm, keys, NewEventManager())
	require.NoError(t, err)

	// Try to delete non-existent document
	err = dm.DeleteDocument(spaceID, "non-existent-doc-id")
	assert.Error(t, err)
}

func TestListDocuments_Empty(t *testing.T) {
	tempDir := t.TempDir()
	keys, err := accountdata.NewRandom()
	require.NoError(t, err)

	sm, err := NewSpaceManager(tempDir, keys, NewEventManager())
	require.NoError(t, err)
	defer sm.Close()

	// Create a space
	err = sm.CreateSpace("test-space", "Test Space", nil)
	require.NoError(t, err)

	spaces := sm.ListSpaces()
	spaceID := spaces[0].SpaceID

	dm, err := NewDocumentManager(sm, keys, NewEventManager())
	require.NoError(t, err)

	// List documents in empty space
	docs, err := dm.ListDocuments(spaceID)
	require.NoError(t, err)
	assert.Len(t, docs, 0)
}

func TestListDocuments_Multiple(t *testing.T) {
	tempDir := t.TempDir()
	keys, err := accountdata.NewRandom()
	require.NoError(t, err)

	sm, err := NewSpaceManager(tempDir, keys, NewEventManager())
	require.NoError(t, err)
	defer sm.Close()

	// Create a space
	err = sm.CreateSpace("test-space", "Test Space", nil)
	require.NoError(t, err)

	spaces := sm.ListSpaces()
	spaceID := spaces[0].SpaceID

	dm, err := NewDocumentManager(sm, keys, NewEventManager())
	require.NoError(t, err)

	// Create multiple documents
	docIDs := make([]string, 3)
	for i := 0; i < 3; i++ {
		docData := []byte("Document " + string(rune('A'+i)))
		docID, err := dm.CreateDocument(spaceID, "Doc "+string(rune('A'+i)), docData, nil)
		require.NoError(t, err)
		docIDs[i] = docID
	}

	// List all documents
	docs, err := dm.ListDocuments(spaceID)
	require.NoError(t, err)
	assert.Len(t, docs, 3)

	// Verify all document IDs are present
	foundIDs := make(map[string]bool)
	for _, doc := range docs {
		foundIDs[doc.DocumentID] = true
	}
	for _, id := range docIDs {
		assert.True(t, foundIDs[id], "Document ID %s should be in list", id)
	}
}

func TestQueryDocuments_NoFilter(t *testing.T) {
	tempDir := t.TempDir()
	keys, err := accountdata.NewRandom()
	require.NoError(t, err)

	sm, err := NewSpaceManager(tempDir, keys, NewEventManager())
	require.NoError(t, err)
	defer sm.Close()

	// Create a space
	err = sm.CreateSpace("test-space", "Test Space", nil)
	require.NoError(t, err)

	spaces := sm.ListSpaces()
	spaceID := spaces[0].SpaceID

	dm, err := NewDocumentManager(sm, keys, NewEventManager())
	require.NoError(t, err)

	// Create documents
	_, err = dm.CreateDocument(spaceID, "Doc A", []byte("A"), nil)
	require.NoError(t, err)
	_, err = dm.CreateDocument(spaceID, "Doc B", []byte("B"), nil)
	require.NoError(t, err)

	// Query with no filter
	docs, err := dm.QueryDocuments(spaceID, nil)
	require.NoError(t, err)
	assert.Len(t, docs, 2)
}

func TestMultipleSpaces(t *testing.T) {
	tempDir := t.TempDir()
	keys, err := accountdata.NewRandom()
	require.NoError(t, err)

	sm, err := NewSpaceManager(tempDir, keys, NewEventManager())
	require.NoError(t, err)
	defer sm.Close()

	// Create two spaces
	err = sm.CreateSpace("space-1", "Space 1", nil)
	require.NoError(t, err)
	err = sm.CreateSpace("space-2", "Space 2", nil)
	require.NoError(t, err)

	spaces := sm.ListSpaces()
	require.Len(t, spaces, 2)

	dm, err := NewDocumentManager(sm, keys, NewEventManager())
	require.NoError(t, err)

	// Create documents in different spaces
	_, err = dm.CreateDocument(spaces[0].SpaceID, "Doc in Space 1", []byte("Data 1"), nil)
	require.NoError(t, err)
	_, err = dm.CreateDocument(spaces[1].SpaceID, "Doc in Space 2", []byte("Data 2"), nil)
	require.NoError(t, err)

	// Verify documents are isolated by space
	docs1, err := dm.ListDocuments(spaces[0].SpaceID)
	require.NoError(t, err)
	assert.Len(t, docs1, 1)
	assert.Equal(t, "Doc in Space 1", docs1[0].Title)

	docs2, err := dm.ListDocuments(spaces[1].SpaceID)
	require.NoError(t, err)
	assert.Len(t, docs2, 1)
	assert.Equal(t, "Doc in Space 2", docs2[0].Title)
}

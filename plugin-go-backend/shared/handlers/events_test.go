package handlers

import (
	"context"
	"testing"
	"time"

	"anysync-backend/shared/anysync"
	pb "anysync-backend/shared/proto/syncspace/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupForEventTests initializes the system for event streaming tests
func setupForEventTests(t *testing.T) context.Context {
	t.Helper()

	// Reset global state
	globalState = &State{}

	// Initialize
	tempDir := t.TempDir()
	initReq := &pb.InitRequest{
		DataDir:   tempDir,
		NetworkId: "test-network",
		DeviceId:  "test-device",
	}

	_, err := Init(context.Background(), initReq)
	require.NoError(t, err)

	return context.Background()
}

// teardownForEventTests cleans up after event tests
func teardownForEventTests(t *testing.T) {
	t.Helper()
	shutdownReq := &pb.ShutdownRequest{}
	_, _ = Shutdown(context.Background(), shutdownReq)
}

// TestSubscribe_DocumentCreatedEvent tests that document creation triggers events
func TestSubscribe_DocumentCreatedEvent(t *testing.T) {
	ctx := setupForEventTests(t)
	defer teardownForEventTests(t)

	// Subscribe to document events
	subscribeReq := &pb.SubscribeRequest{
		EventTypes: []string{"document.created"},
		SpaceIds:   []string{},
	}

	subscriberID, eventChan, err := Subscribe(ctx, subscribeReq)
	require.NoError(t, err)
	require.NotEmpty(t, subscriberID)
	defer Unsubscribe(subscriberID)

	// Create a space first
	createSpaceReq := &pb.CreateSpaceRequest{
		Name: "Test Space",
	}
	createSpaceResp, err := CreateSpace(context.Background(), createSpaceReq)
	require.NoError(t, err)
	spaceResp := createSpaceResp.(*pb.CreateSpaceResponse)
	spaceID := spaceResp.SpaceId

	// Wait a moment for space creation to complete
	time.Sleep(100 * time.Millisecond)

	// Create a document using DocumentManager directly
	docData := []byte("test document content")
	globalState.mu.RLock()
	dm := globalState.documentManager
	globalState.mu.RUnlock()

	documentID, err := dm.CreateDocument(spaceID, "Test Doc", docData, nil)
	require.NoError(t, err)

	// Wait for event
	select {
	case event := <-eventChan:
		assert.Equal(t, anysync.EventDocumentCreated, event.Type)
		assert.Equal(t, spaceID, event.SpaceID)
		assert.Equal(t, documentID, event.Payload["document_id"])
		assert.NotEmpty(t, event.ID)
		assert.NotZero(t, event.Timestamp)
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for document.created event")
	}
}

// TestSubscribe_DocumentUpdatedEvent tests that document updates trigger events
func TestSubscribe_DocumentUpdatedEvent(t *testing.T) {
	ctx := setupForEventTests(t)
	defer teardownForEventTests(t)

	// Create space and document first
	createSpaceReq := &pb.CreateSpaceRequest{
		Name: "Test Space",
	}
	createSpaceResp, err := CreateSpace(context.Background(), createSpaceReq)
	require.NoError(t, err)
	spaceResp := createSpaceResp.(*pb.CreateSpaceResponse)
	spaceID := spaceResp.SpaceId

	time.Sleep(100 * time.Millisecond)

	globalState.mu.RLock()
	dm := globalState.documentManager
	globalState.mu.RUnlock()

	docData := []byte("initial content")
	documentID, err := dm.CreateDocument(spaceID, "Test Doc", docData, nil)
	require.NoError(t, err)

	// Subscribe to update events
	subscribeReq := &pb.SubscribeRequest{
		EventTypes: []string{"document.updated"},
		SpaceIds:   []string{spaceID},
	}

	subscriberID, eventChan, err := Subscribe(ctx, subscribeReq)
	require.NoError(t, err)
	defer Unsubscribe(subscriberID)

	// Update the document
	updatedData := []byte("updated content")
	err = dm.UpdateDocument(spaceID, documentID, updatedData)
	require.NoError(t, err)

	// Wait for event
	select {
	case event := <-eventChan:
		assert.Equal(t, anysync.EventDocumentUpdated, event.Type)
		assert.Equal(t, spaceID, event.SpaceID)
		assert.Equal(t, documentID, event.Payload["document_id"])
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for document.updated event")
	}
}

// TestSubscribe_SpaceDeletedEvent tests that space deletion triggers events
func TestSubscribe_SpaceDeletedEvent(t *testing.T) {
	ctx := setupForEventTests(t)
	defer teardownForEventTests(t)

	// Subscribe to space events
	subscribeReq := &pb.SubscribeRequest{
		EventTypes: []string{"space.deleted"},
		SpaceIds:   []string{},
	}

	subscriberID, eventChan, err := Subscribe(ctx, subscribeReq)
	require.NoError(t, err)
	defer Unsubscribe(subscriberID)

	// Create a space
	createSpaceReq := &pb.CreateSpaceRequest{
		Name: "Test Space",
	}
	createSpaceResp, err := CreateSpace(context.Background(), createSpaceReq)
	require.NoError(t, err)
	spaceResp := createSpaceResp.(*pb.CreateSpaceResponse)
	spaceID := spaceResp.SpaceId

	time.Sleep(100 * time.Millisecond)

	// Delete the space
	deleteSpaceReq := &pb.DeleteSpaceRequest{
		SpaceId: spaceID,
	}
	_, err = DeleteSpace(context.Background(), deleteSpaceReq)
	require.NoError(t, err)

	// Wait for event
	select {
	case event := <-eventChan:
		assert.Equal(t, anysync.EventSpaceDeleted, event.Type)
		assert.Equal(t, spaceID, event.SpaceID)
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for space.deleted event")
	}
}

// TestSubscribe_EventFiltering tests that event type filtering works
func TestSubscribe_EventFiltering(t *testing.T) {
	ctx := setupForEventTests(t)
	defer teardownForEventTests(t)

	// Subscribe only to space.created events
	subscribeReq := &pb.SubscribeRequest{
		EventTypes: []string{"space.created"},
		SpaceIds:   []string{},
	}

	subscriberID, eventChan, err := Subscribe(ctx, subscribeReq)
	require.NoError(t, err)
	defer Unsubscribe(subscriberID)

	// Create a space - should receive event
	createSpaceReq := &pb.CreateSpaceRequest{
		Name: "Test Space 1",
	}
	createSpaceResp, err := CreateSpace(context.Background(), createSpaceReq)
	require.NoError(t, err)
	spaceResp := createSpaceResp.(*pb.CreateSpaceResponse)
	spaceID := spaceResp.SpaceId

	// Wait for space.created event
	select {
	case event := <-eventChan:
		assert.Equal(t, anysync.EventSpaceCreated, event.Type)
		assert.Equal(t, spaceID, event.SpaceID)
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for space.created event")
	}

	time.Sleep(100 * time.Millisecond)

	// Delete the space - should NOT receive event (we only subscribed to space.created)
	deleteSpaceReq := &pb.DeleteSpaceRequest{
		SpaceId: spaceID,
	}
	_, err = DeleteSpace(context.Background(), deleteSpaceReq)
	require.NoError(t, err)

	// Should NOT receive space.deleted event
	select {
	case event := <-eventChan:
		t.Errorf("Should not receive space.deleted event, got: %v", event)
	case <-time.After(200 * time.Millisecond):
		// Expected - no event received
	}
}

// TestSubscribe_MultipleConcurrentSubscribers tests multiple subscribers receiving events
func TestSubscribe_MultipleConcurrentSubscribers(t *testing.T) {
	ctx := setupForEventTests(t)
	defer teardownForEventTests(t)

	// Create 3 subscribers
	subscriber1ID, eventChan1, err := Subscribe(ctx, &pb.SubscribeRequest{
		EventTypes: []string{"space.created"},
	})
	require.NoError(t, err)
	defer Unsubscribe(subscriber1ID)

	subscriber2ID, eventChan2, err := Subscribe(ctx, &pb.SubscribeRequest{
		EventTypes: []string{"space.created"},
	})
	require.NoError(t, err)
	defer Unsubscribe(subscriber2ID)

	subscriber3ID, eventChan3, err := Subscribe(ctx, &pb.SubscribeRequest{
		EventTypes: []string{}, // All events
	})
	require.NoError(t, err)
	defer Unsubscribe(subscriber3ID)

	// Create a space
	createSpaceReq := &pb.CreateSpaceRequest{
		Name: "Test Space",
	}
	createSpaceResp, err := CreateSpace(context.Background(), createSpaceReq)
	require.NoError(t, err)
	spaceResp := createSpaceResp.(*pb.CreateSpaceResponse)
	spaceID := spaceResp.SpaceId

	// All 3 subscribers should receive the event
	receivedCount := 0
	timeout := time.After(1 * time.Second)

	for receivedCount < 3 {
		select {
		case event := <-eventChan1:
			assert.Equal(t, anysync.EventSpaceCreated, event.Type)
			assert.Equal(t, spaceID, event.SpaceID)
			receivedCount++
		case event := <-eventChan2:
			assert.Equal(t, anysync.EventSpaceCreated, event.Type)
			assert.Equal(t, spaceID, event.SpaceID)
			receivedCount++
		case event := <-eventChan3:
			assert.Equal(t, anysync.EventSpaceCreated, event.Type)
			assert.Equal(t, spaceID, event.SpaceID)
			receivedCount++
		case <-timeout:
			t.Fatalf("Timeout: only received %d/3 events", receivedCount)
		}
	}
}

// TestSubscribe_NotInitialized tests subscribing before initialization
func TestSubscribe_NotInitialized(t *testing.T) {
	// Reset global state
	globalState = &State{}

	subscribeReq := &pb.SubscribeRequest{
		EventTypes: []string{},
	}

	_, _, err := Subscribe(context.Background(), subscribeReq)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not initialized")
}

// TestUnsubscribe_NotFound tests unsubscribing with invalid ID
func TestUnsubscribe_NotFound(t *testing.T) {
	_ = setupForEventTests(t)
	defer teardownForEventTests(t)

	err := Unsubscribe("invalid-subscriber-id")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "subscriber not found")
}

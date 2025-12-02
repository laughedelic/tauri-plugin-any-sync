package anysync

import (
	"context"
	"fmt"
	"testing"
	"time"
)

// TestEventManager_Subscribe tests basic subscription functionality.
func TestEventManager_Subscribe(t *testing.T) {
	em := NewEventManager()
	defer em.Close()

	ctx := context.Background()
	filter := EventFilter{
		EventTypes: []EventType{EventDocumentCreated},
		SpaceIDs:   []string{"space1"},
	}

	subscriberID, eventChan, err := em.Subscribe(ctx, filter)
	if err != nil {
		t.Fatalf("Subscribe failed: %v", err)
	}

	if subscriberID == "" {
		t.Error("Expected non-empty subscriber ID")
	}

	if eventChan == nil {
		t.Error("Expected non-nil event channel")
	}

	// Verify subscriber was registered
	if em.GetSubscriberCount() != 1 {
		t.Errorf("Expected 1 subscriber, got %d", em.GetSubscriberCount())
	}
}

// TestEventManager_Unsubscribe tests unsubscription functionality.
func TestEventManager_Unsubscribe(t *testing.T) {
	em := NewEventManager()
	defer em.Close()

	ctx := context.Background()
	filter := EventFilter{}

	subscriberID, eventChan, err := em.Subscribe(ctx, filter)
	if err != nil {
		t.Fatalf("Subscribe failed: %v", err)
	}

	// Unsubscribe
	err = em.Unsubscribe(subscriberID)
	if err != nil {
		t.Fatalf("Unsubscribe failed: %v", err)
	}

	// Verify subscriber was removed
	if em.GetSubscriberCount() != 0 {
		t.Errorf("Expected 0 subscribers, got %d", em.GetSubscriberCount())
	}

	// Verify channel was closed
	select {
	case _, ok := <-eventChan:
		if ok {
			t.Error("Expected channel to be closed")
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Timeout waiting for channel close")
	}

	// Unsubscribing again should return error
	err = em.Unsubscribe(subscriberID)
	if err == nil {
		t.Error("Expected error when unsubscribing non-existent subscriber")
	}
}

// TestEventManager_EmitEvent_NoFilter tests event emission with no filter (all events).
func TestEventManager_EmitEvent_NoFilter(t *testing.T) {
	em := NewEventManager()
	defer em.Close()

	ctx := context.Background()
	filter := EventFilter{} // No filter - all events

	_, eventChan, err := em.Subscribe(ctx, filter)
	if err != nil {
		t.Fatalf("Subscribe failed: %v", err)
	}

	// Emit event
	em.EmitEvent(EventDocumentCreated, "space1", map[string]string{
		"document_id": "doc1",
		"collection":  "notes",
	})

	// Receive event
	select {
	case event := <-eventChan:
		if event.Type != EventDocumentCreated {
			t.Errorf("Expected event type %s, got %s", EventDocumentCreated, event.Type)
		}
		if event.SpaceID != "space1" {
			t.Errorf("Expected space ID 'space1', got '%s'", event.SpaceID)
		}
		if event.Payload["document_id"] != "doc1" {
			t.Errorf("Expected document_id 'doc1', got '%s'", event.Payload["document_id"])
		}
		if event.ID == "" {
			t.Error("Expected non-empty event ID")
		}
		if event.Timestamp == 0 {
			t.Error("Expected non-zero timestamp")
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Timeout waiting for event")
	}
}

// TestEventManager_EmitEvent_WithEventTypeFilter tests event filtering by type.
func TestEventManager_EmitEvent_WithEventTypeFilter(t *testing.T) {
	em := NewEventManager()
	defer em.Close()

	ctx := context.Background()
	filter := EventFilter{
		EventTypes: []EventType{EventDocumentCreated},
	}

	_, eventChan, err := em.Subscribe(ctx, filter)
	if err != nil {
		t.Fatalf("Subscribe failed: %v", err)
	}

	// Emit matching event
	em.EmitEvent(EventDocumentCreated, "space1", map[string]string{
		"document_id": "doc1",
	})

	// Should receive the event
	select {
	case event := <-eventChan:
		if event.Type != EventDocumentCreated {
			t.Errorf("Expected event type %s, got %s", EventDocumentCreated, event.Type)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Timeout waiting for matching event")
	}

	// Emit non-matching event
	em.EmitEvent(EventDocumentUpdated, "space1", map[string]string{
		"document_id": "doc1",
	})

	// Should NOT receive the event
	select {
	case event := <-eventChan:
		t.Errorf("Should not receive non-matching event, got: %v", event)
	case <-time.After(100 * time.Millisecond):
		// Expected - no event received
	}
}

// TestEventManager_EmitEvent_WithSpaceIDFilter tests event filtering by space ID.
func TestEventManager_EmitEvent_WithSpaceIDFilter(t *testing.T) {
	em := NewEventManager()
	defer em.Close()

	ctx := context.Background()
	filter := EventFilter{
		SpaceIDs: []string{"space1", "space2"},
	}

	_, eventChan, err := em.Subscribe(ctx, filter)
	if err != nil {
		t.Fatalf("Subscribe failed: %v", err)
	}

	// Emit matching event (space1)
	em.EmitEvent(EventDocumentCreated, "space1", map[string]string{
		"document_id": "doc1",
	})

	// Should receive the event
	select {
	case event := <-eventChan:
		if event.SpaceID != "space1" {
			t.Errorf("Expected space ID 'space1', got '%s'", event.SpaceID)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Timeout waiting for matching event")
	}

	// Emit matching event (space2)
	em.EmitEvent(EventDocumentCreated, "space2", map[string]string{
		"document_id": "doc2",
	})

	// Should receive the event
	select {
	case event := <-eventChan:
		if event.SpaceID != "space2" {
			t.Errorf("Expected space ID 'space2', got '%s'", event.SpaceID)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Timeout waiting for matching event")
	}

	// Emit non-matching event (space3)
	em.EmitEvent(EventDocumentCreated, "space3", map[string]string{
		"document_id": "doc3",
	})

	// Should NOT receive the event
	select {
	case event := <-eventChan:
		t.Errorf("Should not receive non-matching event, got: %v", event)
	case <-time.After(100 * time.Millisecond):
		// Expected - no event received
	}
}

// TestEventManager_MultipleSubscribers tests broadcasting to multiple subscribers.
func TestEventManager_MultipleSubscribers(t *testing.T) {
	em := NewEventManager()
	defer em.Close()

	ctx := context.Background()

	// Subscribe 3 subscribers with different filters
	_, eventChan1, err := em.Subscribe(ctx, EventFilter{
		EventTypes: []EventType{EventDocumentCreated},
	})
	if err != nil {
		t.Fatalf("Subscribe 1 failed: %v", err)
	}

	_, eventChan2, err := em.Subscribe(ctx, EventFilter{
		SpaceIDs: []string{"space1"},
	})
	if err != nil {
		t.Fatalf("Subscribe 2 failed: %v", err)
	}

	_, eventChan3, err := em.Subscribe(ctx, EventFilter{}) // All events
	if err != nil {
		t.Fatalf("Subscribe 3 failed: %v", err)
	}

	if em.GetSubscriberCount() != 3 {
		t.Errorf("Expected 3 subscribers, got %d", em.GetSubscriberCount())
	}

	// Emit event that matches all subscribers
	em.EmitEvent(EventDocumentCreated, "space1", map[string]string{
		"document_id": "doc1",
	})

	// All 3 should receive the event
	receivedCount := 0
	timeout := time.After(100 * time.Millisecond)

	for receivedCount < 3 {
		select {
		case event := <-eventChan1:
			if event.Type != EventDocumentCreated {
				t.Errorf("Subscriber 1: wrong event type: %s", event.Type)
			}
			receivedCount++
		case event := <-eventChan2:
			if event.SpaceID != "space1" {
				t.Errorf("Subscriber 2: wrong space ID: %s", event.SpaceID)
			}
			receivedCount++
		case event := <-eventChan3:
			if event.Type != EventDocumentCreated {
				t.Errorf("Subscriber 3: wrong event type: %s", event.Type)
			}
			receivedCount++
		case <-timeout:
			t.Fatalf("Timeout: only received %d/3 events", receivedCount)
		}
	}
}

// TestEventManager_ContextCancellation tests automatic unsubscribe on context cancellation.
func TestEventManager_ContextCancellation(t *testing.T) {
	em := NewEventManager()
	defer em.Close()

	ctx, cancel := context.WithCancel(context.Background())
	filter := EventFilter{}

	_, eventChan, err := em.Subscribe(ctx, filter)
	if err != nil {
		t.Fatalf("Subscribe failed: %v", err)
	}

	if em.GetSubscriberCount() != 1 {
		t.Errorf("Expected 1 subscriber, got %d", em.GetSubscriberCount())
	}

	// Cancel context
	cancel()

	// Wait for unsubscribe to complete
	time.Sleep(50 * time.Millisecond)

	// Verify subscriber was removed
	if em.GetSubscriberCount() != 0 {
		t.Errorf("Expected 0 subscribers after context cancellation, got %d", em.GetSubscriberCount())
	}

	// Verify channel was closed
	select {
	case _, ok := <-eventChan:
		if ok {
			t.Error("Expected channel to be closed after context cancellation")
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Timeout waiting for channel close")
	}
}

// TestEventManager_BufferOverflow tests behavior when subscriber channel is full.
func TestEventManager_BufferOverflow(t *testing.T) {
	em := NewEventManager()
	defer em.Close()

	ctx := context.Background()
	filter := EventFilter{}

	_, eventChan, err := em.Subscribe(ctx, filter)
	if err != nil {
		t.Fatalf("Subscribe failed: %v", err)
	}

	// Fill the buffer (100 events) + 10 more to test overflow
	for i := 0; i < 110; i++ {
		em.EmitEvent(EventDocumentCreated, "space1", map[string]string{
			"document_id": fmt.Sprintf("doc%d", i),
		})
	}

	// The first 100 should be in the channel
	receivedCount := 0
	timeout := time.After(100 * time.Millisecond)

drain:
	for {
		select {
		case <-eventChan:
			receivedCount++
		case <-timeout:
			break drain
		}
	}

	// Should have received exactly 100 (buffer size)
	// The last 10 were dropped due to buffer overflow
	if receivedCount != 100 {
		t.Errorf("Expected 100 events (buffer size), got %d", receivedCount)
	}
}

// TestEventManager_Close tests cleanup on close.
func TestEventManager_Close(t *testing.T) {
	em := NewEventManager()

	ctx := context.Background()

	// Create 3 subscribers
	_, eventChan1, _ := em.Subscribe(ctx, EventFilter{})
	_, eventChan2, _ := em.Subscribe(ctx, EventFilter{})
	_, eventChan3, _ := em.Subscribe(ctx, EventFilter{})

	if em.GetSubscriberCount() != 3 {
		t.Errorf("Expected 3 subscribers, got %d", em.GetSubscriberCount())
	}

	// Close the event manager
	err := em.Close()
	if err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	// Verify all subscribers were removed
	if em.GetSubscriberCount() != 0 {
		t.Errorf("Expected 0 subscribers after close, got %d", em.GetSubscriberCount())
	}

	// Verify all channels were closed
	for i, ch := range []<-chan *Event{eventChan1, eventChan2, eventChan3} {
		select {
		case _, ok := <-ch:
			if ok {
				t.Errorf("Expected channel %d to be closed", i+1)
			}
		case <-time.After(100 * time.Millisecond):
			t.Errorf("Timeout waiting for channel %d close", i+1)
		}
	}
}

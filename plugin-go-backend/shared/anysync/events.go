// Package anysync provides Any-Sync integration components.
package anysync

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// EventType represents the type of event.
type EventType string

const (
	// Document events
	EventDocumentCreated EventType = "document.created"
	EventDocumentUpdated EventType = "document.updated"
	EventDocumentDeleted EventType = "document.deleted"

	// Space events
	EventSpaceCreated EventType = "space.created"
	EventSpaceDeleted EventType = "space.deleted"

	// Sync events (for Phase 6)
	EventSyncStarted   EventType = "sync.started"
	EventSyncCompleted EventType = "sync.completed"
	EventSyncError     EventType = "sync.error"
	EventSyncConflict  EventType = "sync.conflict"
)

// Event represents a single event in the system.
type Event struct {
	ID        string            // Unique event ID
	Type      EventType         // Event type
	SpaceID   string            // Space ID (if applicable)
	Timestamp int64             // Unix timestamp
	Payload   map[string]string // Event-specific data
}

// EventFilter defines criteria for filtering events.
type EventFilter struct {
	EventTypes []EventType // Empty means all types
	SpaceIDs   []string    // Empty means all spaces
}

// Subscriber represents a registered event subscriber.
type Subscriber struct {
	ID      string
	Filter  EventFilter
	Channel chan *Event
}

// EventManager manages event subscriptions and broadcasts.
// It provides a local event system for Phase 2E (local-only) and will
// integrate with SyncTree's UpdateListener in Phase 6 (network sync).
type EventManager struct {
	mu          sync.RWMutex
	subscribers map[string]*Subscriber
}

// NewEventManager creates a new EventManager.
func NewEventManager() *EventManager {
	return &EventManager{
		subscribers: make(map[string]*Subscriber),
	}
}

// Subscribe registers a new subscriber with the given filter.
// Returns a subscriber ID and a channel that will receive matching events.
func (em *EventManager) Subscribe(ctx context.Context, filter EventFilter) (string, <-chan *Event, error) {
	em.mu.Lock()
	defer em.mu.Unlock()

	// Generate unique subscriber ID
	subscriberID := uuid.New().String()

	// Create buffered channel to prevent blocking on slow consumers
	// Buffer size of 100 should handle bursts of events
	eventChan := make(chan *Event, 100)

	subscriber := &Subscriber{
		ID:      subscriberID,
		Filter:  filter,
		Channel: eventChan,
	}

	em.subscribers[subscriberID] = subscriber

	// Start goroutine to handle context cancellation
	go func() {
		<-ctx.Done()
		em.Unsubscribe(subscriberID)
	}()

	return subscriberID, eventChan, nil
}

// Unsubscribe removes a subscriber and closes its channel.
func (em *EventManager) Unsubscribe(subscriberID string) error {
	em.mu.Lock()
	defer em.mu.Unlock()

	subscriber, exists := em.subscribers[subscriberID]
	if !exists {
		return fmt.Errorf("subscriber not found: %s", subscriberID)
	}

	// Close the channel to signal no more events
	close(subscriber.Channel)

	// Remove from registry
	delete(em.subscribers, subscriberID)

	return nil
}

// EmitEvent broadcasts an event to all matching subscribers.
// This is a non-blocking operation - if a subscriber's channel is full,
// the event is dropped for that subscriber (fire-and-forget semantics).
func (em *EventManager) EmitEvent(eventType EventType, spaceID string, payload map[string]string) {
	em.mu.RLock()
	defer em.mu.RUnlock()

	// Create the event
	event := &Event{
		ID:        uuid.New().String(),
		Type:      eventType,
		SpaceID:   spaceID,
		Timestamp: time.Now().Unix(),
		Payload:   payload,
	}

	// Broadcast to all matching subscribers
	for _, subscriber := range em.subscribers {
		if em.matchesFilter(event, subscriber.Filter) {
			// Non-blocking send - drop event if channel is full
			select {
			case subscriber.Channel <- event:
				// Event delivered successfully
			default:
				// Channel full - drop event (could log this in production)
			}
		}
	}
}

// matchesFilter checks if an event matches a subscriber's filter.
func (em *EventManager) matchesFilter(event *Event, filter EventFilter) bool {
	// Check event type filter
	if len(filter.EventTypes) > 0 {
		matched := false
		for _, eventType := range filter.EventTypes {
			if event.Type == eventType {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}

	// Check space ID filter
	if len(filter.SpaceIDs) > 0 {
		matched := false
		for _, spaceID := range filter.SpaceIDs {
			if event.SpaceID == spaceID {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}

	return true
}

// GetSubscriberCount returns the current number of subscribers.
// This is useful for testing and monitoring.
func (em *EventManager) GetSubscriberCount() int {
	em.mu.RLock()
	defer em.mu.RUnlock()
	return len(em.subscribers)
}

// Close unsubscribes all subscribers and cleans up resources.
func (em *EventManager) Close() error {
	em.mu.Lock()
	defer em.mu.Unlock()

	// Close all subscriber channels
	for _, subscriber := range em.subscribers {
		close(subscriber.Channel)
	}

	// Clear the subscribers map
	em.subscribers = make(map[string]*Subscriber)

	return nil
}

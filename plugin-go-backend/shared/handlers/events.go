package handlers

import (
	"context"
	"fmt"

	"anysync-backend/shared/anysync"
	pb "anysync-backend/shared/proto/syncspace/v1"
)

// Subscribe creates a subscription to events and returns the subscriber ID and event channel.
// This is a special handler for streaming RPCs that directly accesses the EventManager.
func Subscribe(ctx context.Context, req *pb.SubscribeRequest) (string, <-chan *anysync.Event, error) {
	if err := ensureInitialized(); err != nil {
		return "", nil, err
	}

	globalState.mu.RLock()
	eventManager := globalState.eventManager
	globalState.mu.RUnlock()

	if eventManager == nil {
		return "", nil, fmt.Errorf("event manager not initialized")
	}

	// Convert protobuf event types to anysync.EventType
	eventTypes := make([]anysync.EventType, len(req.EventTypes))
	for i, et := range req.EventTypes {
		eventTypes[i] = anysync.EventType(et)
	}

	// Create filter
	filter := anysync.EventFilter{
		EventTypes: eventTypes,
		SpaceIDs:   req.SpaceIds,
	}

	// Subscribe to events
	subscriberID, eventChan, err := eventManager.Subscribe(ctx, filter)
	if err != nil {
		return "", nil, fmt.Errorf("failed to subscribe: %w", err)
	}

	return subscriberID, eventChan, nil
}

// Unsubscribe removes a subscription by ID.
func Unsubscribe(subscriberID string) error {
	globalState.mu.RLock()
	eventManager := globalState.eventManager
	globalState.mu.RUnlock()

	if eventManager == nil {
		return fmt.Errorf("event manager not initialized")
	}

	return eventManager.Unsubscribe(subscriberID)
}

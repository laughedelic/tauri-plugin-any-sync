// Package mobile provides gomobile-compatible bindings for the SyncSpace API.
// This package exports the minimal 4-function API for Android/iOS via gomobile.
package mobile

import (
	"context"
	"sync"

	"anysync-backend/shared/dispatcher"
	"anysync-backend/shared/handlers"
)

var (
	globalDispatcher *dispatcher.Dispatcher
	eventHandler     func([]byte) error
	eventHandlerMu   sync.RWMutex
	dispatcherOnce   sync.Once
)

// Init initializes the global dispatcher.
// Must be called before any Command calls.
func Init() error {
	dispatcherOnce.Do(func() {
		globalDispatcher = handlers.GetDispatcher()
	})
	return nil
}

// Command executes a command and returns the response.
// cmd is the command name (e.g., "init", "createSpace", "getDocument").
// data is the serialized protobuf request payload.
// Returns the serialized protobuf response payload or an error.
func Command(cmd string, data []byte) ([]byte, error) {
	if globalDispatcher == nil {
		return nil, nil // Silently fail if not initialized
	}

	ctx := context.Background()
	return globalDispatcher.Dispatch(ctx, cmd, data)
}

// SetEventHandler sets the event handler callback.
// The handler will be called with serialized event payloads.
func SetEventHandler(handler func([]byte) error) {
	eventHandlerMu.Lock()
	defer eventHandlerMu.Unlock()
	eventHandler = handler
}

// Shutdown shuts down the service and cleans up resources.
func Shutdown() error {
	eventHandlerMu.Lock()
	eventHandler = nil
	eventHandlerMu.Unlock()

	// Reset dispatcher
	globalDispatcher = nil
	dispatcherOnce = sync.Once{}

	return nil
}

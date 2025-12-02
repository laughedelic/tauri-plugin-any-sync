// Package dispatcher provides a command dispatcher for routing operations to handlers.
package dispatcher

import (
	"context"
	"fmt"

	"google.golang.org/protobuf/proto"
)

// Handler is a function that processes a command.
// It receives the deserialized request message and returns a response message or error.
type Handler func(ctx context.Context, req proto.Message) (proto.Message, error)

// Dispatcher routes commands to their handlers.
type Dispatcher struct {
	handlers map[string]HandlerEntry
}

// HandlerEntry contains the handler function and message types for a command.
type HandlerEntry struct {
	Handler     Handler
	RequestType proto.Message // Used for creating new instances of the request type
}

// New creates a new dispatcher.
func New() *Dispatcher {
	return &Dispatcher{
		handlers: make(map[string]HandlerEntry),
	}
}

// Register registers a handler for a command.
// requestType should be a zero-value instance of the request message type.
func (d *Dispatcher) Register(command string, handler Handler, requestType proto.Message) {
	d.handlers[command] = HandlerEntry{
		Handler:     handler,
		RequestType: requestType,
	}
}

// Dispatch routes a command to its handler.
// Returns the serialized response or an error.
func (d *Dispatcher) Dispatch(ctx context.Context, command string, payload []byte) ([]byte, error) {
	entry, ok := d.handlers[command]
	if !ok {
		return nil, fmt.Errorf("unknown command: %s", command)
	}

	// Create a new instance of the request type
	req := proto.Clone(entry.RequestType)
	if err := proto.Unmarshal(payload, req); err != nil {
		return nil, fmt.Errorf("failed to unmarshal request: %w", err)
	}

	// Call the handler
	resp, err := entry.Handler(ctx, req)
	if err != nil {
		return nil, err
	}

	// Marshal the response
	respBytes, err := proto.Marshal(resp)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal response: %w", err)
	}

	return respBytes, nil
}

// Commands returns the list of registered commands.
func (d *Dispatcher) Commands() []string {
	commands := make([]string, 0, len(d.handlers))
	for cmd := range d.handlers {
		commands = append(commands, cmd)
	}
	return commands
}

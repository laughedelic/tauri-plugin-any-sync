package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"

	transportpb "anysync-backend/desktop/proto/transport/v1"
	"anysync-backend/shared/anysync"
	"anysync-backend/shared/dispatcher"
	"anysync-backend/shared/handlers"
	syncspacepb "anysync-backend/shared/proto/syncspace/v1"
)

var (
	port = flag.Int("port", 0, "Port to listen on (0 for random)")
	host = flag.String("host", "localhost", "Host to bind to")
)

// Server implements the TransportService by calling the dispatcher
type Server struct {
	transportpb.UnimplementedTransportServiceServer
	dispatcher *dispatcher.Dispatcher
}

func NewServer() *Server {
	return &Server{
		dispatcher: handlers.GetDispatcher(),
	}
}

// Init initializes the backend
func (s *Server) Init(ctx context.Context, req *transportpb.InitRequest) (*transportpb.InitResponse, error) {
	// Convert transport.InitRequest to syncspace.InitRequest
	syncspaceReq := &syncspacepb.InitRequest{
		DataDir:   req.StoragePath,
		NetworkId: req.NetworkId,
	}

	// Marshal to bytes
	reqBytes, err := proto.Marshal(syncspaceReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal init request: %w", err)
	}

	// Call dispatcher with PascalCase command name
	respBytes, err := s.dispatcher.Dispatch(ctx, "Init", reqBytes)
	if err != nil {
		return nil, err
	}

	// Unmarshal response
	var syncspaceResp syncspacepb.InitResponse
	if err := proto.Unmarshal(respBytes, &syncspaceResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal init response: %w", err)
	}

	msg := "initialized successfully"
	if !syncspaceResp.Success {
		msg = "initialization failed"
	}

	return &transportpb.InitResponse{
		Message: msg,
	}, nil
}

// Command executes a command through the dispatcher
func (s *Server) Command(ctx context.Context, req *transportpb.CommandRequest) (*transportpb.CommandResponse, error) {
	// req.Data contains protobuf bytes from TypeScript via Rust
	// Dispatch directly - handlers expect and return protobuf bytes
	respBytes, err := s.dispatcher.Dispatch(ctx, req.Cmd, req.Data)
	if err != nil {
		return nil, err
	}

	return &transportpb.CommandResponse{
		Data: respBytes,
	}, nil
}

// Subscribe streams events to the client
func (s *Server) Subscribe(req *transportpb.SubscribeRequest, stream transportpb.TransportService_SubscribeServer) error {
	ctx := stream.Context()

	// Convert transport event types to syncspace event types
	syncspaceReq := &syncspacepb.SubscribeRequest{
		EventTypes: req.EventTypes,
		SpaceIds:   []string{}, // Empty means all spaces
	}

	// Get subscription from handlers (special handling for streaming)
	subscriberID, eventChan, err := handlers.Subscribe(ctx, syncspaceReq)
	if err != nil {
		return fmt.Errorf("failed to subscribe: %w", err)
	}
	defer handlers.Unsubscribe(subscriberID)

	// Stream events to client until context is cancelled or error occurs
	for {
		select {
		case <-ctx.Done():
			// Client disconnected
			return ctx.Err()
		case event, ok := <-eventChan:
			if !ok {
				// Channel closed (unsubscribed)
				return nil
			}

			// Convert anysync.Event to syncspace.SubscribeResponse
			syncspaceEvent := &syncspacepb.SubscribeResponse{
				EventId:   event.ID,
				EventType: string(event.Type),
				SpaceId:   event.SpaceID,
				Timestamp: event.Timestamp,
				Payload:   marshalEventPayload(event),
			}

			// Marshal to bytes
			eventBytes, err := proto.Marshal(syncspaceEvent)
			if err != nil {
				log.Printf("Failed to marshal event: %v", err)
				continue
			}

			// Send to transport stream
			transportEvent := &transportpb.SubscribeResponse{
				Type:      string(event.Type),
				Data:      eventBytes,
				Timestamp: event.Timestamp * 1000, // Convert to milliseconds
			}

			if err := stream.Send(transportEvent); err != nil {
				return fmt.Errorf("failed to send event: %w", err)
			}
		}
	}
}

// marshalEventPayload converts event payload map to protobuf bytes
// For now, this returns empty bytes. In the future, this could marshal
// specific event types (DocumentCreatedEvent, DocumentUpdatedEvent, etc.)
func marshalEventPayload(event *anysync.Event) []byte {
	// TODO: Marshal specific event payload types based on event.Type
	return []byte{}
}

// Shutdown shuts down the backend
func (s *Server) Shutdown(ctx context.Context, req *transportpb.ShutdownRequest) (*transportpb.ShutdownResponse, error) {
	// Convert transport.ShutdownRequest to syncspace.ShutdownRequest
	syncspaceReq := &syncspacepb.ShutdownRequest{}

	// Marshal to bytes
	reqBytes, err := proto.Marshal(syncspaceReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal shutdown request: %w", err)
	}

	// Call dispatcher with PascalCase command name
	respBytes, err := s.dispatcher.Dispatch(ctx, "Shutdown", reqBytes)
	if err != nil {
		return nil, err
	}

	// Unmarshal response
	var syncspaceResp syncspacepb.ShutdownResponse
	if err := proto.Unmarshal(respBytes, &syncspaceResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal shutdown response: %w", err)
	}

	msg := "shutdown successfully"
	if !syncspaceResp.Success {
		msg = "shutdown failed"
	}

	return &transportpb.ShutdownResponse{
		Message: msg,
	}, nil
}

func main() {
	flag.Parse()

	// Create gRPC server
	grpcServer := grpc.NewServer()

	// Register Transport service
	transportServer := NewServer()
	transportpb.RegisterTransportServiceServer(grpcServer, transportServer)

	// Determine listen address
	listenAddr := fmt.Sprintf("%s:%d", *host, *port)
	if *port == 0 {
		listenAddr = fmt.Sprintf("%s:0", *host)
	}

	// Create listener
	lis, err := net.Listen("tcp", listenAddr)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	defer lis.Close()

	// Get actual port
	actualPort := lis.Addr().(*net.TCPAddr).Port
	fmt.Printf("SyncSpace gRPC server listening on %s:%d\n", *host, actualPort)

	// Write port to file for communication with parent process
	if portFile := os.Getenv("ANY_SYNC_PORT_FILE"); portFile != "" {
		if err := os.WriteFile(portFile, []byte(fmt.Sprintf("%d", actualPort)), 0644); err != nil {
			log.Printf("Warning: Failed to write port file: %v", err)
		}
	}

	// Handle signals for graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// Start gRPC server in a goroutine
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Printf("gRPC server error: %v", err)
		}
	}()

	// Wait for shutdown signal
	<-sigCh
	fmt.Println("\nShutting down server...")

	// Graceful stop
	grpcServer.GracefulStop()

	fmt.Println("Server stopped")
}

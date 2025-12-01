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

	"anysync-backend/shared/handlers"
	pb "anysync-backend/shared/proto/syncspace/v1"
)

var (
	port = flag.Int("port", 0, "Port to listen on (0 for random)")
	host = flag.String("host", "localhost", "Host to bind to")
)

// Server implements the SyncSpaceService by calling handlers directly
type Server struct {
	pb.UnimplementedSyncSpaceServiceServer
}

func NewServer() *Server {
	return &Server{}
}

// gRPC service methods - these call handlers directly (no marshaling needed)

func (s *Server) Init(ctx context.Context, req *pb.InitRequest) (*pb.InitResponse, error) {
	resp, err := handlers.Init(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.(*pb.InitResponse), nil
}

func (s *Server) Shutdown(ctx context.Context, req *pb.ShutdownRequest) (*pb.ShutdownResponse, error) {
	resp, err := handlers.Shutdown(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.(*pb.ShutdownResponse), nil
}

func (s *Server) CreateSpace(ctx context.Context, req *pb.CreateSpaceRequest) (*pb.CreateSpaceResponse, error) {
	resp, err := handlers.CreateSpace(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.(*pb.CreateSpaceResponse), nil
}

func (s *Server) JoinSpace(ctx context.Context, req *pb.JoinSpaceRequest) (*pb.JoinSpaceResponse, error) {
	resp, err := handlers.JoinSpace(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.(*pb.JoinSpaceResponse), nil
}

func (s *Server) LeaveSpace(ctx context.Context, req *pb.LeaveSpaceRequest) (*pb.LeaveSpaceResponse, error) {
	resp, err := handlers.LeaveSpace(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.(*pb.LeaveSpaceResponse), nil
}

func (s *Server) ListSpaces(ctx context.Context, req *pb.ListSpacesRequest) (*pb.ListSpacesResponse, error) {
	resp, err := handlers.ListSpaces(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.(*pb.ListSpacesResponse), nil
}

func (s *Server) DeleteSpace(ctx context.Context, req *pb.DeleteSpaceRequest) (*pb.DeleteSpaceResponse, error) {
	resp, err := handlers.DeleteSpace(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.(*pb.DeleteSpaceResponse), nil
}

func (s *Server) CreateDocument(ctx context.Context, req *pb.CreateDocumentRequest) (*pb.CreateDocumentResponse, error) {
	resp, err := handlers.CreateDocument(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.(*pb.CreateDocumentResponse), nil
}

func (s *Server) GetDocument(ctx context.Context, req *pb.GetDocumentRequest) (*pb.GetDocumentResponse, error) {
	resp, err := handlers.GetDocument(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.(*pb.GetDocumentResponse), nil
}

func (s *Server) UpdateDocument(ctx context.Context, req *pb.UpdateDocumentRequest) (*pb.UpdateDocumentResponse, error) {
	resp, err := handlers.UpdateDocument(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.(*pb.UpdateDocumentResponse), nil
}

func (s *Server) DeleteDocument(ctx context.Context, req *pb.DeleteDocumentRequest) (*pb.DeleteDocumentResponse, error) {
	resp, err := handlers.DeleteDocument(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.(*pb.DeleteDocumentResponse), nil
}

func (s *Server) ListDocuments(ctx context.Context, req *pb.ListDocumentsRequest) (*pb.ListDocumentsResponse, error) {
	resp, err := handlers.ListDocuments(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.(*pb.ListDocumentsResponse), nil
}

func (s *Server) QueryDocuments(ctx context.Context, req *pb.QueryDocumentsRequest) (*pb.QueryDocumentsResponse, error) {
	resp, err := handlers.QueryDocuments(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.(*pb.QueryDocumentsResponse), nil
}

func (s *Server) StartSync(ctx context.Context, req *pb.StartSyncRequest) (*pb.StartSyncResponse, error) {
	resp, err := handlers.StartSync(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.(*pb.StartSyncResponse), nil
}

func (s *Server) PauseSync(ctx context.Context, req *pb.PauseSyncRequest) (*pb.PauseSyncResponse, error) {
	resp, err := handlers.PauseSync(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.(*pb.PauseSyncResponse), nil
}

func (s *Server) GetSyncStatus(ctx context.Context, req *pb.GetSyncStatusRequest) (*pb.GetSyncStatusResponse, error) {
	resp, err := handlers.GetSyncStatus(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.(*pb.GetSyncStatusResponse), nil
}

func (s *Server) Subscribe(req *pb.SubscribeRequest, stream pb.SyncSpaceService_SubscribeServer) error {
	// TODO: Implement event streaming
	return fmt.Errorf("event streaming not implemented yet")
}

func main() {
	flag.Parse()

	// Create gRPC server
	grpcServer := grpc.NewServer()

	// Register SyncSpace service
	syncSpaceServer := NewServer()
	pb.RegisterSyncSpaceServiceServer(grpcServer, syncSpaceServer)

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

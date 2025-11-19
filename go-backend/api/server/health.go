package server

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "anysync-backend/api/proto"
)

// HealthServer implements the HealthService gRPC service
type HealthServer struct {
	pb.UnimplementedHealthServiceServer
	serverID  string
	startTime time.Time
}

// NewHealthServer creates a new HealthServer instance
func NewHealthServer() *HealthServer {
	return &HealthServer{
		serverID:  fmt.Sprintf("server-%d", time.Now().Unix()),
		startTime: time.Now(),
	}
}

// Check implements the health check RPC
func (s *HealthServer) Check(ctx context.Context, req *pb.HealthCheckRequest) (*pb.HealthCheckResponse, error) {
	return &pb.HealthCheckResponse{
		Status:  pb.HealthCheckResponse_SERVING,
		Message: fmt.Sprintf("Server %s is running", s.serverID),
	}, nil
}

// Ping implements the ping RPC for round-trip testing
func (s *HealthServer) Ping(ctx context.Context, req *pb.PingRequest) (*pb.PingResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "ping request cannot be nil")
	}

	response := &pb.PingResponse{
		Message:           fmt.Sprintf("Echo: %s", req.GetMessage()),
		RequestTimestamp:  req.GetTimestamp(),
		ResponseTimestamp: time.Now().Unix(),
		ServerId:          s.serverID,
	}

	return response, nil
}

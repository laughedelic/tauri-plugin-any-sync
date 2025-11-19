package server

import (
	"context"
	"testing"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "anysync-backend/api/proto"
)

func TestHealthServer_Check(t *testing.T) {
	server := NewHealthServer()

	// Test successful health check
	req := &pb.HealthCheckRequest{}
	resp, err := server.Check(context.Background(), req)

	if err != nil {
		t.Fatalf("Health check failed: %v", err)
	}

	if resp.Status != pb.HealthCheckResponse_SERVING {
		t.Errorf("Expected SERVING status, got %d", resp.Status)
	}

	if resp.Message == "" {
		t.Error("Expected non-empty message")
	}
}

func TestHealthServer_Ping(t *testing.T) {
	server := NewHealthServer()

	// Test successful ping
	req := &pb.PingRequest{
		Message:   "test message",
		Timestamp: time.Now().Unix(),
	}

	resp, err := server.Ping(context.Background(), req)

	if err != nil {
		t.Fatalf("Ping failed: %v", err)
	}

	// Check response
	expectedMessage := "Echo: test message"
	if resp.Message != expectedMessage {
		t.Errorf("Expected message %s, got %s", expectedMessage, resp.Message)
	}

	if resp.RequestTimestamp != req.Timestamp {
		t.Errorf("Expected request timestamp %d, got %d", req.Timestamp, resp.RequestTimestamp)
	}

	if resp.ResponseTimestamp < req.Timestamp {
		t.Error("Response timestamp should be greater than or equal to request timestamp")
	}

	if resp.ServerId == "" {
		t.Error("Expected non-empty server ID")
	}
}

func TestHealthServer_PingNilRequest(t *testing.T) {
	server := NewHealthServer()

	// Test ping with nil request
	_, err := server.Ping(context.Background(), nil)

	if err == nil {
		t.Error("Expected error for nil request")
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("Expected gRPC status error, got %v", err)
	}

	if st.Code() != codes.InvalidArgument {
		t.Errorf("Expected InvalidArgument code, got %v", st.Code())
	}
}

func TestHealthServer_PingEmptyMessage(t *testing.T) {
	server := NewHealthServer()

	// Test ping with empty message
	req := &pb.PingRequest{
		Message:   "",
		Timestamp: time.Now().Unix(),
	}

	resp, err := server.Ping(context.Background(), req)

	if err != nil {
		t.Fatalf("Ping failed: %v", err)
	}

	// Should handle empty message gracefully
	if resp.Message != "Echo: " {
		t.Errorf("Expected 'Echo: ', got %s", resp.Message)
	}
}

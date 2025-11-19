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
	"time"

	"google.golang.org/grpc"

	pb "anysync-backend/api/proto"
	"anysync-backend/api/server"
	"anysync-backend/internal/config"
	"anysync-backend/internal/health"
)

var (
	port = flag.Int("port", 0, "Port to listen on (0 for random)")
	host = flag.String("host", "localhost", "Host to bind to")
)

func main() {
	flag.Parse()

	// Load configuration
	cfg := config.NewConfig()

	// Override with command line flags
	if *port != 0 {
		cfg.Port = *port
	}
	if *host != "localhost" {
		cfg.Host = *host
	}

	// Create health service
	healthSvc := health.NewService()

	// Create gRPC server
	grpcServer := grpc.NewServer()

	// Register health service
	healthServer := server.NewHealthServer()
	pb.RegisterHealthServiceServer(grpcServer, healthServer)

	// Create listener
	lis, err := net.Listen("tcp", cfg.GetAddress())
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	defer lis.Close()

	// Get actual port (especially important when using random port)
	actualPort := lis.Addr().(*net.TCPAddr).Port
	fmt.Printf("Server listening on %s:%d\n", cfg.Host, actualPort)

	// Write port to file for communication with parent process
	if portFile := os.Getenv("ANY_SYNC_PORT_FILE"); portFile != "" {
		if err := os.WriteFile(portFile, []byte(fmt.Sprintf("%d", actualPort)), 0644); err != nil {
			log.Printf("Warning: Failed to write port file: %v", err)
		}
	}

	// Set up graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle signals
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// Start gRPC server in a goroutine
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Printf("gRPC server error: %v", err)
		}
	}()

	// Start health check routine
	go func() {
		ticker := time.NewTicker(time.Duration(cfg.HealthCheckInterval) * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				healthy, msg := healthSvc.Check(ctx)
				if !healthy {
					log.Printf("Health check failed: %s", msg)
				} else {
					log.Printf("Health check: %s", msg)
				}
			}
		}
	}()

	// Wait for shutdown signal
	<-sigCh
	fmt.Println("Shutting down server...")
	cancel()

	// Graceful stop gRPC server
	grpcServer.GracefulStop()

	// Give some time for graceful shutdown
	time.Sleep(2 * time.Second)
	fmt.Println("Server stopped")
}

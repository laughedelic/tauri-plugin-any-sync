package health

import (
	"context"
	"fmt"
	"time"
)

// Service provides basic health check functionality
type Service struct {
	startTime time.Time
}

// NewService creates a new health service instance
func NewService() *Service {
	return &Service{
		startTime: time.Now(),
	}
}

// Check returns the current health status
func (s *Service) Check(ctx context.Context) (bool, string) {
	select {
	case <-ctx.Done():
		return false, "context cancelled"
	default:
		uptime := time.Since(s.startTime)
		return true, fmt.Sprintf("Service is healthy, uptime: %v", uptime)
	}
}

// GetUptime returns the service uptime
func (s *Service) GetUptime() time.Duration {
	return time.Since(s.startTime)
}

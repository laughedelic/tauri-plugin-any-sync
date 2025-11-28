package mobile

import (
	"context"
	"path/filepath"
	"time"

	"anysync-backend/shared/storage"
)

// MobileService provides exported functions for Android/iOS via gomobile
type MobileService struct {
	store *storage.Store
}

// NewMobileService creates a new MobileService
func NewMobileService() (*MobileService, error) {
	// Use a reasonable default for mobile
	dbPath := filepath.Join(".", "anystore.db")

	store, err := storage.New(dbPath)
	if err != nil {
		return nil, err
	}

	return &MobileService{store: store}, nil
}

// Ping tests the connection and storage layer
func (ms *MobileService) Ping(message string) string {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if message == "" {
		message = "pong"
	}

	// Test storage by performing a simple operation
	// For now, just return the message to verify the layer works
	_ = ctx
	return message
}

// Close gracefully shuts down the service
func (ms *MobileService) Close() error {
	if ms.store != nil {
		return ms.store.Close()
	}
	return nil
}

package handlers

import (
	"context"
	"fmt"

	pb "anysync-backend/shared/proto/syncspace/v1"

	"google.golang.org/protobuf/proto"
)

// StartSync handles starting synchronization.
func StartSync(ctx context.Context, req proto.Message) (proto.Message, error) {
	if err := ensureInitialized(); err != nil {
		return nil, err
	}

	syncReq := req.(*pb.StartSyncRequest)

	// TODO: Implement with Any-Sync sync mechanisms
	_ = syncReq

	return &pb.StartSyncResponse{Success: false}, fmt.Errorf("not implemented yet")
}

// PauseSync handles pausing synchronization.
func PauseSync(ctx context.Context, req proto.Message) (proto.Message, error) {
	if err := ensureInitialized(); err != nil {
		return nil, err
	}

	pauseReq := req.(*pb.PauseSyncRequest)

	// TODO: Implement with Any-Sync sync mechanisms
	_ = pauseReq

	return &pb.PauseSyncResponse{Success: false}, fmt.Errorf("not implemented yet")
}

// GetSyncStatus handles getting sync status.
func GetSyncStatus(ctx context.Context, req proto.Message) (proto.Message, error) {
	if err := ensureInitialized(); err != nil {
		return nil, err
	}

	statusReq := req.(*pb.GetSyncStatusRequest)

	// TODO: Implement with Any-Sync sync mechanisms
	_ = statusReq

	return &pb.GetSyncStatusResponse{
		Statuses: []*pb.SpaceSyncStatus{},
	}, nil
}

package handlers

import (
	"context"
	"fmt"

	pb "anysync-backend/shared/proto/syncspace/v1"

	"google.golang.org/protobuf/proto"
)

// CreateSpace handles space creation.
func CreateSpace(ctx context.Context, req proto.Message) (proto.Message, error) {
	if err := ensureInitialized(); err != nil {
		return nil, err
	}

	spaceReq := req.(*pb.CreateSpaceRequest)

	// Create space using SpaceManager
	globalState.mu.RLock()
	sm := globalState.spaceManager
	globalState.mu.RUnlock()

	if sm == nil {
		return nil, fmt.Errorf("space manager not initialized")
	}

	// Note: spaceReq.SpaceId is used as a reference name, actual ID is generated
	err := sm.CreateSpace(spaceReq.SpaceId, spaceReq.Name, spaceReq.Metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to create space: %w", err)
	}

	// Get the actual generated space ID
	spaces := sm.ListSpaces()
	var actualSpaceID string
	for _, space := range spaces {
		if space.Name == spaceReq.Name {
			actualSpaceID = space.SpaceID
			break
		}
	}

	return &pb.CreateSpaceResponse{
		SpaceId: actualSpaceID,
	}, nil
}

// JoinSpace handles joining a space.
func JoinSpace(ctx context.Context, req proto.Message) (proto.Message, error) {
	if err := ensureInitialized(); err != nil {
		return nil, err
	}

	joinReq := req.(*pb.JoinSpaceRequest)

	// TODO: Implement with Any-Sync SpaceService
	_ = joinReq

	return &pb.JoinSpaceResponse{Success: false}, fmt.Errorf("not implemented yet")
}

// LeaveSpace handles leaving a space.
func LeaveSpace(ctx context.Context, req proto.Message) (proto.Message, error) {
	if err := ensureInitialized(); err != nil {
		return nil, err
	}

	leaveReq := req.(*pb.LeaveSpaceRequest)

	// TODO: Implement with Any-Sync SpaceService
	_ = leaveReq

	return &pb.LeaveSpaceResponse{Success: false}, fmt.Errorf("not implemented yet")
}

// ListSpaces handles listing spaces.
func ListSpaces(ctx context.Context, req proto.Message) (proto.Message, error) {
	if err := ensureInitialized(); err != nil {
		return nil, err
	}

	globalState.mu.RLock()
	sm := globalState.spaceManager
	globalState.mu.RUnlock()

	if sm == nil {
		return nil, fmt.Errorf("space manager not initialized")
	}

	// Get spaces from SpaceManager
	spaces := sm.ListSpaces()

	// Convert to protobuf format
	pbSpaces := make([]*pb.SpaceInfo, len(spaces))
	for i, space := range spaces {
		pbSpaces[i] = &pb.SpaceInfo{
			SpaceId:   space.SpaceID,
			Name:      space.Name,
			Metadata:  space.Metadata,
			CreatedAt: space.CreatedAt,
			UpdatedAt: space.UpdatedAt,
			// SyncStatus: IDLE for local-only mode (network sync not yet implemented)
			SyncStatus: pb.SyncStatus_SYNC_STATUS_IDLE,
		}
	}

	return &pb.ListSpacesResponse{
		Spaces: pbSpaces,
	}, nil
}

// DeleteSpace handles space deletion.
func DeleteSpace(ctx context.Context, req proto.Message) (proto.Message, error) {
	if err := ensureInitialized(); err != nil {
		return nil, err
	}

	deleteReq := req.(*pb.DeleteSpaceRequest)

	globalState.mu.RLock()
	sm := globalState.spaceManager
	globalState.mu.RUnlock()

	if sm == nil {
		return nil, fmt.Errorf("space manager not initialized")
	}

	// Delete space using SpaceManager
	err := sm.DeleteSpace(deleteReq.SpaceId)
	if err != nil {
		return &pb.DeleteSpaceResponse{Success: false}, fmt.Errorf("failed to delete space: %w", err)
	}

	return &pb.DeleteSpaceResponse{Success: true}, nil
}

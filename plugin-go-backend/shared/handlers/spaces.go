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

	// TODO: Implement with Any-Sync SpaceService
	_ = spaceReq

	return nil, fmt.Errorf("not implemented yet")
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

	// TODO: Implement with Any-Sync SpaceService

	return &pb.ListSpacesResponse{
		Spaces: []*pb.SpaceInfo{},
	}, nil
}

// DeleteSpace handles space deletion.
func DeleteSpace(ctx context.Context, req proto.Message) (proto.Message, error) {
	if err := ensureInitialized(); err != nil {
		return nil, err
	}

	deleteReq := req.(*pb.DeleteSpaceRequest)

	// TODO: Implement with Any-Sync SpaceService
	_ = deleteReq

	return &pb.DeleteSpaceResponse{Success: false}, fmt.Errorf("not implemented yet")
}

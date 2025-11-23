package server

import (
	"context"

	pb "anysync-backend/api/proto"
	"anysync-backend/internal/storage"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// StorageServer implements the StorageService gRPC service
type StorageServer struct {
	pb.UnimplementedStorageServiceServer
	store *storage.Store
}

// NewStorageServer creates a new StorageServer
func NewStorageServer(store *storage.Store) *StorageServer {
	return &StorageServer{
		store: store,
	}
}

// Put stores a document in a collection
func (s *StorageServer) Put(ctx context.Context, req *pb.PutRequest) (*pb.PutResponse, error) {
	if req.Collection == "" {
		return nil, status.Error(codes.InvalidArgument, "collection name is required")
	}
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "document ID is required")
	}
	if req.DocumentJson == "" {
		return nil, status.Error(codes.InvalidArgument, "document JSON is required")
	}

	err := s.store.Put(ctx, req.Collection, req.Id, req.DocumentJson)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to store document: %v", err)
	}

	return &pb.PutResponse{Success: true}, nil
}

// Get retrieves a document from a collection
func (s *StorageServer) Get(ctx context.Context, req *pb.GetRequest) (*pb.GetResponse, error) {
	if req.Collection == "" {
		return nil, status.Error(codes.InvalidArgument, "collection name is required")
	}
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "document ID is required")
	}

	documentJSON, err := s.store.Get(ctx, req.Collection, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get document: %v", err)
	}

	found := documentJSON != ""
	return &pb.GetResponse{
		DocumentJson: documentJSON,
		Found:        found,
	}, nil
}

// Delete removes a document from a collection
// This operation is idempotent - deleting a non-existent document succeeds
func (s *StorageServer) Delete(ctx context.Context, req *pb.DeleteRequest) (*pb.DeleteResponse, error) {
	if req.Collection == "" {
		return nil, status.Error(codes.InvalidArgument, "collection name is required")
	}
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "document ID is required")
	}

	existed, err := s.store.Delete(ctx, req.Collection, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete document: %v", err)
	}

	return &pb.DeleteResponse{Existed: existed}, nil
}

// List returns all document IDs in a collection
func (s *StorageServer) List(ctx context.Context, req *pb.ListRequest) (*pb.ListResponse, error) {
	if req.Collection == "" {
		return nil, status.Error(codes.InvalidArgument, "collection name is required")
	}

	ids, err := s.store.List(ctx, req.Collection)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list documents: %v", err)
	}

	return &pb.ListResponse{Ids: ids}, nil
}

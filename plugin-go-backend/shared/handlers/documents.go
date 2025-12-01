package handlers

import (
	"context"
	"fmt"

	pb "anysync-backend/shared/proto/syncspace/v1"
	"google.golang.org/protobuf/proto"
)

// CreateDocument handles document creation.
func CreateDocument(ctx context.Context, req proto.Message) (proto.Message, error) {
	if err := ensureInitialized(); err != nil {
		return nil, err
	}

	docReq := req.(*pb.CreateDocumentRequest)

	// TODO: Implement with Any-Sync ObjectTree
	_ = docReq

	return nil, fmt.Errorf("not implemented yet")
}

// GetDocument retrieves a document.
func GetDocument(ctx context.Context, req proto.Message) (proto.Message, error) {
	if err := ensureInitialized(); err != nil {
		return nil, err
	}

	getReq := req.(*pb.GetDocumentRequest)

	// TODO: Implement with Any-Sync ObjectTree
	_ = getReq

	return &pb.GetDocumentResponse{
		Found: false,
	}, nil
}

// UpdateDocument updates a document.
func UpdateDocument(ctx context.Context, req proto.Message) (proto.Message, error) {
	if err := ensureInitialized(); err != nil {
		return nil, err
	}

	updateReq := req.(*pb.UpdateDocumentRequest)

	// TODO: Implement with Any-Sync ObjectTree
	_ = updateReq

	return nil, fmt.Errorf("not implemented yet")
}

// DeleteDocument deletes a document.
func DeleteDocument(ctx context.Context, req proto.Message) (proto.Message, error) {
	if err := ensureInitialized(); err != nil {
		return nil, err
	}

	deleteReq := req.(*pb.DeleteDocumentRequest)

	// TODO: Implement with Any-Sync ObjectTree
	_ = deleteReq

	return &pb.DeleteDocumentResponse{
		Existed: false,
	}, nil
}

// ListDocuments lists documents.
func ListDocuments(ctx context.Context, req proto.Message) (proto.Message, error) {
	if err := ensureInitialized(); err != nil {
		return nil, err
	}

	listReq := req.(*pb.ListDocumentsRequest)

	// TODO: Implement with Any-Sync ObjectTree
	_ = listReq

	return &pb.ListDocumentsResponse{
		Documents: []*pb.DocumentInfo{},
	}, nil
}

// QueryDocuments queries documents.
func QueryDocuments(ctx context.Context, req proto.Message) (proto.Message, error) {
	if err := ensureInitialized(); err != nil {
		return nil, err
	}

	queryReq := req.(*pb.QueryDocumentsRequest)

	// TODO: Implement with Any-Sync ObjectTree
	_ = queryReq

	return &pb.QueryDocumentsResponse{
		Documents: []*pb.DocumentInfo{},
	}, nil
}

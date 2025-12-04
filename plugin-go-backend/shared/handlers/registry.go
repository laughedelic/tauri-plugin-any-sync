package handlers

import (
	"anysync-backend/shared/dispatcher"

	pb "anysync-backend/shared/proto/syncspace/v1"
)

// RegisterAll registers all handlers with the dispatcher.
func RegisterAll(d *dispatcher.Dispatcher) {
	// Lifecycle - PascalCase to match protobuf service method names
	d.Register("Init", Init, &pb.InitRequest{})
	d.Register("Shutdown", Shutdown, &pb.ShutdownRequest{})

	// Spaces
	d.Register("CreateSpace", CreateSpace, &pb.CreateSpaceRequest{})
	d.Register("JoinSpace", JoinSpace, &pb.JoinSpaceRequest{})
	d.Register("LeaveSpace", LeaveSpace, &pb.LeaveSpaceRequest{})
	d.Register("ListSpaces", ListSpaces, &pb.ListSpacesRequest{})
	d.Register("DeleteSpace", DeleteSpace, &pb.DeleteSpaceRequest{})

	// Documents
	d.Register("CreateDocument", CreateDocument, &pb.CreateDocumentRequest{})
	d.Register("GetDocument", GetDocument, &pb.GetDocumentRequest{})
	d.Register("UpdateDocument", UpdateDocument, &pb.UpdateDocumentRequest{})
	d.Register("DeleteDocument", DeleteDocument, &pb.DeleteDocumentRequest{})
	d.Register("ListDocuments", ListDocuments, &pb.ListDocumentsRequest{})
	d.Register("QueryDocuments", QueryDocuments, &pb.QueryDocumentsRequest{})

	// Sync
	d.Register("StartSync", StartSync, &pb.StartSyncRequest{})
	d.Register("PauseSync", PauseSync, &pb.PauseSyncRequest{})
	d.Register("GetSyncStatus", GetSyncStatus, &pb.GetSyncStatusRequest{})
}

// GetDispatcher creates and returns a dispatcher with all handlers registered.
func GetDispatcher() *dispatcher.Dispatcher {
	d := dispatcher.New()
	RegisterAll(d)
	return d
}

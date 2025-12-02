package handlers

import (
	"anysync-backend/shared/dispatcher"

	pb "anysync-backend/shared/proto/syncspace/v1"
)

// RegisterAll registers all handlers with the dispatcher.
func RegisterAll(d *dispatcher.Dispatcher) {
	// Lifecycle
	d.Register("init", Init, &pb.InitRequest{})
	d.Register("shutdown", Shutdown, &pb.ShutdownRequest{})

	// Spaces
	d.Register("createSpace", CreateSpace, &pb.CreateSpaceRequest{})
	d.Register("joinSpace", JoinSpace, &pb.JoinSpaceRequest{})
	d.Register("leaveSpace", LeaveSpace, &pb.LeaveSpaceRequest{})
	d.Register("listSpaces", ListSpaces, &pb.ListSpacesRequest{})
	d.Register("deleteSpace", DeleteSpace, &pb.DeleteSpaceRequest{})

	// Documents
	d.Register("createDocument", CreateDocument, &pb.CreateDocumentRequest{})
	d.Register("getDocument", GetDocument, &pb.GetDocumentRequest{})
	d.Register("updateDocument", UpdateDocument, &pb.UpdateDocumentRequest{})
	d.Register("deleteDocument", DeleteDocument, &pb.DeleteDocumentRequest{})
	d.Register("listDocuments", ListDocuments, &pb.ListDocumentsRequest{})
	d.Register("queryDocuments", QueryDocuments, &pb.QueryDocumentsRequest{})

	// Sync
	d.Register("startSync", StartSync, &pb.StartSyncRequest{})
	d.Register("pauseSync", PauseSync, &pb.PauseSyncRequest{})
	d.Register("getSyncStatus", GetSyncStatus, &pb.GetSyncStatusRequest{})
}

// GetDispatcher creates and returns a dispatcher with all handlers registered.
func GetDispatcher() *dispatcher.Dispatcher {
	d := dispatcher.New()
	RegisterAll(d)
	return d
}

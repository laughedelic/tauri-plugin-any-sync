module anysync-backend/desktop

go 1.25.0

require (
	anysync-backend/shared v0.0.0
	google.golang.org/grpc v1.77.0
	google.golang.org/protobuf v1.36.10
)

replace anysync-backend/shared => ../shared

require (
	golang.org/x/net v0.47.0 // indirect
	golang.org/x/sys v0.38.0 // indirect
	golang.org/x/text v0.31.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20251022142026-3a174f9686a8 // indirect
)

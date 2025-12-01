module anysync-backend/mobile

go 1.25.0

require (
	anysync-backend/shared v0.0.0
	golang.org/x/mobile v0.0.0-20251126181937-5c265dc024c4
)

replace anysync-backend/shared => ../shared

require (
	golang.org/x/mod v0.30.0 // indirect
	golang.org/x/sync v0.18.0 // indirect
	golang.org/x/tools v0.39.0 // indirect
	google.golang.org/protobuf v1.36.10 // indirect
)

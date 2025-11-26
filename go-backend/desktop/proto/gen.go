package proto

//go:generate sh -c "cd .. && protoc --go_out=. --go-grpc_out=. --go_opt=paths=source_relative --go-grpc_opt=paths=source_relative proto/health.proto proto/storage.proto"

// Package proto contains generated protobuf message and service definitions.

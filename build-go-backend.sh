#!/bin/bash

# Build script for Go backend with cross-compilation support
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if Go is installed
check_go() {
    if ! command -v go &> /dev/null; then
        print_error "Go is not installed or not in PATH"
        exit 1
    fi
    
    GO_VERSION=$(go version | awk '{print $3}')
    print_status "Found Go version: $GO_VERSION"
}

# Check if protoc is installed
check_protoc() {
    if ! command -v protoc &> /dev/null; then
        print_error "protoc is not installed or not in PATH"
        exit 1
    fi
    
    PROTOC_VERSION=$(protoc --version)
    print_status "Found protoc version: $PROTOC_VERSION"
}

# Check if protobuf Go plugins are installed
check_protoc_plugins() {
    export PATH=$PATH:$(go env GOPATH)/bin
    
    if ! command -v protoc-gen-go &> /dev/null; then
        print_warning "protoc-gen-go not found, installing..."
        go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
    fi
    
    if ! command -v protoc-gen-go-grpc &> /dev/null; then
        print_warning "protoc-gen-go-grpc not found, installing..."
        go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
    fi
}

# Generate protobuf code
generate_proto() {
    print_status "Generating protobuf code..."
    cd go-backend
    
    # Add Go bin to PATH for protoc plugins
    export PATH=$PATH:$(go env GOPATH)/bin
    
    # Generate protobuf code and check for errors
    if ! protoc --go_out=. --go-grpc_out=. api/proto/*.proto 2>&1; then
        cd ..
        print_error "Failed to generate protobuf code"
        exit 1
    fi
    
    cd ..
    print_status "Protobuf code generated successfully"
}

# Build for a specific target
build_target() {
    local target=$1
    local output_name=${2:-any-sync-$1}
    
    print_status "Building for target: $target"
    
    cd go-backend
    
    # Set GOOS and GOARCH for cross-compilation
    case $target in
        "x86_64-apple-darwin")
            export GOOS=darwin GOARCH=amd64
            ;;
        "aarch64-apple-darwin")
            export GOOS=darwin GOARCH=arm64
            ;;
        "x86_64-unknown-linux-gnu")
            export GOOS=linux GOARCH=amd64
            ;;
        "aarch64-unknown-linux-gnu")
            export GOOS=linux GOARCH=arm64
            ;;
        "x86_64-pc-windows-msvc")
            export GOOS=windows GOARCH=amd64
            output_name="${output_name}.exe"
            ;;
        *)
            print_error "Unknown target: $target"
            return 1
            ;;
    esac
    
    # Create output directory
    mkdir -p ../binaries
    
    # Build
    # Set ldflags to strip debug symbols and disable DWARF generation in CI builds for smaller binaries
    go build${CI:+ -ldflags "-s -w"} -o "../binaries/${output_name}" ./cmd/server
    
    # Reset environment
    unset GOOS GOARCH
    
    cd ..
    print_status "Built binary: binaries/${output_name}"
}

# Main build function
main() {
    print_status "Starting Go backend build process..."
    
    # Check dependencies
    check_go
    check_protoc
    check_protoc_plugins
    
    # Generate protobuf code
    generate_proto
    
    # Build for common platforms if cross-compilation is requested
    if [[ "$1" == "--cross" ]]; then
        print_status "Cross-compiling for multiple platforms..."
        rm -f binaries/*
        
        build_target "x86_64-apple-darwin"
        build_target "aarch64-apple-darwin"
        build_target "x86_64-unknown-linux-gnu"
        build_target "aarch64-unknown-linux-gnu"
        build_target "x86_64-pc-windows-msvc"
    elif [[ -n "$1" ]]; then
        print_status "Building for specified target: $1"
        build_target "$1"
    else
        HOST_TARGET=$(rustc -Vv | grep host | cut -f2 -d' ')
        print_status "Building for host platform: $HOST_TARGET"
        build_target "$HOST_TARGET"
    fi
    
    print_status "Build process completed successfully!"
    
    # List built binaries
    print_status "Built binaries:"
    ls -la binaries/
}

# Show usage
usage() {
    echo "Usage: $0 [--cross]"
    echo "  --cross    Build for all supported platforms"
    echo ""
    echo "Supported platforms:"
    echo "  - x86_64-apple-darwin"
    echo "  - aarch64-apple-darwin"
    echo "  - x86_64-unknown-linux-gnu"
    echo "  - aarch64-unknown-linux-gnu"
    echo "  - x86_64-pc-windows-msvc"
}

# Parse command line arguments
case "$1" in
    "--help"|"-h")
        usage
        exit 0
        ;;
    *)
        main "$@"
        ;;
esac
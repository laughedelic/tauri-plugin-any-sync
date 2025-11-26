#!/usr/bin/env bash

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
GO_BACKEND_DIR="${SCRIPT_DIR}/go-backend"
BINARIES_DIR="${SCRIPT_DIR}/binaries"

# Output file name
AAR_NAME="any-sync-android.aar"

echo -e "${GREEN}Building Android .aar with gomobile${NC}"

# Check if gomobile is installed
check_gomobile() {
    if ! command -v gomobile &> /dev/null; then
        echo -e "${RED}Error: gomobile is not installed${NC}"
        echo "Install it with: go install golang.org/x/mobile/cmd/gomobile@latest"
        echo "Then run: gomobile init"
        exit 1
    fi
    echo -e "${GREEN}✓ gomobile found${NC}"
}

# Check if Android NDK is available
check_ndk() {
    if [ -n "${ANDROID_HOME:-}" ]; then
        echo -e "${GREEN}✓ ANDROID_HOME is set: ${ANDROID_HOME}${NC}"
    elif [ -d "$HOME/Library/Android/sdk" ]; then
        export ANDROID_HOME="$HOME/Library/Android/sdk"
        echo -e "${YELLOW}Setting ANDROID_HOME to: ${ANDROID_HOME}${NC}"
    else
        echo -e "${YELLOW}Warning: ANDROID_HOME not set and default path not found${NC}"
        echo "Android NDK may not be available"
    fi
}

# Build Android .aar
build_android_aar() {
    echo -e "${GREEN}Building Android .aar...${NC}"
    
    cd "${GO_BACKEND_DIR}/mobile"
    
    # Build for Android with API level 21 (minimum supported by NDK)
    gomobile bind \
        -target=android \
        -androidapi=21 \
        -o "${BINARIES_DIR}/${AAR_NAME}" \
        .
    
    cd "${SCRIPT_DIR}"
    
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✓ Successfully built ${AAR_NAME}${NC}"
    else
        echo -e "${RED}Error: Failed to build ${AAR_NAME}${NC}"
        exit 1
    fi
}

# Generate SHA256 checksum
generate_checksum() {
    echo -e "${GREEN}Generating SHA256 checksum...${NC}"
    
    cd "${BINARIES_DIR}"
    
    if command -v sha256sum &> /dev/null; then
        sha256sum "${AAR_NAME}" > "${AAR_NAME}.sha256"
    elif command -v shasum &> /dev/null; then
        shasum -a 256 "${AAR_NAME}" > "${AAR_NAME}.sha256"
    else
        echo -e "${RED}Error: Neither sha256sum nor shasum found${NC}"
        exit 1
    fi
    
    echo -e "${GREEN}✓ Checksum generated${NC}"
    cat "${AAR_NAME}.sha256"
}

# Get file size in human-readable format
get_file_size() {
    local file="$1"
    if [[ "$OSTYPE" == "darwin"* ]]; then
        stat -f "%z" "$file" | awk '{ 
            if ($1 > 1024*1024*1024) printf "%.2f GB\n", $1/1024/1024/1024
            else if ($1 > 1024*1024) printf "%.2f MB\n", $1/1024/1024
            else if ($1 > 1024) printf "%.2f KB\n", $1/1024
            else printf "%d B\n", $1
        }'
    else
        stat -c "%s" "$file" | awk '{ 
            if ($1 > 1024*1024*1024) printf "%.2f GB\n", $1/1024/1024/1024
            else if ($1 > 1024*1024) printf "%.2f MB\n", $1/1024/1024
            else if ($1 > 1024) printf "%.2f KB\n", $1/1024
            else printf "%d B\n", $1
        }'
    fi
}

# Inspect .aar contents
inspect_aar() {
    echo -e "${GREEN}Inspecting .aar contents...${NC}"
    
    cd "${BINARIES_DIR}"
    
    # Show file size
    local size=$(get_file_size "${AAR_NAME}")
    echo -e "${GREEN}File size: ${size}${NC}"
    
    # List contents (aar is just a zip file)
    if command -v unzip &> /dev/null; then
        echo -e "\n${GREEN}Contents:${NC}"
        unzip -l "${AAR_NAME}" | grep -E "(\.so|\.class|AndroidManifest\.xml)"
    else
        echo -e "${YELLOW}unzip not available, skipping inspection${NC}"
    fi
}

# Main execution
main() {
    echo "========================================"
    echo "  AnySync Android Build"
    echo "========================================"
    echo
    
    # Create binaries directory if it doesn't exist
    mkdir -p "${BINARIES_DIR}"
    
    # Run checks and build
    check_gomobile
    check_ndk
    build_android_aar
    generate_checksum
    inspect_aar
    
    echo
    echo -e "${GREEN}========================================"
    echo -e "  Build Complete!"
    echo -e "========================================${NC}"
    echo
    echo -e "Output: ${BINARIES_DIR}/${AAR_NAME}"
    echo -e "Checksum: ${BINARIES_DIR}/${AAR_NAME}.sha256"
    echo
}

# Run main function
main "$@"

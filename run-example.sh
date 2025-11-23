#!/bin/bash
# Script to run the example app with local binaries and debug logging

export ANY_SYNC_GO_BINARIES_DIR="$(pwd)/binaries"
export RUST_LOG=${1:-debug}

echo "RUST_LOG=$RUST_LOG"

cd examples/tauri-app
npm run tauri dev

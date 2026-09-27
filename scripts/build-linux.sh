#!/usr/bin/env bash
# Build gitagger for Linux (amd64 + arm64).
# Usage: sh scripts/build-linux.sh
set -eu
cd "$(dirname "$0")/.."
mkdir -p build
echo "==> Building gitagger for linux/amd64"
GOOS=linux GOARCH=amd64 go build -trimpath -o build/gitagger-linux-amd64 ./cmd/gitagger
echo "==> Building gitagger for linux/arm64"
GOOS=linux GOARCH=arm64 go build -trimpath -o build/gitagger-linux-arm64 ./cmd/gitagger
echo "OK:"
ls -la build/gitagger-linux-*

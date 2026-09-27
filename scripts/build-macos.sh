#!/usr/bin/env bash
# Build gitagger for macOS (Intel amd64 + Apple Silicon arm64).
# Usage: sh scripts/build-macos.sh
set -eu
cd "$(dirname "$0")/.."
mkdir -p dist
echo "==> Building gitagger for darwin/amd64 (Intel)"
GOOS=darwin GOARCH=amd64 go build -trimpath -o dist/gitagger-macos-amd64 ./cmd/gitagger
echo "==> Building gitagger for darwin/arm64 (Apple Silicon)"
GOOS=darwin GOARCH=arm64 go build -trimpath -o dist/gitagger-macos-arm64 ./cmd/gitagger
echo "OK:"
ls -la dist/gitagger-macos-*

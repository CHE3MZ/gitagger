#!/usr/bin/env bash
# Build gitagger for every release platform (linux, macOS, Windows; amd64 + arm64).
# Usage: sh scripts/build.sh
set -eu
cd "$(dirname "$0")/.."
mkdir -p build

build() {
  echo "==> Building gitagger for $1/$2"
  GOOS=$1 GOARCH=$2 CGO_ENABLED=0 go build -trimpath -o "build/gitagger-$1-$2$3" ./cmd/gitagger
}

build linux amd64 ""
build linux arm64 ""
build darwin amd64 ""
build darwin arm64 ""
build windows amd64 .exe
build windows arm64 .exe

echo "OK:"
ls -la build/gitagger-*

#!/usr/bin/env bash
# Single test automation: vet + full test suite.
# Usage: sh scripts/test.sh
# Mirrors CI (.github/workflows): `go test ./...` must pass before every commit.
set -eu
cd "$(dirname "$0")/.."
mkdir -p logs
echo "==> go vet ./..."
go vet ./... 2>&1 | tee logs/go-test-vet.txt
echo "==> go test ./... -v"
go test ./... -v 2>&1 | tee logs/go-test.txt
echo "All tests passed. Logs in logs/go-test.txt"

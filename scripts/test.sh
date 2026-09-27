#!/usr/bin/env bash
# Single test automation: vet + full test suite.
# Usage: sh scripts/test.sh
# Mirrors CI (.github/workflows) and plan.md 12.3: `go test ./...` must pass.
set -eu
cd "$(dirname "$0")/.."
mkdir -p logs
echo "==> go vet ./..."
go vet ./... 2>&1 | tee logs/go-test-vet.txt
echo "==> go test ./... -v"
go test ./... -v 2>&1 | tee logs/go-test.txt
echo "All tests passed. Logs in logs/go-test.txt"

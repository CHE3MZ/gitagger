#!/usr/bin/env bash

set -o pipefail

# Always run from the repo root regardless of where the script is invoked.
cd "$(dirname "$0")/../.." || exit

mkdir -p logs

failed=0

run_check() {
    local name="$1"
    shift

    echo "==> Running $name"

    if "$@" 2>&1 | tee "logs/go-bugcheck-${name}.txt"; then
        echo "    $name: OK"
    else
        echo "    $name: FAILED"
        failed=1
    fi

    echo
}

run_check golangci-lint golangci-lint run ./...

# Collect Go files without bash-only mapfile so this works under sh too.
set --
while IFS= read -r f; do
    set -- "$@" "$f"
done <<EOF
$(find . -name '*.go' -type f)
EOF

if [ "$#" -gt 0 ]; then
    run_check gopls gopls check "$@"
else
    echo "No Go files found; skipping gopls check."
fi

run_check govulncheck govulncheck ./...
run_check staticcheck staticcheck ./...
run_check errcheck errcheck ./...
run_check vet go vet ./...
run_check test go test ./...

exit "$failed"

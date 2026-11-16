#!/usr/bin/env sh
set -u

failed=0
sh scripts/ops/go-bugcheck.sh || failed=1
sh scripts/ops/go-seccheck.sh || failed=1
exit "$failed"

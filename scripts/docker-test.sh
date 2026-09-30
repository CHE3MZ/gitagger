#!/usr/bin/env sh
# Functional tests for the gitagger container image.
# Usage: sh scripts/docker-test.sh [image]   (default: gitagger:test)
# The image must already exist (CI builds it first with docker buildx).
# Needs docker + git on PATH. Temp repos only — never the repo's own .git.
# Push/remote paths are out of scope here (same binary the Go suite already
# covers); this proves the container-specific layer: mounts, users, shells,
# hooks, config, and that git + certs are present at runtime.
set -eu
cd "$(dirname "$0")/.."

IMAGE="${1:-gitagger:test}"
mkdir -p logs
LOG="logs/docker-test.txt"
: > "$LOG"
passed=0
failed=0

say() {
    printf '%s\n' "$1"
    printf '%s\n' "$1" >> "$LOG"
}
pass() {
    passed=$((passed + 1))
    say "PASS: $1"
}
fail() {
    failed=$((failed + 1))
    say "FAIL: $1"
}

need() {
    if ! command -v "$1" >/dev/null 2>&1; then
        say "missing $1 on PATH — cannot test containers"
        exit 1
    fi
}
need docker
need git
if ! docker info >/dev/null 2>&1; then
    say "docker daemon unreachable"
    exit 1
fi
if ! docker image inspect "$IMAGE" >/dev/null 2>&1; then
    say "image $IMAGE not found — build it first (docker build -t $IMAGE .)"
    exit 1
fi

# dk runs gitagger with the fixture repo mounted at /repo.
dk() {
    docker run --rm -v "$REPO:/repo" -w /repo "$IMAGE" "$@"
}

new_repo() {
    REPO="$(mktemp -d)"
    git -C "$REPO" init -q
    git -C "$REPO" config user.email "gitagger-docker@example.com"
    git -C "$REPO" config user.name "gitagger-docker"
    git -C "$REPO" config commit.gpgsign false
    echo hi > "$REPO/f.txt"
    git -C "$REPO" add .
    git -C "$REPO" commit -qm first
}

write_config() {
    printf '%s\n' "$1" > "$REPO/.gitagger.yml"
}

say "==> image: $IMAGE"

# 1: version smoke.
if out="$(docker run --rm "$IMAGE" version 2>&1)"; then
    case "$out" in *gitagger*) pass "version" ;; *) fail "version output: $out" ;; esac
else
    fail "version exits nonzero"
fi

# 2: help lists the real commands (image matches these sources).
if out="$(docker run --rm "$IMAGE" --help 2>&1)"; then
    case "$out" in *view*remove*) pass "help" ;; *) fail "help output: $out" ;; esac
else
    fail "help exits nonzero"
fi

# 3: fresh repo roundtrip — container writes the tag into the mount.
new_repo
if dk --no-push >/dev/null 2>&1; then
    if [ "$(git -C "$REPO" tag --list)" = "v1.0.0" ]; then
        pass "fresh tag v1.0.0"
    else
        fail "tags after run: $(git -C "$REPO" tag --list)"
    fi
else
    fail "fresh tag run exits nonzero"
fi

# 4: list + view see it.
if out="$(dk list 2>&1)"; then
    case "$out" in *v1.0.0*) pass "list" ;; *) fail "list output: $out" ;; esac
else
    fail "list exits nonzero"
fi
if out="$(dk view v1.0.0 2>&1)"; then
    case "$out" in *lightweight*) pass "view lightweight" ;; *) fail "view output: $out" ;; esac
else
    fail "view exits nonzero"
fi

# 5: annotated tag message survives the container.
new_repo
if dk --no-push -m "hello tag" >/dev/null 2>&1; then
    if out="$(dk view v1.0.0 2>&1)"; then
        case "$out" in *hello\ tag*) pass "annotated message" ;; *) fail "view output: $out" ;; esac
    else
        fail "annotated view exits nonzero"
    fi
else
    fail "annotated tag run exits nonzero"
fi

# 6: interactive confirm path (piped y deletes).
if printf 'y\n' | docker run --rm -i -v "$REPO:/repo" -w /repo "$IMAGE" remove v1.0.0 >/dev/null 2>&1; then
    if [ -z "$(git -C "$REPO" tag --list)" ]; then
        pass "prompted remove"
    else
        fail "tag still present after y"
    fi
else
    fail "prompted remove exits nonzero"
fi

# 7: -c removes without asking.
new_repo
git -C "$REPO" tag v1.0.0
if dk remove v1.0.0 -c >/dev/null 2>&1; then
    if [ -z "$(git -C "$REPO" tag --list)" ]; then
        pass "confirm remove"
    else
        fail "tag still present after -c"
    fi
else
    fail "confirm remove exits nonzero"
fi

# 8: missing tag exits 1.
if dk view v9.9.9 >/dev/null 2>&1; then
    fail "missing tag should fail"
else
    code=$?
    if [ "$code" -eq 1 ]; then pass "missing tag exit 1"; else fail "missing tag exit $code"; fi
fi

# 9: typo suggests (exit 3).
if out="$(dk hlp 2>&1)"; then
    fail "hlp should fail"
else
    code=$?
    if [ "$code" -eq 3 ]; then
        case "$out" in *"most similar"*) pass "did-you-mean" ;; *) fail "hlp output: $out" ;; esac
    else
        fail "hlp exit $code"
    fi
fi

# 10: config is respected (pre flavor).
new_repo
write_config "pre: rc"
if dk --no-push >/dev/null 2>&1; then
    if [ "$(git -C "$REPO" tag --list)" = "v1.0.0-rc" ]; then
        pass "config pre rc"
    else
        fail "tags after run: $(git -C "$REPO" tag --list)"
    fi
else
    fail "config run exits nonzero"
fi

# 11: success hooks run on the container shell.
new_repo
write_config 'push: false
on:
  success:
    - run: echo hi > hook-ran'
if dk >/dev/null 2>&1; then
    if [ -f "$REPO/hook-ran" ]; then
        pass "success hook"
    else
        fail "hook marker missing"
    fi
else
    fail "hook run exits nonzero"
fi

# 12: check validates the config.
if dk check -v >/dev/null 2>&1; then
    pass "check"
else
    fail "check exits nonzero"
fi

# 13: arbitrary UID works (mounted repos are owned by someone else).
new_repo
if dk --no-push >/dev/null 2>&1; then
    if out="$(docker run --rm --user 1001:1001 -v "$REPO:/repo" -w /repo "$IMAGE" list 2>&1)"; then
        case "$out" in *v1.0.0*) pass "arbitrary uid" ;; *) fail "uid list output: $out" ;; esac
    else
        fail "arbitrary uid exits nonzero"
    fi
else
    fail "uid setup run exits nonzero"
fi

say ""
say "docker tests: $passed passed, $failed failed. Log in $LOG"
if [ "$failed" -ne 0 ]; then
    exit 1
fi

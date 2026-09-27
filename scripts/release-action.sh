#!/usr/bin/env bash
# Move floating marketplace tags (vX, vX.Y) to a release tag.
# Usage: sh scripts/release-action.sh [tag]   (default: latest tag)
# Only stable vX.Y.Z tags move floats — prereleases are skipped by design.
# Safe to re-run: refuses to move a float backwards.
set -eu
cd "$(dirname "$0")/.."

tag="${1:-$(git describe --tags --abbrev=0)}"
git rev-parse --verify --quiet "refs/tags/$tag^{commit}" >/dev/null || {
  echo "error: no such tag: $tag" >&2
  exit 1
}
case "$tag" in
  *-*) echo "error: $tag looks like a prerelease — floats stay where they are" >&2; exit 1 ;;
esac
if ! printf '%s' "$tag" | grep -Eq '^v[0-9]+\.[0-9]+\.[0-9]+$'; then
  echo "error: $tag is not vX.Y.Z — floats need semver releases" >&2
  exit 1
fi

major=${tag%%.*}
minor=${tag%.*}

move_float() {
  float=$1
  if git ls-remote --exit-code origin "refs/tags/$float" >/dev/null 2>&1; then
    git fetch -q origin "tag $float" 2>/dev/null || true
    if ! git merge-base --is-ancestor "$float" "$tag" 2>/dev/null; then
      echo "error: refusing to move $float backwards ($tag is not ahead)" >&2
      exit 1
    fi
    echo "$float: $(git rev-parse --short "$float") -> $(git rev-parse --short "$tag")"
  else
    echo "$float: (new) -> $(git rev-parse --short "$tag")"
  fi
  git tag -f "$float" "$tag"
  git push --force --quiet origin "$float"
  echo "moved $float to $tag"
}

move_float "$major"
move_float "$minor"
echo "marketplace pointers now follow $tag"

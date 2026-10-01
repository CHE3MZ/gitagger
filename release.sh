#!/usr/bin/env bash
# Fail fast: a broken build or upload must never move the marketplace tags.
set -eu

# Correctly capture the latest git tag into the variable
echo "Getting the latest tag..."
taglatest=$(git describe --tags --abbrev=0)

# Run GoReleaser skipping SCM validation and publishing
echo "Running Goreleaser..."
goreleaser release --clean --skip=publish

# Changelog for the release notes: commits since the previous tag.
# Needs git-cliff (https://git-cliff.org, see cliff.toml); falls back
# to a static line when it is not installed. This runs by hand and on
# the release workflow runner (via the success hook in .gitagger.yml),
# so git-cliff must exist in both places — see the install step in
# .github/workflows/gitagger.yml.
notes_file="$(mktemp)"
if command -v git-cliff >/dev/null 2>&1; then
  echo "Generating changelog..."
  prev_tag="$(git describe --tags --abbrev=0 "$taglatest^" 2>/dev/null || true)"
  if [ -n "$prev_tag" ]; then
    git cliff "$prev_tag..$taglatest" > "$notes_file"
  else
    git cliff --tag "$taglatest" > "$notes_file"
  fi
else
  echo "git-cliff not found — using static release notes."
  printf '%s\n' "This release was generated with GoReleaser!" > "$notes_file"
fi

# Create the release and upload whatever files exist in the dist folder
echo "Uploading release binaries to github via the gh CLI..."
gh release create "$taglatest" \
  ./dist/*.tar.gz \
  ./dist/*.zip \
  ./dist/*.txt \
  ./dist/metadata.json \
  --title "$taglatest" \
  --notes-file "$notes_file"
rm -f "$notes_file"

# Point the marketplace float tags (vX, vX.Y) at this release.
echo "Moving marketplace tags..."
sh scripts/release-action.sh "$taglatest"

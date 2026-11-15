#!/usr/bin/env bash
# Generates Formula/gitagger.rb for the CHE3MZ/homebrew-tools tap.
# Usage: scripts/homebrew-formula.sh <vX.Y.Z tag> <checksums.txt> <output.rb>
set -euo pipefail

if [ "$#" -ne 3 ]; then
  echo "usage: $0 <vX.Y.Z tag> <checksums.txt> <output.rb>" >&2
  exit 2
fi

tag=$1
checksums=$2
output=$3

if [[ ! "$tag" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  echo "refusing Homebrew formula for non-stable tag: $tag" >&2
  exit 1
fi

if [ ! -f "$checksums" ]; then
  echo "checksums file not found: $checksums" >&2
  exit 1
fi

sha_for() {
  local file=$1 sha
  sha=$(tr -d '\r' < "$checksums" | awk -v file="$file" '$2 == file { print $1; exit }')
  if [[ ! "$sha" =~ ^[0-9a-f]{64}$ ]]; then
    echo "missing or invalid checksum for $file" >&2
    exit 1
  fi
  printf '%s' "$sha"
}

darwin_arm64=$(sha_for gitagger_Darwin_arm64.tar.gz)
darwin_x86_64=$(sha_for gitagger_Darwin_x86_64.tar.gz)
linux_arm64=$(sha_for gitagger_Linux_arm64.tar.gz)
linux_x86_64=$(sha_for gitagger_Linux_x86_64.tar.gz)

version=${tag#v}
mkdir -p "$(dirname "$output")"

cat > "$output" <<EOF
class Gitagger < Formula
  desc "Create and push Git tags for your repository"
  homepage "https://che3mz.github.io/gitagger/"
  version "$version"
  license "Apache-2.0"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/CHE3MZ/gitagger/releases/download/$tag/gitagger_Darwin_arm64.tar.gz"
      sha256 "$darwin_arm64"
    else
      url "https://github.com/CHE3MZ/gitagger/releases/download/$tag/gitagger_Darwin_x86_64.tar.gz"
      sha256 "$darwin_x86_64"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/CHE3MZ/gitagger/releases/download/$tag/gitagger_Linux_arm64.tar.gz"
      sha256 "$linux_arm64"
    else
      url "https://github.com/CHE3MZ/gitagger/releases/download/$tag/gitagger_Linux_x86_64.tar.gz"
      sha256 "$linux_x86_64"
    end
  end

  livecheck do
    url :stable
    strategy :github_latest
  end

  def install
    bin.install "gitagger"
  end

  test do
    assert_match version.to_s, shell_output("#{bin}/gitagger version")
  end
end
EOF

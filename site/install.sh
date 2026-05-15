#!/bin/sh
# loby installer. Autodetects OS + arch, downloads the latest release from
# GitHub, verifies the checksum, and installs to a directory on $PATH.
#
# Usage:
#   curl -fsSL https://loby.voska.org/install.sh | sh
#   curl -fsSL https://loby.voska.org/install.sh | LOBY_VERSION=v1.0.0 sh
#   curl -fsSL https://loby.voska.org/install.sh | LOBY_PREFIX=$HOME/.local sh
#
# Override defaults via env vars:
#   LOBY_VERSION   tag to install (default: latest)
#   LOBY_PREFIX    install prefix (default: /usr/local; falls back to $HOME/.local)
#   LOBY_BIN_DIR   bin directory (default: $LOBY_PREFIX/bin)
set -eu

REPO="voska/loby"
BINARY="loby"
VERSION="${LOBY_VERSION:-latest}"

# Detect OS / arch
os=$(uname -s | tr '[:upper:]' '[:lower:]')
case "$os" in
  linux|darwin) ;;
  msys*|mingw*|cygwin*) os="windows" ;;
  *) echo "loby: unsupported OS: $os" >&2; exit 1 ;;
esac
arch=$(uname -m)
case "$arch" in
  x86_64|amd64) arch="x86_64" ;;
  arm64|aarch64) arch="arm64" ;;
  *) echo "loby: unsupported architecture: $arch" >&2; exit 1 ;;
esac

# Resolve tag
if [ "$VERSION" = "latest" ]; then
  tag=$(curl -fsSL "https://api.github.com/repos/$REPO/releases/latest" \
    | sed -n 's/.*"tag_name": *"\([^"]*\)".*/\1/p' | head -n1)
else
  tag="$VERSION"
fi
[ -n "${tag:-}" ] || { echo "loby: failed to resolve release tag" >&2; exit 1; }
ver="${tag#v}"

# Title-case the OS for the archive name (goreleaser default)
case "$os" in
  linux)   os_title="Linux" ;;
  darwin)  os_title="Darwin" ;;
  windows) os_title="Windows" ;;
esac
ext="tar.gz"
[ "$os" = "windows" ] && ext="zip"
archive="${BINARY}_${ver}_${os_title}_${arch}.${ext}"
url="https://github.com/$REPO/releases/download/$tag/$archive"
sums_url="https://github.com/$REPO/releases/download/$tag/checksums.txt"

# Choose install dir
prefix="${LOBY_PREFIX:-/usr/local}"
if [ ! -w "$prefix" ] && [ "$(id -u)" -ne 0 ]; then
  prefix="$HOME/.local"
fi
bin_dir="${LOBY_BIN_DIR:-$prefix/bin}"
mkdir -p "$bin_dir"

# Download into a tempdir
tmp=$(mktemp -d 2>/dev/null || mktemp -d -t lobyinstall)
trap 'rm -rf "$tmp"' EXIT
printf '%s\n' "Downloading $archive ($tag) → $bin_dir"
curl -fsSL -o "$tmp/$archive" "$url"

# Verify checksum if available
if curl -fsSL -o "$tmp/checksums.txt" "$sums_url" 2>/dev/null; then
  expected=$(grep " $archive\$" "$tmp/checksums.txt" | awk '{print $1}')
  if [ -n "$expected" ]; then
    actual=$(shasum -a 256 "$tmp/$archive" 2>/dev/null | awk '{print $1}' \
      || sha256sum "$tmp/$archive" | awk '{print $1}')
    if [ "$expected" != "$actual" ]; then
      echo "loby: checksum mismatch" >&2
      exit 1
    fi
  fi
fi

# Extract + install
cd "$tmp"
if [ "$ext" = "tar.gz" ]; then
  tar -xzf "$archive"
else
  unzip -q "$archive"
fi
binname="$BINARY"
[ "$os" = "windows" ] && binname="${BINARY}.exe"
install -m 0755 "$binname" "$bin_dir/$binname" 2>/dev/null \
  || cp "$binname" "$bin_dir/$binname"
chmod 0755 "$bin_dir/$binname"

printf 'Installed %s to %s/%s\n' "$BINARY" "$bin_dir" "$binname"
case ":$PATH:" in
  *":$bin_dir:"*) ;;
  *) printf 'Add %s to your PATH: export PATH="%s:$PATH"\n' "$bin_dir" "$bin_dir" ;;
esac
"$bin_dir/$binname" version 2>/dev/null || true

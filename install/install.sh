#!/usr/bin/env sh
set -eu

repo="${JLINK_CLI_REPO:-guajun/jlink-cli}"
install_dir="${JLINK_CLI_INSTALL_DIR:-$HOME/.local/bin}"
api_url="https://api.github.com/repos/$repo/releases/latest"

os="$(uname -s | tr '[:upper:]' '[:lower:]')"
case "$os" in
  linux) os="linux" ;;
  *) echo "jlink-cli installer supports Linux. Detected: $os" >&2; exit 1 ;;
esac

machine="$(uname -m)"
case "$machine" in
  x86_64|amd64) arch="amd64" ;;
  aarch64|arm64) arch="arm64" ;;
  *) echo "Unsupported architecture: $machine" >&2; exit 1 ;;
esac

require() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "Missing required command: $1" >&2
    exit 1
  fi
}

require curl
require tar

if command -v python3 >/dev/null 2>&1; then
  release_json="$(curl -fsSL "$api_url")"
  asset_url="$(printf '%s' "$release_json" | python3 -c 'import json, re, sys; data=json.load(sys.stdin); pattern=re.compile(r"linux[_-]" + sys.argv[1] + r"\.tar\.gz$", re.IGNORECASE); matches=[asset["browser_download_url"] for asset in data.get("assets", []) if pattern.search(asset.get("name", ""))]; print(matches[0] if matches else "")' "$arch")"
  tag="$(printf '%s' "$release_json" | python3 -c 'import json, sys; print(json.load(sys.stdin).get("tag_name", "latest"))')"
else
  require grep
  require sed
  release_json="$(curl -fsSL "$api_url")"
  asset_url="$(printf '%s' "$release_json" | grep 'browser_download_url' | grep -Ei "linux[_-]${arch}\.tar\.gz" | sed -E 's/.*"browser_download_url": "([^"]+)".*/\1/' | head -n 1)"
  tag="$(printf '%s' "$release_json" | grep '"tag_name"' | sed -E 's/.*"tag_name": "([^"]+)".*/\1/' | head -n 1)"
fi

if [ -z "$asset_url" ]; then
  echo "No Linux $arch release asset found for $repo." >&2
  exit 1
fi

tmp_dir="$(mktemp -d)"
cleanup() {
  rm -rf "$tmp_dir"
}
trap cleanup EXIT INT TERM

archive="$tmp_dir/jlink-cli.tar.gz"
curl -fsSL "$asset_url" -o "$archive"
tar -xzf "$archive" -C "$tmp_dir"

binary="$(find "$tmp_dir" -type f -name jlink-cli | head -n 1)"
if [ -z "$binary" ]; then
  echo "Downloaded archive did not contain jlink-cli." >&2
  exit 1
fi

mkdir -p "$install_dir"
install -m 0755 "$binary" "$install_dir/jlink-cli"
echo "Installed jlink-cli ${tag:-latest} to $install_dir"

case ":$PATH:" in
  *":$install_dir:"*) ;;
  *) echo "Add $install_dir to PATH, or run $install_dir/jlink-cli directly." ;;
esac

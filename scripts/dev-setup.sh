#!/usr/bin/env bash
set -euo pipefail

# Dev setup: fetch rclone into ./bin/rclone, prepare ./data, print status.
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")"/.. && pwd)"
BIN_DIR="$ROOT_DIR/bin"
DATA_DIR="$ROOT_DIR/data"
mkdir -p "$BIN_DIR" "$DATA_DIR"

# Detect OS/ARCH → rclone download suffix
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
case "$OS" in
  linux) os_part=linux ;;
  darwin) os_part=osx ;;  # rclone uses 'osx' archives
  msys*|cygwin*|mingw*) os_part=windows ;;
  *) os_part=linux ;;
esac
case "$ARCH" in
  x86_64|amd64) arch_part=amd64 ;;
  aarch64|arm64) arch_part=arm64 ;;
  armv7l) arch_part=arm ;;
  *) arch_part=amd64 ;;
esac

# Map darwin naming used by rclone: rclone-<ver>-osx-<arch>.zip
if [[ "$os_part" == "osx" ]]; then
  archive_suffix="osx-${arch_part}"
else
  archive_suffix="${os_part}-${arch_part}"
fi

# Prefer current release URL
URL="https://downloads.rclone.org/rclone-current-${archive_suffix}.zip"
TMP_ZIP="$(mktemp -t rclone.zip.XXXXXX)"

echo "[dev-setup] Downloading rclone from: $URL"
curl -fsSL "$URL" -o "$TMP_ZIP"
TMP_DIR="$(mktemp -d -t rclone.XXXXXX)"
unzip -q "$TMP_ZIP" -d "$TMP_DIR"
# Find rclone binary inside extracted dir
RCLONE_BIN_PATH=$(find "$TMP_DIR" -type f -name rclone -o -name rclone.exe | head -n1)
if [[ -z "${RCLONE_BIN_PATH:-}" ]]; then
  echo "[dev-setup] rclone binary not found in archive" >&2
  exit 1
fi
# Place to ./bin/rclone (or rclone.exe on Windows Git Bash)
TARGET="$BIN_DIR/rclone"
cp "$RCLONE_BIN_PATH" "$TARGET"
chmod +x "$TARGET" || true
rm -f "$TMP_ZIP" && rm -rf "$TMP_DIR"

# Ensure a writable config path exists
RCLONE_CONF="$DATA_DIR/rclone.conf"
if [[ ! -f "$RCLONE_CONF" ]]; then
  echo "{}" > "$RCLONE_CONF" || true
fi

echo "[dev-setup] Done. Binaries & paths:"
echo "  rclone: $TARGET"
echo "  data dir: $DATA_DIR"
echo "  rclone.conf: $RCLONE_CONF"
echo "[dev-setup] You can now run: make run"

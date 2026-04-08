#!/bin/sh
set -e

DATA_DIR="${APP_DATA_DIR:-/app/data}"
mkdir -p "$DATA_DIR"

# Best-effort chown (handles first run and host-mounted dirs)
chown -R 1000:1000 "$DATA_DIR" 2>/dev/null || true

# Check writability; if not writable (e.g., SELinux or root-owned mount), fallback to root
if touch "$DATA_DIR/.writable_test" 2>/dev/null; then
  rm -f "$DATA_DIR/.writable_test"
  echo "[entrypoint] Using appuser(1000) with writable data dir: $DATA_DIR"
  exec su-exec appuser:appuser /app/server
else
  echo "[entrypoint] Warning: $DATA_DIR not writable by uid 1000. Running as root."
  echo "[entrypoint] Suggest: chown -R 1000:1000 <host-data-dir> or use :Z (SELinux) or set user: '1000:1000' in compose."
  exec /app/server
fi

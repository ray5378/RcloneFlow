#!/bin/sh
set -e

# Ensure data dir exists and is owned by appuser(1000)
mkdir -p /app/data
# Best-effort chown (handles first run and host-mounted dirs)
chown -R 1000:1000 /app/data 2>/dev/null || true

# Run as appuser (uid:1000)
exec su-exec appuser:appuser /app/server

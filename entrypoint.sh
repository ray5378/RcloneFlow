#!/bin/sh
PUID=${PUID:-1000}
PGID=${PGID:-1000}

addgroup -g "$PGID" appuser 2>/dev/null || true
adduser -D -u "$PUID" -G appuser appuser 2>/dev/null || true

chown -R appuser:appuser /app/data 2>/dev/null || true

exec su-exec appuser:appuser "$@"
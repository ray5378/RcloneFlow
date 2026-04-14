# RcloneFlow

RcloneFlow = Rclone + Web UI + Scheduler + History.

Primary goals: stable progress for active runs, JWT-protected APIs, frozen history model. On this branch, directory browsing uses RC; file operations use CLI.

## Key Concepts
- Stable progress (preparing/transferring/between_files/finalizing) during active runs
- Frozen history: backend writes `summary.finalSummary` when a run finishes; history renders only this finalized data
- JWT auth: required for log download and all APIs

## Frozen History (summary.finalSummary)
A one-shot final summary is written to DB when a run finishes (success/failure). The history view renders the finalized snapshot only.

Fields:
- startAt / finishedAt (RFC3339)
- durationSec / durationText
- result (success/failed)
- transferredBytes / totalBytes / avgSpeedBps
- counts { copied, deleted, skipped, failed, total }
- files: array of { path, action, status, sizeBytes?, at? } (retention-based cleanup, not truncated by default)

## APIs
- GET /api/runs — list with durationSeconds/durationText and summary.finalSummary
- GET /api/runs/{id} — single run
- GET /api/runs/active — active-only stable progress (not stored)
- GET /api/runs/{id}/files — raw/parsed log (history UI does not depend on it)

## Retention & Cleanup
- FINAL_SUMMARY_RETENTION_DAYS — keep runs/finalSummary (default 7)
- CLEANUP_INTERVAL_HOURS — cleanup frequency (default 24h)
- Log retention is separate (if configured)

## Settings (/api/settings + APP_DATA_DIR/settings.json)
Precedence: Env > settings.json > built-in defaults. Managed via UI “Settings → Defaults”.

Common keys:
- ACCESS_TOKEN_TTL (e.g., 24h)
- REFRESH_TOKEN_TTL (e.g., 90d)
- PRECHECK_MODE: none | size
- PROGRESS_FLUSH_INTERVAL (e.g., 5s)
- PROGRESS_FLUSH_MIN_DELTA_PCT (0-100; decimals ok)
- PROGRESS_FLUSH_MIN_DELTA_BYTES (bytes; UI shows MB)
- FINISH_WAIT_INTERVAL (e.g., 5s)
- FINISH_WAIT_TIMEOUT (e.g., 5h)
- FINAL_SUMMARY_RETENTION_DAYS / CLEANUP_INTERVAL_HOURS
- WEBHOOK_MAX_FILES: 0 = unlimited (default)

### Webhook
- Payload includes task/run/summary and file names (up to N; 0 = all)
- WeCom + Generic Markdown supported

## Build & Run
- docker build --no-cache -t ray5378/rcloneflow:next-master -f rcloneflow/Dockerfile rcloneflow
- Vite outDir = ../web; final image serves /app/web
- Do NOT mount over /app/web (you’ll get stale frontend). The runtime image is shell-less; for inspection use `docker create+cp` to extract /app/web.

## Docker Compose Example
```yaml
version: '3.8'
services:
  rcloneflow:
    image: ray5378/rcloneflow:next-master
    container_name: rcloneflow
    ports:
      - "17871:17871"
      - "17872:17872"
      - "17873:17873"
    environment:
      - APP_ADDR=:17871
      - APP_DATA_DIR=/app/data
      - LOG_LEVEL=info
      - PRECHECK_MODE=none
      - PROGRESS_FLUSH_INTERVAL=5s
      - PROGRESS_FLUSH_MIN_DELTA_PCT=1
      - PROGRESS_FLUSH_MIN_DELTA_BYTES=52428800
      - FINISH_WAIT_INTERVAL=5s
      - FINISH_WAIT_TIMEOUT=5h
      - WEBHOOK_MAX_FILES=0
    volumes:
      - ./data:/app/data
    restart: unless-stopped
```

## Frontend Behavior
- History renders only finalSummary (start/end/duration/result/avg/bytes/files)
- Active runs render stable progress (phase/percentage/speed) without persisting realtime
- Dark theme progress bar uses high-contrast white fill; after 100% the list refreshes after 20s (two passes) and a 25s stuck-frame watchdog triggers a refresh

## Browsing & File Ops
- Browsing: RC (ListPath) — more stable for SMB/WebDAV
- File ops: CLI (/api/fs/* via rclone CLI) — copy/move/rename/delete/mkdir/public-link
  - WebDAV MOVE fallback: copy(+dir)+delete/purge with visibility/gone waits
  - SMB share dedupe on specific errors; sanitize only trailing ASCII colons per path segment
- Storage management: RC (/api/remotes, /api/config, /api/providers, /api/usage/{fs}, /api/fsinfo/{fs})

Notes:
- Browsing = RC; File ops = CLI (fixed). No RC/CLI toggles.

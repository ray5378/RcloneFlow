# RcloneFlow

RcloneFlow = Rclone + Web UI + Scheduler + History.

This fork focuses on stable progress, JWT-protected APIs, and a frozen history model.

## Key Concepts
- Stable progress (preparing/transferring/between_files/finalizing) for active runs
- Frozen history: summary.finalSummary is persisted when a run finishes
- JWT auth for log download and APIs

## Frozen History (summary.finalSummary)
When a run finishes (success or failure), the backend writes a one-shot final summary to DB. Frontend history renders only this finalized data.

Schema:
- startAt (RFC3339)
- finishedAt (RFC3339)
- durationSec (number)
- durationText (string, e.g., "1小时2分3秒")
- result ("success" | "failed")
- transferredBytes (number)
- totalBytes (number)
- avgSpeedBps (number)
- counts { copied, deleted, skipped, failed, total }
- files: full array of { path, action, status, sizeBytes?, at? }

Notes:
- files is full (no default cap). Retention is time-based.
- Original transfer log remains downloadable for raw inspection.

## APIs
- GET /api/runs → list; each entry includes durationSeconds, durationText and summary.finalSummary
- GET /api/runs/{id} → single run with the same fields
- GET /api/runs/active → realtime only (not stored); runRecord includes durationSeconds, durationText; stableProgress describes phase/percentage/speed
- GET /api/runs/{id}/files → raw/parsed log for legacy inspection (history UI no longer depends on it)

## Retention
- FINAL_SUMMARY_RETENTION_DAYS (env): days to keep runs + finalSummary (default 7)
- CLEANUP_INTERVAL_HOURS (env): cleanup frequency hours (default 24)
- Log retention is separate: LOG_RETENTION_DAYS / LOG_CLEANUP_INTERVAL_HOURS

## Build
- docker build --no-cache -t ray5378/rcloneflow:next-master -f rcloneflow/Dockerfile rcloneflow

## Frontend Behavior
- History pages render only finalSummary: start/end, frozen duration, result, average speed, transferred/total, file details.
- During an active run, the UI shows a lightweight realtime banner (phase/percentage/speed); nothing from realtime is stored to DB.

See docs/FINAL_SUMMARY.md for details.

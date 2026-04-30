# Run Final Summary (Frozen History)

> 注：当前仓库已进入一轮阶段性收口状态。若后续对运行历史链路继续调整，应优先做契约补充或验收报点修复，不再做无边界扩改。

RcloneFlow persists a frozen run summary to DB when a run finishes (success or failure). Frontend history pages render only this finalized data — no realtime fields are stored.

## Schema (summary.finalSummary)
- startAt: RFC3339
- finishedAt: RFC3339
- durationSec: number (seconds)
- durationText: string (e.g., "1小时2分3秒")
- result: "success" | "failed"
- transferredBytes: number
- totalBytes: number
- avgSpeedBps: number (transferredBytes / durationSec; 0 when durationSec == 0)
- counts: { copied, deleted, skipped, failed, total }
- files: Array<{
  - path: string
  - action: "Copied" | "Deleted" | "Skipped" | "Error"
  - status: "success" | "skipped" | "failed"
  - sizeBytes?: number (best-effort; filled from file-progress lines)
  - at?: string (timestamp from log line when present)
}>

Notes:
- files is full (no default cap). Use FINAL_SUMMARY_RETENTION_DAYS to control retention.
- Original transfer log remains downloadable for raw inspection.

## API Contracts
- GET /api/runs → each run includes durationSeconds, durationText (frozen for finished runs) and summary.finalSummary.
- GET /api/runs/{id} → same as above for a single run.
- GET /api/runs/active → runRecord also includes computed durationSeconds, durationText (from now - startedAt); realtime only, not stored.

## Retention
- FINAL_SUMMARY_RETENTION_DAYS (env) controls how many days run history is kept. Default: 7.
- CLEANUP_INTERVAL_HOURS (env) controls cleanup frequency. Default: 24.
- When run history is cleaned, the corresponding rclone history log files are removed together.

## Frontend Behavior
- History views render only summary.finalSummary.
- "Run Detail" shows: start/end, frozen duration, result, average speed, transferred/total, and file details.
- During an active run, the frontend prefers `/api/runs/active.progress` for live UI. Backend log parsing writes live progress into `summary.progress`; history/detail views use `summary.finalSummary`.

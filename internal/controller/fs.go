package controller

// Deprecated: File-system operations have been migrated to CLI-only (see fs_cli.go).
// This placeholder file documents the deprecation and intentionally contains no handlers.
//
// Notes
// - Directory listing (browser) stays on RC in browser.go
// - Storage management (add/edit/delete remotes, config import/export, usage/fsinfo) stays on RC
// - Do NOT add /api/fs/* routes here; all FS ops must go through fs_cli.go (rclone CLI)
// - RC-based FS helpers in internal/rclone/client.go are kept only for diagnostics/future fallback and are not wired to routes

// Package cli provides a CLI-based rclone runner (scaffold).
// Final goal: replace RC with local rclone subprocess controlled by the backend.
// This scaffold is intentionally minimal and safe to land; implementation follows in subsequent commits.
package cli

// Placeholder types to outline structure without affecting current build paths.
// Real implementations will be added incrementally.

type Runner struct{}

type StartOptions struct{
	// TODO: map frontend TaskOptions -> rclone CLI flags here
}

type RunHandle struct{
	RunID int64
	PID   int
}

func NewRunner() *Runner { return &Runner{} }

// Start will be implemented to spawn a rclone subprocess with safe args.
func (r *Runner) Start(opts StartOptions) (*RunHandle, error) { return nil, nil }

// Stop will be implemented to gracefully terminate the subprocess (INT->TERM->KILL).
func (r *Runner) Stop(handle *RunHandle) error { return nil }

// TODO: progress parser (stats json / one-line), log rolling, and derived progress feed.

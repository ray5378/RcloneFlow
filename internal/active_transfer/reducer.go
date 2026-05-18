package active_transfer

func (m *Manager) OnFileProgress(runID int64, path string, bytes, total, speed int64, pct *float64) {
	m.UpdateCurrentFile(runID, path, bytes, total, speed, pct)
}

func (m *Manager) OnFileCopied(runID int64, path string) {
	m.MarkCompleted(runID, path, FileStatusCopied, "")
}

func (m *Manager) OnFileCASMatched(runID int64, path string) {
	m.MarkCompleted(runID, path, FileStatusCASMatched, "")
}

func (m *Manager) OnFileFailed(runID int64, path string, message string) {
	m.MarkCompleted(runID, path, FileStatusFailed, message)
}

func (m *Manager) OnFileSkipped(runID int64, path string, message string) {
	m.MarkCompleted(runID, path, FileStatusSkipped, message)
}

func (m *Manager) OnFileDeleted(runID int64, path string) {
	m.MarkCompleted(runID, path, FileStatusDeleted, "")
}

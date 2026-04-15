package service

import (
	"rcloneflow/internal/store"
)

type storeRunAdapter struct {
	db *store.DB
}

func newStoreRunAdapter(db *store.DB) *storeRunAdapter {
	return &storeRunAdapter{db: db}
}

func (a *storeRunAdapter) ListRuns(page, pageSize int) ([]RunRecord, int, error) {
	return a.db.ListRuns(page, pageSize)
}

func (a *storeRunAdapter) ListRunsByTask(taskId int64) ([]RunRecord, error) {
	return a.db.ListRunsByTask(taskId)
}

func (a *storeRunAdapter) ListActiveRuns() ([]RunRecord, error) {
	return a.db.ListActiveRuns()
}

func (a *storeRunAdapter) GetActiveRunByTaskID(taskID int64) (RunRecord, error) {
	return a.db.GetActiveRunByTaskID(taskID)
}

func (a *storeRunAdapter) UpdateRun(id int64, updateFn func(*RunRecord)) {
	a.db.UpdateRun(id, updateFn)
}

func (a *storeRunAdapter) DeleteRun(id int64) error {
	return a.db.DeleteRun(id)
}

func (a *storeRunAdapter) DeleteAllRuns() error {
	return a.db.DeleteAllRuns()
}

func (a *storeRunAdapter) DeleteRunsByTask(taskId int64) error {
	return a.db.DeleteRunsByTask(taskId)
}

func (a *storeRunAdapter) CleanOldRuns(days int) (int64, error) {
	return a.db.CleanOldRuns(days)
}

func (a *storeRunAdapter) UpdateRunStatusByJobId(jobId int64, status, errorMsg string) error {
	return a.db.UpdateRunStatusByJobId(jobId, status, errorMsg)
}

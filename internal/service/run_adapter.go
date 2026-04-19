package service

import (
	"encoding/json"
	"time"

	"rcloneflow/internal/store"
)

type storeRunAdapter struct {
	db *store.DB
}

func NewStoreRunAdapter(db *store.DB) *storeRunAdapter {
	return &storeRunAdapter{db: db}
}

func formatTime(t time.Time) string {
	return t.Format(time.RFC3339)
}

func formatOptTime(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format(time.RFC3339)
}

func toRunRecord(r store.Run) RunRecord {
	summaryStr := ""
	if r.Summary != nil {
		bs, _ := json.Marshal(r.Summary)
		summaryStr = string(bs)
	}
	finAt := formatOptTime(r.FinishedAt)
	if finAt == "" && r.Summary != nil {
		if v, ok := r.Summary["finishedAt"].(string); ok && v != "" {
			finAt = v
		}
	}
	return RunRecord{
		ID:               r.ID,
		TaskID:           r.TaskID,
		Status:           r.Status,
		Trigger:          r.Trigger,
		StartedAt:        formatTime(r.CreatedAt),
		FinishedAt:       finAt,
		TaskName:         r.TaskName,
		TaskMode:         r.TaskMode,
		SourceRemote:     r.SourceRemote,
		SourcePath:       r.SourcePath,
		TargetRemote:     r.TargetRemote,
		TargetPath:       r.TargetPath,
		BytesTransferred: r.BytesTransferred,
		Speed:            r.Speed,
		Error:            r.Error,
		Summary:          summaryStr,
	}
}

func toRunRecords(runs []store.Run) []RunRecord {
	result := make([]RunRecord, len(runs))
	for i, r := range runs {
		result[i] = toRunRecord(r)
	}
	return result
}

func (a *storeRunAdapter) ListRuns(page, pageSize int) ([]RunRecord, int, error) {
	records, total, err := a.db.ListRuns(page, pageSize)
	return toRunRecords(records), total, err
}

func (a *storeRunAdapter) ListRunsByTask(taskId int64) ([]RunRecord, error) {
	records, err := a.db.ListRunsByTask(taskId)
	return toRunRecords(records), err
}

func (a *storeRunAdapter) ListActiveRuns() ([]RunRecord, error) {
	records, err := a.db.ListActiveRuns()
	return toRunRecords(records), err
}

func (a *storeRunAdapter) GetActiveRunByTaskID(taskID int64) (RunRecord, error) {
	r, err := a.db.GetActiveRunByTaskID(taskID)
	if err != nil {
		return RunRecord{}, err
	}
	return toRunRecord(r), nil
}

func (a *storeRunAdapter) UpdateRun(id int64, updateFn func(*RunRecord)) {
	a.db.UpdateRun(id, func(r *store.Run) {
		rec := toRunRecord(*r)
		updateFn(&rec)
		r.Status = rec.Status
		r.FinishedAt = nil
		if rec.FinishedAt != "" {
			if t, err := time.Parse("2006-01-02T15:04:05Z", rec.FinishedAt); err == nil {
				r.FinishedAt = &t
			}
		}
		r.Speed = rec.Speed
		r.BytesTransferred = rec.BytesTransferred
		r.Error = rec.Error
		if rec.Summary != "" {
			var summary map[string]any
			json.Unmarshal([]byte(rec.Summary), &summary)
			r.Summary = summary
		}
	})
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


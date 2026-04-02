package service

import (
	"encoding/json"

	"rcloneflow/internal/store"
)

// storeRunAdapter store.DB到RunServiceInterface的适配器
type storeRunAdapter struct {
	db *store.DB
}

// NewStoreRunAdapter 创建适配器
func NewStoreRunAdapter(db *store.DB) RunServiceInterface {
	return &storeRunAdapter{db: db}
}

// ListRuns 获取所有运行记录
func (a *storeRunAdapter) ListRuns() ([]RunRecord, error) {
	runs, err := a.db.ListRuns()
	if err != nil {
		return nil, err
	}
	result := make([]RunRecord, len(runs))
	for i, r := range runs {
		summaryStr := ""
		if r.Summary != nil {
			bs, _ := json.Marshal(r.Summary)
			summaryStr = string(bs)
		}
		result[i] = RunRecord{
			ID:         r.ID,
			TaskID:     r.TaskID,
			RcJobID:    r.RcJobID,
			Status:     r.Status,
			Trigger:    r.Trigger,
			StartedAt:  r.CreatedAt.Format("2006-01-02T15:04:05Z"),
			Summary:    summaryStr,
			Error:      r.Error,
		}
	}
	return result, nil
}

// ListActiveRuns 获取所有运行中的任务
func (a *storeRunAdapter) ListActiveRuns() ([]RunRecord, error) {
	runs, err := a.db.ListActiveRuns()
	if err != nil {
		return nil, err
	}
	result := make([]RunRecord, len(runs))
	for i, r := range runs {
		summaryStr := ""
		if r.Summary != nil {
			bs, _ := json.Marshal(r.Summary)
			summaryStr = string(bs)
		}
		result[i] = RunRecord{
			ID:         r.ID,
			TaskID:     r.TaskID,
			RcJobID:    r.RcJobID,
			Status:     r.Status,
			Trigger:    r.Trigger,
			StartedAt:  r.CreatedAt.Format("2006-01-02T15:04:05Z"),
			Summary:    summaryStr,
			Error:      r.Error,
		}
	}
	return result, nil
}

// UpdateRun 更新运行记录
func (a *storeRunAdapter) UpdateRun(id int64, updateFn func(*RunRecord)) {
	a.db.UpdateRun(id, func(r *store.Run) {
		record := &RunRecord{
			ID:         r.ID,
			TaskID:     r.TaskID,
			RcJobID:    r.RcJobID,
			Status:     r.Status,
			Trigger:    r.Trigger,
			StartedAt:  r.CreatedAt.Format("2006-01-02T15:04:05Z"),
			Error:      r.Error,
		}
		if r.Summary != nil {
			bs, _ := json.Marshal(r.Summary)
			record.Summary = string(bs)
		}
		updateFn(record)
		r.Status = record.Status
		if record.Summary != "" {
			var summary map[string]any
			json.Unmarshal([]byte(record.Summary), &summary)
			r.Summary = summary
		}
		r.Error = record.Error
	})
}

// DeleteRun 删除运行记录
func (a *storeRunAdapter) DeleteRun(id int64) error {
	return a.db.DeleteRun(id)
}

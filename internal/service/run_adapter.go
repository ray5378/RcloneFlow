package service

import (
	"encoding/json"
	"time"

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

// formatTime 格式化时间
func formatTime(t time.Time) string {
	return t.Format("2006-01-02T15:04:05Z")
}

// formatOptTime 格式化可选时间
func formatOptTime(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format("2006-01-02T15:04:05Z")
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
			ID:               r.ID,
			TaskID:           r.TaskID,
			RcJobID:          r.RcJobID,
			Status:           r.Status,
			Trigger:         r.Trigger,
			StartedAt:        formatTime(r.CreatedAt),
			FinishedAt:       formatOptTime(r.FinishedAt),
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
	return result, nil
}

func (a *storeRunAdapter) ListRunsByTask(taskId int64) ([]RunRecord, error) {
	runs, err := a.db.ListRunsByTask(taskId)
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
			ID:               r.ID,
			TaskID:           r.TaskID,
			RcJobID:          r.RcJobID,
			Status:           r.Status,
			Trigger:         r.Trigger,
			StartedAt:        formatTime(r.CreatedAt),
			FinishedAt:       formatOptTime(r.FinishedAt),
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
			ID:               r.ID,
			TaskID:           r.TaskID,
			RcJobID:          r.RcJobID,
			Status:           r.Status,
			Trigger:         r.Trigger,
			StartedAt:        formatTime(r.CreatedAt),
			FinishedAt:       formatOptTime(r.FinishedAt),
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
	return result, nil
}

// UpdateRun 更新运行记录
func (a *storeRunAdapter) UpdateRun(id int64, updateFn func(*RunRecord)) {
	a.db.UpdateRun(id, func(r *store.Run) {
		record := &RunRecord{
			ID:               r.ID,
			TaskID:           r.TaskID,
			RcJobID:          r.RcJobID,
			Status:           r.Status,
			Trigger:         r.Trigger,
			StartedAt:        formatTime(r.CreatedAt),
			FinishedAt:       formatOptTime(r.FinishedAt),
			TaskName:         r.TaskName,
			TaskMode:         r.TaskMode,
			SourceRemote:     r.SourceRemote,
			SourcePath:       r.SourcePath,
			TargetRemote:     r.TargetRemote,
			TargetPath:       r.TargetPath,
			BytesTransferred: r.BytesTransferred,
			Speed:            r.Speed,
			Error:            r.Error,
		}
		if r.Summary != nil {
			bs, _ := json.Marshal(r.Summary)
			record.Summary = string(bs)
		}
		updateFn(record)
		r.Status = record.Status
		r.FinishedAt = nil
		if record.FinishedAt != "" {
			t, _ := time.Parse("2006-01-02T15:04:05Z", record.FinishedAt)
			r.FinishedAt = &t
		}
		r.Speed = record.Speed
		r.BytesTransferred = record.BytesTransferred
		r.Error = record.Error
		if record.Summary != "" {
			var summary map[string]any
			json.Unmarshal([]byte(record.Summary), &summary)
			r.Summary = summary
		}
	})
}

// DeleteRun 删除运行记录
func (a *storeRunAdapter) DeleteRun(id int64) error {
	return a.db.DeleteRun(id)
}

func (a *storeRunAdapter) DeleteAllRuns() error {
	return a.db.DeleteAllRuns()
}

func (a *storeRunAdapter) DeleteRunsByTask(taskId int64) error {
	return a.db.DeleteRunsByTask(taskId)
}

// CleanOldRuns 删除指定天数之前的运行记录
func (a *storeRunAdapter) CleanOldRuns(days int) (int64, error) {
	return a.db.CleanOldRuns(days)
}

// UpdateRunProgress 更新运行进度（bytes和speed）
func (a *storeRunAdapter) UpdateRunProgress(id int64, bytesTransferred int64, speed string) error {
	return a.db.UpdateRunProgress(id, bytesTransferred, speed)
}

// UpdateRunStatusByJobId 根据 JobID 更新运行状态
func (a *storeRunAdapter) UpdateRunStatusByJobId(jobId int64, status, errorMsg string) error {
	return a.db.UpdateRunStatusByJobId(jobId, status, errorMsg)
}

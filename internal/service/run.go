package service

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"

	"go.uber.org/zap"
	"rcloneflow/internal/logger"
)

type RunRecord struct {
	ID               int64  `json:"id"`
	TaskID           int64  `json:"taskId"`
	Status           string `json:"status"`
	Trigger          string `json:"trigger"`
	StartedAt        string `json:"startedAt"`
	FinishedAt       string `json:"finishedAt,omitempty"`
	TaskName         string `json:"taskName,omitempty"`
	TaskMode         string `json:"taskMode,omitempty"`
	SourceRemote     string `json:"sourceRemote,omitempty"`
	SourcePath       string `json:"sourcePath,omitempty"`
	TargetRemote     string `json:"targetRemote,omitempty"`
	TargetPath       string `json:"targetPath,omitempty"`
	BytesTransferred int64  `json:"bytesTransferred,omitempty"`
	Speed            string `json:"speed,omitempty"`
	Error            string `json:"error,omitempty"`
	Summary          string `json:"summary,omitempty"`
}

type RunServiceInterface interface {
	ListRuns(page, pageSize int) ([]RunRecord, int, error)
	ListRunsByTask(taskId int64) ([]RunRecord, error)
	ListActiveRuns() ([]RunRecord, error)
	GetActiveRunByTaskID(taskID int64) (RunRecord, error)
	GetRun(id int64) (RunRecord, error)
	UpdateRun(id int64, updateFn func(*RunRecord))
	DeleteRun(id int64) error
	DeleteAllRuns() error
	DeleteRunsByTask(taskId int64) error
	DeleteRunsByIDs(ids []int64) error
	CleanOldRuns(days int) (int64, error)
	Vacuum() error
	ResetSequence(tableName string) error
}

type RunService struct {
	db RunServiceInterface
}

func NewRunService(db RunServiceInterface) *RunService {
	return &RunService{db: db}
}

func (s *RunService) ListRuns(page, pageSize int) ([]RunRecord, int, error) {
	return s.db.ListRuns(page, pageSize)
}

func (s *RunService) ListRunsByTask(taskId int64) ([]RunRecord, error) {
	return s.db.ListRunsByTask(taskId)
}

func (s *RunService) ListActiveRuns() ([]RunRecord, error) {
	return s.db.ListActiveRuns()
}

func (s *RunService) GetRun(id int64) (RunRecord, error) {
	return s.db.GetRun(id)
}

func (s *RunService) GetActiveRunByTaskID(taskID int64) (RunRecord, error) {
	return s.db.GetActiveRunByTaskID(taskID)
}

func (s *RunService) UpdateRunStatus(id int64, summary map[string]any) {
	if len(summary) == 0 {
		return
	}
	s.db.UpdateRun(id, func(r *RunRecord) {
		var old map[string]any
		if r.Summary != "" {
			if err := json.Unmarshal([]byte(r.Summary), &old); err != nil {
				logger.Error("unmarshal run summary", zap.Error(err))
				old = nil
			}
		}
		merged := deepMerge(old, summary)
		if bs, err := jsonMarshal(merged); err == nil {
			r.Summary = string(bs)
		}
		finished, _ := merged["finished"].(bool)
		success, _ := merged["success"].(bool)
		if finished {
			if success {
				r.Status = "finished"
				r.Error = ""
			} else {
				r.Status = "failed"
				if errMsg, ok := merged["error"].(string); ok {
					r.Error = errMsg
				}
			}
		}
	})
}

var jsonBufPool = sync.Pool{
	New: func() any {
		return &bytes.Buffer{}
	},
}

func jsonMarshal(v any) ([]byte, error) {
	buf := jsonBufPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer jsonBufPool.Put(buf)
	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err != nil {
		return nil, err
	}
	result := make([]byte, buf.Len()-1)
	copy(result, buf.Bytes())
	return result, nil
}

func deepMerge(a, b map[string]any) map[string]any {
	if a == nil {
		a = map[string]any{}
	}
	for k, v := range b {
		if vm, ok := v.(map[string]any); ok {
			if am, ok2 := a[k].(map[string]any); ok2 {
				deepMerge(am, vm)
			} else {
				a[k] = deepMerge(map[string]any{}, vm)
			}
		} else {
			a[k] = v
		}
	}
	return a
}

func (s *RunService) DeleteRun(id int64) error {
	run, err := s.db.GetRun(id)
	if err != nil {
		return err
	}
	cleanupRunLog(run)
	return s.db.DeleteRun(id)
}

func (s *RunService) DeleteAllRuns() error {
	page := 1
	pageSize := 500
	for {
		runs, total, err := s.db.ListRuns(page, pageSize)
		if err != nil {
			return err
		}
		for _, run := range runs {
			cleanupRunLog(run)
		}
		if len(runs) == 0 || page*pageSize >= total {
			break
		}
		page++
	}
	if err := s.db.DeleteAllRuns(); err != nil {
		return err
	}
	if err := s.db.ResetSequence("runs"); err != nil {
		return err
	}
	return s.Vacuum()
}

func (s *RunService) DeleteRunsByTask(taskId int64) error {
	runs, err := s.db.ListRunsByTask(taskId)
	if err != nil {
		return err
	}
	for _, run := range runs {
		cleanupRunLog(run)
	}
	return s.db.DeleteRunsByTask(taskId)
}

func cleanupRunLog(run RunRecord) {
	if run.Summary == "" {
		return
	}
	var summary map[string]any
	if err := json.Unmarshal([]byte(run.Summary), &summary); err != nil || summary == nil {
		return
	}
	p, _ := summary["stderrFile"].(string)
	if p == "" {
		return
	}
	_ = os.Remove(p)
	dir := filepath.Dir(p)
	if entries, err := os.ReadDir(dir); err == nil && len(entries) == 0 {
		_ = os.Remove(dir)
	}
}

func (s *RunService) CleanOldRuns(days int) (int64, error) {
	if days <= 0 {
		return 0, nil
	}
	cutoff := time.Now().AddDate(0, 0, -days)
	page := 1
	pageSize := 500
	deleted := int64(0)
	for {
		runs, total, err := s.db.ListRuns(page, pageSize)
		if err != nil {
			return deleted, err
		}
		if len(runs) == 0 {
			break
		}
		var idsToDelete []int64
		for _, run := range runs {
			startedAt, err := time.Parse(time.RFC3339, run.StartedAt)
			if err != nil {
				cleanupRunLog(run)
				idsToDelete = append(idsToDelete, run.ID)
				continue
			}
			if startedAt.Before(cutoff) {
				cleanupRunLog(run)
				idsToDelete = append(idsToDelete, run.ID)
			}
		}
		if len(idsToDelete) > 0 {
			if err := s.db.DeleteRunsByIDs(idsToDelete); err != nil {
				return deleted, err
			}
			deleted += int64(len(idsToDelete))
		}
		if page*pageSize >= total {
			break
		}
		page++
	}
	return deleted, nil
}

func (s *RunService) Vacuum() error {
	return s.db.Vacuum()
}

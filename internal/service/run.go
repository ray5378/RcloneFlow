package service

import "fmt"

// RunRecord 运行记录结构
type RunRecord struct {
	ID               int64  `json:"id"`
	TaskID           int64  `json:"taskId"`
	RcJobID          int64  `json:"rcJobId"`
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

// RunServiceInterface 运行记录服务接口
type RunServiceInterface interface {
	ListRuns() ([]RunRecord, error)
	ListActiveRuns() ([]RunRecord, error)
	UpdateRun(id int64, updateFn func(*RunRecord))
	DeleteRun(id int64) error
}

// RunService 运行记录服务层
type RunService struct {
	db RunServiceInterface
}

// NewRunService 创建运行记录服务
func NewRunService(db RunServiceInterface) *RunService {
	return &RunService{db: db}
}

// ListRuns 获取所有运行记录
func (s *RunService) ListRuns() ([]RunRecord, error) {
	return s.db.ListRuns()
}

// ListActiveRuns 获取所有运行中的任务
func (s *RunService) ListActiveRuns() ([]RunRecord, error) {
	return s.db.ListActiveRuns()
}

// UpdateRunStatus 更新运行状态
func (s *RunService) UpdateRunStatus(id int64, summary map[string]any) {
	s.db.UpdateRun(id, func(r *RunRecord) {
		r.Summary = fmt.Sprintf("%v", summary)
		if finished, ok := summary["finished"].(bool); ok && finished {
			r.Status = "finished"
		}
		if success, ok := summary["success"].(bool); ok && !success {
			r.Status = "failed"
		}
	})
}

func (s *RunService) DeleteRun(id int64) error {
	return s.db.DeleteRun(id)
}

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
	ListRunsByTask(taskId int64) ([]RunRecord, error)
	ListActiveRuns() ([]RunRecord, error)
	UpdateRun(id int64, updateFn func(*RunRecord))
	DeleteRun(id int64) error
	DeleteAllRuns() error
	DeleteRunsByTask(taskId int64) error
	CleanOldRuns(days int) (int64, error)
	UpdateRunStatusByJobId(jobId int64, status, errorMsg string) error
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

func (s *RunService) ListRunsByTask(taskId int64) ([]RunRecord, error) {
	return s.db.ListRunsByTask(taskId)
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

func (s *RunService) DeleteAllRuns() error {
	return s.db.DeleteAllRuns()
}

func (s *RunService) DeleteRunsByTask(taskId int64) error {
	return s.db.DeleteRunsByTask(taskId)
}

// CleanOldRuns 删除指定天数之前的运行记录，返回删除的记录数
func (s *RunService) CleanOldRuns(days int) (int64, error) {
	return s.db.CleanOldRuns(days)
}

// UpdateRunStatusByJobId 根据 JobID 更新运行状态
func (s *RunService) UpdateRunStatusByJobId(jobId int64, status, errorMsg string) error {
	return s.db.UpdateRunStatusByJobId(jobId, status, errorMsg)
}

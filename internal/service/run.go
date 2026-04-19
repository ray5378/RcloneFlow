package service

import "encoding/json"

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
	ListRuns(page, pageSize int) ([]RunRecord, int, error)
	ListRunsByTask(taskId int64) ([]RunRecord, error)
	ListActiveRuns() ([]RunRecord, error)
	GetActiveRunByTaskID(taskID int64) (RunRecord, error)
	UpdateRun(id int64, updateFn func(*RunRecord))
	DeleteRun(id int64) error
	DeleteAllRuns() error
	DeleteRunsByTask(taskId int64) error
	CleanOldRuns(days int) (int64, error)
}

// RunService 运行记录服务层
type RunService struct {
	db RunServiceInterface
}

// NewRunService 创建运行记录服务
func NewRunService(db RunServiceInterface) *RunService {
	return &RunService{db: db}
}

// ListRuns 获取所有运行记录（分页）
func (s *RunService) ListRuns(page, pageSize int) ([]RunRecord, int, error) {
	return s.db.ListRuns(page, pageSize)
}

func (s *RunService) ListRunsByTask(taskId int64) ([]RunRecord, error) {
	return s.db.ListRunsByTask(taskId)
}

// ListActiveRuns 获取所有运行中的任务
func (s *RunService) ListActiveRuns() ([]RunRecord, error) {
	return s.db.ListActiveRuns()
}

// GetActiveRunByTaskID 获取任务当前运行中的记录
func (s *RunService) GetActiveRunByTaskID(taskID int64) (RunRecord, error) {
	return s.db.GetActiveRunByTaskID(taskID)
}

// UpdateRunStatus 更新运行状态
func (s *RunService) UpdateRunStatus(id int64, summary map[string]any) {
	s.db.UpdateRun(id, func(r *RunRecord) {
		// 读取旧 summary
		var old map[string]any
		if r.Summary != "" {
			_ = json.Unmarshal([]byte(r.Summary), &old)
		}
		if old == nil {
			old = map[string]any{}
		}
		// 合并：src 覆盖 dst 的同名键；map 递归
		merged := deepMerge(old, summary)
		if bs, err := json.Marshal(merged); err == nil {
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

// deepMerge merges b into a (map[string]any); for nested maps it recurses.
func deepMerge(a, b map[string]any) map[string]any {
	if a == nil {
		a = map[string]any{}
	}
	for k, v := range b {
		if vm, ok := v.(map[string]any); ok {
			if am, ok2 := a[k].(map[string]any); ok2 {
				a[k] = deepMerge(am, vm)
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


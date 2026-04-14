package service

import (
	"rcloneflow/internal/store"
)

// ScheduleService 定时任务服务层
type ScheduleService struct {
	db *store.DB
}

// NewScheduleService 创建定时任务服务
func NewScheduleService(db *store.DB) *ScheduleService {
	return &ScheduleService{db: db}
}

// ListSchedules 获取所有定时任务
func (s *ScheduleService) ListSchedules() ([]store.Schedule, error) {
	return s.db.ListSchedules()
}

// CreateSchedule 创建定时任务
func (s *ScheduleService) CreateSchedule(taskID int64, spec string, enabled bool) (store.Schedule, error) {
	return s.db.AddSchedule(store.Schedule{
		TaskID:  taskID,
		Spec:    spec,
		Enabled: enabled,
	})
}

// UpdateSpec 更新定时规则
func (s *ScheduleService) UpdateSpec(id int64, spec string) error {
	return s.db.UpdateScheduleSpec(id, spec)
}

// DeleteSchedule 删除定时任务
func (s *ScheduleService) DeleteSchedule(id int64) error {
	return s.db.DeleteSchedule(id)
}

// SetScheduleEnabled 启用/禁用定时任务
func (s *ScheduleService) SetScheduleEnabled(id int64, enabled bool) error {
	return s.db.SetScheduleEnabled(id, enabled)
}

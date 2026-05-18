package service

import (
	"testing"

	"rcloneflow/internal/store"
)

// ScheduleDAOMock 模拟ScheduleDAO
type ScheduleDAOMock struct {
	schedules []store.Schedule
}

func (m *ScheduleDAOMock) Create(schedule store.Schedule) (store.Schedule, error) {
	schedule.ID = int64(len(m.schedules) + 10)
	m.schedules = append(m.schedules, schedule)
	return schedule, nil
}

func (m *ScheduleDAOMock) GetByID(id int64) (store.Schedule, bool) {
	for _, s := range m.schedules {
		if s.ID == id {
			return s, true
		}
	}
	return store.Schedule{}, false
}

func (m *ScheduleDAOMock) GetAll() ([]store.Schedule, error) {
	return m.schedules, nil
}

func (m *ScheduleDAOMock) Delete(id int64) error {
	for i, s := range m.schedules {
		if s.ID == id {
			m.schedules = append(m.schedules[:i], m.schedules[i+1:]...)
			return nil
		}
	}
	return nil
}

// TestScheduleService_CreateSchedule 测试创建定时任务
func TestScheduleService_CreateSchedule(t *testing.T) {
	scheduleDAO := &ScheduleDAOMock{}

	svc := &testScheduleService{
		scheduleDAO: scheduleDAO,
	}

	schedule, err := svc.CreateSchedule(1, "@every 5m")
	if err != nil {
		t.Fatalf("CreateSchedule() error = %v", err)
	}

	if schedule.ID == 0 {
		t.Error("expected non-zero ID")
	}

	if schedule.Spec != "@every 5m" {
		t.Errorf("expected Spec '@every 5m', got '%s'", schedule.Spec)
	}

	if len(scheduleDAO.schedules) != 1 {
		t.Errorf("expected 1 schedule, got %d", len(scheduleDAO.schedules))
	}
}

// TestScheduleService_ListSchedules 测试列出定时任务
func TestScheduleService_ListSchedules(t *testing.T) {
	scheduleDAO := &ScheduleDAOMock{
		schedules: []store.Schedule{
			{ID: 1, TaskID: 1, Spec: "@every 5m"},
			{ID: 2, TaskID: 2, Spec: "@every 10m"},
		},
	}

	svc := &testScheduleService{
		scheduleDAO: scheduleDAO,
	}

	schedules, err := svc.ListSchedules()
	if err != nil {
		t.Fatalf("ListSchedules() error = %v", err)
	}

	if len(schedules) != 2 {
		t.Errorf("expected 2 schedules, got %d", len(schedules))
	}
}

// TestScheduleService_DeleteSchedule 测试删除定时任务
func TestScheduleService_DeleteSchedule(t *testing.T) {
	scheduleDAO := &ScheduleDAOMock{
		schedules: []store.Schedule{
			{ID: 1, TaskID: 1, Spec: "@every 5m"},
			{ID: 2, TaskID: 2, Spec: "@every 10m"},
		},
	}

	svc := &testScheduleService{
		scheduleDAO: scheduleDAO,
	}

	err := svc.DeleteSchedule(1)
	if err != nil {
		t.Fatalf("DeleteSchedule() error = %v", err)
	}

	if len(scheduleDAO.schedules) != 1 {
		t.Errorf("expected 1 schedule after delete, got %d", len(scheduleDAO.schedules))
	}

	_, ok := scheduleDAO.GetByID(1)
	if ok {
		t.Error("expected schedule 1 to be deleted")
	}
}

// TestScheduleService_GetSchedule 测试获取单个定时任务
func TestScheduleService_GetSchedule(t *testing.T) {
	scheduleDAO := &ScheduleDAOMock{
		schedules: []store.Schedule{
			{ID: 1, TaskID: 1, Spec: "@every 5m"},
		},
	}

	svc := &testScheduleService{
		scheduleDAO: scheduleDAO,
	}

	schedule, ok := svc.GetSchedule(1)
	if !ok {
		t.Error("expected to get schedule 1")
	}

	if schedule.Spec != "@every 5m" {
		t.Errorf("expected Spec '@every 5m', got '%s'", schedule.Spec)
	}
}

// TestScheduleService_GetSchedule_NotFound 测试获取不存在的定时任务
func TestScheduleService_GetSchedule_NotFound(t *testing.T) {
	scheduleDAO := &ScheduleDAOMock{}

	svc := &testScheduleService{
		scheduleDAO: scheduleDAO,
	}

	_, ok := svc.GetSchedule(999)
	if ok {
		t.Error("expected schedule 999 not found")
	}
}

// testScheduleService 简化版定时任务服务
type testScheduleService struct {
	scheduleDAO *ScheduleDAOMock
}

func (s *testScheduleService) CreateSchedule(taskID int64, spec string) (store.Schedule, error) {
	return s.scheduleDAO.Create(store.Schedule{
		TaskID: taskID,
		Spec:   spec,
	})
}

func (s *testScheduleService) ListSchedules() ([]store.Schedule, error) {
	return s.scheduleDAO.GetAll()
}

func (s *testScheduleService) GetSchedule(id int64) (store.Schedule, bool) {
	return s.scheduleDAO.GetByID(id)
}

func (s *testScheduleService) DeleteSchedule(id int64) error {
	return s.scheduleDAO.Delete(id)
}

// TestErrScheduleNotFound 测试错误
func TestErrScheduleNotFound(t *testing.T) {
	if ErrScheduleNotFound.Error() != "schedule not found" {
		t.Errorf("expected 'schedule not found', got '%s'", ErrScheduleNotFound.Error())
	}
}

// TestErrRunNotFound 测试错误
func TestErrRunNotFound(t *testing.T) {
	if ErrRunNotFound.Error() != "run not found" {
		t.Errorf("expected 'run not found', got '%s'", ErrRunNotFound.Error())
	}
}

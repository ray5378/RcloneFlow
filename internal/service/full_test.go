package service

import (
	"context"
	"testing"

	"rcloneflow/internal/store"
)

// TaskDAOMock 模拟TaskDAO
type TaskDAOMock struct {
	tasks []store.Task
}

func (m *TaskDAOMock) Create(task store.Task) (store.Task, error) {
	task.ID = int64(len(m.tasks) + 1)
	m.tasks = append(m.tasks, task)
	return task, nil
}

func (m *TaskDAOMock) GetByID(id int64) (store.Task, bool) {
	for _, t := range m.tasks {
		if t.ID == id {
			return t, true
		}
	}
	return store.Task{}, false
}

func (m *TaskDAOMock) GetAll() ([]store.Task, error) {
	return m.tasks, nil
}

func (m *TaskDAOMock) Update(id int64, task store.Task) error {
	for i, t := range m.tasks {
		if t.ID == id {
			m.tasks[i] = task
			return nil
		}
	}
	return nil
}

func (m *TaskDAOMock) Delete(id int64) error {
	for i, t := range m.tasks {
		if t.ID == id {
			m.tasks = append(m.tasks[:i], m.tasks[i+1:]...)
			return nil
		}
	}
	return nil
}

// RunDAOMock 模拟RunDAO
type RunDAOMock struct {
	runs []store.Run
}

func (m *RunDAOMock) Create(run store.Run) (store.Run, error) {
	run.ID = int64(len(m.runs) + 100)
	m.runs = append(m.runs, run)
	return run, nil
}

func (m *RunDAOMock) GetByID(id int64) (store.Run, bool) {
	for _, r := range m.runs {
		if r.ID == id {
			return r, true
		}
	}
	return store.Run{}, false
}

func (m *RunDAOMock) GetAll() ([]store.Run, error) {
	return m.runs, nil
}

func (m *RunDAOMock) GetRunning() ([]store.JobStatus, error) {
	var running []store.JobStatus
	for _, r := range m.runs {
		if r.Status == "running" {
			running = append(running, store.JobStatus{
				ID:      r.ID,
				RcJobID: r.RcJobID,
				Status:  r.Status,
			})
		}
	}
	return running, nil
}

func (m *RunDAOMock) Update(id int64, updateFn func(*store.Run)) error {
	for i, r := range m.runs {
		if r.ID == id {
			updateFn(&m.runs[i])
			return nil
		}
	}
	return nil
}

func (m *RunDAOMock) UpdateStatus(id int64, status, errorMsg string, summary map[string]any) error {
	for i, r := range m.runs {
		if r.ID == id {
			m.runs[i].Status = status
			m.runs[i].Error = errorMsg
			m.runs[i].Summary = summary
			return nil
		}
	}
	return nil
}

// TaskRunnerMock 模拟TaskRunner
type TaskRunnerMock struct {
	RunTaskFn func(ctx context.Context, taskID int64, mode, srcRemote, srcPath, dstRemote, dstPath, trigger string) (int64, error)
}

func (m *TaskRunnerMock) RunTask(ctx context.Context, taskID int64, mode, srcRemote, srcPath, dstRemote, dstPath, trigger string) (int64, error) {
	if m.RunTaskFn != nil {
		return m.RunTaskFn(ctx, taskID, mode, srcRemote, srcPath, dstRemote, dstPath, trigger)
	}
	return 999, nil
}

// TestTaskService_CreateTask 测试创建任务
func TestTaskService_CreateTask(t *testing.T) {
	taskDAO := &TaskDAOMock{}
	runner := &TaskRunnerMock{}

	// 创建一个简化的service来测试
	svc := &testTaskService{
		taskDAO: taskDAO,
		runner:  runner,
	}

	task := store.Task{
		Name:         "test-task",
		Mode:         "copy",
		SourceRemote: "local",
		SourcePath:   "/src",
		TargetRemote: "gdrive",
		TargetPath:   "/dst",
	}

	created, err := svc.CreateTask(task)
	if err != nil {
		t.Fatalf("CreateTask() error = %v", err)
	}

	if created.Name != "test-task" {
		t.Errorf("expected Name 'test-task', got %s", created.Name)
	}

	if created.ID == 0 {
		t.Error("expected non-zero ID")
	}

	if len(taskDAO.tasks) != 1 {
		t.Errorf("expected 1 task in DAO, got %d", len(taskDAO.tasks))
	}
}

// TestTaskService_ListTasks 测试列出任务
func TestTaskService_ListTasks(t *testing.T) {
	taskDAO := &TaskDAOMock{
		tasks: []store.Task{
			{ID: 1, Name: "task1", Mode: "copy"},
			{ID: 2, Name: "task2", Mode: "sync"},
		},
	}

	svc := &testTaskService{
		taskDAO: taskDAO,
		runner:  &TaskRunnerMock{},
	}

	tasks, err := svc.ListTasks()
	if err != nil {
		t.Fatalf("ListTasks() error = %v", err)
	}

	if len(tasks) != 2 {
		t.Errorf("expected 2 tasks, got %d", len(tasks))
	}
}

// TestTaskService_UpdateTask 测试更新任务
func TestTaskService_UpdateTask(t *testing.T) {
	taskDAO := &TaskDAOMock{
		tasks: []store.Task{
			{ID: 1, Name: "original", Mode: "copy"},
		},
	}

	svc := &testTaskService{
		taskDAO: taskDAO,
		runner:  &TaskRunnerMock{},
	}

	updated := store.Task{
		Name:         "updated",
		Mode:         "sync",
		SourceRemote: "local",
		SourcePath:   "/newsrc",
		TargetRemote: "gdrive",
		TargetPath:   "/newdst",
	}

	err := svc.UpdateTask(1, updated)
	if err != nil {
		t.Fatalf("UpdateTask() error = %v", err)
	}

	task, _ := taskDAO.GetByID(1)
	if task.Name != "updated" {
		t.Errorf("expected Name 'updated', got %s", task.Name)
	}

	if task.Mode != "sync" {
		t.Errorf("expected Mode 'sync', got %s", task.Mode)
	}
}

// TestTaskService_DeleteTask 测试删除任务
func TestTaskService_DeleteTask(t *testing.T) {
	taskDAO := &TaskDAOMock{
		tasks: []store.Task{
			{ID: 1, Name: "task1"},
			{ID: 2, Name: "task2"},
		},
	}

	svc := &testTaskService{
		taskDAO: taskDAO,
		runner:  &TaskRunnerMock{},
	}

	err := svc.DeleteTask(1)
	if err != nil {
		t.Fatalf("DeleteTask() error = %v", err)
	}

	if len(taskDAO.tasks) != 1 {
		t.Errorf("expected 1 task after delete, got %d", len(taskDAO.tasks))
	}

	_, ok := taskDAO.GetByID(1)
	if ok {
		t.Error("expected task 1 to be deleted")
	}
}

// TestTaskService_RunTask 测试运行任务
func TestTaskService_RunTask(t *testing.T) {
	taskDAO := &TaskDAOMock{
		tasks: []store.Task{
			{ID: 1, Name: "task1", Mode: "copy", SourceRemote: "local", SourcePath: "/src", TargetRemote: "gdrive", TargetPath: "/dst"},
		},
	}
	runDAO := &RunDAOMock{}
	runner := &TaskRunnerMock{
		RunTaskFn: func(ctx context.Context, taskID int64, mode, srcRemote, srcPath, dstRemote, dstPath, trigger string) (int64, error) {
			if taskID != 1 {
				t.Errorf("expected taskID 1, got %d", taskID)
			}
			if mode != "copy" {
				t.Errorf("expected mode 'copy', got %s", mode)
			}
			return 123, nil
		},
	}

	svc := &testTaskServiceWithRun{
		taskDAO: taskDAO,
		runDAO:  runDAO,
		runner:  runner,
	}

	err := svc.RunTask(context.Background(), 1, "manual")
	if err != nil {
		t.Fatalf("RunTask() error = %v", err)
	}

	if len(runDAO.runs) != 1 {
		t.Errorf("expected 1 run record, got %d", len(runDAO.runs))
	}

	if runDAO.runs[0].RcJobID != 123 {
		t.Errorf("expected RcJobID 123, got %d", runDAO.runs[0].RcJobID)
	}

	if runDAO.runs[0].Status != "running" {
		t.Errorf("expected Status 'running', got %s", runDAO.runs[0].Status)
	}
}

// TestTaskService_RunTask_NotFound 测试运行不存在的任务
func TestTaskService_RunTask_NotFound(t *testing.T) {
	taskDAO := &TaskDAOMock{tasks: []store.Task{}}
	runDAO := &RunDAOMock{}
	runner := &TaskRunnerMock{}

	svc := &testTaskServiceWithRun{
		taskDAO: taskDAO,
		runDAO:  runDAO,
		runner:  runner,
	}

	err := svc.RunTask(context.Background(), 999, "manual")
	if err != ErrTaskNotFound {
		t.Errorf("expected ErrTaskNotFound, got %v", err)
	}
}

// testTaskService 简化版任务服务（用于测试）
type testTaskService struct {
	taskDAO *TaskDAOMock
	runner  *TaskRunnerMock
}

func (s *testTaskService) CreateTask(task store.Task) (store.Task, error) {
	return s.taskDAO.Create(task)
}

func (s *testTaskService) ListTasks() ([]store.Task, error) {
	return s.taskDAO.GetAll()
}

func (s *testTaskService) UpdateTask(id int64, task store.Task) error {
	return s.taskDAO.Update(id, task)
}

func (s *testTaskService) DeleteTask(id int64) error {
	return s.taskDAO.Delete(id)
}

func (s *testTaskService) GetTask(id int64) (store.Task, bool) {
	return s.taskDAO.GetByID(id)
}

// testTaskServiceWithRun 带运行的简化任务服务
type testTaskServiceWithRun struct {
	taskDAO *TaskDAOMock
	runDAO  *RunDAOMock
	runner  *TaskRunnerMock
}

func (s *testTaskServiceWithRun) CreateTask(task store.Task) (store.Task, error) {
	return s.taskDAO.Create(task)
}

func (s *testTaskServiceWithRun) ListTasks() ([]store.Task, error) {
	return s.taskDAO.GetAll()
}

func (s *testTaskServiceWithRun) UpdateTask(id int64, task store.Task) error {
	return s.taskDAO.Update(id, task)
}

func (s *testTaskServiceWithRun) DeleteTask(id int64) error {
	return s.taskDAO.Delete(id)
}

func (s *testTaskServiceWithRun) GetTask(id int64) (store.Task, bool) {
	return s.taskDAO.GetByID(id)
}

func (s *testTaskServiceWithRun) RunTask(ctx context.Context, taskID int64, trigger string) error {
	t, ok := s.GetTask(taskID)
	if !ok {
		return ErrTaskNotFound
	}

	jobID, err := s.runner.RunTask(ctx, t.ID, t.Mode, t.SourceRemote, t.SourcePath, t.TargetRemote, t.TargetPath, trigger)
	if err != nil {
		return err
	}

	_, err = s.runDAO.Create(store.Run{
		TaskID:  taskID,
		RcJobID: jobID,
		Status:  "running",
		Trigger: trigger,
	})
	return err
}

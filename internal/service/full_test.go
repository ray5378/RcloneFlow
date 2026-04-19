package service

import (
	"context"
	"testing"

	"rcloneflow/internal/store"
)

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

func (m *TaskDAOMock) GetAll() ([]store.Task, error) { return m.tasks, nil }

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

func (m *RunDAOMock) GetAll() ([]store.Run, error) { return m.runs, nil }

func (m *RunDAOMock) Update(id int64, updateFn func(*store.Run)) error {
	for i := range m.runs {
		if m.runs[i].ID == id {
			updateFn(&m.runs[i])
			return nil
		}
	}
	return nil
}

func (m *RunDAOMock) UpdateStatus(id int64, status, errorMsg string, summary map[string]any) error {
	for i := range m.runs {
		if m.runs[i].ID == id {
			m.runs[i].Status = status
			m.runs[i].Error = errorMsg
			m.runs[i].Summary = summary
			return nil
		}
	}
	return nil
}

type TaskRunnerMock struct {
	RunTaskFn func(ctx context.Context, taskID int64, mode, srcRemote, srcPath, dstRemote, dstPath, trigger string) (int64, error)
}

func (m *TaskRunnerMock) RunTask(ctx context.Context, taskID int64, mode, srcRemote, srcPath, dstRemote, dstPath, trigger string) (int64, error) {
	if m.RunTaskFn != nil {
		return m.RunTaskFn(ctx, taskID, mode, srcRemote, srcPath, dstRemote, dstPath, trigger)
	}
	return 999, nil
}

func TestTaskServiceFull_CreateTask(t *testing.T) {
	taskDAO := &TaskDAOMock{}
	svc := &testTaskService{taskDAO: taskDAO, runner: &TaskRunnerMock{}}

	created, err := svc.CreateTask(store.Task{
		Name:         "test-task",
		Mode:         "copy",
		SourceRemote: "local",
		SourcePath:   "/src",
		TargetRemote: "gdrive",
		TargetPath:   "/dst",
	})
	if err != nil {
		t.Fatalf("CreateTask() error = %v", err)
	}
	if created.ID == 0 {
		t.Fatal("expected non-zero ID")
	}
}

func TestTaskServiceFull_RunTask(t *testing.T) {
	taskDAO := &TaskDAOMock{tasks: []store.Task{{ID: 1, Name: "task1", Mode: "copy", SourceRemote: "local", SourcePath: "/src", TargetRemote: "gdrive", TargetPath: "/dst"}}}
	runDAO := &RunDAOMock{}
	runner := &TaskRunnerMock{
		RunTaskFn: func(ctx context.Context, taskID int64, mode, srcRemote, srcPath, dstRemote, dstPath, trigger string) (int64, error) {
			return 123, nil
		},
	}

	svc := &testTaskServiceWithRun{taskDAO: taskDAO, runDAO: runDAO, runner: runner}
	if err := svc.RunTask(context.Background(), 1, "manual"); err != nil {
		t.Fatalf("RunTask() error = %v", err)
	}
	if len(runDAO.runs) != 1 {
		t.Fatalf("expected 1 run record, got %d", len(runDAO.runs))
	}
	if runDAO.runs[0].Status != "running" {
		t.Errorf("expected Status running, got %s", runDAO.runs[0].Status)
	}
}

func TestTaskServiceFull_RunTask_NotFound(t *testing.T) {
	svc := &testTaskServiceWithRun{taskDAO: &TaskDAOMock{}, runDAO: &RunDAOMock{}, runner: &TaskRunnerMock{}}
	if err := svc.RunTask(context.Background(), 999, "manual"); err != ErrTaskNotFound {
		t.Errorf("expected ErrTaskNotFound, got %v", err)
	}
}

type testTaskService struct {
	taskDAO *TaskDAOMock
	runner  *TaskRunnerMock
}

func (s *testTaskService) CreateTask(task store.Task) (store.Task, error) { return s.taskDAO.Create(task) }
func (s *testTaskService) ListTasks() ([]store.Task, error)               { return s.taskDAO.GetAll() }
func (s *testTaskService) UpdateTask(id int64, task store.Task) error     { return s.taskDAO.Update(id, task) }
func (s *testTaskService) DeleteTask(id int64) error                      { return s.taskDAO.Delete(id) }
func (s *testTaskService) GetTask(id int64) (store.Task, bool)            { return s.taskDAO.GetByID(id) }

type testTaskServiceWithRun struct {
	taskDAO *TaskDAOMock
	runDAO  *RunDAOMock
	runner  *TaskRunnerMock
}

func (s *testTaskServiceWithRun) CreateTask(task store.Task) (store.Task, error) {
	return s.taskDAO.Create(task)
}
func (s *testTaskServiceWithRun) ListTasks() ([]store.Task, error) { return s.taskDAO.GetAll() }
func (s *testTaskServiceWithRun) UpdateTask(id int64, task store.Task) error {
	return s.taskDAO.Update(id, task)
}
func (s *testTaskServiceWithRun) DeleteTask(id int64) error           { return s.taskDAO.Delete(id) }
func (s *testTaskServiceWithRun) GetTask(id int64) (store.Task, bool) { return s.taskDAO.GetByID(id) }
func (s *testTaskServiceWithRun) RunTask(ctx context.Context, taskID int64, trigger string) error {
	t, ok := s.GetTask(taskID)
	if !ok {
		return ErrTaskNotFound
	}
	_, err := s.runner.RunTask(ctx, t.ID, t.Mode, t.SourceRemote, t.SourcePath, t.TargetRemote, t.TargetPath, trigger)
	if err != nil {
		return err
	}
	_, err = s.runDAO.Create(store.Run{TaskID: taskID, Status: "running", Trigger: trigger})
	return err
}

package service

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"go.uber.org/zap"

	"rcloneflow/internal/active_transfer"
	"rcloneflow/internal/config"
	"rcloneflow/internal/logger"
	"rcloneflow/internal/store"
)

type TaskService struct {
	db        *store.DB
	activeMgr *active_transfer.Manager
	tagSvc    *TagService
}

func NewTaskService(db *store.DB, activeMgr ...*active_transfer.Manager) *TaskService {
	var mgr *active_transfer.Manager
	if len(activeMgr) > 0 {
		mgr = activeMgr[0]
	}
	return &TaskService{db: db, activeMgr: mgr}
}

func (s *TaskService) SetTagService(svc *TagService) {
	s.tagSvc = svc
}

func (s *TaskService) recalcTags() {
	if s.tagSvc != nil {
		if err := s.tagSvc.RecalcTags(); err != nil {
			logger.Error("recalc tags failed", zap.Error(err))
		}
	}
}

func (s *TaskService) ListTasks() ([]store.Task, error) {
	return s.db.ListTasks()
}

func (s *TaskService) CreateTask(task store.Task) (store.Task, error) {
	if err := s.ensureTaskNameUnique(task.Name, 0); err != nil {
		return store.Task{}, err
	}
	result, err := s.db.AddTask(task)
	if err == nil {
		s.recalcTags()
	}
	return result, err
}

func (s *TaskService) UpdateTask(id int64, task store.Task) error {
	cur, ok := s.db.GetTask(id)
	if !ok {
		return ErrTaskNotFound
	}
	merged := cur
	if strings.TrimSpace(task.Name) != "" {
		merged.Name = task.Name
	}
	if strings.TrimSpace(task.Mode) != "" {
		merged.Mode = task.Mode
	}
	if strings.TrimSpace(task.SourceRemote) != "" {
		merged.SourceRemote = task.SourceRemote
	}
	if strings.TrimSpace(task.SourcePath) != "" {
		merged.SourcePath = task.SourcePath
	}
	if strings.TrimSpace(task.TargetRemote) != "" {
		merged.TargetRemote = task.TargetRemote
	}
	if strings.TrimSpace(task.TargetPath) != "" {
		merged.TargetPath = task.TargetPath
	}
	if len(task.Options) > 0 {
		merged.Options = task.Options
	}
	if len(task.BisyncOptions) > 0 {
		merged.BisyncOptions = task.BisyncOptions
	}
	if cur.Mode == "bisync" && merged.Mode != "bisync" {
		_ = s.cleanupBisyncDir(cur.Name)
	}
	if err := s.ensureTaskNameUnique(merged.Name, id); err != nil {
		return err
	}
	if err := s.db.UpdateTask(id, merged); err != nil {
		return err
	}
	s.recalcTags()
	return nil
}

func (s *TaskService) ensureTaskNameUnique(name string, excludeID int64) error {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return nil
	}
	tasks, err := s.db.ListTasks()
	if err != nil {
		return err
	}
	for _, task := range tasks {
		if task.ID == excludeID {
			continue
		}
		if strings.EqualFold(strings.TrimSpace(task.Name), trimmed) {
			return ErrTaskNameExists
		}
	}
	return nil
}

func (s *TaskService) UpdateTaskSortOrders(orders map[int64]int64, priorityTaskID int64) error {
	tasks, err := s.db.ListTasks()
	if err != nil {
		return err
	}

	byID := make(map[int64]store.Task, len(tasks))
	used := make(map[int64]int64, len(tasks))
	for _, task := range tasks {
		byID[task.ID] = task
	}

	if priorityTaskID != 0 {
		if _, ok := byID[priorityTaskID]; !ok {
			return ErrTaskNotFound
		}
	}

	orderedIDs := make([]int64, 0, len(orders))
	if priorityTaskID != 0 {
		if _, ok := orders[priorityTaskID]; ok {
			orderedIDs = append(orderedIDs, priorityTaskID)
		}
	}
	for _, task := range tasks {
		if task.ID == priorityTaskID {
			continue
		}
		if _, ok := orders[task.ID]; ok {
			orderedIDs = append(orderedIDs, task.ID)
		}
	}

	for _, taskID := range orderedIDs {
		requested := orders[taskID]
		if _, ok := byID[taskID]; !ok {
			return ErrTaskNotFound
		}
		current := requested
		for {
			if _, exists := used[current]; !exists {
				used[current] = taskID
				break
			}
			current++
		}
	}

	for _, task := range tasks {
		if _, ok := orders[task.ID]; ok {
			continue
		}
		current := task.SortOrder
		if current == 0 {
			current = task.ID
		}
		for {
			if _, exists := used[current]; !exists {
				used[current] = task.ID
				break
			}
			current++
		}
	}

	finalIDs := make([]int64, 0, len(used))
	for _, taskID := range used {
		finalIDs = append(finalIDs, taskID)
	}

	updates := make(map[int64]int64, len(finalIDs))
	type pair struct {
		sortOrder int64
		taskID    int64
	}
	pairs := make([]pair, 0, len(used))
	for sortOrder, taskID := range used {
		pairs = append(pairs, pair{sortOrder: sortOrder, taskID: taskID})
	}
	for i := 0; i < len(pairs); i++ {
		for j := i + 1; j < len(pairs); j++ {
			if pairs[j].sortOrder < pairs[i].sortOrder {
				pairs[i], pairs[j] = pairs[j], pairs[i]
			}
		}
	}
	for index, item := range pairs {
		updates[item.taskID] = int64(index + 1)
	}

	return s.db.UpdateTaskSortOrders(updates)
}

func (s *TaskService) UpdateTaskOptions(id int64, opts map[string]any) error {
	t, ok := s.db.GetTask(id)
	if !ok {
		return ErrTaskNotFound
	}
	merged := map[string]any{}
	if len(t.Options) > 0 {
		var cur map[string]any
		if json.Unmarshal(t.Options, &cur) == nil && cur != nil {
			for k, v := range cur {
				merged[k] = v
			}
		}
	}
	for k, v := range opts {
		merged[k] = v
	}
	b, err := json.Marshal(merged)
	if err != nil {
		return err
	}
	t.Options = b
	return s.db.UpdateTask(id, t)
}

func (s *TaskService) DeleteTask(id int64) error {
	task, ok := s.db.GetTask(id)
	if !ok {
		return ErrTaskNotFound
	}
	runs, err := s.db.ListRunsByTask(id)
	if err != nil {
		return err
	}

	logDirs := make(map[string]struct{})
	for _, run := range runs {
		if run.Summary == nil {
			continue
		}
		if p, ok := run.Summary["stderrFile"].(string); ok && p != "" {
			_ = os.Remove(p)
			logDirs[filepath.Dir(p)] = struct{}{}
		}
	}
	for dir := range logDirs {
		if entries, err := os.ReadDir(dir); err == nil && len(entries) == 0 {
			_ = os.Remove(dir)
		}
	}

	logsBase := config.DataDir()
	logsDir := filepath.Join(logsBase, "logs")
	trimmedName := strings.TrimSpace(task.Name)
	if trimmedName != "" {
		pattern := filepath.Join(logsDir, trimmedName+"-*")
		if matches, err := filepath.Glob(pattern); err == nil {
			for _, dir := range matches {
				_ = os.RemoveAll(dir)
			}
		}
	}

	if task.Mode == "bisync" {
		_ = s.cleanupBisyncDir(task.Name)
	}

	if err := s.db.DeleteRunsByTask(id); err != nil {
		return err
	}

	err = s.db.DeleteTask(id)
	if err == nil {
		s.recalcTags()
	}
	return err
}

func (s *TaskService) GetTask(id int64) (store.Task, bool) {
	return s.db.GetTask(id)
}

func (s *TaskService) getBisyncDir(taskName string) string {
	dataDir := config.DataDir()
	safeName := func(s string) string {
		s = strings.TrimSpace(s)
		if s == "" {
			s = "task"
		}
		invalid := regexp.MustCompile(`[^a-zA-Z0-9\p{Han}_-]+`)
		s = invalid.ReplaceAllString(s, "_")
		return s
	}(taskName)
	return filepath.Join(dataDir, "bisync", safeName)
}

func (s *TaskService) cleanupBisyncDir(taskName string) error {
	dir := s.getBisyncDir(taskName)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil
	}
	return os.RemoveAll(dir)
}

func (s *TaskService) GetBisyncLstFiles(taskID int64) ([]string, error) {
	task, ok := s.db.GetTask(taskID)
	if !ok {
		return nil, ErrTaskNotFound
	}
	dir := s.getBisyncDir(task.Name)
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}
	files := []string{}
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".lst") {
			files = append(files, entry.Name())
		}
	}
	return files, nil
}

func (s *TaskService) DeleteBisyncLstFile(taskID int64, filename string) error {
	task, ok := s.db.GetTask(taskID)
	if !ok {
		return ErrTaskNotFound
	}
	dir := s.getBisyncDir(task.Name)
	filePath := filepath.Join(dir, filename)
	if filepath.Base(filename) != filename {
		return fmt.Errorf("invalid filename")
	}
	return os.Remove(filePath)
}

func (s *TaskService) RollbackBisyncLstFile(taskID int64, filename string) error {
	task, ok := s.db.GetTask(taskID)
	if !ok {
		return ErrTaskNotFound
	}
	dir := s.getBisyncDir(task.Name)
	srcPath := filepath.Join(dir, filename)
	if filepath.Base(filename) != filename {
		return fmt.Errorf("invalid filename")
	}
	if _, err := os.Stat(srcPath); os.IsNotExist(err) {
		return fmt.Errorf("file not found")
	}
	now := time.Now().Format("20060102-150405")
	currentFiles := []string{"path1.lst", "path2.lst"}
	for _, f := range currentFiles {
		currentPath := filepath.Join(dir, f)
		if _, err := os.Stat(currentPath); err == nil {
			_ = os.Rename(currentPath, filepath.Join(dir, f+"."+now+".bak"))
		}
	}
	if strings.Contains(filename, "path1") {
		return os.Rename(srcPath, filepath.Join(dir, "path1.lst"))
	} else if strings.Contains(filename, "path2") {
		return os.Rename(srcPath, filepath.Join(dir, "path2.lst"))
	}
	return fmt.Errorf("unrecognized lst file type")
}

func (s *TaskService) ResyncBisync(taskID int64) error {
	task, ok := s.db.GetTask(taskID)
	if !ok {
		return ErrTaskNotFound
	}
	var opts map[string]any
	if len(task.BisyncOptions) > 0 {
		_ = json.Unmarshal(task.BisyncOptions, &opts)
	}
	if opts == nil {
		opts = map[string]any{}
	}
	opts["resync"] = true
	b, _ := json.Marshal(opts)
	task.BisyncOptions = b
	if err := s.db.UpdateTask(taskID, task); err != nil {
		return err
	}
	_, err := s.RunTask(context.Background(), taskID, "manual")
	return err
}
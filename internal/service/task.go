package service

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
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

type BisyncLstVersion struct {
	ID        string    `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	Path1Lst  string    `json:"path1Lst"`
	Path2Lst  string    `json:"path2Lst"`
}

func (s *TaskService) GetBisyncLstFiles(taskID int64) ([]BisyncLstVersion, error) {
	task, ok := s.db.GetTask(taskID)
	if !ok {
		return nil, ErrTaskNotFound
	}
	dir := s.getBisyncDir(task.Name)
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return []BisyncLstVersion{}, nil
		}
		return nil, err
	}
	
	versions := make(map[string]*BisyncLstVersion)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		
		if (strings.HasSuffix(name, ".path1.lst") || strings.HasSuffix(name, ".path2.lst")) && strings.Contains(name, ".lst.") {
			parts := strings.Split(name, ".lst.")
			if len(parts) == 2 {
				timestampStr := strings.TrimSuffix(parts[1], ".bak")
				if timestamp, err := time.Parse("20060102-150405", timestampStr); err == nil {
					versionID := timestampStr
					if _, exists := versions[versionID]; !exists {
						versions[versionID] = &BisyncLstVersion{
							ID:        versionID,
							Timestamp: timestamp,
							Path1Lst:  "",
							Path2Lst:  "",
						}
					}
					if strings.Contains(name, ".path1.lst") {
						versions[versionID].Path1Lst = name
					} else if strings.Contains(name, ".path2.lst") {
						versions[versionID].Path2Lst = name
					}
				}
			}
		}
	}
	
	result := make([]BisyncLstVersion, 0, len(versions))
	for _, v := range versions {
		result = append(result, *v)
	}
	
	sort.Slice(result, func(i, j int) bool {
		return result[i].Timestamp.After(result[j].Timestamp)
	})
	
	return result, nil
}

func (s *TaskService) DeleteBisyncLstVersion(taskID int64, versionID string) error {
	task, ok := s.db.GetTask(taskID)
	if !ok {
		return ErrTaskNotFound
	}
	dir := s.getBisyncDir(task.Name)
	
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	
	pattern := fmt.Sprintf(".lst.%s.bak", versionID)
	for _, entry := range entries {
		if !entry.IsDir() && strings.Contains(entry.Name(), pattern) {
			filePath := filepath.Join(dir, entry.Name())
			_ = os.Remove(filePath)
		}
	}
	
	return nil
}

func (s *TaskService) RollbackBisyncLstVersion(taskID int64, versionID string) error {
	task, ok := s.db.GetTask(taskID)
	if !ok {
		return ErrTaskNotFound
	}
	dir := s.getBisyncDir(task.Name)
	
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("bisync directory not found")
		}
		return err
	}
	
	pattern := fmt.Sprintf(".lst.%s.bak", versionID)
	path1File := ""
	path2File := ""
	for _, entry := range entries {
		if !entry.IsDir() && strings.Contains(entry.Name(), pattern) {
			if strings.Contains(entry.Name(), ".path1.lst.") {
				path1File = entry.Name()
			} else if strings.Contains(entry.Name(), ".path2.lst.") {
				path2File = entry.Name()
			}
		}
	}
	
	if path1File == "" && path2File == "" {
		return fmt.Errorf("version not found")
	}
	
	now := time.Now().Format("20060102-150405")
	currentFiles := []string{"path1.lst", "path2.lst"}
	for _, f := range currentFiles {
		currentPath := filepath.Join(dir, f)
		if _, err := os.Stat(currentPath); err == nil {
			_ = os.Rename(currentPath, filepath.Join(dir, f+"."+now+".bak"))
		}
	}
	
	if path1File != "" {
		if err := os.Rename(filepath.Join(dir, path1File), filepath.Join(dir, "path1.lst")); err != nil {
			return err
		}
	}
	if path2File != "" {
		if err := os.Rename(filepath.Join(dir, path2File), filepath.Join(dir, "path2.lst")); err != nil {
			return err
		}
	}
	
	return nil
}

func (s *TaskService) BackupCurrentLstFiles(taskID int64) error {
	task, ok := s.db.GetTask(taskID)
	if !ok {
		return ErrTaskNotFound
	}
	dir := s.getBisyncDir(task.Name)
	
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	
	now := time.Now().Format("20060102-150405")
	currentFiles := []string{"path1.lst", "path2.lst"}
	for _, f := range currentFiles {
		srcPath := filepath.Join(dir, f)
		if _, err := os.Stat(srcPath); err == nil {
			dstPath := filepath.Join(dir, f+"."+now+".bak")
			if err := os.Rename(srcPath, dstPath); err != nil {
				_ = os.Copy(srcPath, dstPath)
			}
		}
	}
	
	return s.cleanupOldLstBackups(taskID)
}

func (s *TaskService) cleanupOldLstBackups(taskID int64) error {
	task, ok := s.db.GetTask(taskID)
	if !ok {
		return ErrTaskNotFound
	}
	dir := s.getBisyncDir(task.Name)
	
	backupCount := 5
	if len(task.BisyncOptions) > 0 {
		var opts store.BisyncOptions
		if json.Unmarshal(task.BisyncOptions, &opts) == nil && opts.LstBackupCount > 0 {
			backupCount = opts.LstBackupCount
		}
	}
	
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	
	type backupFile struct {
		name      string
		timestamp time.Time
	}
	backups := make([]backupFile, 0)
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".bak") {
			continue
		}
		name := entry.Name()
		parts := strings.Split(name, ".lst.")
		if len(parts) == 2 {
			timestampStr := strings.TrimSuffix(parts[1], ".bak")
			if timestamp, err := time.Parse("20060102-150405", timestampStr); err == nil {
				backups = append(backups, backupFile{name: name, timestamp: timestamp})
			}
		}
	}
	
	sort.Slice(backups, func(i, j int) bool {
		return backups[i].timestamp.After(backups[j].timestamp)
	})
	
	if len(backups) > backupCount*2 {
		timestampSet := make(map[string]bool)
		for _, b := range backups {
			timestampParts := strings.Split(b.name, ".lst.")
			if len(timestampParts) == 2 {
				timestampSet[strings.TrimSuffix(timestampParts[1], ".bak")] = true
			}
		}
		
		timestamps := make([]string, 0, len(timestampSet))
		for ts := range timestampSet {
			timestamps = append(timestamps, ts)
		}
		
		sort.Slice(timestamps, func(i, j int) bool {
			t1, _ := time.Parse("20060102-150405", timestamps[i])
			t2, _ := time.Parse("20060102-150405", timestamps[j])
			return t1.After(t2)
		})
		
		if len(timestamps) > backupCount {
			for i := backupCount; i < len(timestamps); i++ {
				pattern := fmt.Sprintf(".lst.%s.bak", timestamps[i])
				for _, entry := range entries {
					if !entry.IsDir() && strings.Contains(entry.Name(), pattern) {
						_ = os.Remove(filepath.Join(dir, entry.Name()))
					}
				}
			}
		}
	}
	
	return nil
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
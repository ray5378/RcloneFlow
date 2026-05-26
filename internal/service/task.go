package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
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

func (s *TaskService) validateTask(task store.Task, excludeID int64) error {
	if err := s.ensureTaskNameUnique(task.Name, excludeID); err != nil {
		return err
	}
	return nil
}

func (s *TaskService) CreateTask(task store.Task) (store.Task, error) {
	if err := s.validateTask(task, 0); err != nil {
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
	if err := s.validateTask(merged, id); err != nil {
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
		invalid := regexp.MustCompile(`[^a-zA-Z0-9\p{Han}_.-]+`)
		s = invalid.ReplaceAllString(s, "_")
		return s
	}(taskName)
	return filepath.Join(dataDir, "bisync", safeName)
}

func (s *TaskService) getBakDir(taskName string) string {
	return filepath.Join(s.getBisyncDir(taskName), "bak")
}

func (s *TaskService) cleanupBisyncDir(taskName string) error {
	dir := s.getBisyncDir(taskName)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil
	}
	return os.RemoveAll(dir)
}

type BisyncLstVersion struct {
	ID              string    `json:"id"`
	Timestamp       time.Time `json:"timestamp"`
	Path1Lst        string    `json:"path1Lst"`
	Path2Lst        string    `json:"path2Lst"`
	Type            string    `json:"type"`
	Conflict1       string    `json:"conflict1,omitempty"`
	Conflict2       string    `json:"conflict2,omitempty"`
	Path1Size       int64     `json:"path1Size,omitempty"`
	Path2Size       int64     `json:"path2Size,omitempty"`
	Conflict1Size   int64     `json:"conflict1Size,omitempty"`
	Conflict2Size   int64     `json:"conflict2Size,omitempty"`
	Conflict1Mtime  time.Time `json:"conflict1Mtime,omitempty"`
	Conflict2Mtime  time.Time `json:"conflict2Mtime,omitempty"`
}

func (s *TaskService) GetBisyncLstFiles(taskID int64) ([]BisyncLstVersion, error) {
	task, ok := s.db.GetTask(taskID)
	if !ok {
		return nil, ErrTaskNotFound
	}
	dir := s.getBisyncDir(task.Name)
	
	var entries []fs.DirEntry
	var err error
	
	entries, err = os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			safeNameOld := func(s string) string {
				s = strings.TrimSpace(s)
				if s == "" {
					s = "task"
				}
				invalid := regexp.MustCompile(`[^a-zA-Z0-9\p{Han}_-]+`)
				return invalid.ReplaceAllString(s, "_")
			}(task.Name)
			dirOld := filepath.Join(config.DataDir(), "bisync", safeNameOld)
			entries, err = os.ReadDir(dirOld)
			if err != nil {
				if os.IsNotExist(err) {
					return []BisyncLstVersion{}, nil
				}
				return nil, err
			}
		} else {
			return nil, err
		}
	}
	
	versions := make(map[string]*BisyncLstVersion)
	
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		
		if name == "path1.lst" || name == "path2.lst" {
			versionID := "current"
			info, _ := entry.Info()
			size := int64(0)
			modTime := time.Now()
			if info != nil {
				size = info.Size()
				modTime = info.ModTime()
			}
			if _, exists := versions[versionID]; !exists {
				versions[versionID] = &BisyncLstVersion{
					ID:        versionID,
					Timestamp: modTime,
					Path1Lst:  "",
					Path2Lst:  "",
					Type:      "current",
				}
			} else if modTime.After(versions[versionID].Timestamp) {
				versions[versionID].Timestamp = modTime
			}
			if name == "path1.lst" {
				versions[versionID].Path1Lst = name
				versions[versionID].Path1Size = size
			} else if name == "path2.lst" {
				versions[versionID].Path2Lst = name
				versions[versionID].Path2Size = size
			}
			continue
		}
		
		if strings.Contains(name, ".conflict1") || strings.Contains(name, ".conflict2") {
			prefix := name
			if strings.Contains(name, ".conflict1") {
				prefix = strings.TrimSuffix(name, ".conflict1")
			} else if strings.Contains(name, ".conflict2") {
				prefix = strings.TrimSuffix(name, ".conflict2")
			}
			
			info, _ := entry.Info()
			size := int64(0)
			if info != nil {
				size = info.Size()
			}
			
			if _, exists := versions[prefix]; !exists {
				timestamp := time.Now()
				if info != nil {
					timestamp = info.ModTime()
				}
				versions[prefix] = &BisyncLstVersion{
					ID:        prefix,
					Timestamp: timestamp,
					Path1Lst:  "",
					Path2Lst:  "",
					Type:      "conflict",
				}
			}
			
			if strings.Contains(name, ".conflict1") {
				versions[prefix].Conflict1 = name
				versions[prefix].Conflict1Size = size
				versions[prefix].Conflict1Mtime = info.ModTime()
			} else if strings.Contains(name, ".conflict2") {
				versions[prefix].Conflict2 = name
				versions[prefix].Conflict2Size = size
				versions[prefix].Conflict2Mtime = info.ModTime()
			}
			continue
		}
		
		if strings.HasSuffix(name, "-old") {
			info, _ := entry.Info()
			size := int64(0)
			if info != nil {
				size = info.Size()
			}
			timestamp := time.Now()
			if info != nil {
				timestamp = info.ModTime()
			}
			
			if strings.Contains(name, ".path1.lst") {
				prefix := strings.TrimSuffix(name, ".path1.lst-old")
				if _, exists := versions[prefix]; !exists {
					versions[prefix] = &BisyncLstVersion{
						ID:        prefix,
						Timestamp: timestamp,
						Path1Lst:  "",
						Path2Lst:  "",
						Type:      "backup",
					}
				}
				versions[prefix].Path1Lst = name
				versions[prefix].Path1Size = size
			} else if strings.Contains(name, ".path2.lst") {
				prefix := strings.TrimSuffix(name, ".path2.lst-old")
				if _, exists := versions[prefix]; !exists {
					versions[prefix] = &BisyncLstVersion{
						ID:        prefix,
						Timestamp: timestamp,
						Path1Lst:  "",
						Path2Lst:  "",
						Type:      "backup",
					}
				}
				versions[prefix].Path2Lst = name
				versions[prefix].Path2Size = size
			}
			continue
		}
		
		if name == "path1.lst" || name == "path2.lst" || strings.HasSuffix(name, ".path1.lst") || strings.HasSuffix(name, ".path2.lst") {
			versionID := "current"
			info, _ := entry.Info()
			size := int64(0)
			modTime := time.Now()
			if info != nil {
				size = info.Size()
				modTime = info.ModTime()
			}
			if _, exists := versions[versionID]; !exists {
				versions[versionID] = &BisyncLstVersion{
					ID:        versionID,
					Timestamp: modTime,
					Path1Lst:  "",
					Path2Lst:  "",
					Type:      "current",
				}
			} else if modTime.After(versions[versionID].Timestamp) {
				versions[versionID].Timestamp = modTime
			}
			if strings.HasSuffix(name, ".path1.lst") {
				versions[versionID].Path1Lst = name
				versions[versionID].Path1Size = size
			} else {
				versions[versionID].Path2Lst = name
				versions[versionID].Path2Size = size
			}
		}
	}
	
	bakDir := s.getBakDir(task.Name)
	if bakEntries, err := os.ReadDir(bakDir); err == nil {
		for _, entry := range bakEntries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".bak") {
				continue
			}
			name := entry.Name()
			if strings.HasSuffix(name, "-old") {
				continue
			}
			info, _ := entry.Info()
			size := int64(0)
			if info != nil {
				size = info.Size()
			}
			tsStr := ""
			base := strings.TrimSuffix(name, ".bak")
			if idx := strings.LastIndex(base, ".lst."); idx >= 0 {
				tsStr = base[idx+5:]
			} else if idx := strings.LastIndex(base, ".path1.lst."); idx >= 0 {
				tsStr = base[idx+11:]
			} else if idx := strings.LastIndex(base, ".path2.lst."); idx >= 0 {
				tsStr = base[idx+11:]
			}
			if tsStr != "" {
				if timestamp, err := time.ParseInLocation("20060102-150405", tsStr, time.Local); err == nil {
					if _, exists := versions[tsStr]; !exists {
						versions[tsStr] = &BisyncLstVersion{
							ID:        tsStr,
							Timestamp: timestamp,
							Path1Lst:  "",
							Path2Lst:  "",
							Type:      "backup",
						}
					}
					if strings.HasPrefix(base, "path1.lst.") || strings.Contains(base, ".path1.lst.") {
						versions[tsStr].Path1Lst = name
						versions[tsStr].Path1Size = size
					} else if strings.HasPrefix(base, "path2.lst.") || strings.Contains(base, ".path2.lst.") {
						versions[tsStr].Path2Lst = name
						versions[tsStr].Path2Size = size
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
	
	var entries []fs.DirEntry
	var err error
	
	entries, err = os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			safeNameOld := func(s string) string {
				s = strings.TrimSpace(s)
				if s == "" {
					s = "task"
				}
				invalid := regexp.MustCompile(`[^a-zA-Z0-9\p{Han}_-]+`)
				return invalid.ReplaceAllString(s, "_")
			}(task.Name)
			dirOld := filepath.Join(config.DataDir(), "bisync", safeNameOld)
			entries, err = os.ReadDir(dirOld)
			if err != nil {
				if os.IsNotExist(err) {
					return nil
				}
				return err
			}
			dir = dirOld
		} else {
			return err
		}
	}
	
	if versionID == "current" {
		return ErrBisyncCurrentVersion
	}
	
	if strings.Contains(versionID, ".conflict") {
		conflict1 := strings.TrimSuffix(versionID, ".conflict2") + ".conflict1"
		conflict2 := strings.TrimSuffix(versionID, ".conflict1") + ".conflict2"
		for _, entry := range entries {
			if !entry.IsDir() && (entry.Name() == conflict1 || entry.Name() == conflict2) {
				filePath := filepath.Join(dir, entry.Name())
				_ = os.Remove(filePath)
			}
		}
	} else if strings.Contains(versionID, ".path1.lst-old") {
		prefix := strings.TrimSuffix(versionID, ".path1.lst-old")
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasPrefix(entry.Name(), prefix) && strings.HasSuffix(entry.Name(), "-old") {
				filePath := filepath.Join(dir, entry.Name())
				_ = os.Remove(filePath)
			}
		}
	} else if strings.Contains(versionID, ".path2.lst-old") {
		prefix := strings.TrimSuffix(versionID, ".path2.lst-old")
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasPrefix(entry.Name(), prefix) && strings.HasSuffix(entry.Name(), "-old") {
				filePath := filepath.Join(dir, entry.Name())
				_ = os.Remove(filePath)
			}
		}
	} else {
		found := false
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasPrefix(entry.Name(), versionID) && strings.HasSuffix(entry.Name(), "-old") {
				_ = os.Remove(filepath.Join(dir, entry.Name()))
				found = true
			}
		}
		if !found {
			pattern := fmt.Sprintf(".lst.%s.bak", versionID)
			bakDir := s.getBakDir(task.Name)
			if bakEntries, err := os.ReadDir(bakDir); err == nil {
				for _, entry := range bakEntries {
					if !entry.IsDir() && strings.Contains(entry.Name(), pattern) {
						_ = os.Remove(filepath.Join(bakDir, entry.Name()))
					}
				}
			}
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
	
	var entries []fs.DirEntry
	var err error
	
	entries, err = os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			safeNameOld := func(s string) string {
				s = strings.TrimSpace(s)
				if s == "" {
					s = "task"
				}
				invalid := regexp.MustCompile(`[^a-zA-Z0-9\p{Han}_-]+`)
				return invalid.ReplaceAllString(s, "_")
			}(task.Name)
			dirOld := filepath.Join(config.DataDir(), "bisync", safeNameOld)
			entries, err = os.ReadDir(dirOld)
			if err != nil {
				if os.IsNotExist(err) {
					return fmt.Errorf("bisync directory not found")
				}
				return err
			}
			dir = dirOld
		} else {
			return err
		}
	}
	
	path1File := ""
	path2File := ""
	
	if strings.Contains(versionID, ".path1.lst-old") || strings.Contains(versionID, ".path2.lst-old") {
		prefix := ""
		if strings.Contains(versionID, ".path1.lst-old") {
			prefix = strings.TrimSuffix(versionID, ".path1.lst-old")
		} else if strings.Contains(versionID, ".path2.lst-old") {
			prefix = strings.TrimSuffix(versionID, ".path2.lst-old")
		}
		
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasPrefix(entry.Name(), prefix) && strings.HasSuffix(entry.Name(), "-old") {
				if strings.Contains(entry.Name(), ".path1.lst-old") {
					path1File = entry.Name()
				} else if strings.Contains(entry.Name(), ".path2.lst-old") {
					path2File = entry.Name()
				}
			}
		}
	} else {
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasPrefix(entry.Name(), versionID) && strings.HasSuffix(entry.Name(), "-old") {
				if strings.Contains(entry.Name(), ".path1.lst-old") {
					path1File = entry.Name()
				} else if strings.Contains(entry.Name(), ".path2.lst-old") {
					path2File = entry.Name()
				}
			}
		}
		if path1File == "" && path2File == "" {
			pattern := fmt.Sprintf(".lst.%s.bak", versionID)
			bakDir := s.getBakDir(task.Name)
			if bakEntries, err := os.ReadDir(bakDir); err == nil {
				for _, entry := range bakEntries {
					if entry.IsDir() || !strings.Contains(entry.Name(), pattern) {
						continue
					}
					name := entry.Name()
					if strings.Contains(name, ".path1.lst.") || strings.HasPrefix(name, "path1.lst.") {
						path1File = filepath.Join(bakDir, name)
					} else if strings.Contains(name, ".path2.lst.") || strings.HasPrefix(name, "path2.lst.") {
						path2File = filepath.Join(bakDir, name)
					}
				}
			}
		}
	}
	
	if path1File == "" && path2File == "" {
		return fmt.Errorf("version not found")
	}
	
	bakDir := s.getBakDir(task.Name)
	_ = os.MkdirAll(bakDir, 0o755)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasSuffix(name, "-old") || strings.HasSuffix(name, ".bak") {
			continue
		}
		if strings.HasSuffix(name, ".path1.lst") || name == "path1.lst" {
			srcPath := filepath.Join(dir, name)
			if info, err := os.Stat(srcPath); err == nil {
				ts := info.ModTime().Format("20060102-150405")
				_ = os.Rename(srcPath, filepath.Join(bakDir, "path1.lst."+ts+".bak"))
			}
		} else if strings.HasSuffix(name, ".path2.lst") || name == "path2.lst" {
			srcPath := filepath.Join(dir, name)
			if info, err := os.Stat(srcPath); err == nil {
				ts := info.ModTime().Format("20060102-150405")
				_ = os.Rename(srcPath, filepath.Join(bakDir, "path2.lst."+ts+".bak"))
			}
		}
	}
	
	if path1File != "" {
		src := path1File
		if !filepath.IsAbs(src) {
			src = filepath.Join(dir, src)
		}
		if _, err := os.Stat(src); err == nil {
			if err := os.Rename(src, filepath.Join(dir, "path1.lst")); err != nil {
				return err
			}
		}
	}
	if path2File != "" {
		src := path2File
		if !filepath.IsAbs(src) {
			src = filepath.Join(dir, src)
		}
		if _, err := os.Stat(src); err == nil {
			if err := os.Rename(src, filepath.Join(dir, "path2.lst")); err != nil {
				return err
			}
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
	bakDir := s.getBakDir(task.Name)
	
	if err := os.MkdirAll(bakDir, 0o755); err != nil {
		return err
	}
	
	currentFiles := []string{"path1.lst", "path2.lst"}
	for _, f := range currentFiles {
		srcPath := filepath.Join(dir, f)
		if info, err := os.Stat(srcPath); err == nil {
			ts := info.ModTime().Format("20060102-150405")
			dstPath := filepath.Join(bakDir, f+"."+ts+".bak")
			if err := os.Rename(srcPath, dstPath); err != nil {
				srcFile, err := os.Open(srcPath)
				if err == nil {
					defer srcFile.Close()
					dstFile, err := os.Create(dstPath)
					if err == nil {
						defer dstFile.Close()
						_, _ = io.Copy(dstFile, srcFile)
						os.Remove(srcPath)
					}
				}
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
	bakDir := s.getBakDir(task.Name)
	
	backupCount := 5
	if len(task.BisyncOptions) > 0 {
		var opts store.BisyncOptions
		if json.Unmarshal(task.BisyncOptions, &opts) == nil && opts.LstBackupCount > 0 {
			backupCount = opts.LstBackupCount
		}
	}
	
	entries, err := os.ReadDir(bakDir)
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
			if timestamp, err := time.ParseInLocation("20060102-150405", timestampStr, time.Local); err == nil {
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
			t1, _ := time.ParseInLocation("20060102-150405", timestamps[i], time.Local)
			t2, _ := time.ParseInLocation("20060102-150405", timestamps[j], time.Local)
			return t1.After(t2)
		})
		
		if len(timestamps) > backupCount {
			for i := backupCount; i < len(timestamps); i++ {
				pattern := fmt.Sprintf(".lst.%s.bak", timestamps[i])
				for _, entry := range entries {
					if !entry.IsDir() && strings.Contains(entry.Name(), pattern) {
						_ = os.Remove(filepath.Join(bakDir, entry.Name()))
					}
				}
			}
		}
	}
	
	return nil
}

func (s *TaskService) GetBisyncLstContent(taskID int64, fileName string) (string, error) {
	task, ok := s.db.GetTask(taskID)
	if !ok {
		return "", ErrTaskNotFound
	}
	dir := s.getBisyncDir(task.Name)
	filePath := filepath.Join(dir, fileName)

	if strings.HasSuffix(fileName, ".bak") {
		bakDir := s.getBakDir(task.Name)
		bakPath := filepath.Join(bakDir, fileName)
		if _, err := os.Stat(bakPath); err == nil {
			filePath = bakPath
		}
	}

	if !strings.HasPrefix(filePath, filepath.Clean(dir)+string(filepath.Separator)) &&
		!strings.HasPrefix(filePath, filepath.Clean(s.getBakDir(task.Name))+string(filepath.Separator)) {
		return "", fmt.Errorf("invalid file name")
	}

	info, err := os.Stat(filePath)
	if err != nil {
		return "", fmt.Errorf("file not found")
	}
	if info.IsDir() {
		return "", fmt.Errorf("not a file")
	}
	if info.Size() > 10*1024*1024 {
		return "", fmt.Errorf("file too large")
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read file")
	}
	return string(data), nil
}

func (s *TaskService) ResolveBisyncConflict(taskID int64, versionID, keepFile string) error {
	task, ok := s.db.GetTask(taskID)
	if !ok {
		return ErrTaskNotFound
	}
	dir := s.getBisyncDir(task.Name)
	if _, err := os.Stat(dir); err != nil {
		return fmt.Errorf("bisync directory not found")
	}

	var keepName, deleteName string
	if keepFile == "conflict1" {
		keepName = versionID + ".conflict1"
		deleteName = versionID + ".conflict2"
	} else {
		keepName = versionID + ".conflict2"
		deleteName = versionID + ".conflict1"
	}

	keepPath := filepath.Join(dir, keepName)
	delPath := filepath.Join(dir, deleteName)
	if !strings.HasPrefix(filepath.Clean(keepPath), filepath.Clean(dir)+string(os.PathSeparator)) ||
		!strings.HasPrefix(filepath.Clean(delPath), filepath.Clean(dir)+string(os.PathSeparator)) {
		return fmt.Errorf("invalid file path")
	}

	if err := os.Remove(delPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove conflict file: %w", err)
	}

	origPath := filepath.Join(dir, versionID)
	if err := os.Rename(keepPath, origPath); err != nil {
		return fmt.Errorf("failed to rename file: %w", err)
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


package service

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"rcloneflow/internal/store"
)

func (s *TaskService) ExportTasks() (map[string]any, error) {
	tasks, err := s.db.ListTasks()
	if err != nil {
		return nil, err
	}
	schedules, err := s.db.ListSchedules()
	if err != nil {
		return nil, err
	}

	exportTasks := make([]map[string]any, 0, len(tasks))
	for _, t := range tasks {
		taskMap := map[string]any{
			"name":         t.Name,
			"mode":         t.Mode,
			"sourceRemote": t.SourceRemote,
			"sourcePath":   t.SourcePath,
			"targetRemote": t.TargetRemote,
			"targetPath":   t.TargetPath,
			"sortOrder":    t.SortOrder,
		}
		if len(t.Options) > 0 {
			var opts map[string]any
			if json.Unmarshal(t.Options, &opts) == nil {
				taskMap["options"] = opts
			}
		}
		if len(t.BisyncOptions) > 0 {
			var bisyncOpts map[string]any
			if json.Unmarshal(t.BisyncOptions, &bisyncOpts) == nil {
				taskMap["bisyncOptions"] = bisyncOpts
			}
		}
		exportTasks = append(exportTasks, taskMap)
	}

	taskNameByID := make(map[int64]string)
	for _, t := range tasks {
		taskNameByID[t.ID] = t.Name
	}

	exportSchedules := make([]map[string]any, 0, len(schedules))
	for _, sc := range schedules {
		taskName, ok := taskNameByID[sc.TaskID]
		if !ok {
			continue
		}
		exportSchedules = append(exportSchedules, map[string]any{
			"taskName": taskName,
			"spec":     sc.Spec,
			"enabled":  sc.Enabled,
		})
	}

	return map[string]any{
		"version":    1,
		"exportedAt": time.Now().Format(time.RFC3339),
		"tasks":      exportTasks,
		"schedules":  exportSchedules,
	}, nil
}

func (s *TaskService) ImportTasks(data map[string]any, strategy string) (imported, skipped, overwritten int, err error) {
	tasksData, ok := data["tasks"].([]any)
	if !ok {
		return 0, 0, 0, fmt.Errorf("无效的任务数据")
	}
	schedulesData, _ := data["schedules"].([]any)

	existingTasks, err := s.db.ListTasks()
	if err != nil {
		return 0, 0, 0, err
	}
	existingNames := make(map[string]int64)
	for _, t := range existingTasks {
		existingNames[strings.ToLower(strings.TrimSpace(t.Name))] = t.ID
	}

	type importedTask struct {
		name string
		id   int64
	}
	importedTasks := make([]importedTask, 0)

	for _, td := range tasksData {
		taskMap, ok := td.(map[string]any)
		if !ok {
			skipped++
			continue
		}

		name, _ := taskMap["name"].(string)
		if name == "" {
			skipped++
			continue
		}

		mode, _ := taskMap["mode"].(string)
		sourceRemote, _ := taskMap["sourceRemote"].(string)
		sourcePath, _ := taskMap["sourcePath"].(string)
		targetRemote, _ := taskMap["targetRemote"].(string)
		targetPath, _ := taskMap["targetPath"].(string)

		if mode == "" || sourceRemote == "" || targetRemote == "" {
			skipped++
			continue
		}

		newTask := store.Task{
			Name:         name,
			Mode:         mode,
			SourceRemote: sourceRemote,
			SourcePath:   sourcePath,
			TargetRemote: targetRemote,
			TargetPath:   targetPath,
		}

		if opts, ok := taskMap["options"]; ok {
			optsBytes, err := json.Marshal(opts)
			if err == nil {
				newTask.Options = optsBytes
			}
		}
		if bisyncOpts, ok := taskMap["bisyncOptions"]; ok {
			bisyncOptsBytes, err := json.Marshal(bisyncOpts)
			if err == nil {
				newTask.BisyncOptions = bisyncOptsBytes
			}
		}

		lowerName := strings.ToLower(strings.TrimSpace(name))
		if existingID, exists := existingNames[lowerName]; exists {
			if strategy == "overwrite" {
				cur, ok := s.db.GetTask(existingID)
				if !ok {
					skipped++
					continue
				}
				if newTask.SourcePath == "" {
					newTask.SourcePath = cur.SourcePath
				}
				if newTask.TargetPath == "" {
					newTask.TargetPath = cur.TargetPath
				}
				if len(newTask.Options) == 0 {
					newTask.Options = cur.Options
				}
				if len(newTask.BisyncOptions) == 0 {
					newTask.BisyncOptions = cur.BisyncOptions
				}
				if err := s.db.UpdateTask(existingID, newTask); err != nil {
					skipped++
					continue
				}
				overwritten++
				importedTasks = append(importedTasks, importedTask{name: name, id: existingID})
			} else {
				skipped++
			}
		} else {
			created, err := s.db.AddTask(newTask)
			if err != nil {
				skipped++
				continue
			}
			imported++
			importedTasks = append(importedTasks, importedTask{name: name, id: created.ID})
			existingNames[lowerName] = created.ID
		}
	}

	if len(schedulesData) > 0 {
		taskIDByName := make(map[string]int64)
		for _, t := range importedTasks {
			taskIDByName[t.name] = t.id
		}
		for _, t := range existingTasks {
			if _, exists := taskIDByName[t.Name]; !exists {
				taskIDByName[t.Name] = t.ID
			}
		}

		allSchedules, err := s.db.ListSchedules()
		if err != nil {
			return imported, skipped, overwritten, err
		}
		existingSchedKeys := make(map[string]bool)
		for _, sc := range allSchedules {
			existingSchedKeys[fmt.Sprintf("%d:%s", sc.TaskID, sc.Spec)] = true
		}

		for _, sd := range schedulesData {
			schedMap, ok := sd.(map[string]any)
			if !ok {
				continue
			}
			taskName, _ := schedMap["taskName"].(string)
			spec, _ := schedMap["spec"].(string)
			enabled, _ := schedMap["enabled"].(bool)

			if taskName == "" || spec == "" {
				continue
			}

			taskID, exists := taskIDByName[taskName]
			if !exists {
				continue
			}

			scheduleKey := fmt.Sprintf("%d:%s", taskID, spec)
			if existingSchedKeys[scheduleKey] {
				continue
			}

			_, _ = s.db.AddSchedule(store.Schedule{
				TaskID:  taskID,
				Spec:    spec,
				Enabled: enabled,
			})
		}
	}

	return imported, skipped, overwritten, nil
}

func (s *TaskService) ClearAllTasks() error {
	tasks, err := s.db.ListTasks()
	if err != nil {
		return err
	}

	for _, t := range tasks {
		if err := s.DeleteTask(t.ID); err != nil {
			return err
		}
	}

	if err := s.db.Vacuum(); err != nil {
		return err
	}
	if err := s.db.ResetSequence("tasks"); err != nil {
		return err
	}
	if err := s.db.ResetSequence("runs"); err != nil {
		return err
	}
	if err := s.db.ResetSequence("schedules"); err != nil {
		return err
	}
	return nil
}
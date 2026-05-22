package service

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"go.uber.org/zap"

	"rcloneflow/internal/active_transfer"
	"rcloneflow/internal/adapter"
	"rcloneflow/internal/config"
	"rcloneflow/internal/logger"
	"rcloneflow/internal/runnercli"
	"rcloneflow/internal/settings"
	"rcloneflow/internal/store"
)

type TaskRunResult struct {
	Started bool   `json:"started"`
	Reason  string `json:"reason,omitempty"`
	Message string `json:"message,omitempty"`
	TaskID  int64  `json:"taskId,omitempty"`
}

func (s *TaskService) RunTask(ctx context.Context, taskID int64, trigger string) (TaskRunResult, error) {
	t, ok := s.db.GetTask(taskID)
	if !ok {
		return TaskRunResult{}, ErrTaskNotFound
	}

	var opts *adapter.TaskOptions
	if len(t.Options) > 0 {
		if taskOpts, err := adapter.ParseTaskOptionsCompat(t.Options); err == nil {
			opts = taskOpts
		}
	}

	opts = adapter.MergeTaskOptions(opts)

	effectiveOptions := map[string]any{}
	if bs, err := json.Marshal(opts); err == nil {
		if err := json.Unmarshal(bs, &effectiveOptions); err != nil {
			logger.Error("unmarshal effective options", zap.Error(err))
		}
	}
	if len(t.Options) > 0 {
		var raw map[string]any
		if err := json.Unmarshal(t.Options, &raw); err == nil && raw != nil {
			for k, v := range raw {
				effectiveOptions[k] = v
			}
		}
	}

	streamingEnabled := true
	if v, ok := effectiveOptions["enableStreaming"].(bool); ok {
		streamingEnabled = v
	}

	if activeRun, err := s.db.GetActiveRunByTaskID(taskID); err == nil && activeRun.ID > 0 {
		return TaskRunResult{
			Started: false,
			Reason:  "already_running",
			Message: "任务已在运行中，跳过本次执行",
			TaskID:  taskID,
		}, nil
	}

	singletonMode, isSingleton := effectiveOptions["singletonMode"].(bool)

	newRun := store.Run{
		TaskID:  taskID,
		Status:  "running",
		Trigger: trigger,
		Summary: map[string]any{
			"streamingEnabled": streamingEnabled,
			"effectiveOptions": effectiveOptions,
		},
		TaskName:     t.Name,
		TaskMode:     t.Mode,
		SourceRemote: t.SourceRemote,
		SourcePath:   t.SourcePath,
		TargetRemote: t.TargetRemote,
		TargetPath:   t.TargetPath,
	}

	if isSingleton && singletonMode {
		run, existed, err := s.db.TryAcquireRun(&newRun)
		if err != nil {
			return TaskRunResult{}, fmt.Errorf("单例模式：申请运行记录失败，%w", err)
		}
		if existed {
			_, _ = s.db.AddRun(store.Run{
				TaskID:  taskID,
				Status:  "skipped",
				Trigger: trigger,
				Summary: map[string]any{
					"finalSummary": map[string]any{
						"message": "单例模式：有其他任务正在运行，跳过本次执行",
					},
				},
				TaskName:     t.Name,
				TaskMode:     t.Mode,
				SourceRemote: t.SourceRemote,
				SourcePath:   t.SourcePath,
				TargetRemote: t.TargetRemote,
				TargetPath:   t.TargetPath,
			})
			return TaskRunResult{
				Started: false,
				Reason:  "singleton_blocked",
				Message: "单例模式：有其他任务正在运行，跳过本次执行",
				TaskID:  taskID,
			}, nil
		}
		if ts, err := settings.Load(); err == nil {
			_ = s.db.UpdateRun(run.ID, func(rr *store.Run) {
				if rr.Summary == nil {
					rr.Summary = map[string]any{}
				}
				rr.Summary["transferDefaults"] = ts
			})
		}
		if s.activeMgr != nil {
			mode := active_transfer.TrackingModeNormal
			if opts != nil && opts.OpenlistCasCompatible {
				mode = active_transfer.TrackingModeCAS
			}
			cfg := os.Getenv("RCLONE_CONFIG")
			if cfg == "" {
				cfg = filepath.Join(config.DataDir(), "rclone.conf")
			}
			src := t.SourceRemote + ":" + strings.TrimPrefix(t.SourcePath, "/")
			dst := t.TargetRemote + ":" + strings.TrimPrefix(t.TargetPath, "/")
			s.activeMgr.InitState(run.ID, taskID, mode, nil)
			if opts != nil && opts.Transfers > 0 {
				s.activeMgr.SetTransferSlots(run.ID, opts.Transfers)
			}
			go func(runID, taskID int64, mode active_transfer.TrackingMode, cfg, src, dst string, opts *adapter.TaskOptions) {
				defer func() {
					if r := recover(); r != nil {
						logger.Error("goroutine panic", zap.Any("panic", r))
					}
				}()
				candidates, err := active_transfer.BuildCandidateFiles(context.Background(), cfg, src, dst, opts)
				if err != nil {
					s.activeMgr.SetPreflightResult(runID, err)
					return
				}
				s.activeMgr.MergeCandidates(runID, candidates)
			}(run.ID, taskID, mode, cfg, src, dst, opts)
		}
		go func() {
			defer func() {
				if r := recover(); r != nil {
					logger.Error("goroutine panic", zap.Any("panic", r))
				}
			}()
			_ = runnercli.New(&runnercli.StoreDBAdapter{DB: s.db}, &runnercli.WSBroadcaster{}, s.activeMgr).Start(context.Background(), *run, t.Mode, t.SourceRemote, t.SourcePath, t.TargetRemote, t.TargetPath)
		}()
		return TaskRunResult{Started: true, TaskID: taskID}, nil
	}

	run, err := s.db.AddRun(newRun)
	if err != nil {
		return TaskRunResult{}, err
	}
	if ts, err := settings.Load(); err == nil {
		_ = s.db.UpdateRun(run.ID, func(rr *store.Run) {
			if rr.Summary == nil {
				rr.Summary = map[string]any{}
			}
			rr.Summary["transferDefaults"] = ts
		})
	}
	if s.activeMgr != nil {
		mode := active_transfer.TrackingModeNormal
		if opts != nil && opts.OpenlistCasCompatible {
			mode = active_transfer.TrackingModeCAS
		}
		cfg := os.Getenv("RCLONE_CONFIG")
		if cfg == "" {
			cfg = filepath.Join(config.DataDir(), "rclone.conf")
		}
		src := t.SourceRemote + ":" + strings.TrimPrefix(t.SourcePath, "/")
		dst := t.TargetRemote + ":" + strings.TrimPrefix(t.TargetPath, "/")
		s.activeMgr.InitState(run.ID, taskID, mode, nil)
		if opts != nil && opts.Transfers > 0 {
			s.activeMgr.SetTransferSlots(run.ID, opts.Transfers)
		}
		go func(runID, taskID int64, mode active_transfer.TrackingMode, cfg, src, dst string, opts *adapter.TaskOptions) {
			defer func() {
				if r := recover(); r != nil {
					logger.Error("goroutine panic", zap.Any("panic", r))
				}
			}()
			candidates, err := active_transfer.BuildCandidateFiles(context.Background(), cfg, src, dst, opts)
			if err != nil {
				s.activeMgr.SetPreflightResult(runID, err)
				return
			}
			s.activeMgr.MergeCandidates(runID, candidates)
		}(run.ID, taskID, mode, cfg, src, dst, opts)
	}
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Error("goroutine panic", zap.Any("panic", r))
			}
		}()
		_ = runnercli.New(&runnercli.StoreDBAdapter{DB: s.db}, &runnercli.WSBroadcaster{}, s.activeMgr).Start(context.Background(), run, t.Mode, t.SourceRemote, t.SourcePath, t.TargetRemote, t.TargetPath)
	}()
	return TaskRunResult{Started: true, TaskID: taskID}, nil
}
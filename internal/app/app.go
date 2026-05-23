package app

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"rcloneflow/internal/active_transfer"
	"rcloneflow/internal/adapter"
	"rcloneflow/internal/config"
	"rcloneflow/internal/controller"
	"rcloneflow/internal/logger"
	"rcloneflow/internal/router"
	"rcloneflow/internal/scheduler"
	"rcloneflow/internal/service"
	"rcloneflow/internal/store"
	"rcloneflow/internal/websocket"

	"go.uber.org/zap"
)

type taskServiceSchedulerRunner struct {
	svc *service.TaskService
}

func (r taskServiceSchedulerRunner) RunTask(ctx context.Context, taskID int64, trigger string) error {
	_, err := r.svc.RunTask(ctx, taskID, trigger)
	return err
}

// Run 启动服务器（默认监听 SIGINT/SIGTERM）
func Run(cfg *config.Config) error {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(quit)
	return RunWithShutdown(cfg, quit)
}

// RunWithShutdown 启动服务器，通过 stop channel 控制关闭
func RunWithShutdown(cfg *config.Config, stop <-chan os.Signal) error {
	// 初始化日志
	if err := logger.Init(cfg.GetLogLevel(), cfg.GetLogOutput()); err != nil {
		return err
	}
	defer logger.Sync()

	logger.Info("启动RcloneFlow服务",
		zap.String("addr", cfg.GetServerAddr()),
		zap.String("data_dir", cfg.GetDataDir()),
		zap.String("log_level", cfg.GetLogLevel()),
	)

	// 初始化数据库
	db, err := store.Open(cfg.GetDataDir())
	if err != nil {
		logger.Error("数据库初始化失败", zap.Error(err))
		return err
	}

	// 初始化 logger
	maybeStartEmbeddedRC()
	// 初始化rclone客户端
	rc := adapter.NewRcloneClient(nil)

	// 清空所有运行状态（容器重启后恢复，防止单例模式误判）
	if err := db.ClearAllRunningStatus(); err != nil {
		logger.Warn("清空运行状态失败", zap.Error(err))
	}

	// 初始化服务层（定时任务采用 TaskService 以使用 CLI Runner，确保产生日志 stderrFile）
	activeMgr := active_transfer.NewManager()
	activeMgr.SetPersistFunc(func(runID int64, snap active_transfer.ActiveTransferSnapshot) {
		_ = db.UpdateRun(runID, func(rr *store.Run) {
			if rr.Summary == nil {
				rr.Summary = map[string]any{}
			}
			for k, v := range active_transfer.SnapshotEnvelope(snap) {
				rr.Summary[k] = v
			}
		})
		websocket.Broadcast("active_transfer_snapshot", map[string]any{
			"run_id":   runID,
			"task_id":  snap.TaskID,
			"snapshot": snap,
		})
	})
	taskSvc := service.NewTaskService(db, activeMgr)
	scheduleSvc := service.NewScheduleService(db)
	runSvc := service.NewRunService(service.NewStoreRunAdapter(db))

	// 初始化控制器
	remoteCtrl := controller.NewRemoteController(rc)
	taskCtrl := controller.NewTaskController(taskSvc, scheduleSvc, runSvc, rc)
	browserCtrl := controller.NewBrowserController(rc)
	activeTransferCtrl := controller.NewActiveTransferController(activeMgr, runSvc)

	// 初始化调度器(需要在controller之前,以便传递)
	// 使用 TaskService 作为 Runner，以统一走 CLI Runner（生成 stderr 日志文件）
	sched := scheduler.NewWithRunner(db, taskServiceSchedulerRunner{svc: taskSvc})
	if err := sched.Start(); err != nil {
		logger.Error("调度器初始化失败", zap.Error(err))
		return err
	}

	scheduleCtrl := controller.NewScheduleController(scheduleSvc, sched)
	runCtrl := controller.NewRunController(runSvc, rc)
	fsCtrl := controller.NewFsController(rc)
	authSvc := service.NewAuthService(db)
	authCtrl := controller.NewAuthController(authSvc)

	// 标签服务
	tagSvc := service.NewTagService(db)
	taskSvc.SetTagService(tagSvc)
	if err := tagSvc.RecalcTags(); err != nil {
		logger.Error("标签初始化失败", zap.Error(err))
	}
	tagCtrl := controller.NewTagController(tagSvc)

	// 版本信息
	versionCtrl := controller.NewVersionController(rc)

	// 初始化路由
	r := router.New(remoteCtrl, taskCtrl, browserCtrl, scheduleCtrl, runCtrl, fsCtrl, authCtrl, activeTransferCtrl, tagCtrl, versionCtrl, cfg.GetStaticDir())

	// 注入 settings → cleanup 重排钩子（在声明服务之后再赋值）
	var cleanupSvc *service.CleanupService
	var logCleanupSvc *service.LogCleanupService
	controller.ReplanCleanupHook = func(intervalHours int, retentionDays int) {
		if cleanupSvc != nil {
			cleanupSvc.Replan(intervalHours, retentionDays)
		}
	}
	controller.ReplanLogCleanupHook = func(retentionDays int) {
		if logCleanupSvc != nil {
			logCleanupSvc.Replan(retentionDays)
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 启动历史记录清理服务
	// 清理服务（单实例），供设置保存后重排
	if cfg.GetCleanupInterval() > 0 && cfg.GetCleanupRetention() > 0 {
		cleanupSvc = service.NewCleanupService(
			runSvc,
			time.Duration(cfg.GetCleanupInterval())*time.Hour,
			cfg.GetCleanupRetention(),
		)
		go func() {
			defer func() {
				if r := recover(); r != nil {
					logger.Error("goroutine panic", zap.Any("panic", r))
				}
			}()
			cleanupSvc.Start(ctx)
		}()
		logger.Info("历史记录清理服务已启动",
			zap.Int("interval_hours", cfg.GetCleanupInterval()),
			zap.Int("retention_days", cfg.GetCleanupRetention()))
	}

	// 启动日志清理服务（产品语义上跟随历史记录保留天数）
	logsDir := filepath.Join(cfg.GetDataDir(), "logs")
	logRetention := cfg.GetCleanupRetention()
	if logRetention <= 0 {
		logRetention = 7 // 默认7天
	}
	logCleanupSvc = service.NewLogCleanupService(logsDir, 24*time.Hour, logRetention) // 每天检查一次
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Error("goroutine panic", zap.Any("panic", r))
			}
		}()
		logCleanupSvc.Start(ctx)
	}()

	// 设置路由
	mux := http.NewServeMux()
	r.Setup(mux)

	addr := cfg.GetServerAddr()
	srv := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// 启动服务器
	go func() {
		logger.Info("服务监听中", zap.String("addr", addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("服务器启动失败", zap.Error(err))
		}
	}()

	// 等待关闭信号
	sig := <-stop
	logger.Info("收到关闭信号，开始优雅关闭", zap.String("signal", sig.String()))

	cancel()
	logger.Info("已通知后台服务停止")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("服务器关闭失败", zap.Error(err))
		return err
	}
	logger.Info("服务器已安全关闭")
	return nil
}


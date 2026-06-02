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
	"rcloneflow/internal/webdavserver"
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

	// 初始化嵌入式 rclone
	maybeStartEmbeddedRC()
	
	// 初始化服务、控制器、路由
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	services, err := initServices(cfg, db, ctx)
	if err != nil {
		return err
	}

	mux := setupRouter(cfg, services)
	srv := startHTTPServer(cfg, mux)

	// 自动恢复 WebDAV 服务
	restoreWebDAV(services)

	// 等待关闭信号
	sig := <-stop
	logger.Info("收到关闭信号，开始优雅关闭", zap.String("signal", sig.String()))

	shutdown(services, srv)
	return nil
}

// appServices 包含应用程序的所有服务和控制器
type appServices struct {
	db              *store.DB
	taskSvc         *service.TaskService
	scheduleSvc     *service.ScheduleService
	runSvc          *service.RunService
	authSvc         *service.AuthService
	tagSvc          *service.TagService
	cleanupSvc      *service.CleanupService
	logCleanupSvc   *service.LogCleanupService
	sched           *scheduler.Scheduler
	activeMgr       *active_transfer.Manager
	webdavManager   *webdavserver.Manager
	controllers     []any
}

// initServices 初始化所有服务和控制器
func initServices(cfg *config.Config, db *store.DB, ctx context.Context) (*appServices, error) {
	// 初始化 rclone 客户端
	rc := adapter.NewRcloneClient(nil)

	// 清空所有运行状态（容器重启后恢复，防止单例模式误判）
	if err := db.ClearAllRunningStatus(); err != nil {
		logger.Warn("清空运行状态失败", zap.Error(err))
	}

	// 初始化服务层
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
	authSvc := service.NewAuthService(db)
	tagSvc := service.NewTagService(db)
	
	taskSvc.SetTagService(tagSvc)
	if err := tagSvc.RecalcTags(); err != nil {
		logger.Error("标签初始化失败", zap.Error(err))
	}

	// 初始化调度器
	sched := scheduler.NewWithRunner(db, taskServiceSchedulerRunner{svc: taskSvc})
	if err := sched.Start(); err != nil {
		logger.Error("调度器初始化失败", zap.Error(err))
		return nil, err
	}

	// 初始化控制器
	remoteCtrl := controller.NewRemoteController(rc)
	taskCtrl := controller.NewTaskController(taskSvc, scheduleSvc, runSvc, rc)
	browserCtrl := controller.NewBrowserController(rc)
	activeTransferCtrl := controller.NewActiveTransferController(activeMgr, runSvc)
	scheduleCtrl := controller.NewScheduleController(scheduleSvc, sched)
	runCtrl := controller.NewRunController(runSvc, rc)
	fsCtrl := controller.NewFsController(rc)
	authCtrl := controller.NewAuthController(authSvc)
	tagCtrl := controller.NewTagController(tagSvc)
	versionCtrl := controller.NewVersionController(rc)

	webdavManager := webdavserver.NewManager(cfg.GetDataDir())
	webdavCtrl := controller.NewWebdavController(webdavManager)

	controller.WebDAVRestartHook = func() error {
		if !webdavManager.IsEnabled() {
			return nil
		}
		return webdavManager.Restart()
	}
	controller.WebDAVCacheReplanHook = func(interval string) {
		webdavManager.StopCacheCleanupScheduler()
		webdavManager.StartCacheCleanupScheduler(ctx)
	}
	controller.WebDAVCacheCleanupNowHook = func() error {
		return webdavManager.CacheCleanupNow()
	}

	// 初始化清理服务
	var cleanupSvc *service.CleanupService
	var logCleanupSvc *service.LogCleanupService
	
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

	logsDir := filepath.Join(cfg.GetDataDir(), "logs")
	logRetention := cfg.GetCleanupRetention()
	if logRetention <= 0 {
		logRetention = 7 // 默认7天
	}
	logCleanupSvc = service.NewLogCleanupService(logsDir, 24*time.Hour, logRetention)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Error("goroutine panic", zap.Any("panic", r))
			}
		}()
		logCleanupSvc.Start(ctx)
	}()

	// 注入 settings → cleanup 重排钩子
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

	webdavManager.StartCacheCleanupScheduler(ctx)

	return &appServices{
		db:              db,
		taskSvc:         taskSvc,
		scheduleSvc:     scheduleSvc,
		runSvc:          runSvc,
		authSvc:         authSvc,
		tagSvc:          tagSvc,
		cleanupSvc:      cleanupSvc,
		logCleanupSvc:   logCleanupSvc,
		sched:           sched,
		activeMgr:       activeMgr,
		webdavManager:   webdavManager,
		controllers:     []any{remoteCtrl, taskCtrl, browserCtrl, scheduleCtrl, runCtrl, fsCtrl, authCtrl, activeTransferCtrl, tagCtrl, versionCtrl, webdavCtrl},
	}, nil
}

// setupRouter 设置 HTTP 路由
func setupRouter(cfg *config.Config, services *appServices) *http.ServeMux {
	controllers := services.controllers
	remoteCtrl := controllers[0].(*controller.RemoteController)
	taskCtrl := controllers[1].(*controller.TaskController)
	browserCtrl := controllers[2].(*controller.BrowserController)
	scheduleCtrl := controllers[3].(*controller.ScheduleController)
	runCtrl := controllers[4].(*controller.RunController)
	fsCtrl := controllers[5].(*controller.FsController)
	authCtrl := controllers[6].(*controller.AuthController)
	activeTransferCtrl := controllers[7].(*controller.ActiveTransferController)
	tagCtrl := controllers[8].(*controller.TagController)
	versionCtrl := controllers[9].(*controller.VersionController)
	webdavCtrl := controllers[10].(*controller.WebdavController)

	r := router.New(remoteCtrl, taskCtrl, browserCtrl, scheduleCtrl, runCtrl, fsCtrl, authCtrl, activeTransferCtrl, tagCtrl, versionCtrl, webdavCtrl, services.webdavManager, cfg.GetStaticDir())

	mux := http.NewServeMux()
	r.Setup(mux)
	return mux
}

// startHTTPServer 启动 HTTP 服务器
func startHTTPServer(cfg *config.Config, mux *http.ServeMux) *http.Server {
	addr := cfg.GetServerAddr()
	srv := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  5 * time.Minute,
		WriteTimeout: 30 * time.Minute,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		logger.Info("服务监听中", zap.String("addr", addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("服务器启动失败", zap.Error(err))
		}
	}()

	return srv
}

// shutdown 优雅关闭所有服务
func shutdown(services *appServices, srv *http.Server) {
	logger.Info("已通知后台服务停止")

	// 关闭 WebDAV 服务（保留启用状态，容器重启后自动恢复）
	if services.webdavManager != nil {
		services.webdavManager.Shutdown()
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()
	
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("服务器关闭失败", zap.Error(err))
	}
	logger.Info("服务器已安全关闭")
}

func restoreWebDAV(services *appServices) {
	if services.webdavManager == nil || !services.webdavManager.IsEnabled() {
		return
	}

	services.webdavManager.AutoRestore()
}


package app

import (
	"context"
	"net/http"

	"rcloneflow/internal/config"
	"rcloneflow/internal/controller"
	"rcloneflow/internal/logger"
	"rcloneflow/internal/rclone"
	"rcloneflow/internal/router"
	"rcloneflow/internal/scheduler"
	"rcloneflow/internal/service"
	"rcloneflow/internal/store"

	"go.uber.org/zap"
)

// Run 启动服务器
func Run(cfg *config.Config) error {
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

	// 初始化rclone客户端
	rc := rclone.NewFromEnv()

	// 初始化服务层
	taskRunner := adapter.NewTaskRunner(rc)
	taskSvc := service.NewTaskService(db, taskRunner)
	scheduleSvc := service.NewScheduleService(db)
	runSvc := service.NewRunService(service.NewStoreRunAdapter(db))

	// 初始化控制器
	remoteCtrl := controller.NewRemoteController(rc)
	taskCtrl := controller.NewTaskController(taskSvc, rc)
	browserCtrl := controller.NewBrowserController(rc)
	scheduleCtrl := controller.NewScheduleController(scheduleSvc)
	runCtrl := controller.NewRunController(runSvc, rc)
	fsCtrl := controller.NewFsController(rc)

	// 初始化路由
	r := router.New(remoteCtrl, taskCtrl, browserCtrl, scheduleCtrl, runCtrl, fsCtrl)

	// 初始化调度器
	sched := scheduler.New(db, taskRunner)
	if err := sched.Start(); err != nil {
		logger.Error("调度器初始化失败", zap.Error(err))
		return err
	}

	// 启动任务状态同步服务（定期从rclone job API同步状态到数据库）
	jobSync := service.NewJobSyncService(db, rc, cfg.GetPoolInterval())
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go jobSync.Start(ctx)

	// 设置路由
	mux := http.NewServeMux()
	r.Setup(mux)

	// 添加中间件
	handler := withCORS(mux)

	// 启动服务器
	addr := cfg.GetServerAddr()
	logger.Info("服务监听中", zap.String("addr", addr))
	return http.ListenAndServe(addr, handler)
}

// withCORS CORS中间件
func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,OPTIONS")
		if r.Method == http.MethodOptions {
			return
		}
		next.ServeHTTP(w, r)
	})
}

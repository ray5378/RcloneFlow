package app

import (
	"context"
	"net/http"
	"time"

	"rcloneflow/internal/config"
	"rcloneflow/internal/controller"
	"rcloneflow/internal/logger"
	"rcloneflow/internal/rclone"
	"rcloneflow/internal/router"
	"rcloneflow/internal/scheduler"
	"rcloneflow/internal/service"
	"rcloneflow/internal/store"
	clirunner "rcloneflow/internal/runner/cli"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
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

	// 创建默认管理员账户
	createDefaultAdmin(db)

	// 确保事件采样表存在（用于 CLI 进度落库，可幂等）
	if _, err := db.Raw().Exec(`CREATE TABLE IF NOT EXISTS run_events (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		run_id INTEGER NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		bytes REAL DEFAULT 0,
		total_bytes REAL DEFAULT 0,
		percent REAL DEFAULT 0,
		speed_bps REAL DEFAULT 0,
		eta_sec INTEGER DEFAULT 0,
		FOREIGN KEY (run_id) REFERENCES runs(id) ON DELETE CASCADE
	)`); err != nil {
		logger.Warn("run_events 创建失败（不影响主流程）", zap.Error(err))
	} else {
		_, _ = db.Raw().Exec(`CREATE INDEX IF NOT EXISTS idx_run_events_run_id ON run_events(run_id)`) 
		_, _ = db.Raw().Exec(`CREATE INDEX IF NOT EXISTS idx_run_events_created_at ON run_events(created_at)`) 
	}

	// 初始化rclone客户端
	rc := rclone.NewFromEnv()

	// 初始化服务层（任务运行切到 CLI 直控）
	taskSvc := service.NewTaskService(db, clirunner.NewTaskRunnerAdapter())
	scheduleSvc := service.NewScheduleService(db)
	runSvc := service.NewRunService(service.NewStoreRunAdapter(db))

	// 初始化控制器
	configPath := db.PathJoin("rclone.conf")
	remoteCtrl := controller.NewRemoteController(rc, configPath)
	taskCtrl := controller.NewTaskController(taskSvc, rc)
	browserCtrl := controller.NewBrowserController(rc)
	
	// 初始化调度器（改为注入 CLI Runner）
	sched := scheduler.New(db, scheduler.NewTaskRunner(db, clirunner.NewTaskRunnerAdapter()))
	if err := sched.Start(); err != nil {
		logger.Error("调度器初始化失败", zap.Error(err))
		return err
	}
	
	scheduleCtrl := controller.NewScheduleController(scheduleSvc, sched)
	runCtrl := controller.NewRunController(runSvc, rc)
	fsCtrl := controller.NewFsController(rc)
	authCtrl := controller.NewAuthController(db)

	// 初始化路由
	r := router.New(remoteCtrl, taskCtrl, browserCtrl, scheduleCtrl, runCtrl, fsCtrl, authCtrl)

	// 启动任务状态同步服务（RC 同步仍保留以兼容老数据；CLI 路径通过内存事件主动回写 DB）
	jobSync := service.NewJobSyncService(db, rc, cfg.GetPoolInterval())
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go jobSync.Start(ctx)
	// 附加 CLI 进度持久化监听器（写回 runs.bytes/speed + 事件表）
	_ = service.AttachProgressPersistence(service.NewStoreRunAdapter(db), db)

	// 启动历史记录清理服务
	if cfg.GetCleanupInterval() > 0 && cfg.GetCleanupRetention() > 0 {
		cleanupSvc := service.NewCleanupService(
			runSvc,
			db,
			time.Duration(cfg.GetCleanupInterval())*time.Hour,
			cfg.GetCleanupRetention(),
		)
		go cleanupSvc.Start(ctx)
		logger.Info("历史记录清理服务已启动",
			zap.Int("interval_hours", cfg.GetCleanupInterval()),
			zap.Int("retention_days", cfg.GetCleanupRetention()))
	}

	// 设置路由
	mux := http.NewServeMux()
	r.Setup(mux)
	// 挂载 CLI 运行器的最小路由（临时接入，后续统一风格并替换 RC 实现）

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
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,OPTIONS")
		if r.Method == http.MethodOptions {
			return
		}
		next.ServeHTTP(w, r)
	})
}

// createDefaultAdmin 创建默认管理员账户
func createDefaultAdmin(db *store.DB) {
	// 检查数据库中是否已有任何用户
	users, err := db.ListUsers()
	if err != nil {
		logger.Error("检查用户列表失败", zap.Error(err))
		return
	}

	// 如果已有用户，不创建默认账户
	if len(users) > 0 {
		return
	}

	// 创建默认管理员
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("admin"), bcrypt.DefaultCost)
	if err != nil {
		logger.Error("创建默认管理员密码失败", zap.Error(err))
		return
	}

	_, err = db.CreateUser("admin", string(hashedPassword))
	if err != nil {
		logger.Error("创建默认管理员账户失败", zap.Error(err))
		return
	}

	logger.Info("已创建默认管理员账户: admin / admin")
}

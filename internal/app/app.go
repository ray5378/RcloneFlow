package app

import (
	"log"
	"net/http"
	"os"

	"rcloneflow/internal/controller"
	"rcloneflow/internal/rclone"
	"rcloneflow/internal/router"
	"rcloneflow/internal/scheduler"
	"rcloneflow/internal/service"
	"rcloneflow/internal/store"
)

// Run 启动服务器
func Run() error {
	// 初始化数据库
	dbDir := os.Getenv("APP_DATA_DIR")
	if dbDir == "" {
		dbDir = "./data"
	}
	db, err := store.Open(dbDir)
	if err != nil {
		return err
	}

	// 初始化rclone客户端
	rc := rclone.NewFromEnv()

	// 初始化服务层
	taskSvc := service.NewTaskService(db, rc)
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
	sched := scheduler.New(db, rc)
	if err := sched.Start(); err != nil {
		return err
	}

	// 设置路由
	mux := http.NewServeMux()
	r.Setup(mux)

	// 添加中间件
	handler := withCORS(mux)

	// 启动服务器
	addr := os.Getenv("APP_ADDR")
	if addr == "" {
		addr = ":17870"
	}
	log.Printf("listening on %s", addr)
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

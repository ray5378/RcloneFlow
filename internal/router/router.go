package router

import (
	"net/http"
	"strings"

	"rcloneflow/internal/auth"
	"rcloneflow/internal/controller"
)

// Router 路由定义
type Router struct {
	remoteCtrl   *controller.RemoteController
	taskCtrl     *controller.TaskController
	browserCtrl  *controller.BrowserController
	scheduleCtrl *controller.ScheduleController
	runCtrl      *controller.RunController
	fsCtrl       *controller.FsController
	authCtrl     *controller.AuthController
}

// New 创建路由实例
func New(
	rc *controller.RemoteController,
	taskCtrl *controller.TaskController,
	browserCtrl *controller.BrowserController,
	scheduleCtrl *controller.ScheduleController,
	runCtrl *controller.RunController,
	fsCtrl *controller.FsController,
	authCtrl *controller.AuthController,
) *Router {
	return &Router{
		remoteCtrl:   rc,
		taskCtrl:    taskCtrl,
		browserCtrl: browserCtrl,
		scheduleCtrl: scheduleCtrl,
		runCtrl:     runCtrl,
		fsCtrl:      fsCtrl,
		authCtrl:    authCtrl,
	}
}

// Setup 注册所有路由
func (r *Router) Setup(mux *http.ServeMux) {
	// 健康检查（公开）
	mux.HandleFunc("/healthz", r.remoteCtrl.Healthz)

	// 认证相关（公开）
	mux.HandleFunc("/api/auth/login", r.authCtrl.Login)
	mux.HandleFunc("/api/auth/refresh", r.authCtrl.Refresh)

	// 需要认证的API路由
	apiMux := http.NewServeMux()

	// 修改密码
	apiMux.HandleFunc("/api/auth/change-password", r.authCtrl.ChangePassword)

	// 远程存储相关
	apiMux.HandleFunc("/api/remotes", r.remoteCtrl.HandleRemotes)
	apiMux.HandleFunc("/api/remotes/config/", r.remoteCtrl.HandleRemoteConfig)
	apiMux.HandleFunc("/api/remotes/test", r.remoteCtrl.HandleRemoteTest)
	apiMux.HandleFunc("/api/providers", r.remoteCtrl.HandleProviders)
	apiMux.HandleFunc("/api/config/dump", r.remoteCtrl.HandleConfigDump)
	apiMux.HandleFunc("/api/config/", r.remoteCtrl.HandleConfigActions)
	apiMux.HandleFunc("/api/usage/", r.remoteCtrl.HandleUsage)
	apiMux.HandleFunc("/api/fsinfo/", r.remoteCtrl.HandleFsInfo)

	// 文件浏览器
	apiMux.HandleFunc("/api/browser/list", r.browserCtrl.HandleList)

	// 文件系统操作
	apiMux.HandleFunc("/api/fs/mkdir", r.fsCtrl.HandleMkdir)
	apiMux.HandleFunc("/api/fs/delete", r.fsCtrl.HandleDeleteFile)
	apiMux.HandleFunc("/api/fs/purge", r.fsCtrl.HandlePurge)
	apiMux.HandleFunc("/api/fs/move", r.fsCtrl.HandleMove)
	apiMux.HandleFunc("/api/fs/copy", r.fsCtrl.HandleCopy)
	apiMux.HandleFunc("/api/fs/copyDir", r.fsCtrl.HandleCopyDir)
	apiMux.HandleFunc("/api/fs/moveDir", r.fsCtrl.HandleMoveDir)
	apiMux.HandleFunc("/api/fs/publiclink", r.fsCtrl.HandlePublicLink)

	// 任务管理
	apiMux.HandleFunc("/api/tasks", r.taskCtrl.HandleTasks)
	apiMux.HandleFunc("/api/tasks/", r.taskCtrl.HandleTaskActions)

	// 定时任务
	apiMux.HandleFunc("/api/schedules", r.scheduleCtrl.HandleSchedules)
	apiMux.HandleFunc("/api/schedules/", r.scheduleCtrl.HandleSchedules)

	// 运行记录
	apiMux.HandleFunc("/api/runs", r.runCtrl.HandleRuns)
	apiMux.HandleFunc("/api/runs/active", r.runCtrl.HandleActiveRuns)
	apiMux.HandleFunc("/api/stats/global", r.runCtrl.HandleGlobalStats)
	// CLI 扩展接口：停止与日志下载
	apiMux.HandleFunc("/api/runs/", func(w http.ResponseWriter, req *http.Request){
		if strings.HasSuffix(req.URL.Path, "/stop") { r.runCtrl.HandleRunStopCLI(w, req); return }
		if strings.HasSuffix(req.URL.Path, "/log") { r.runCtrl.HandleRunLog(w, req); return }
		r.runCtrl.HandleRunStatus(w, req)
	})
	apiMux.HandleFunc("/api/jobs/{jobId}/status", r.runCtrl.HandleJobStatus)
	apiMux.HandleFunc("/api/jobs/{jobId}/stop", r.runCtrl.HandleJobStop)
	apiMux.HandleFunc("/api/runs/task/", r.runCtrl.HandleRunsByTask)

	// 应用JWT中间件保护API路由
	protectedMux := auth.JWTMiddleware(apiMux)

	// 注册受保护的路由
	mux.Handle("/api/", protectedMux)

	// 静态文件（公开）
	mux.Handle("/", http.FileServer(http.Dir("./web")))
}

package router

import (
	"net/http"

	"rcloneflow/internal/controller"
)

// Router 路由定义
type Router struct {
	remoteCtrl  *controller.RemoteController
	taskCtrl   *controller.TaskController
	browserCtrl *controller.BrowserController
	scheduleCtrl *controller.ScheduleController
	runCtrl    *controller.RunController
	fsCtrl     *controller.FsController
}

// New 创建路由实例
func New(rc *controller.RemoteController, taskCtrl *controller.TaskController, browserCtrl *controller.BrowserController, scheduleCtrl *controller.ScheduleController, runCtrl *controller.RunController, fsCtrl *controller.FsController) *Router {
	return &Router{
		remoteCtrl:   rc,
		taskCtrl:    taskCtrl,
		browserCtrl: browserCtrl,
		scheduleCtrl: scheduleCtrl,
		runCtrl:     runCtrl,
		fsCtrl:      fsCtrl,
	}
}

// Setup 注册所有路由
func (r *Router) Setup(mux *http.ServeMux) {
	// 健康检查
	mux.HandleFunc("/healthz", r.remoteCtrl.Healthz)

	// 远程存储相关
	mux.HandleFunc("/api/remotes", r.remoteCtrl.HandleRemotes)
	mux.HandleFunc("/api/remotes/config/", r.remoteCtrl.HandleRemoteConfig)
	mux.HandleFunc("/api/remotes/test", r.remoteCtrl.HandleRemoteTest)
	mux.HandleFunc("/api/providers", r.remoteCtrl.HandleProviders)
	mux.HandleFunc("/api/config/dump", r.remoteCtrl.HandleConfigDump)
	mux.HandleFunc("/api/config/", r.remoteCtrl.HandleConfigActions)
	mux.HandleFunc("/api/usage/", r.remoteCtrl.HandleUsage)
	mux.HandleFunc("/api/fsinfo/", r.remoteCtrl.HandleFsInfo)

	// 文件浏览器
	mux.HandleFunc("/api/browser/list", r.browserCtrl.HandleList)

	// 文件系统操作
	mux.HandleFunc("/api/fs/mkdir", r.fsCtrl.HandleMkdir)
	mux.HandleFunc("/api/fs/delete", r.fsCtrl.HandleDeleteFile)
	mux.HandleFunc("/api/fs/purge", r.fsCtrl.HandlePurge)
	mux.HandleFunc("/api/fs/move", r.fsCtrl.HandleMove)
	mux.HandleFunc("/api/fs/copy", r.fsCtrl.HandleCopy)
	mux.HandleFunc("/api/fs/copyDir", r.fsCtrl.HandleCopyDir)
	mux.HandleFunc("/api/fs/moveDir", r.fsCtrl.HandleMoveDir)
	mux.HandleFunc("/api/fs/publiclink", r.fsCtrl.HandlePublicLink)

	// 任务管理
	mux.HandleFunc("/api/tasks", r.taskCtrl.HandleTasks)
	mux.HandleFunc("/api/tasks/", r.taskCtrl.HandleTaskActions)

	// 定时任务
	mux.HandleFunc("/api/schedules", r.scheduleCtrl.HandleSchedules)
	mux.HandleFunc("/api/schedules/", r.scheduleCtrl.HandleSchedules)

	// 运行记录
	mux.HandleFunc("/api/runs", r.runCtrl.HandleRuns)
	mux.HandleFunc("/api/runs/active", r.runCtrl.HandleActiveRuns)
	mux.HandleFunc("/api/runs/", r.runCtrl.HandleRunStatus)

	// 静态文件
	mux.Handle("/", http.FileServer(http.Dir("./web")))
}

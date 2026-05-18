package router

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"rcloneflow/internal/auth"
	"rcloneflow/internal/controller"
	"rcloneflow/internal/websocket"
)

// Router 路由定义
type Router struct {
	remoteCtrl   *controller.RemoteController
	taskCtrl     *controller.TaskController
	browserCtrl  *controller.BrowserController
	scheduleCtrl *controller.ScheduleController
	runCtrl            *controller.RunController
	fsCtrl             *controller.FsController
	authCtrl           *controller.AuthController
	activeTransferCtrl *controller.ActiveTransferController
	staticDir          string
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
	activeTransferCtrl *controller.ActiveTransferController,
	staticDir string,
) *Router {
	return &Router{
		remoteCtrl:   rc,
		taskCtrl:     taskCtrl,
		browserCtrl:  browserCtrl,
		scheduleCtrl: scheduleCtrl,
		runCtrl:            runCtrl,
		fsCtrl:             fsCtrl,
		authCtrl:           authCtrl,
		activeTransferCtrl: activeTransferCtrl,
		staticDir:          staticDir,
	}
}

// Setup 注册所有路由
func (r *Router) Setup(mux *http.ServeMux) {
	// 健康检查（公开）
	mux.HandleFunc("/healthz", r.remoteCtrl.Healthz)

	// WebSocket（公开，用于实时推送）
	mux.Handle("/ws", websocket.NewHandler(websocket.GetHub()))

	// Webhook 触发（公开）
	mux.HandleFunc("/webhook/", controller.NewWebhookController(r.taskCtrl.Service()).HandleTrigger)

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
	apiMux.HandleFunc("/api/tasks/bootstrap", r.taskCtrl.HandleBootstrap)
	apiMux.HandleFunc("/api/tasks", r.taskCtrl.HandleTasks)
	apiMux.HandleFunc("/api/tasks/", func(w http.ResponseWriter, req *http.Request) {
		if strings.HasSuffix(req.URL.Path, "/active-transfer/completed") {
			r.activeTransferCtrl.HandleCompleted(w, req)
			return
		}
		if strings.HasSuffix(req.URL.Path, "/active-transfer/pending") {
			r.activeTransferCtrl.HandlePending(w, req)
			return
		}
		if strings.HasSuffix(req.URL.Path, "/active-transfer") {
			r.activeTransferCtrl.HandleOverview(w, req)
			return
		}
		if strings.HasSuffix(req.URL.Path, "/kill") {
			r.runCtrl.HandleTaskKill(w, req)
			return
		}
		r.taskCtrl.HandleTaskActions(w, req)
	})

	// 定时任务
	apiMux.HandleFunc("/api/schedules", r.scheduleCtrl.HandleSchedules)
	apiMux.HandleFunc("/api/schedules/", r.scheduleCtrl.HandleSchedules)

	// 运行记录
	apiMux.HandleFunc("/api/runs", r.runCtrl.HandleRuns)
	apiMux.HandleFunc("/api/runs/active", r.runCtrl.HandleActiveRuns)
	apiMux.HandleFunc("/api/stats/global", r.runCtrl.HandleGlobalStats)

	// 设置中心
	apiMux.HandleFunc("/api/settings", controller.NewSettingsController().HandleSettings)
	// CLI 扩展接口：停止/强杀/日志下载/文件明细
	apiMux.HandleFunc("/api/runs/", func(w http.ResponseWriter, req *http.Request) {
		if strings.HasSuffix(req.URL.Path, "/stop") {
			r.runCtrl.HandleRunStopCLI(w, req)
			return
		}
		if strings.HasSuffix(req.URL.Path, "/kill") {
			r.runCtrl.HandleRunKillCLI(w, req)
			return
		}
		if strings.HasSuffix(req.URL.Path, "/files") {
			r.runCtrl.HandleRunFiles(w, req)
			return
		}
		if strings.HasSuffix(req.URL.Path, "/log") {
			r.runCtrl.HandleRunLog(w, req)
			return
		}
		r.runCtrl.HandleRunStatus(w, req)
	})
	apiMux.HandleFunc("/api/runs/task/", r.runCtrl.HandleRunsByTask)

	// 应用JWT中间件保护API路由
	protectedMux := auth.JWTMiddleware(apiMux)

	// 注册受保护的路由
	mux.Handle("/api/", protectedMux)

	// 静态文件（公开）
	mux.Handle("/", staticFileHandler(r.staticDir))
}

func staticFileHandler(staticDir string) http.Handler {
	if _, err := os.Stat(staticDir); err != nil {
		message := fmt.Sprintf(`<!doctype html><html lang="zh-CN"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>RcloneFlow</title><style>body{font-family:system-ui,-apple-system,BlinkMacSystemFont,"Segoe UI",sans-serif;margin:0;min-height:100vh;display:flex;align-items:center;justify-content:center;background:#111827;color:#e5e7eb}main{max-width:720px;padding:32px}code{background:#1f2937;padding:2px 6px;border-radius:4px}a{color:#93c5fd}</style></head><body><main><h1>前端构建产物缺失</h1><p>当前静态目录 <code>%s</code> 不存在。请先执行 <code>cd frontend && npm run build</code>，或使用 Docker 镜像启动。</p></main></body></html>`, staticDir)
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte(message))
		})
	}
	return http.FileServer(http.Dir(staticDir))
}

# AGENTS.md — RcloneFlow

RcloneFlow = rclone + Web UI。Go (Gin) 后端 + Vue 3 / TypeScript / Vite 前端。SQLite 数据库。默认端口 `17870`。

## 分支策略

仅保留 `dev`（开发）与 `master`（稳定）两个分支。发布标签仅保留 `v1.0.0`。

## 开发者命令

### 后端
```bash
go build -o server ./cmd/server   # 编译（自动嵌入 git hash）
./server                           # 运行（读取 config.yaml 或环境变量）
```

### 前端
```bash
cd frontend
npm install
npm run dev     # 开发服务器 :4200，代理 /api → localhost:17870
npm run build   # 输出到 ../web/
npm run test    # vitest 运行（happy-dom，覆盖 src/api/）
```

### Docker
```bash
docker compose up -d --build       # 完整构建 + 启动（自动嵌入 git hash）
docker build --no-cache -t ray5378/rcloneflow:dev .
```

### CI (GitHub Actions)
推送到 `master` 或 `dev` 触发 `.github/workflows/docker.yml`：Go 1.25 预检构建 → 推送 Docker Hub。需要 `DOCKERHUB_USERNAME` + `DOCKERHUB_TOKEN` 密钥。

## 架构概览

```
cmd/server/main.go          → config.Load → app.Run()
internal/app/app.go         → DB 初始化、内置 RC、调度器、HTTP 服务器
internal/router/router.go   → 路由注册
internal/controller/*       → HTTP 处理器
internal/service/*          → 业务逻辑
internal/runnercli/         → CLI 执行器（执行 rclone，解析 stdout 进度）
internal/rclone/            → rclone RC 客户端
internal/store/             → SQLite（单连接：SetMaxOpenConns(1)）
internal/scheduler/         → cron 调度器
internal/websocket/         → WS 广播 active_transfer_snapshot
frontend/src/views/         → TaskView, RunView, BrowserView, ScheduleView, LoginView
frontend/src/composables/   → 60+ 组合函数，重点：useRunningHint, useActiveRunLookup, activeRunProgress, useWebSocket, useTaskViewRuntime, useTaskFormRuntime
frontend/src/api/           → API 请求层（15 个文件）
web/                        → 前端构建产物（自动生成，已 gitignore）
```

## 关键约定与易错点

### 进度数据链（最高风险区域）
运行中 UI 必须读取 `/api/runs/active.progress`（`progress` 字段）。
- `progress` = 实时解析日志帧 — **运行中任务的主数据源**
- `finalSummary` = 运行结束时冻结 — 仅用于历史页面
- `preflight` = 预估值 — 不得驱动运行中展示
- 解析仅接受完整聚合单行统计；文件级行（`Copied (new)`、`Deleted` 等）不得当作总进度
- 大小配对解析需要显式字节单位（MiB/GiB），避免匹配到时间戳片段如 `2026/04`
- `stableProgress` 已从代码中删除，不再使用

### SQLite 单连接模型
`store.Open()` 使用 `SetMaxOpenConns(1)` / `SetMaxIdleConns(1)`。任何新增的 DB 打开点必须沿用此模式。即使启用 WAL + busy_timeout 也不得放松连接池。

### 前端弹窗
- **禁止使用** `alert()`、`confirm()`、`prompt()`
- 所有用户提示必须使用自定义 Vue Modal 或 Toast 组件

### WebSocket 匹配
`run_progress` WS 更新必须按 `runRecord.id` 匹配，而非 task ID。

### 大型 Vue 文件 — 尾部碎片风险
`TaskView.vue`、`TaskCard.vue`、`RunItem.vue`、`TaskHistoryViewShell.vue` 是大型文件，编辑后容易出现重复 `</style>` 或孤立片段。编辑后务必检查文件尾部结构。如已损坏，重写为单份干净结尾 — 不要在坏尾上修补。

### TaskView.vue 组合函数顺序依赖
必须保持以下顺序：
1. `useTaskFormRuntime(...)` → 然后 `useTaskViewModalBindings(...)`
2. `useTaskViewAuxRuntime(...)` → 然后 `useRunningHintRuntime(...)`

不得为了美观重排声明顺序 — 会导致生产构建报 `Cannot access 'X' before initialization`。

### Browser 与 FS 控制器
- `internal/controller/browser.go` → rclone RC 浏览
- `internal/controller/fs_cli.go` → CLI 文件操作
不要混淆这两条链路。

### 构建产物
`web/` 由 `frontend/` 中 `npm run build` 生成，已 gitignore。本地非 Docker 运行时需先构建前端，否则服务器会提供过期或缺失的静态资源。

### 内置 rclone RC
默认通过 `EMBED_RC=true` 启用，用于 remotes/providers/config/browser。可通过 `EMBED_RC=false` 关闭。

### 默认账号
用户名：`admin` / 密码：`admin`（无用户时首次启动自动创建）。

### Docker 镜像
运行时镜像基于 Alpine 3.19，安装了 `bash`、`busybox`、`curl`、`wget`、`sqlite-libs`。需要提取文件时可直接 `docker exec`。

### 构建哈希自动嵌入
Go 编译时通过 `internal/version.CommitHash` 嵌入 git hash。Docker 构建自动从 `.git/refs/heads/*` 获取；本地构建由 `init()` 回退到 `git rev-parse`。无需手动传参。

### 不要挂载覆盖 `/app/web`
覆盖 `/app/web` 的卷挂载会隐藏内置前端。

## 环境变量

| 变量 | 默认值 | 说明 |
|---|---|---|
| `APP_ADDR` | `:17870` | 服务器监听地址 |
| `APP_DATA_DIR` | `/app/data` | 数据目录（SQLite、日志） |
| `APP_STATIC_DIR` | `./web` | 前端静态文件 |
| `RCLONE_RC_URL` | `http://127.0.0.1:5572` | rclone RC 端点 |
| `RCLONE_RC_USER` | | RC 用户名（可选） |
| `RCLONE_RC_PASS` | | RC 密码（可选） |
| `EMBED_RC` | `true` | 启动内置 rclone RC |
| `LOG_LEVEL` | `info` | debug\|info\|warn\|error |
| `LOG_OUTPUT` | `stdout` | stdout\|stderr\|文件路径 |

## 高风险文件（编辑时需格外小心）

- `frontend/src/views/TaskView.vue`
- `internal/controller/run.go`
- `internal/runnercli/runner.go`
- `frontend/src/composables/useWebSocket.ts`

## 延伸阅读

- `docs/KNOWN_PITFALLS.md` — 反复出现的坑点及示例
- `docs/ARCHITECTURE_OVERVIEW.md` — 详细架构说明
- `README-dev.md` — API 端点、数据库结构、Webhook 格式
- `docs/TESTING_GUIDE.md` — 测试规范

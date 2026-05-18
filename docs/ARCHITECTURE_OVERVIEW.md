# ARCHITECTURE_OVERVIEW.md

项目架构总览。

这份文档的目标，是帮助开发者快速理解这个项目的整体结构、主要链路和高风险区域，而不是一上来就陷入具体代码细节。

---

## 1. 项目整体形态

这是一个包含前端页面、后端接口、rclone 执行链、任务调度与实时状态更新的完整应用。

从整体上看，可以分成五层：

1. **前端界面层**
2. **HTTP / WebSocket 接口层**
3. **业务聚合层**
4. **执行与适配层**
5. **存储与配置层**

---

## 2. 入口与主链

### 2.1 后端入口
- `cmd/server/main.go`
- 启动主链：`config.Load("") -> app.Run(cfg)`

### 2.2 路由入口
- `internal/router/router.go`
- 这里集中注册 HTTP 路由与相关控制器

### 2.3 前端入口
- `frontend/src/main.ts`
- 根组件：`frontend/src/App.vue`

---

## 3. 前端结构概览

### 3.1 页面层
主要页面位于：
- `frontend/src/views/TaskView.vue`
- `frontend/src/views/RunView.vue`
- `frontend/src/views/BrowserView.vue`
- `frontend/src/views/ScheduleView.vue`
- `frontend/src/views/LoginView.vue`

### 3.2 组件层
可复用 UI 组件位于：
- `frontend/src/components/*`
- 其中任务相关拆分已形成：
  - `components/task/RunningHintModal.vue`
  - `components/task/runningHint.ts`

### 3.3 组合逻辑层
可复用状态与组合逻辑位于：
- `frontend/src/composables/*`
- 当前运行中相关重点：
  - `useRunningHint.ts`
  - `useActiveRunLookup.ts`
  - `activeRunProgress.ts`
  - `useWebSocket.ts`

### 3.4 API 请求层
前端接口封装位于：
- `frontend/src/api/*`

---

## 4. 后端结构概览

### 4.1 controller 层
位于：
- `internal/controller/*`

负责：
- HTTP 请求与响应
- 接口字段组装
- 部分状态聚合与契约输出

当前运行中任务链最关键的控制器之一：
- `internal/controller/run.go`

### 4.2 service 层
位于：
- `internal/service/*`

负责：
- 业务规则
- 状态聚合
- 任务、运行、调度等流程协调

### 4.3 执行与适配层
位于：
- `internal/runnercli/*`
- `internal/adapter/*`
- `internal/rclone/*`

其中当前主执行链重点是：
- `internal/runnercli/runner.go`

### 4.4 存储层
位于：
- `internal/store/*`
- `internal/dao/*`

负责：
- 数据存储
- 任务、运行、调度记录读写

### 4.5 其他基础层
- 配置：`internal/config/*`
- 鉴权：`internal/auth/*`
- 日志：`internal/logger/*`
- WebSocket：`internal/websocket/*`
- 调度：`internal/scheduler/*`

---

## 5. 当前最关键的数据链之一：运行中进度链

这是目前项目里最敏感、也最容易出回归的链路之一。

固定排查顺序：

`日志原文 -> summary.progress -> /api/runs/active.progress -> 前端展示`

### 5.1 日志来源
rclone one-line progress 日志是运行中真实进度的重要来源。

### 5.2 runner 解析层
核心位置：
- `internal/runnercli/runner.go`

负责：
- 消费日志输出
- 解析 one-line progress
- 写入运行摘要中的 `summary.progress`

### 5.3 active runs 接口层
核心位置：
- `internal/controller/run.go`
- 接口：`/api/runs/active`

负责：
- 把运行记录与 progress 聚合成前端实时展示主数据
- 输出 `progress` / `progressLine` / `progressCheck` 等字段

### 5.4 前端使用层
核心位置：
- `TaskView.vue`
- `useRunningHint.ts`
- `useActiveRunLookup.ts`
- `activeRunProgress.ts`

当前约定：
- 运行中 UI 主数据源是 `/api/runs/active.progress`
- 任务卡片完成态现在由前端单份冻结帧 `completedFreezeByTask` 承接
- 前端运行中 helper / runtime / refresh 命名已开始收口到 `runningProgress` / `progress` 语义
- `preflight` 只保留预估语义

---

## 6. 另外两条重要链路

### 6.1 文件浏览 / 文件操作链路
当前已明确区分：
- `internal/controller/browser.go`：走 rclone RC 浏览
- `internal/controller/fs_cli.go`：走 CLI 文件操作

不要混淆“浏览”和“文件操作”两条链。

### 6.2 实时更新链路
前端运行中刷新来自两部分：
- WebSocket 实时推送
- 轻量轮询兜底

关键前端位置：
- `useWebSocket.ts`
- `TaskView.vue`

关键约定：
- `run_progress` 更新 active run 时，必须按 `runRecord.id` 匹配

---

## 7. 当前已知高风险区域

以下区域改动时默认高风险：

- `frontend/src/views/TaskView.vue`
- `internal/controller/run.go`
- `internal/runnercli/runner.go`
- `frontend/src/composables/useWebSocket.ts`
- 运行中相关 UI 组件与 composable

原因通常包括：
- 多层状态链汇聚
- 用户可直接看到效果
- 回退逻辑较多
- 历史兼容层仍未完全清理

---

## 8. 当前结构性注意事项

### 8.1 `TaskView.vue` 仍偏大
虽然已经拆出一部分 running hint 与 active run 逻辑，但仍属于需要继续拆分的高风险文件。

### 8.2 `progress / finalSummary / preflight` 容易被误用
这三个字段语义不同，后续开发时必须严格区分。

### 8.3 仓库中仍有旧 runner 痕迹
当前主执行链重点是：
- `internal/runnercli/runner.go`

而：
- `internal/app/runner.go`

属于容易误导维护者的旧路径，后续需继续标注或清理。

---

## 9. 推荐理解顺序

如果你是第一次接触这个项目，建议按下面顺序理解：

1. `cmd/server/main.go`
2. `internal/router/router.go`
3. 前端主页面 `TaskView.vue`
4. `internal/controller/run.go`
5. `internal/runnercli/runner.go`
6. 再回看相关文档：
   - `ENGINEERING_RULES.md`
   - `TECH_DEBT.md`
   - `FRONTEND_RULES.md`
   - `BACKEND_RULES.md`

---

## 10. 一个简单原则

如果你正在改某个问题，但还说不清：
- 数据从哪里来
- 中间经过哪些层
- 最后在哪里展示

那通常说明你还不该急着改代码，应该先把链路走通。
通。
��路走通。
通。
��。

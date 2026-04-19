# TECH_DEBT.md

开发向技术债与重构清单。这个文档不面向最终用户，主要用于避免后续开发把当前已理顺的链路重新绕乱。

## 当前已确认的开发约定

### 分支约定
- `dev`：开发分支
- `master`：主线 / 稳定分支
- 正确流程：`dev` 开发 -> 验证 -> `dev -> master`
- 不再使用 `next-master` 作为主线
- 不应再执行 `master -> dev` 反向合并

### 运行中进度链约定
- 运行中 UI 主数据源：`/api/runs/active.progress`
- `progress`：运行中的实时进度（live frame）
- `stableProgress`：兼容字段 / 完成态固化，不应再作为运行中主数据源
- `preflight`：预估总量（来自 `rclone size` + 过滤条件），仅用于预估展示，不应再作为运行中主展示主源
- 当前 active runs 调试字段：
  - `progressLine`
  - `progressSource`
  - `progressMismatch`
  - `progressCheck`

### 运行中 ETA 约定
- 当前运行中 ETA 只使用 live progress：
  - `live.totalBytes`
  - `live.bytes`
  - `live.speed`
- 不再回退 `preflight.totalBytes`

## 已确认并已修复的问题（供后续避免回归）

### 1. 运行中体积显示错成 `4 B / 4 B`
根因：
- `parseOneLineProgress()` 中 `bytesPairRe` 过于宽松
- 会把日志时间戳里的 `2026/04` 误匹配成 `bytes/totalBytes`

现状：
- 已要求 aggregate size pair 必须带明确字节单位（如 `MiB/GiB`）
- 已补单元测试防止回归

### 2. 任务卡片不自动刷新
根因：
- `run_progress` 的 WebSocket 更新按错误字段匹配 active run
- 之前用了 `r.id === msg.data.run_id`
- 实际结构应为 `r.runRecord.id`

现状：
- 已修成按 `runRecord.id` 更新
- 已增加轻量兜底轮询

### 3. 运行中进度链来源混乱
根因：
- 前端曾对 `stableProgress` 做二次拼接
- 后端和前端都存在 `progress / stableProgress / preflight` 混用

现状：
- 运行中主展示已统一优先使用 `progress`
- `stableProgress` 仅保留兼容 / 完成态固化
- `preflight` 仅保留预估语义

### 4. 前端 build warning
已处理：
- 删除无效的 `queueApi` 包装和残留引用
- 修复 `frontend/src/styles/global.css` 中缺失的 `}`

## 临时测试提速方案（后续应剔除）

### 当前临时方案目的
为了在网络不稳定、Docker Hub / 包管理源偶发超时的情况下，快速验证构建流程，仓库当前临时引入了“本地缓存优先”的测试提速方案。

补充当前测试工作流约定：
- 每次真正到达用户测试点时，先执行 `docker build --no-cache -t ray5378/rcloneflow:latest .`
- 当前规则改为：构建开始后无需全程盯看，默认在约 5 分钟时主动检查一次结果；无论该次检查结果是成功、失败还是仍未完成，都直接把结果告知用户，不再自动重试
- 该 5 分钟节点的检查与通知必须由 Agent 主动触发，不能依赖用户再次追问构建状态
- Agent 需要先自己检查构建结果，再决定是通知用户测试还是告知用户失败/未完成状态

这套方案的目标是：
- 降低 `docker build` 对外网的依赖
- 方便快速回归测试
- 不是长期默认工程方案

### 当前临时缓存目录
- `third_party/docker/`：基础镜像 tar 缓存
- `third_party/npm-cache/`：npm cache
- `third_party/go-mod-cache/`：Go module cache
- `third_party/apk-cache/`：Alpine apk 包缓存

### 当前临时接线点
- `Dockerfile`
  - webbuilder：优先使用 `third_party/npm-cache`
  - gobuilder：优先使用 `third_party/apk-cache`
  - gobuilder：优先使用 `third_party/go-mod-cache`
- `scripts/docker/load-base-images.sh`
  - 预加载 `node:18-alpine`
  - 预加载 `golang:1.25-alpine`
  - 预加载 `alpine:3.19`

### 后续剔除原则
当以下条件满足时，应优先回收这套临时方案：
- Docker 基础镜像拉取稳定
- npm / Go / apk 镜像源稳定
- 不再需要把大缓存文件长期保留在仓库中

优先剔除顺序建议：
1. `third_party/npm-cache/`
2. `third_party/go-mod-cache/`
3. `third_party/apk-cache/`
4. `third_party/docker/*.tar`
5. `Dockerfile` 中对应的临时本地缓存优先逻辑

## 拆分进度（截至当前主线）

补充说明：
- 更适合长期持续更新的拆分作战视图，已单独整理到：`docs/TASKVIEW_SPLIT_MAP.md`
- `TECH_DEBT.md` 继续侧重记录：已拆结果、关键回归、待重构方向
- `TASKVIEW_SPLIT_MAP.md` 继续侧重记录：已拆/半拆/未拆状态、推荐拆分顺序、每一步重点测试项

### `frontend/src/views/TaskView.vue` 已完成的拆分
以下内容已经从 `TaskView.vue` 中拆出或收口：

补充到当前阶段：
- 历史详情弹窗相关主干已进一步拆为四条链：
  - `frontend/src/composables/useRunDetailComputed.ts`
  - `frontend/src/composables/useRunDetailFiles.ts`
  - `frontend/src/composables/useRunDetailState.ts`
  - `frontend/src/composables/useRunDetailEntry.ts`
- 当前 `TaskView.vue` 在历史详情弹窗这一块已更接近装配层；页面层主要保留模板装配、事件转发和少量页面级连接线
- 创建任务表单区当前也已进入更完整的半拆阶段：`AddTaskForm.vue` 已承接表单主模板，`useTaskFormState.ts` 已承接基础表单状态与新建/编辑入口回填链，`useTaskFormPrepare.ts` 已承接命令行模式提交前准备与统一前置校验，`useTaskCommandParse.ts` 已承接命令解析链，`useTaskFormSubmit.ts` 已承接提交编排主干与成功/失败收尾，`useTaskFormFlow.ts` 已承接 `createTask()` 最后一层入口编排，`useTaskPathBrowse.ts` 已承接路径浏览状态、breadcrumb、加载、重置与编辑态恢复，`useTaskFormEntry.ts` 已承接新建/编辑任务入口编排，`frontend/src/components/task/scheduleOptions.ts` 已承接定时规则纯逻辑与字段全集
- 任务列表外层装配链当前也已进入半拆阶段：`useTaskListView.ts` 已承接搜索/分页/跳页/派生列表，`useTaskViewUi.ts` 已承接菜单与确认弹窗 UI 壳，`useTaskHistoryActions.ts` 已承接历史删除动作主干，`useTaskListActions.ts` 已承接任务列表行为入口壳，`useTaskRunActions.ts` 已承接运行/停止入口壳
- 页面级数据加载 / WebSocket / 刷新协调当前也已开始半拆：`useTaskViewDataSync.ts` 已承接 `loadData()`、`loadActiveRuns()`、`loadGlobalStats()`、`openGlobalStats()` 与 `run_status` / `run_progress` WebSocket 接线第一层；`useTaskProgressSync.ts` 已继续承接 `getDbProgressStable()`、`getDeNoisedStableByRun()`、`getDeNoisedStableByTask()`、`formatBps()`、`calcEtaFromAvg()`、`triggerAutoRefresh()`；本轮新增 `useTaskViewRefreshLifecycle.ts`，已继续承接 stuck 检测定时器、activeRuns 轮询兜底定时器和生命周期清理。当前 `TaskView.vue` 在这块更接近页面装配层，只剩少量 glue code 与零散连接线
- 本轮再次验证出一条必须长期保留的经验：`TaskView.vue` 在新增 composable 后，最危险的不是实现本身，而是页面接线顺序、顶部 import、旧本地实现残留和文件尾脏片；这些问题会直接引发 `xxx is not defined`、`Cannot access ... before initialization`、页面空白、任务列表空白、`Invalid end tag.`
- 本轮拆分反复验证出一条必须长期保留的经验：新增 composable 后，必须立即核对 `import`、页面解构、旧状态/旧函数清理、调用点是否都切到新来源，否则极易出现 `xxx is not defined`、页面空白、详情打不开等真实运行时回归；本轮就先后出现过 `loadActiveRuns is not defined`、`loadData is not defined`、`useTaskProgressSync is not defined`，后续这条经验也应继续套用到 `useTaskViewRefreshLifecycle.ts`
- 另一条本轮新增经验：页面模板拆成子组件后，原样式不会自动跟随；对于 `AdvancedOptions.vue` 这类内部布局复杂的子组件，必须显式迁移内部样式，否则会出现勾选框错位、标签与输入框布局颠倒、输入项宽度异常等明显 UI 回归

#### 1. 运行中提示小窗 UI
- 已拆出：`frontend/src/components/task/RunningHintModal.vue`
- 目的：将小窗模板、按钮、调试展开区从页面主文件中分离

#### 2. 运行中提示相关格式化 / 调试 helper
- 已拆出：`frontend/src/components/task/runningHint.ts`
- 当前承载：
  - `getActiveProgress()`
  - `getActiveProgressText()`
  - `getActiveProgressLine()`
  - `getActiveProgressCheck()`
  - `getActiveProgressCheckText()`
  - `getActiveProgressJson()`
  - `getRunningHintDebug()`

#### 3. 运行中提示状态管理
- 已拆出：`frontend/src/composables/useRunningHint.ts`
- 当前承载：
  - 小窗开关状态
  - 当前 run 绑定
  - debug 开关状态
  - phase / progress / debugInfo 组装
  - 打开日志动作

#### 4. active run 基础读取
- 已拆出：`frontend/src/composables/useActiveRunLookup.ts`
- 当前承载：
  - `getActiveRunByTaskId()`
  - `getActiveProgressByTaskId()`
  - `getActiveProgressTextByTaskId()`

#### 5. active run fallback / 数值归一化
- 已拆出：`frontend/src/composables/activeRunProgress.ts`
- 当前承载：
  - `getDeNoisedStableByRun()`
  - `getDeNoisedStableByTask()`

### 拆分过程中的关键回归修复（已完成）
以下问题已在拆分过程中修正并验证：
- running hint 小窗强制弹出
- running hint 内容为空 / 调试详情全为 `-`
- running hint 新旧接线混用导致页面空白
- `useRunningHint` / `useActiveRunLookup` 漏导入导致页面运行时报错
- 拆分过程中 `TaskView.vue` 残留旧引用或文件尾部脏片导致 build 失败

### 运行中总量/总数/百分比源头修复（已完成）
这部分不是前端 masking，而是已回到后端源头修复：
- `/api/runs/active` 不再让 early live progress 的偏小 `totalBytes / plannedFiles` 覆盖正确 `preflight`
- 当 `preflight` 接管总量时，百分比按 `bytes / totalBytes` 重新计算
- 已移除前端为止血临时添加的“总量/总数防倒退显示保护”

## 待重构清单

### 高优先级

#### 1. 继续拆分 `frontend/src/views/TaskView.vue`
现状：
- 虽然运行中提示与 active run 相关逻辑已经拆出一部分，但文件整体仍然偏大
- 仍混合了承担任务列表、历史记录、创建任务、弹窗、WebSocket 更新等职责

下一步建议拆分优先块：
- 历史记录区域
- 任务列表区域
- 创建任务区域 / 表单块
- 页面级 WebSocket / 列表刷新协调逻辑

目标：
- 进一步降低耦合
- 让 `TaskView.vue` 更接近页面装配层，而不是功能承载层

#### 2. 明确 `progress / stableProgress / preflight` 的字段边界
现状：
- 虽然主链已经理顺，但仍存在兼容回退逻辑
- 未来维护者仍可能误用字段

建议：
- 在接口类型、注释、文档中继续强化语义约束
- 尽量限制运行中页面直接接触 `stableProgress`

#### 3. 标注或清理旧 runner
现状：
- 当前主执行链是 `internal/runnercli/runner.go`
- 仓库里仍保留 `internal/app/runner.go`

风险：
- 容易误导后续修改者改错地方

建议：
- 至少加注释标明谁是主链、谁是遗留实现
- 长期看应逐步移除旧实现

### 中优先级

#### 4. 运行中调试信息开关化
现状：
- 小窗已支持“展开调试详情”
- 但仍属于偏开发向能力

建议：
- 可进一步改为开发模式开关 / 管理员开关 / 配置项开关

#### 5. 统一任务卡片与历史记录卡片的视觉基底
现状：
- 视觉效果看似相似，但样式来源不完全一致
- 改 hover 时已经暴露出不一致问题

建议：
- 抽出统一的列表项 hover/active 基底样式
- 任务卡片 / 历史项在其上做差异化扩展

#### 6. 规范 `web/` 构建产物流程
现状：
- 分支切换和前端 build 后，`web/index.html` 与 `web/assets/*` 容易弄脏工作区

建议：
- 明确是否提交构建产物
- 若继续提交，应规范清理流程
- 若不提交，应调整构建/发布流程，避免噪音改动

### 低优先级

#### 7. 优化 preflight 文案与用户预期
现状：
- 预估总量与运行中实时总量并非天然完全一致
- 用户可能误以为两者必须完全相等

建议：
- 在 UI 文案上进一步区分：
  - `预估总量`
  - `当前总量`
  - `实时进度`

#### 8. 继续清理历史兼容层
现状：
- 当前为了平滑迁移，仍保留一定兼容回退逻辑

建议：
- 在运行中链路彻底稳定后，再逐步减少不必要回退

#### 9. 把工程规则继续下沉到代码注释 / 契约层
现状：
- 当前 `docs/ENGINEERING_RULES.md` 已有较完整的开发规范

建议：
- 后续关键链路继续加短注释，避免只靠外部文档记忆
��链路继续加短注释，避免只靠外部文档记忆
续加短注释，避免只靠外部文档记忆

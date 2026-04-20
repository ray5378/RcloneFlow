# TECH_DEBT.md

开发向技术债与重构清单。

这份文档不面向最终用户，主要用来回答三件事：
- 哪些历史问题已经解决，后续不要回滚
- 哪些旧结论已经过时，不要继续按旧方向推进
- 当前还剩哪些真正值得做的技术债

---

## 1. 当前已确认的开发约定

### 分支约定
- `dev`：开发分支
- `master`：主线 / 稳定分支
- 正确流程：`dev` 开发 -> 验证 -> `dev -> master`
- 不再使用 `next-master` 作为主线
- 不应再执行 `master -> dev` 反向合并

### 任务运行进度链约定
当前任务进度主链只保留两条：
- WebSocket `run_progress`
- `/api/runs/active`

字段语义：
- `progress`：运行中的实时进度（live frame），是运行中 UI 主数据源
- 任务卡片完成态由前端冻结帧承接，不参与运行中主链
- `preflight`：已从任务卡片主数据链、`/api/runs/active` 兜底链、运行详情主展示链退场，不再参与总量 / 总数 / 百分比主数据计算

当前 active runs 调试字段：
- `progressLine`
- `progressSource`
- `progressMismatch`
- `progressCheck`

### 运行中 ETA / 总量约定
- 当前运行中 ETA 只使用 live progress：
  - `live.totalBytes`
  - `live.bytes`
  - `live.speed`
- 不再回退 `preflight.totalBytes`
- WebSocket `run_progress` 与 `/api/runs/active` 两条链都要求：
  - `totalBytes` 单调不减
  - `totalCount` 单调不减

### 删除任务约定
删除任务不只是删除 `tasks` 主记录。

当前目标语义：
- `DeleteTask` = 删除任务 + 清理历史记录 + 清理关联运行日志文件 + 清空空日志目录壳

当前明确要求：
- 删除任务时，除数据库级联删除 `tasks / schedules / runs` 外，还应显式清理该任务关联 run 的 `summary.stderrFile`

### 日志清理链约定
- 任务运行日志清理：已支持 `LOG_RETENTION_DAYS` 热更新，修改默认配置后无需重启即可重排并立刻执行一次清理
- 程序日志清理：当前仍不算完整应用内能力；若 `LOG_OUTPUT` 走 stdout/stderr，应由 Docker / 宿主日志轮转负责
- 不要把“任务运行日志清理正常”误解成“所有程序日志文件都已由应用统一自动清理”

---

## 2. 已解决，后续不要回滚的问题

### 2.1 运行中体积显示错成 `4 B / 4 B`
根因：
- `parseOneLineProgress()` 中 `bytesPairRe` 过于宽松
- 会把日志时间戳里的 `2026/04` 误匹配成 `bytes/totalBytes`

现状：
- 已要求 aggregate size pair 必须带明确字节单位（如 `MiB/GiB`）
- 已补单元测试防止回归

### 2.2 任务卡片不自动刷新
根因：
- `run_progress` 的 WebSocket 更新按错误字段匹配 active run
- 之前用了 `r.id === msg.data.run_id`
- 实际结构应为 `r.runRecord.id`

现状：
- 已修成按 `runRecord.id` 更新
- 已增加轻量兜底轮询

### 2.3 运行中进度链来源混乱
根因：
- 前端曾对完成态/过渡态字段做二次拼接
- 后端和前端都存在 `progress / finalSummary / preflight` 语义混用

现状：
- 运行中主展示已统一优先使用 `progress`
- 当前代码已删除 `stableProgress`；完成态详情职责收敛到 `finalSummary`，任务卡片完成态则改由前端冻结帧承接
- `preflight` 已从主展示链退场

### 2.4 前端 build warning
已处理：
- 删除无效的 `queueApi` 包装和残留引用
- 修复 `frontend/src/styles/global.css` 中缺失的 `}`

### 2.5 TaskView / 壳组件拆分过程中反复出现的真实回归
以下问题已多轮修复，后续必须优先防回归：
- barrel import 循环依赖（`from './'` / `from '../components/task'` / `from '../components/toast'`）会在压缩后表现为 `Cannot access 'Ir/Bl/... before initialization'`
- `setup` 顶层先用后定义（TDZ）会在压缩后表现为同类 `before initialization` 错误
- modal 下沉做一半会导致“按钮点击无反应”
- modal 表单若直接读未初始化对象，会出现：
  - `Cannot read properties of undefined (reading 'triggerId')`
  - `Cannot set properties of undefined (setting 'manual'/'schedule'/...)`
- `ref` 型表单对象若 setter 没写回 `.value`，会出现“看着改了、实际保存还是旧值”的假保存
- `TaskView.vue` / `RunDetailModal.vue` / 其他大 SFC 在多轮 edit 后容易出现尾部模板/样式残片；遇到 `Invalid end tag.` 时优先整段回正，不做碎补

### 2.6 `rc_job_id` 运行时残留查询
以下问题不是前端假象，而是后端真实残留，已修复：
- `runs.rc_job_id` 字段从 schema 退场后，运行中接口 / 历史接口 / 强制停止链路里仍残留旧 SQL 查询

当前约束：
- 任何运行时 SQL 都不得再读取 `rc_job_id`
- 迁移 v1 历史定义可保留，但当前业务查询不允许再引用

### 2.7 precheck / preflight 半退场状态
已完成清理：
- 设置页 `PRECHECK_MODE` 已删除
- preflight 已从任务主进度链和运行详情主展示链退场

后续要求：
- 不再保留 preflight / precheck 的“半退场”状态
- 前后端一起清理，不做旧壳兼容展示

### 2.8 删除任务只删主记录、不清关联数据
已修复：
- `TaskService.DeleteTask(id int64)` 现在会先 `ListRunsByTask(id)`
- 对关联 run 的 `summary.stderrFile` 执行文件删除
- 目录清空后删除空目录
- 最后再调用 `s.db.DeleteTask(id)`，让数据库继续级联删 `runs / schedules`

---

## 3. 已过时 / 已降级的问题

这些内容曾经成立，但按当前代码状态，已经不该继续作为主要技术债推进。

### 3.1 继续大拆 `TaskView.vue`
已过时。

当前 `TaskView.vue` 已从“单文件页面骨架 + 运行时装配”收敛为：
- 页面总装配层
- 3 个页面子壳：
  - `TaskListViewShell.vue`
  - `TaskHistoryViewShell.vue`
  - `TaskEditorViewShell.vue`

后续方向已经不是“继续大拆”，而是：
- 少量 glue code 收尾
- 防止头部 import / wiring block 反弹
- 防止尾部模板 / 样式块反弹
- 稳定 3 个 shell 边界

更细的拆分现状与边界，以 `docs/TASKVIEW_SPLIT_MAP.md` 为准。

### 3.2 优化 preflight 文案与用户预期
已降级。

原因：
- preflight 已从任务主进度链退场
- 当前主链已经明确只保留 WebSocket `run_progress` + `/api/runs/active`

因此这条不再是当前主要矛盾。

### 3.3 继续大规模清理运行态 RC 旧链
已基本完成收口。

当前边界：
- 保留 RC：文件浏览、文件操作、添加存储等仍在使用的能力
- 其余已有 CLI 主链的运行态 RC 旧链，不再保留兼容壳

后续不再把这条当成本阶段主要技术债。

---

## 4. 当前真实未解决的技术债

### 高优先级

#### 4.1 `TaskView.vue` 收尾优化与防反弹
现状：
- 页面主链已基本收敛为装配层
- 当前主要矛盾不再是“继续拆”，而是稳定现有结构

仍值得继续做的点：
- 清理少量残留 glue code
- 继续减少页面顶层装配噪音
- 稳住 3 个 shell 边界，不让 `TaskView.vue` 回退去直连旧 section
- 持续防止头部 import / wiring block 和尾部模板 / 样式块反弹

当前已确认的顺序 guardrail：
- `useRunningHintRuntime(...)` 必须放在 `useTaskViewAuxRuntime(...)` 之后，因为它依赖 `openRunLogFromHint -> openRunLog`
- `useTaskViewModalBindings(...)` 必须放在 `useTaskFormRuntime(...)` 之后，因为它依赖 `commandMode` / `commandText` / `showAdvancedOptions`
- 这两条都属于“看起来只是排版整理，实际上会触发 TDZ / before initialization”的高风险点；后续 4.1 只允许补分段标记、注释、局部低风险 glue 清理，不再重排这几段声明顺序

#### 4.2 固化 `progress / completedFreezeByTask / finalSummary` 契约边界
现状：
- 主链已经理顺
- 但类型层 / 注释层 / 接口契约层仍可继续加固

目标：
- 在接口类型、注释、文档中继续强化语义约束
- 严格限制运行中页面只接触 `progress`
- 任务卡片完成态只接触 `completedFreezeByTask`
- 历史详情只接触 `finalSummary`
- 降低后续维护者误用字段职责的风险

当前推进建议：
- 第一轮先做低风险契约加固：补前端类型、关键 composable 就地 guardrail 注释、接口文档同步
- 第二轮继续把约束下沉到后端注释与自动化测试，明确禁止 `finalSummary` 回流到 `/api/runs/active`
- 这一轮不改 UI 行为，不再继续整理 `TaskView.vue` 结构

#### 4.3 旧 runner 路径已完成清理
已完成：
- 当前主执行链固定为 `internal/runnercli/runner.go`
- 旧的 `internal/app/runner.go` 已删除

后续要求：
- 任务执行、进度解析、日志链路、停止链路只维护 `runnercli` 主链
- 不再恢复或重新引入并行旧 runner 实现

### 中优先级

#### 4.4 运行中调试信息开关化（已落地）
现状：
- running hint 小窗的调试详情已改为默认关闭
- 已通过默认配置项 `RUNNING_HINT_DEBUG_ENABLED` 控制是否允许展开
- 设置页已支持保存、重开回显与保存后立即生效

当前规则：
- 默认用户界面不应长期裸露调试详情
- 调试详情不属于运行中主展示链，只用于排障
- 如需开启，优先通过明确配置项控制，而不是把调试入口直接暴露给所有用户

#### 4.5 统一任务卡片与历史记录卡片的视觉基底
现状：
- 视觉效果看似相似，但样式来源不完全一致
- 改 hover 时已经暴露出不一致问题

建议：
- 抽出统一的列表项 hover / active 基底样式
- 任务卡片 / 历史项在其上做差异化扩展

#### 4.6 `web/` 构建产物流程已明确，后续重点是执行一致性
现状：
- 当前仓库继续跟踪并提交 `web/index.html` 与 `web/assets/*`
- 这意味着前端源码改动后，若重新 build 产生新产物，应同步提交对应 `web/` 产物

当前规则：
- 前端源码改动若影响打包结果，必须同步更新并提交 `web/` 产物
- 不允许只提交前端源码、不提交对应产物
- 也不允许把与本轮无关的旧产物噪音一起混入提交

后续重点：
- 让开发流程、发布流程、提交检查都严格执行这条规则
- 若未来决定改成“不再提交构建产物”，应单独作为一轮发布流程改造，不在日常改动中半切换

### 低优先级

#### 4.7 继续清理剩余历史兼容层
现状：
- 当前仍保留少量兼容回退逻辑
- 但已经不是主矛盾

建议：
- 在现有运行中链路稳定后，再逐步减少不必要回退

#### 4.8 把工程规则继续下沉到代码注释 / 契约层
现状：
- 当前已有较完整的外部文档
- 但关键链路的代码就地注释仍偏少

建议：
- 在关键链路继续加短注释
- 避免只靠外部文档记忆工程规则

---

## 5. 临时测试提速方案（后续应剔除）

### 当前临时方案目的
为了在网络不稳定、Docker Hub / 包管理源偶发超时的情况下，快速验证构建流程，仓库当前临时引入了“本地缓存优先”的测试提速方案。

补充当前测试工作流约定：
- 每次真正到达用户测试点时，先执行 `docker build --no-cache -t ray5378/rcloneflow:latest .`
- 构建开始后无需全程盯看，默认在约 5 分钟时主动检查一次结果；无论该次检查结果是成功、失败还是仍未完成，都直接把结果告知用户，不再自动重试
- 该 5 分钟节点的检查与通知必须由 Agent 主动触发，不能依赖用户再次追问构建状态
- Agent 需要先自己检查构建结果，再决定是通知用户测试还是告知用户失败 / 未完成状态

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
��缓存优先逻辑

# ENGINEERING_RULES.md

开发约束与技术要求。用于约束后续开发行为，避免把当前已经理顺的链路重新改乱。

## 1. 分支规则（必须遵守）

- `dev`：开发分支
- `master`：主线 / 稳定分支
- 正确流程：`dev` 开发 -> 测试验证 -> `dev -> master`
- 不再使用 `next-master` 作为主线
- 不允许从 `master` 直接开发
- 不允许执行 `master -> dev` 的反向合并

## 2. 运行中进度源规则（必须遵守）

### 2.1 运行中 UI 主数据源
运行中 UI 必须优先使用：
- `/api/runs/active.progress`

包括但不限于：
- 任务卡片
- 运行中提示小窗
- 运行中 ETA
- 运行中进度文案

### 2.2 字段语义
- `progress`：运行中的实时进度（live frame）
- `stableProgress`：兼容字段 / 完成态固化，不得再作为运行中 UI 主数据源
- `preflight`：预估总量（来自 `rclone size` + 过滤条件），仅用于预估展示

### 2.3 禁止事项
- 不要再让运行中 UI 优先读取 `stableProgress`
- 不要再让运行中 ETA 回退到 `preflight`
- 不要在前端对运行中进度做“历史最大值拼接”或二次抗噪合并
- 不要把 `preflight` 当作运行中总量真源

## 3. 运行中 ETA 规则（必须遵守）

运行中 ETA 只允许基于 live progress 计算：
- `live.totalBytes`
- `live.bytes`
- `live.speed`

如果 `live.totalBytes` 缺失：
- 直接不显示 ETA / 返回空
- 不再回退 `preflight.totalBytes`

## 4. 日志解析规则（必须遵守）

### 4.1 aggregate progress 解析规则
聚合进度解析只允许接受完整的 aggregate one-line 统计行。

以下内容不得再被当成总进度行：
- 文件级进度行
- `Copied (new)`
- `Deleted`
- 其他不完整的碎片日志

### 4.2 size pair 解析规则
aggregate size pair 必须要求显式字节单位（例如 `MiB/GiB`）。

禁止再次允许这种宽松匹配：
- 把时间戳里的 `2026/04` 误解析成 `bytes/totalBytes`

## 5. Active Runs 接口规则（必须遵守）

`/api/runs/active` 当前是运行中 UI 的唯一事实来源。

接口约定：
- `progress`：运行中实时进度主字段
- `stableProgress`：兼容字段 / 完成态固化
- `progressLine`：最后成功解析到的原始进度日志行
- `progressSource`：当前进度来源
- `progressMismatch` / `progressCheck`：后端一致性自检结果

后续开发不得：
- 删除 `progress` 并要求前端重新切回 `stableProgress`
- 让 `stableProgress` 再次承担运行中主展示职责

## 6. WebSocket / 前端刷新规则（必须遵守）

- `run_progress` 更新 active run 时，必须按 `runRecord.id` 匹配
- 不要再按错误顶层字段匹配（例如 `r.id`）
- 运行中刷新允许存在轻量轮询兜底，但 WebSocket 仍是首选实时来源

## 7. 预检（preflight）规则（必须遵守）

### 7.1 允许用途
`preflight` 只允许用于：
- 预估总量展示
- 预估文件数展示
- 缺失时作为极小范围的兼容兜底（仅在后端链路明确需要时）

### 7.2 不允许用途
`preflight` 不应再直接驱动：
- 运行中任务卡片主进度
- 运行中 ETA
- 运行中小窗主展示

## 8. 前端组件规则（建议强约束）

### 8.1 禁止继续把大块功能堆进单文件视图
`frontend/src/views/TaskView.vue` 过大的根本原因，是长期把多个职责直接堆进一个页面文件，而没有及时拆成组件或 composable。

后续必须避免重复这个模式：
- 不要继续把任务列表、历史记录、运行中提示、弹窗、进度格式化、WebSocket 更新、调试逻辑继续直接堆进同一个 view 文件
- 新增功能如果已经形成独立 UI 区块或独立状态链，应优先拆分为组件或 composable
- 当某个 view 同时承担“展示 + 状态管理 + 网络更新 + 调试逻辑”时，应视为需要拆分的危险信号

### 8.2 拆分优先原则
新增或重构前端逻辑时，优先按以下边界拆分：
- 独立弹窗 -> 独立组件
- 独立卡片区域 -> 独立组件
- 可复用的进度/格式化/调试逻辑 -> helper 或 composable
- 与主页面弱耦合的交互块 -> 子组件

### 8.3 明确禁止事项
- 不要为了“先快点做完”而把新功能直接追加进 `TaskView.vue` 尾部
- 不要通过继续增加局部状态、临时函数、模板条件分支来维持超大页面文件
- 不要把本该在子组件中的 hover/弹窗/调试逻辑长期滞留在父级 view 中

### 8.4 其他前端约束
- 新增运行中相关展示时，优先考虑独立组件或 helper/composable
- 视觉相似的卡片，不要依赖偶然命中的全局类名去复用 hover 效果

## 9. 构建产物规则（建议强约束）

- 前端构建后的 `web/index.html` 与 `web/assets/*` 容易污染工作区
- 修改源码后，应避免把构建产物误当成业务代码改动提交
- 分支切换后若出现构建产物脏状态，应先清理再继续开发

## 10. 遇到运行中进度问题时的排查顺序（必须遵守）

排查顺序固定为：

`日志原文 -> summary.progress -> /api/runs/active.progress -> 任务卡片 / 运行中提示小窗`

不要直接从以下字段逆推运行中 UI：
- `preflight`
- `stableProgress`

除非你明确是在处理：
- 完成态固化
- 历史兼容逻辑
- 缺失字段兜底

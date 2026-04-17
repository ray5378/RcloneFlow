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

## 待重构清单

### 高优先级

#### 1. 拆分 `frontend/src/views/TaskView.vue`
现状：
- 文件过大
- 混合了承担任务列表、历史记录、运行中提示、弹窗、进度格式化、WebSocket 更新等职责

建议拆分优先块：
- `RunningHintModal.vue`
- active progress 相关 helper / composable
- 历史记录区域
- 任务列表区域

目标：
- 降低耦合
- 减少后续修改时误伤进度链

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

## 当前不建议乱动的地方
- 运行中日志解析主链（除非有明确 bug）
- `progress` 作为 active UI 主源的约定
- `stableProgress` 的完成态固化逻辑（除非同时检查历史展示）
- `preflight` 的存在本身（问题不是它存在，而是不该再混入运行中主展示）

## 后续修改建议
当后续再改运行中展示时，优先按这条链排查：

`日志原文 -> summary.progress -> /api/runs/active.progress -> 任务卡片/运行中提示小窗`

不要直接从 `preflight` 或 `stableProgress` 入手推导运行中 UI，除非明确是在处理回退或完成态。

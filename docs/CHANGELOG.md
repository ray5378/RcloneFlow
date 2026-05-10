# CHANGELOG.md

阶段性变更记录。

这份文档不追求完整版本发布体系，而是先记录项目在关键阶段完成了哪些重要变化，便于后续回顾与交接。

---

## 2026-05

### JSON 日志运行中进度修复
- 修复 JSON log 模式下运行中 `progress` 统计缺失问题
- 恢复 `percentage` / `plannedFiles` / `completedFiles` 的实时回填
- 正确处理 `xfr#0/N` 场景，避免已完成文件数被误判为非 0
- 保持默认 `--stats-one-line` 主链不变，避免为了当前文件进度改成 `--progress` 带来额外风险

### 历史完成态与传输明细修复
- 修复历史完成态 `finalSummary` 无法消费 JSON 日志的问题
- 历史 run 详情与任务历史卡片现在可从 JSON log 正确恢复：
  - `counts`
  - `files`
  - 最终统计预览
- `/api/runs/:id/files` 增加对 JSON 日志行的解析支持
- 运行详情里的时间改为可读时间格式
- 运行详情里的大小显示改为正常字节文案，`0` 不再显示为 `-`
- JSON 日志里的 `size` 会透传到运行详情明细接口

### move 任务明细去重
- 修复 move 任务在运行详情里把 `Copied + Deleted` 重复当成两条成功记录的问题
- 后端会将同一路径的 `Copied + Deleted` 合并为单条 `Moved`
- 历史完成态 `finalSummary.files` 与 `/api/runs/:id/files` 都应用同一语义，避免前后端口径不一致

### active transfer 启动体验优化
- 将 active transfer preflight 改为异步，不再阻塞任务真正启动
- 新流程为：
  - 先初始化空状态
  - 立即启动传输
  - 后台补齐候选文件、待传列表、文件大小
- 回填候选集时不会覆盖已开始传输的 current file，也不会覆盖已完成项
- active transfer summary / snapshot 新增 preflight 状态字段，支持前端识别“正在扫描文件列表”阶段
- 前端在 preflight 进行中新增轻量提示：数量和待传清单会逐步补齐

### 验证与交付
- 针对历史 JSON 日志解析、move 明细合并、active transfer 异步 preflight 补充了后端测试
- 针对运行详情大小显示补充了前端格式化测试
- 本阶段相关提交：
  - `6725545` `fix: normalize run detail history display`
  - `8e29223` `feat: make transfer preflight async`

## 2026-04

### 工程与文档体系
- 补齐并强化 `ENGINEERING_RULES.md`
- 新增 `DEVELOPMENT_CHECKLIST.md`
- 新增 `FRONTEND_RULES.md`
- 新增 `BACKEND_RULES.md`
- 新增 `AGENT_COLLABORATION_GUIDE.md`
- 新增 `ARCHITECTURE_OVERVIEW.md`
- 新增 `API_CONTRACT_GUIDE.md`
- 新增 `DEBUGGING_PLAYBOOK.md`
- 新增 `GLOSSARY.md`
- 新增 `REFACTORING_PLAYBOOK.md`
- 新增 `TESTING_GUIDE.md`
- 新增 `RELEASE_GUIDE.md`
- 新增 `DECISION_LOG.md`
- 新增文档索引 `README.md`

### 运行中进度链治理
- 统一运行中 UI 主真源为 `/api/runs/active.progress`
- 明确运行中主展示真源固定为 `progress`，任务卡片完成态由前端冻结帧 `completedFreezeByTask` 承接
- 明确 `preflight` 保留预估语义，不直接驱动运行中主展示
- 运行中 ETA 不再回退 `preflight`

### 运行中显示问题修复
- 修复 early live progress 导致总量/总数先对后错再恢复的问题
- 修复总量修正后百分比仍不一致的问题
- 修复方向回到后端源头，而不是依赖前端 masking

### 前端拆分进展
- 将 running hint 相关逻辑从 `TaskView.vue` 中拆出：
  - `RunningHintModal.vue`
  - `runningHint.ts`
  - `useRunningHint.ts`
  - `useActiveRunLookup.ts`
  - `activeRunProgress.ts`
- 同时完成相关接线修复和回归修复

### 添加存储文案专项收尾
- 对 `frontend/src/components/AddRemoteModal.vue` 的 `optionLabelMap` / `optionHelpMap` 做了一次集中收整
- 重点收掉中文模式下“添加存储”真实会上屏的 option label / help 漏翻、错位、重复 key、英文透出问题
- 本轮集中整理覆盖了 SMB、WebDAV、crypt、local、OneDrive 以及一批通用 option
- 这条线当前已从“持续专项清理”切换为“阶段性收口 + 验收报点修复”

### 仓库状态收尾
- 分支已收敛为只保留 `dev` / `master`
- `dev` / `master` 当前已对齐到同一提交
- 历史 tag 已清理并重建，当前只保留 `v1.0.0`
- 发布 / 分支 / README 口径已同步到文档

---

## 使用建议

后续如果项目继续有明确阶段性成果，可以继续按月份或阶段补充到这里。

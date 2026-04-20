# CHANGELOG.md

阶段性变更记录。

这份文档不追求完整版本发布体系，而是先记录项目在关键阶段完成了哪些重要变化，便于后续回顾与交接。

---

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

---

## 使用建议

后续如果项目继续有明确阶段性成果，可以继续按月份或阶段补充到这里。

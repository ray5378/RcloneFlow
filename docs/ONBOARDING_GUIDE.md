# ONBOARDING_GUIDE.md

新接手项目指南。

这份文档面向第一次接触该项目的人，目标不是让你立刻看懂所有代码，而是帮助你快速建立基本认知，知道应该先看什么、先理解什么、先不要乱动什么。

---

## 1. 第一件事：先看文档，不要先盲改代码

推荐顺序：
1. `docs/README.md`
2. `docs/ENGINEERING_RULES.md`
3. `docs/DEVELOPMENT_CHECKLIST.md`
4. `docs/ARCHITECTURE_OVERVIEW.md`
5. `docs/TECH_DEBT.md`

如果你主要改前端，再看：
- `docs/FRONTEND_RULES.md`

如果你主要改后端，再看：
- `docs/BACKEND_RULES.md`

---

## 2. 第二件事：先理解主链，不要先理解所有细节

优先理解这些主链：
- 后端启动链
- 路由入口
- 前端主页面
- 运行中进度链
- WebSocket 更新链

如果你能先说清：
- 数据从哪里来
- 中间经过哪些层
- 最后在哪里展示

那你就已经跨过最难的第一步了。

---

## 3. 先知道哪些地方是高风险区

当前高风险区域主要包括：
- `frontend/src/views/TaskView.vue`
- `internal/controller/run.go`
- `internal/runnercli/runner.go`
- 运行中相关 composable 和弹窗
- WebSocket 接线

这些地方改动前应优先：
- 明确目标
- 缩小范围
- 小步推进

---

## 4. 先知道哪些规则不能随便破

当前几条最关键的约束：
- `dev -> master` 是固定主线流程
- 运行中 UI 主真源是 `/api/runs/active.progress`
- 运行中主展示真源固定为 `progress`，任务卡片完成态由前端冻结帧 `completedFreezeByTask` 承接
- `preflight` 保留预估语义，不直接驱动运行中主展示
- 高风险改动优先小步推进、每步停测

---

## 5. 如果你是第一次动代码

建议先从以下类型的任务开始：
- 文档补充
- 小范围 helper / composable 整理
- 低风险样式修复
- 明确链路的小 bug 修复

短期不建议一上来就：
- 大拆主页面
- 大改实时进度链
- 大改接口契约
- 同时改前端和后端多个高风险点

---

## 6. 如果你遇到问题，不知道从哪查

优先看：
- `docs/DEBUGGING_PLAYBOOK.md`
- `docs/API_CONTRACT_GUIDE.md`
- `docs/GLOSSARY.md`

这三份通常能帮你快速判断：
- 这是哪类问题
- 该查哪条链
- 关键术语到底是什么意思

---

## 7. 一个简单建议

先把自己训练成：
- 能看懂一条链
- 能分清一层职责
- 能描述一个现象
- 能判断一个改动风险

再去追求“马上看懂全部代码”。

这样进入项目会快得多，也更不容易被复杂度劝退。

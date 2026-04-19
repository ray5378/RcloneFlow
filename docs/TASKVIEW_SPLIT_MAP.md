# TASKVIEW_SPLIT_MAP.md

`frontend/src/views/TaskView.vue` 拆分地图。

这份文档不再记录早期“大拆分想象图”，而是对齐当前真实代码状态，明确：
- 哪些 runtime / shell / modal 已经接回主链
- `TaskView.vue` 现在还承担什么
- 拆成 3 个页面子壳后，页面总装配层与子壳边界如何划分
- 后续只建议做哪些低风险收尾或防反弹加固

---

## 1. 当前阶段判断

当前结论已经更新为：

> `frontend/src/views/TaskView.vue` 已从“单文件页面骨架 + 运行时装配”进一步收敛为“**页面总装配层 + 3 个页面子壳**”。

当前真实状态：
- `frontend/src/views/TaskView.vue` 当前约 **654 行**
- 三大主界面已经拆成 3 个页面子壳，并已接入 `TaskView.vue` 主装配链
- `TaskView.vue` 不再直接装配 `TaskListSection` / `TaskHistorySection` / `AddTaskForm`
- 当前主要矛盾已经从“继续拆骨架”转为：
  - 保持 3 壳接线稳定
  - 防止 `TaskView.vue` 头尾块反弹
  - 决定 modal 层是否还值得继续下沉

这意味着：
- 这条线已经不再是“继续大拆 `TaskView.vue`”
- 而是“**总装配层稳定化 + 子壳边界定型**”

---

## 2. 当前已经完成的拆分/收口

### A. 页面级 state / runtime 已接入

当前 `TaskView.vue` 已接入的页面级状态与 runtime：
- `useTaskViewState`
- `useTaskViewRuntimeState`
- `useTaskViewRuntime`
- `useTaskViewAuxRuntime`
- `useTaskListView`
- `useTaskListRuntime`
- `useTaskFormNormalize`
- `useTaskFormRuntime`
- `useTaskHistoryRuntime`
- `useRunDetailRuntime`
- `useRunDetailEntry`
- `useRunningHintRuntime`
- `useToastCenter`

当前 `TaskView.vue` 的页面层职责已收敛为：
- 组合页面级状态与 runtime
- 组织模块切换
- 把 props / actions 分发给 3 个页面子壳
- 保留 modal 层装配
- 保留少量页面级 setter / close handler / glue code

---

### B. 三大主界面已经拆成 3 个页面子壳

当前新增并接实的 3 个页面子壳：
- `TaskListViewShell.vue`
- `TaskHistoryViewShell.vue`
- `TaskEditorViewShell.vue`

当前边界：

#### `TaskListViewShell.vue`
负责：
- 任务列表主界面骨架
- `TaskListSection` props / events 装配

#### `TaskHistoryViewShell.vue`
负责：
- 历史主界面骨架
- `TaskHistorySection` props / events 装配
- 详情 / 分页 / 过滤等历史壳层动作接线

#### `TaskEditorViewShell.vue`
负责：
- 添加 / 编辑任务主界面骨架
- `AddTaskForm` props / events 装配

当前 `TaskView.vue` 已不再直接装配：
- `TaskListSection.vue`
- `TaskHistorySection.vue`
- `AddTaskForm.vue`

这一步是当前结构变化里最关键的一步。

---

### C. modal 层已进入“按归属部分下沉”阶段

当前 modal 层已经不再全部留在 `TaskView.vue``。

#### 已下沉到 `TaskListViewShell.vue`
- `WebhookConfigModal.vue`
- `SingletonConfigModal.vue`

#### 已下沉到 `TaskHistoryViewShell.vue`
- `RunLogModal.vue`
- `RunningHintModal.vue`

#### 当前仍保留在 `TaskView.vue`
- `GlobalStatsModal.vue`
- `ConfirmModal.vue`

当前边界判断：
- 任务列表操作强相关弹层，归到列表壳
- 运行/历史链强相关弹层，归到历史壳
- 全局性/跨模块弹层，继续留在页面总装配层

这是当前阶段更合理的最小扰动结构，不必为了“全下沉”继续强推进。

---

### D. 运行详情链已完成阶段性收口

当前运行详情相关已接线能力包括：
- `openRunDetailModal`
- `closeRunDetailModal`
- `openRunDetailFiles`
- `showRunDetail`
- `closeRunDetail`
- `getFinalSummary`
- `getPreflight`
- `setFinalFilter`
- `goPrevFinalFilesPage`
- `goNextFinalFilesPage`
- `jumpFinalFilesPage`
- `goPrevFilesPage`
- `goNextFilesPage`

当前职责划分：
- 详情运行时主链仍由既有 composable 承担
- 历史主界面骨架由 `TaskHistoryViewShell.vue` 装配
- `TaskView.vue` 只保留总装配层职责

---

## 3. 当前 `TaskView.vue` 还应该承担什么

当前 `TaskView.vue` 合理保留的内容：

### A. 页面总装配层
- runtime / state 汇总
- 模块切换控制
- props / action 下发给 3 个页面子壳

### B. modal 层装配
当前仍可合理保留：
- 各类 modal 的 visible / data / action 接线
- 少量 close handler

### C. 少量页面级 glue code
例如：
- `closeWebhookModal`
- `closeSingletonModal`
- `closeLogModal`
- `closeGlobalStatsModal`
- `setTaskSearch`
- `setTasksJumpPageValue`
- `setHistoryStatusFilter`
- `setJumpPageValue`
- `setFinalFilesJumpValue`

这类内容继续保留在 `TaskView.vue` 是合理的。

---

## 4. 当前不建议继续推进的方向

### A. 不再为了“继续拆”而继续碎拆
当前不建议：
- 再继续把轻量 setter / close handler 外移
- 再为了压行数而把总装配层切得更碎
- 再继续增加大量 props / emits 链路

原因：
- 收益已经明显下降
- 结构会开始变碎
- 更容易让总装配层可读性下降

### B. 不碰稳定业务主链
当前不建议主动触碰：
- 运行主链
- 历史详情业务语义
- 任务执行 / 停止主行为
- 页面级刷新主语义
- 已稳定的 runtime 边界

---

## 5. 当前真正高频的脆弱点

### A. `TaskView.vue` 头部 import / wiring block
已知问题：
- 旧导入反弹
- 新导入缺失
- 顶部 import block 重复
- 解构返回项漏接

处理规则：
- 每次结构调整后，都先检查头部 import / 解构 / 模板绑定是否一致
- 头部块出问题时，优先整块回正，不做碎补

### B. `TaskView.vue` style 尾部反弹
已知表现：
- build 报 `Invalid end tag.`
- 文件尾部反弹出多份 `</style>`
- 或残留截断 CSS 片段

固定处理策略：
1. 检查：
   - `grep -n "<style scoped>\|</style>" frontend/src/views/TaskView.vue`
2. 如异常，不做猜测式补丁
3. 直接从 `<style scoped>` 到文件尾整段重写为：
   - 单份 `<style scoped>`
   - 单份 `</style>`
   - 干净、完整、可构建版本

这条规则仍然是当前最稳定的修法。

---

## 6. 后续推荐动作顺序

如果继续推进，推荐顺序已经更新为：

1. **保持 3 个页面子壳接线稳定**
   - 不让 `TaskView.vue` 再直连旧 section
   - 不让壳层再回退成半接线状态

2. **维护 `TaskView.vue` 头尾块稳定**
   - 重点盯 import / wiring block
   - 重点盯 style 尾部 block

3. **决定 modal 层是否继续下沉**
   - 仅在确认收益明显、风险可控时再做
   - 否则停在当前结构点即可

4. **文档持续同步，不拖到最后**
   - `docs/TASKVIEW_SPLIT_MAP.md`
   - `docs/TECH_DEBT.md`

---

## 7. 每轮改动后的固定验证

每轮 `TaskView.vue` / view shell 调整后，固定执行：

- `cd /root/.openclaw/workspace/rcloneflow/frontend && npx tsc --noEmit -p tsconfig.json`
- `cd /root/.openclaw/workspace/rcloneflow/frontend && npm run build`

若前端源码改动触发构建产物变更，还要同步提交：
- `web/index.html`
- `web/assets/*`

提交后必须再次确认：
- `git status --short` 最终为空

---

## 8. 当前一句话路线图

当前正确路线图已经不是“继续大拆 `TaskView.vue`”，而是：

`TaskView.vue 退回总装配层 -> 三大主界面通过 3 个 view shell 接入 -> 维持壳层边界稳定 -> 只做防反弹与必要文档同步`

也就是说，当前结构目标已经从：
- “把所有东西塞回一个 view 里做清理”

转成：
- “让总装配层、页面子壳、modal 层各自停在稳定边界上”

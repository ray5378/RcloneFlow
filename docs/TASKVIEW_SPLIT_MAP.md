# TASKVIEW_SPLIT_MAP.md

`frontend/src/views/TaskView.vue` 拆分地图。

这份文档不再记录早期“大拆分想象图”，而是对齐当前真实代码状态，明确：
- 哪些 runtime / section / modal 已经接回主链
- `TaskView.vue` 现在还承担什么
- 后续只建议做哪些低风险页面壳收尾
- 哪些高扰动方向当前不建议再推进

---

## 1. 当前阶段判断

当前结论已经明确：

> `frontend/src/views/TaskView.vue` 的拆分已经进入**阶段性完成**，后续重点不再是继续大拆，而是做**页面壳级低风险收尾优化**。

当前真实状态：
- `frontend/src/views/TaskView.vue` 当前约 **632 行**
- 相比更早的 900+ 行阶段，页面已经明显从“功能承载层”收敛到“页面装配层”
- 主要 runtime、section、modal 已经下沉并重新接回主链
- 当前主要矛盾不再是“缺少结构拆分”，而是“顶层页面壳还有少量 glue code / 模板桥接 / 导入噪音待收尾”

当前不建议再把目标表述成：
- “继续大拆 `TaskView.vue`”
- “继续切更多 composable”
- “继续做结构性迁移”

更准确的表述应该是：
- **保持主链稳定**
- **持续压薄页面壳**
- **把顶层模板和 script 中剩余的低价值桥接继续清掉**
- **文档与真实代码保持同步**

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

当前页面层职责已经收敛为：
- 组合页面级状态与 runtime
- 把 props / events 接到各个 section / modal
- 保留极少量页面壳 setter / close handler / 轻量 glue code

这和早期“大量逻辑直接堆在 `TaskView.vue`”已经不是一个阶段。

---

### B. 独立 modal 壳已经完成回收

当前已接回组件链的 modal：
- `RunningHintModal.vue`
- `RunDetailModal.vue`
- `GlobalStatsModal.vue`
- `RunLogModal.vue`
- `SingletonConfigModal.vue`
- `WebhookConfigModal.vue`
- `ConfirmModal.vue`

当前页面层对这些 modal 的职责主要是：
- 传 `visible`
- 传业务数据 / 展示 props
- 传 `close` / `save` / `confirm` / `test` 等事件

也就是说，这一层已经基本符合“页面装配壳”定位。

---

### C. 轻量 section / 页面骨架已经完成接线

当前已下沉并接回的页面骨架组件：
- `TaskListSection.vue`
- `TaskHistorySection.vue`
- `AddTaskForm.vue`
- `TaskListHeader.vue`
- `TaskListPagination.vue`

当前边界：
- `TaskView.vue` 负责页面级数据装配
- section 负责各自区域模板骨架
- 业务行为入口仍由已有 runtime / action composable 提供

这意味着当前主要工作已经从“拆 section”转成“清理 section 与页面壳之间的冗余桥接”。

---

### D. 运行详情链已经完成阶段性收口

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

当前页面层在详情链中的职责主要是：
- 入口判断
- 详情弹窗装配
- 少量分页/过滤事件接线

说明这块也已经不适合再用“未拆主块”来描述。

---

## 3. 当前仍属于页面壳收尾的内容

后续允许继续推进的，应该只剩这些**低风险页面壳收尾项**。

### A. 模板透明 wrapper / 内联赋值桥接

典型例子：
- `@foo="bar($event)"` → `@foo="bar"`
- `@close="xxx = false"` → `@close="closeXxxModal"`
- `@update:xxx="state.xxx = $event"` → 命名明确的轻量 setter

目标：
- 继续减少模板里的即时表达式
- 让页面壳更多体现“装配”，少体现“局部逻辑书写”

这是当前最合适、最稳定的推进方向。

---

### B. 顶层导入 / 解构 / 接线反弹问题

这一类属于高频反弹项：
- 旧导入又回来了，但页面没在用
- 页面模板已切到新函数，script 里没补导入/解构
- composable 新增返回项后，页面没接回来

过去多轮已经反复遇到的典型对象包括：
- `handleError`
- `showSuccess`
- `formatDuration`
- `showConfirm`
- `useTaskFormNormalize`

后续规则：
- 每次改动页面壳后，都要顺手检查 import / 解构 / 模板绑定是否一致
- 这类问题优先级高于继续“拆结构”

---

### C. style 尾部脏片反弹

这是 `TaskView.vue` 当前最典型的高频脆弱点之一。

已知表现：
- build 报 `Invalid end tag.`
- 文件尾部反弹出多份 `</style>`
- 或残留被截断的 CSS 片段

固定处理策略：
1. 先检查：
   - `grep -n "<style scoped>\|</style>" frontend/src/views/TaskView.vue`
2. 如出现异常，不做猜测式补丁
3. 直接从 `<style scoped>` 到文件尾整段重写为：
   - 单份 `<style scoped>`
   - 单份 `</style>`
   - 干净、完整、可构建版本

这条规则已经被多轮验证为最稳定的修法。

---

## 4. 当前不建议继续推进的方向

### A. 不再为了“继续拆”而继续切更多 composable

当前阶段不建议：
- 把稳定主链再强行切成更多小块
- 为了文件更短而继续迁移已稳定逻辑
- 在没有明确收益时重排 runtime 边界

原因：
- 当前收益已经明显下降
- 风险却在上升
- 更容易打破现有稳定接线

---

### B. 不碰稳定业务主链

当前不建议主动触碰：
- 运行主链
- 历史详情真实业务语义
- 任务执行/停止主行为
- 页面级刷新主语义
- 已经稳定的 runtime 边界

当前目标不是“重新设计页面”，而是“把页面壳收干净”。

---

## 5. 后续推荐动作顺序

如果继续推进，推荐顺序应该是：

1. **继续扫描 `TaskView.vue` 当前真实文件状态**
   - 不依赖旧快照
   - 只改当前仍真实存在的问题

2. **继续压模板透明 wrapper / 内联赋值桥接**
   - 优先改成直接传函数引用
   - modal close 统一收成轻量 close handler

3. **继续清理顶层导入 / 解构 / 透传噪音**
   - 删除未使用导入
   - 补回真实缺失接线
   - 保持 composable 返回项与页面装配一致

4. **持续盯住 style 尾部脏片反弹**
   - 一旦反弹，直接按固定规则整段修复

5. **文档持续同步，不拖到最后**
   - `docs/TASKVIEW_SPLIT_MAP.md`
   - `docs/TECH_DEBT.md`

---

## 6. 每轮改动后的固定验证

每轮 `TaskView.vue` 页面壳收尾后，固定执行：

- `cd /root/.openclaw/workspace/rcloneflow/frontend && npx tsc --noEmit -p tsconfig.json`
- `cd /root/.openclaw/workspace/rcloneflow/frontend && npm run build`

若前端源码改动触发构建产物变更，还要同步提交：
- `web/index.html`
- `web/assets/*`

提交后必须再次确认：
- `git status --short` 最终为空

---

## 7. 当前一句话路线图

当前正确路线图已经不是“继续大拆”，而是：

`runtime / section / modal 已接回主链 -> TaskView.vue 进入阶段性完成 -> 后续只做页面壳低风险收尾优化`

目标是让 `TaskView.vue` 最终稳定停留在：
- 页面装配层
- 数据入口层
- 少量轻量 setter / close handler / glue code

而不是再回到“大体量功能承载文件”或继续做高扰动结构改造。

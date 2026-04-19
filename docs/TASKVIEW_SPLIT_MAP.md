# TASKVIEW_SPLIT_MAP.md

`frontend/src/views/TaskView.vue` 拆分地图。

这份文档的目标，不是重复记录所有技术债，而是把 `TaskView.vue` 的拆分工作整理成一张可以持续维护的“作战地图”，让后续每拆一步都能明确知道：
- 已经拆了什么
- 还剩什么没拆
- 哪一块处于半拆状态
- 下一刀最适合先砍哪里
- 每一刀会影响哪些功能、要重点测什么

---

## 1. 当前阶段判断

当前拆分进度还处于：
- **已完成第一批关键逻辑抽离**
- **但尚未进入页面主体的大块拆分完成阶段**

更准确地说：
- 运行中提示相关逻辑已经拆出
- active run 读取与部分抗噪逻辑已经拆出
- 但 `TaskView.vue` 仍然承担了页面主体的大量职责

当前整体状态应理解为：
> 已经打开拆分局面，但主页面仍未真正瘦身完成。

---

## 2. 已拆出的模块

### A. 运行中提示小窗 UI
状态：**已拆完**

已拆文件：
- `frontend/src/components/task/RunningHintModal.vue`

当前职责：
- 小窗模板
- 小窗按钮
- 调试展开区

影响范围：
- 运行中提示弹窗
- 小窗中的调试信息显示

重点测试：
- 小窗是否正常打开/关闭
- 小窗内容是否随当前 run 正常切换
- 调试展开区是否显示正确

---

### B. 运行中提示 helper
状态：**已拆完**

已拆文件：
- `frontend/src/components/task/runningHint.ts`

当前职责：
- active progress 读取
- progress 文本拼装
- progress line / check / debug JSON 组装

影响范围：
- 运行中提示文本
- 调试详情格式化结果

重点测试：
- 进度文本是否与 active run 一致
- progress line / check 是否不为空
- 调试信息字段是否完整

---

### C. 运行中提示状态管理
状态：**已拆完**

已拆文件：
- `frontend/src/composables/useRunningHint.ts`

当前职责：
- 小窗开关状态
- 当前 run 绑定
- debug 开关状态
- phase / progress / debugInfo 组装
- 打开日志动作

影响范围：
- 小窗与当前运行中任务的绑定
- debug 状态切换

重点测试：
- 点击运行中提示入口后是否打开正确 run
- 日志按钮是否可用
- debug 状态切换是否稳定

---

### D. active run 基础读取
状态：**已拆完**

已拆文件：
- `frontend/src/composables/useActiveRunLookup.ts`

当前职责：
- `getActiveRunByTaskId()`
- `getActiveProgressByTaskId()`
- `getActiveProgressTextByTaskId()`

影响范围：
- 任务卡片读取 active run
- running hint 读取 active run

重点测试：
- taskId 到 active run 的映射是否正确
- 运行中任务卡片是否能正确显示 active progress

---

### E. active run fallback / 数值归一化
状态：**已拆完**

已拆文件：
- `frontend/src/composables/activeRunProgress.ts`

当前职责：
- `getDeNoisedStableByRun()`
- `getDeNoisedStableByTask()`

影响范围：
- active run 相关的抗噪读取
- 完成态 / 兼容态数值归一

重点测试：
- 数值显示是否稳定
- 不同 run 状态下是否没有明显倒退或错乱

---

## 3. 当前半拆状态区域

### A. 页面级 active run 接线
状态：**半拆**

现状：
- 底层读取 helper / composable 已经有了
- 但页面仍然承担大量接线与组合逻辑

剩余问题：
- 页面层仍知道太多 active run 细节
- 页面仍可能继续堆出新的双轨读取逻辑

后续方向：
- 继续把页面层保留为装配入口
- 尽量让具体读取和拼装继续下沉

---

### B. 运行中相关展示联动
状态：**半拆**

现状：
- 小窗已经拆出
- 但任务卡片、页面状态、WebSocket 刷新联动仍部分留在主页面

剩余问题：
- 运行中展示更新链路仍较长
- 页面中仍存在较多联动点

后续方向：
- 后续拆任务列表块时，把运行中展示联动进一步收口

---

## 4. 仍未拆的主块

### A. 历史记录区域
状态：**半拆**
优先级：**高**

现状：
- 历史页主区块（页头 / 筛选 / 分页 / 列表）已抽到独立组件 `TaskHistoryPanel.vue`
- 第一刀当前已验证稳定：进入历史页、筛选、翻页、删除单条、删除所有、删除后刷新均已恢复正常
- 运行详情弹窗模板主块已抽到独立组件 `RunDetailModal.vue`
- 详情弹窗状态与行为逻辑当前仍主要留在 `TaskView.vue`
- 历史记录删除后的即时刷新问题本轮已定位并修到直接根因：`useApi.ts` 中删除接口成功返回值语义错误，导致前端误回滚
- 历史记录区域第二刀曾尝试把状态/删除刷新逻辑继续迁出 `TaskView.vue`，但因一次迁移范围过大、出现新旧双轨接线与模板残片问题，已明确撤回

原因：
- UI 区块相对独立
- 与运行中主链耦合较低
- 拆分收益高、风险相对可控

已暴露的结构性风险：
- 当前历史页仍存在 `runs` 与 `taskRuns` 双轨状态
- 如果当前显示源与删除/刷新命中的更新源不是同一份状态，就会出现：
  - 后端已经删除成功
  - 但前端当前视图没有立即变化
- 这与此前“任务卡片进度不更新”属于同类问题：
  - 页面显示依赖 A
  - 更新动作只命中 B
  - A 没同步，UI 就卡住

拆后目标：
- 抽成独立组件
- 页面仅负责传入数据、回调和少量状态
- 后续继续收敛历史页显示源与更新源，减少双轨状态导致的刷新错位

当前收口结论：
- 第一刀保留并成立：`TaskHistoryPanel.vue`
- 第二刀已按 A / B / C 三小步重做并全部通过验证
- 历史记录主区当前已完成阶段性收口：UI / computed / loader / actions 已拆出

当前已拆出的历史记录相关文件：
- `frontend/src/components/task/TaskHistoryPanel.vue`
- `frontend/src/components/task/RunDetailModal.vue`
- `frontend/src/composables/useTaskHistoryComputed.ts`
- `frontend/src/composables/useTaskHistoryLoader.ts`
- `frontend/src/composables/useTaskHistoryActions.ts`
- `frontend/src/composables/useRunDetailComputed.ts`
- `frontend/src/composables/useRunDetailFiles.ts`
- `frontend/src/composables/useRunDetailState.ts`
- `frontend/src/composables/useRunDetailEntry.ts`

本轮新增的拆分经验（必须保留）：
- `TaskView.vue` 属于高脆弱文件；当新增 composable 并把页面逻辑切到新 composable 时，最容易出现的真实回归不是“逻辑写错”，而是“接线没落完整”
- 这类回归的典型表现包括：`xxx is not defined`、任务列表空白、点击详情无反应、详情打不开
- 因此后续每次新增 composable 后，必须立刻逐项核对：
  - `import` 是否补齐
  - 页面解构是否已接回
  - 页面旧状态是否删干净
  - 调用点是否都已切到新来源
- 上述核对不通过前，不应把该步视为已稳定完成，也不应继续往下一刀扩改

本轮运行详情链新增收口：
- `frontend/src/composables/useRunDetailComputed.ts` 当前已承接：
  - `getFinalSummary`
  - `getPreflight`
  - `finalFiles`
  - `finalCountAll`
  - `finalCountSuccess`
  - `finalCountFailed`
  - `finalCountOther`
  - `setFinalFilter`
  - `finalFilesTotal`
  - `totalFinalFilesPages`
  - `pagedFinalFiles`
  - `finalFilesJump`
  - `goPrevFinalFilesPage()`
  - `goNextFinalFilesPage()`
  - `jumpFinalFilesPage()`
- `frontend/src/composables/useRunDetailFiles.ts` 当前已承接：
  - `runFilesPage`
  - `openRunDetailFiles(run)`
  - `pagedRunFiles`
  - `totalRunFilesPages`
  - `goPrevFilesPage()`
  - `goNextFilesPage()`
  - 以及已下沉但页面当前不再直接使用的底层状态：`runFiles` / `runFilesTotal` / `runFilesPageSize` / `reloadRunFiles()`
- `frontend/src/composables/useRunDetailState.ts` 当前已承接：
  - `showDetailModal`
  - `runDetail`
  - `openRunDetailModal(run)`
  - `closeRunDetailModal()`
- 当前页面层 `TaskView.vue` 对历史详情弹窗的职责已进一步收敛为：
  - `showRunDetail(run)`：只做入口判断（`running` -> `openRunningHint`，非 running -> 进入历史详情）
  - `closeRunDetail()`：只保留关闭入口
  - 模板装配与事件转发

本轮新增收口：
- 历史详情弹窗模板主块已从 `TaskView.vue` 抽出到 `RunDetailModal.vue`
- 当前采取的是“先拆模板、后拆状态/行为”的最小风险路线
- 详情弹窗相关状态、筛选、分页、明细刷新等逻辑当前仍由页面层持有
- 本轮拆分中曾出现一次典型回归：父组件 `scoped` 样式未跟随模板迁移，导致总结区网格布局与文件明细表格样式失效；后续已在 `RunDetailModal.vue` 内补齐专属样式并通过用户回归测试
- `FileItem.vue` 也已同步调整列宽与左右 padding，使明细行与表头对齐
- 历史页标题栏“空白处返回任务卡片”交互也已在 `TaskHistoryPanel.vue` 内修复：空白区域可返回，但筛选 / 分页 / header actions 等交互区不会误触返回

下一步建议：
- 在 `RunDetailModal.vue` 已稳定的前提下，再考虑继续下沉详情弹窗内部状态与行为逻辑
- 优先评估是否把详情筛选 / 明细分页 / 明细派生计算继续抽到 composable
- 目标是继续减少 `TaskView.vue` 中历史记录剩余的大块状态承载，而不是一次性大迁移

重点测试：
- 历史列表显示
- 展开/收起
- 状态文案
- 与运行中任务切换时是否互不干扰
- 删除单条后是否立即从当前列表消失
- 删除全部后是否立即清空当前历史视图
- 分页页码在删除后是否仍正确

---

### B. 任务列表区域
状态：**未拆**
优先级：**高**

原因：
- 这是页面主区块之一
- 同时承载运行中状态、按钮动作、卡片交互
- 后续大部分结构治理都绕不过它

拆后目标：
- 把任务卡片渲染和卡片交互从页面层分离
- 页面层更多只负责列表组织与数据提供

重点测试：
- 任务卡片渲染
- 运行中状态显示
- hover/选中态
- 各按钮动作
- 与 WebSocket 刷新后的状态一致性

---

### C. 创建任务区域 / 表单块
状态：**半拆**
优先级：**高**

当前已拆出的创建任务相关文件：
- `frontend/src/components/task/AddTaskForm.vue`
- `frontend/src/composables/useTaskFormState.ts`

当前已下沉职责：
- `AddTaskForm.vue`
  - 创建任务表单主模板
  - 复用 `ScheduleOptions.vue`
  - 复用 `AdvancedOptions.vue`
- `useTaskFormState.ts`
  - `createForm`
  - `commandMode`
  - `commandText`
  - `editingTask`
  - `showAdvancedOptions`
  - `resetTaskFormForCreate()`
  - `fillTaskFormForEdit(task, scheduleSpec?)`
- `useTaskFormPrepare.ts`
  - `prepareTaskFormSubmit()`
  - `validateTaskFormBeforeSubmit()`
- `useTaskCommandParse.ts`
  - `parseRcloneCommand()`
  - `parseRemotePath()`
  - `stripQuotes()`
  - `toCamel()`
- `useTaskFormSubmit.ts`
  - `handleTaskFormDoneClick()`
  - `validateTaskForm()`
  - `buildTaskPayload()`
  - `buildScheduleSpec()`
  - `submitTaskForm()`
  - `completeTaskFormSubmit()`
  - `resetTaskFormSubmitState()`
  - `executeTaskFormSubmit()`
- `useTaskFormFlow.ts`
  - `runTaskFormFlow()`
- `useTaskPathBrowse.ts`
  - `sourcePathOptions`
  - `targetPathOptions`
  - `showSourcePathInput`
  - `showTargetPathInput`
  - `sourceCurrentPath`
  - `targetCurrentPath`
  - `sourceBreadcrumbs`
  - `targetBreadcrumbs`
  - `setShowSourcePathInput()`
  - `setShowTargetPathInput()`
  - `resetTaskPathBrowse()`
  - `restoreTaskPathBrowse(task)`
  - `onSourceRemoteChange()`
  - `onTargetRemoteChange()`
  - `onSourceBreadcrumbClick()`
  - `onTargetBreadcrumbClick()`
  - `loadSourcePath()`
  - `loadTargetPath()`
  - `onSourceClick()`
  - `onSourceArrow()`
  - `onTargetClick()`
  - `onTargetArrow()`

当前页面层 `TaskView.vue` 仍保留：
- `createTask()` 极薄入口壳（只负责调用 flow 与 toast）
- `creatingState`
- 更外层的页面级数据加载 / WebSocket / 刷新协调

本轮已完成并验证的创建任务入口链：
- 进入“添加任务”
- 进入“编辑任务”
- 创建/编辑切换时表单基础状态重置与回填
- 高级选项区样式回归修复
- 任务卡片进入编辑时的入口报错修复

本轮创建任务拆分的关键经验：
- `TaskView.vue` 在新增表单相关 composable 后，同样容易出现“页面已调用但 import 未补齐”的真实回归
- 这类回归会直接表现为：页面空白、任务列表空白、编辑入口失效、`xxx is not defined`
- 表单模板抽成组件后，原样式不会自动跟随；对子组件内部布局（尤其 `AdvancedOptions.vue`）必须显式迁移样式，否则会出现布局严重错位

本轮新增完成并验证的任务列表外层装配链：
- `useTaskListView.ts` 已承接任务列表搜索 / 分页 / 跳页 / 派生列表链
- `useTaskViewUi.ts` 已承接菜单开关态、确认弹窗状态、确认打开/关闭/确认动作
- `useTaskHistoryActions.ts` 已真实接入页面主链，承接 `clearRun()` / `clearAllRuns()`
- `useTaskListActions.ts` 已承接 `deleteTask()` / `toggleSchedule()` / `deleteSchedule()` / `clearAllRunsWithConfirm()` 等行为入口壳
- `useTaskRunActions.ts` 已承接 `runTask()` / `stopTaskAny()` / `runningTaskId` / `stoppedTaskId`
- 用户已确认：任务列表显示、删除任务、删除定时任务、清空历史、运行任务、停止任务均正常
- 同时已修复并验证一整组拆分回归：`getScheduleByTaskId is not defined`、`useTaskRunActions is not defined`、`confirmAndClose is not a function`、`Cannot access ... before initialization`

建议下一步拆分顺序：
1. 先收口本轮创建任务表单区 + 路径浏览深链 + 入口编排链 + 定时规则治理 + 任务列表外层装配链成果
2. 再评估更深的页面级数据加载 / WebSocket / 刷新协调是否值得继续下沉
3. 最后再决定是否触碰更深的任务列表行为细节或页面总装配层

重点测试：
- 新建任务
- 编辑任务
- 新建/编辑切换
- 命令行模式
- 定时任务选项
- 高级选项
- 路径浏览与路径回填

---

### D. 页面级 WebSocket / 列表刷新协调逻辑
状态：**已开始半拆**
优先级：**高，继续拆时仍需谨慎**

当前已拆出：
- `frontend/src/composables/useTaskViewDataSync.ts`
  - `loadData()`
  - `loadActiveRuns()`
  - `loadGlobalStats()`
  - `openGlobalStats()`
  - `setupRealtimeSync()`
- `frontend/src/composables/useTaskProgressSync.ts`
  - `getDbProgressStable()`
  - `getDeNoisedStableByRun()`
  - `getDeNoisedStableByTask()`
  - `formatBps()`
  - `calcEtaFromAvg()`
  - `triggerAutoRefresh()`

当前页面层仍保留：
- stuck 检测 / 轮询兜底定时器
- 页面级生命周期挂接壳
- 活跃进度稳态/去噪/ETA 估算的剩余 timer / signature 协调逻辑

拆后目标：
- 让页面层只负责挂接刷新能力
- 将刷新协调与消息处理继续收口到更清晰的 composable / service 风格逻辑中

重点测试：
- 任务卡片自动刷新
- 历史记录刷新
- run_progress 更新接线
- 页面不空白
- 无重复轮询或失控刷新

---

## 5. 推荐拆分顺序

### 第 1 刀：历史记录区域
原因：
- 相对独立
- 风险低于任务列表主区
- 适合作为继续拆分的热身刀

### 第 2 刀：创建任务区域 / 表单块
原因：
- 功能边界清楚
- 适合和任务列表、历史区块分开

### 第 3 刀：任务列表区域
原因：
- 收益最大
- 但接线更多，适合在前两刀稳定后再进

### 第 4 刀：页面级 WebSocket / 刷新协调逻辑
原因：
- 风险最高
- 应放在主要 UI 区块拆稳以后再收口

---

## 6. 每次拆分后都要更新的内容

以后每完成一步拆分，应同步更新本文件以下内容：
- 对应区块状态：未拆 / 半拆 / 已拆完
- 已拆出的文件名
- 当前职责变化
- 影响范围
- 重点测试项
- 下一步推荐顺序是否变化

如果某一步拆分过程中出现回归，也应补到这里，避免重复踩坑。

---

## 7. 当前不建议直接先动的地方

在没有明确目标时，不建议优先从这些地方下手：
- 运行中进度主链本身
- `progress / stableProgress / preflight` 主语义
- 旧回退逻辑的深层清理
- 页面级 WebSocket 主链（除非这轮目标就是它）

原因：
- 这些区域风险更高
- 更适合作为“后续收口”，而不是下一刀的第一选择

---

## 9. 当前一句话路线图

当前拆分路线建议保持为：

`运行中提示已拆出 -> 历史记录 -> 创建任务表单 -> 任务列表 -> 页面级 WebSocket/刷新协调`

这条路线的目标，是让 `TaskView.vue` 最终收敛为：
- 页面装配层
- 数据入口层
- 少量状态协调层

而不是继续作为一个超大功能承载文件。
合作为“后续收口”，而不是下一刀的第一选择

---

## 8. 当前一句话路线图

当前拆分路线建议保持为：

`运行中提示已拆出 -> 历史记录 -> 创建任务表单 -> 任务列表 -> 页面级 WebSocket/刷新协调`

这条路线的目标，是让 `TaskView.vue` 最终收敛为：
- 页面装配层
- 数据入口层
- 少量状态协调层

而不是继续作为一个超大功能承载文件。
是让 `TaskView.vue` 最终收敛为：
- 页面装配层
- 数据入口层
- 少量状态协调层

而不是继续作为一个超大功能承载文件。

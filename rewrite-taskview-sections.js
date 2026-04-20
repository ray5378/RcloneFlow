const { execSync } = require('child_process')
const fs = require('fs')

let text = execSync('git show e95393b:frontend/src/views/TaskView.vue', { encoding: 'utf8' })

text = text.replace(
  `const { toasts, showToast } = useToastCenter()\nconst { normalizeTaskOptions } = useTaskFormNormalize()\n\n// Set up global error handler for composables\nsetErrorHandler((message, type) => {\n  showToast(message, type as 'info' | 'success' | 'error')\n})\n\nconst {`,
  `const { toasts, showToast } = useToastCenter()\nconst { normalizeTaskOptions } = useTaskFormNormalize()\n\n// 页面级错误统一走 toast，避免各个 composable 自己维护分散提示。\nsetErrorHandler((message, type) => {\n  showToast(message, type as 'info' | 'success' | 'error')\n})\n\n// 1) 基础页面状态\nconst {`
)

text = text.replace(
  `} = useTaskViewState()\n\nconst {\n  activeRuns,`,
  `} = useTaskViewState()\n\n// 2) 运行时状态（active runs / 全局统计）\nconst {\n  activeRuns,`
)

text = text.replace(
  `} = useTaskViewRuntimeState()\n\nconst {\n  tasksPage,`,
  `} = useTaskViewRuntimeState()\n\n// 3) 任务列表视图态\nconst {\n  tasksPage,`
)

text = text.replace(
  `} = useTaskListView(tasks)\n\nconst {\n  showDetailModal,`,
  `} = useTaskListView(tasks)\n\n// 4) 运行详情 / 最终总结链\nconst {\n  showDetailModal,`
)

text = text.replace(
  `} = useRunDetailRuntime({ runApi })\n\nconst openRunLogFromHint = (run: any) => openRunLog(run)\n\n// 已移除"实时进度"弹窗逻辑，卡片直接显示稳态进度\nconst {`,
  `} = useRunDetailRuntime({ runApi })\n\n// 5) 主数据加载与展示计算\nconst {`
)

text = text.replace(
  `} = useTaskViewRuntime({\n  tasks,`,
  `} = useTaskViewRuntime({\n  tasks,`
)

text = text.replace(
  `})\n\nconst {\n  setTaskSearch,`,
  `})\n\n// 6) 弹窗 / 辅助动作运行时\nconst {\n  showWebhookModal,\n  webhookForm,\n  setWebhook,\n  saveWebhook,\n  testWebhook,\n  showSingletonModal,\n  singletonForm,\n  setSingletonMode,\n  saveSingleton,\n  showLogModal,\n  logModalTitle,\n  logContent,\n  openRunLog,\n  openMenuId,\n  showConfirm,\n  confirmModal,\n  closeConfirm,\n  confirmAndClose,\n  formatTime,\n  getStatusClass,\n  getStatusText,\n} = useTaskViewAuxRuntime({\n  loadData,\n  showToast,\n  taskApi,\n  getFinalSummary: getFinalSummaryFromComposable,\n})\n\n// 7) 页面级 bridge：分页 / 返回 / 跳页\nconst {\n  setTaskSearch,`
)

const auxBlock = `const {\n  showWebhookModal,\n  webhookForm,\n  setWebhook,\n  saveWebhook,\n  testWebhook,\n  showSingletonModal,\n  singletonForm,\n  setSingletonMode,\n  saveSingleton,\n  showLogModal,\n  logModalTitle,\n  logContent,\n  openRunLog,\n  openMenuId,\n  showConfirm,\n  confirmModal,\n  closeConfirm,\n  confirmAndClose,\n  formatTime,\n  getStatusClass,\n  getStatusText,\n} = useTaskViewAuxRuntime({\n  loadData,\n  showToast,\n  taskApi,\n  getFinalSummary: getFinalSummaryFromComposable,\n})\n\n`
text = text.replace(auxBlock, '')

text = text.replace(
  `})\n\nconst {\n  closeWebhookModal,`,
  `})\n\n// 8) 页面级 bridge：modal 字段绑定与开关态\nconst {\n  closeWebhookModal,`
)

text = text.replace(
  `})\n\nconst {\n  createForm,`,
  `})\n\nconst openRunLogFromHint = (run: any) => openRunLog(run)\n\n// 9) 运行中提示小窗：只消费 active runs 和调试开关，不反向承担主进度链职责\nconst {\n  runningHintVisible,\n  runningHintRun,\n  runningHintDebugOpen,\n  runningHintPhaseText,\n  runningHintProgressText,\n  runningHintDebugInfo,\n  openRunningHint,\n  closeRunningHint,\n  toggleRunningHintDebug,\n  openRunningHintLog,\n} = useRunningHintRuntime(activeRuns, openRunLogFromHint, props.runningHintDebugEnabled === true)\n\n// 10) 运行详情入口编排\nconst {\n  showRunDetail,\n  closeRunDetail,\n} = useRunDetailEntry({\n  openRunningHint,\n  openRunDetailModal,\n  openRunDetailFiles,\n  closeRunDetailModal,\n})\n\n// move 模式时，成功数量代表 Moved 条数；已在后端合并 Copied+Deleted 为 Moved\n\n// 11) 任务编辑器运行时\nconst {\n  createForm,`
)

const oldHintBlock = `const openRunLogFromHint = (run: any) => openRunLog(run)\n\n// 已移除"实时进度"弹窗逻辑，卡片直接显示稳态进度\nconst {\n  runningHintVisible,\n  runningHintRun,\n  runningHintDebugOpen,\n  runningHintPhaseText,\n  runningHintProgressText,\n  runningHintDebugInfo,\n  openRunningHint,\n  closeRunningHint,\n  toggleRunningHintDebug,\n  openRunningHintLog,\n} = useRunningHintRuntime(activeRuns, openRunLogFromHint, props.runningHintDebugEnabled === true)\n\nconst {\n  showRunDetail,\n  closeRunDetail,\n} = useRunDetailEntry({\n  openRunningHint,\n  openRunDetailModal,\n  openRunDetailFiles,\n  closeRunDetailModal,\n})\n`
text = text.replace(oldHintBlock, '')

text = text.replace(
  `} = useTaskFormRuntime({\n  schedules,`,
  `} = useTaskFormRuntime({\n  schedules,`
)

text = text.replace(
  `})\n\nconst {\n  filteredRuns,`,
  `})\n\n// 12) 历史列表运行时\nconst {\n  filteredRuns,`
)

text = text.replace(
  `})\n\nconst {\n  deleteTask,`,
  `})\n\n// 13) 任务列表动作运行时\nconst {\n  deleteTask,`
)

fs.writeFileSync('frontend/src/views/TaskView.vue', text)

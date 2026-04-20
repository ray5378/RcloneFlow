const fs = require('fs')

const p = 'frontend/src/views/TaskView.vue'
let t = fs.readFileSync(p, 'utf8')

const hintBlock = `const openRunLogFromHint = (run: any) => openRunLog(run)

// 已移除"实时进度"弹窗逻辑，卡片直接显示稳态进度
const {
  runningHintVisible,
  runningHintRun,
  runningHintDebugOpen,
  runningHintPhaseText,
  runningHintProgressText,
  runningHintDebugInfo,
  openRunningHint,
  closeRunningHint,
  toggleRunningHintDebug,
  openRunningHintLog,
} = useRunningHintRuntime(activeRuns, openRunLogFromHint, props.runningHintDebugEnabled === true)

const {
  showRunDetail,
  closeRunDetail,
} = useRunDetailEntry({
  openRunningHint,
  openRunDetailModal,
  openRunDetailFiles,
  closeRunDetailModal,
})
`
if (!t.includes(hintBlock)) throw new Error('hint block not found')
t = t.replace(hintBlock, '')

const modalBindingsBlock = `const {
  closeWebhookModal,
  closeSingletonModal,
  closeLogModal,
  closeGlobalStatsModal,
  setWebhookTriggerId,
  setWebhookPostUrl,
  setWebhookWecomUrl,
  setWebhookNotifyManual,
  setWebhookNotifySchedule,
  setWebhookNotifyWebhook,
  setWebhookStatusSuccess,
  setWebhookStatusFailed,
  setSingletonEnabled,
  setCommandMode,
  setCommandText,
  setShowAdvancedOptions,
} = useTaskViewModalBindings({
  showWebhookModal,
  webhookForm,
  showSingletonModal,
  singletonForm,
  showLogModal,
  commandMode,
  commandText,
  showAdvancedOptions,
  showGlobalStatsModal,
})

`
if (!t.includes(modalBindingsBlock)) throw new Error('modal bindings block not found')
t = t.replace(modalBindingsBlock, '')

const auxAnchor = `} = useTaskViewAuxRuntime({
  loadData,
  showToast,
  taskApi,
  getFinalSummary: getFinalSummaryFromComposable,
})

`
if (!t.includes(auxAnchor)) throw new Error('aux anchor not found')

t = t.replace(
  auxAnchor,
  auxAnchor + `const openRunLogFromHint = (run: any) => openRunLog(run)

// 已移除"实时进度"弹窗逻辑，卡片直接显示稳态进度。
// 注意：这里依赖 openRunLog，因此必须放在 useTaskViewAuxRuntime 之后。
const {
  runningHintVisible,
  runningHintRun,
  runningHintDebugOpen,
  runningHintPhaseText,
  runningHintProgressText,
  runningHintDebugInfo,
  openRunningHint,
  closeRunningHint,
  toggleRunningHintDebug,
  openRunningHintLog,
} = useRunningHintRuntime(activeRuns, openRunLogFromHint, props.runningHintDebugEnabled === true)

const {
  showRunDetail,
  closeRunDetail,
} = useRunDetailEntry({
  openRunningHint,
  openRunDetailModal,
  openRunDetailFiles,
  closeRunDetailModal,
})

// webhook / singleton / editor modal 绑定桥只负责字段级 UI 接线；
// 这里同样依赖 aux runtime 暴露出的表单 ref，因此必须放在其后。
` + modalBindingsBlock
)

fs.writeFileSync(p, t)

import { ref, computed, watch } from 'vue'
import { useTaskViewRuntimeState } from './useTaskViewRuntimeState'
import { useTaskViewRuntime } from './useTaskViewRuntime'
import { useTaskViewAuxRuntime } from './useTaskViewAuxRuntime'
import { useTaskViewPagingBridge } from './useTaskViewPagingBridge'
import { useTaskViewModalBindings } from './useTaskViewModalBindings'
import { useTaskListView } from './useTaskListView'
import { useTaskHistoryRuntime } from './useTaskHistoryRuntime'
import { useTaskListRuntime } from './useTaskListRuntime'
import { useTaskFormRuntime } from './useTaskFormRuntime'
import { useTaskFormNormalize } from './useTaskFormNormalize'
import { useTaskTags } from './useTaskTags'
import { useToastCenter } from './useToastCenter'
import { useActiveTransferDetail } from './useActiveTransferDetail'
import { useRunningHintRuntime } from './useRunningHintRuntime'
import { useRunDetailRuntime } from './useRunDetailRuntime'
import { useRunDetailEntry } from './useRunDetailEntry'
import { useScheduleConfigModal } from './useScheduleConfigModal'
import { taskApi, remoteApi, runApi, jobApi, scheduleApi } from './useApi'
import { setErrorHandler } from './useError'
import { parseRcloneCommand } from './useTaskCommandParse'
import { t } from '../i18n'
import { formatBytes, formatBytesPerSec, formatEta } from '../utils/format'

export interface UseTaskViewReturn {
  toasts: ReturnType<typeof useToastCenter>['toasts']
  showToast: ReturnType<typeof useToastCenter>['showToast']
  currentModule: ReturnType<typeof useTaskViewRuntimeState>['currentModule']
  tasks: ReturnType<typeof useTaskViewRuntimeState>['tasks']
  schedules: ReturnType<typeof useTaskViewRuntimeState>['schedules']
  remotes: ReturnType<typeof useTaskViewRuntimeState>['remotes']
  filteredTasks: ReturnType<typeof useTaskListView>['filteredTasks']
  tasksTotal: ReturnType<typeof useTaskListView>['tasksTotal']
  tasksPage: ReturnType<typeof useTaskListView>['tasksPage']
  tasksPageSize: ReturnType<typeof useTaskListView>['tasksPageSize']
  currentTasksPages: ReturnType<typeof useTaskListView>['currentTasksPages']
  tasksJumpPage: ReturnType<typeof useTaskListView>['tasksJumpPage']
  taskSearch: ReturnType<typeof useTaskListView>['taskSearch']
  filteredTasksRaw: ReturnType<typeof useTaskListView>['filteredTasksRaw']
  setTaskSearch: ReturnType<typeof useTaskViewPagingBridge>['setTaskSearch']
  setTasksJumpPageValue: ReturnType<typeof useTaskViewPagingBridge>['setTasksJumpPageValue']
  prevTasksPage: ReturnType<typeof useTaskViewPagingBridge>['prevTasksPage']
  nextTasksPage: ReturnType<typeof useTaskViewPagingBridge>['nextTasksPage']
  jumpToTasksPage: ReturnType<typeof useTaskListView>['jumpToTasksPage']
  actionTags: ReturnType<typeof useTaskTags>['actionTags']
  keywordTags: ReturnType<typeof useTaskTags>['keywordTags']
  selectedKeywordTags: ReturnType<typeof useTaskTags>['selectedKeywordTags']
  suggestedTags: ReturnType<typeof useTaskTags>['suggestedTags']
  tagManagerVisible: ReturnType<typeof ref<boolean>>
  openTagManager: () => void
  closeTagManager: () => void
  handleCreateTag: (tag: string) => Promise<void>
  handleSelectTag: (tag: string) => Promise<void>
  handleUnselectTag: (tag: string) => Promise<void>
  handleDeleteTag: (tag: string) => Promise<void>
  bisyncLstManagerVisible: ReturnType<typeof ref<boolean>>
  bisyncLstManagerTaskId: ReturnType<typeof ref<number | null>>
  bisyncLstManagerOptions: ReturnType<typeof ref<any>>
  openBisyncLstManager: (task: any) => void
  closeBisyncLstManager: () => void
  runningTaskId: ReturnType<typeof useTaskListRuntime>['runningTaskId']
  stoppedTaskId: ReturnType<typeof useTaskListRuntime>['stoppedTaskId']
  runTask: ReturnType<typeof useTaskListRuntime>['runTask']
  goToAddTask: ReturnType<typeof useTaskListRuntime>['goToAddTask']
  editTask: ReturnType<typeof useTaskListRuntime>['editTask']
  deleteTask: ReturnType<typeof useTaskListRuntime>['deleteTask']
  stopTaskAny: ReturnType<typeof useTaskListRuntime>['stopTaskAny']
  saveTaskSortOrders: ReturnType<typeof useTaskListRuntime>['saveTaskSortOrders']
  openScheduleConfigForTask: ReturnType<typeof useScheduleConfigModal>['openScheduleConfigForTask']
  getScheduleByTaskId: ReturnType<typeof useTaskFormRuntime>['getScheduleByTaskId']
  getTaskCardProgressByTask: ReturnType<typeof useTaskViewRuntime>['getTaskCardProgressByTask']
  openMenuId: ReturnType<typeof useTaskViewAuxRuntime>['openMenuId']
  showConfirm: ReturnType<typeof useTaskViewAuxRuntime>['showConfirm']
  confirmModal: ReturnType<typeof useTaskViewAuxRuntime>['confirmModal']
  closeConfirm: ReturnType<typeof useTaskViewAuxRuntime>['closeConfirm']
  confirmAndClose: ReturnType<typeof useTaskViewAuxRuntime>['confirmAndClose']
  clearAllRunsWithConfirm: ReturnType<typeof useTaskListRuntime>['clearAllRunsWithConfirm']
  loadData: ReturnType<typeof useTaskViewRuntime>['loadData']
  loadActiveRuns: ReturnType<typeof useTaskViewRuntime>['loadActiveRuns']
  runs: ReturnType<typeof useTaskViewRuntimeState>['runs']
  runsTotal: ReturnType<typeof useTaskViewRuntimeState>['runsTotal']
  taskRuns: ReturnType<typeof useTaskViewRuntimeState>['taskRuns']
  runsPage: ReturnType<typeof useTaskViewRuntimeState>['runsPage']
  runsPageSize: ReturnType<typeof useTaskViewRuntimeState>['runsPageSize']
  jumpPage: ReturnType<typeof useTaskViewRuntimeState>['jumpPage']
  filteredRuns: ReturnType<typeof useTaskHistoryRuntime>['filteredRuns']
  currentTotal: ReturnType<typeof useTaskHistoryRuntime>['currentTotal']
  currentTotalPages: ReturnType<typeof useTaskHistoryRuntime>['currentTotalPages']
  historyFilterTaskId: ReturnType<typeof useTaskViewRuntimeState>['historyFilterTaskId']
  historyStatusFilter: ReturnType<typeof useTaskViewRuntimeState>['historyStatusFilter']
  setHistoryStatusFilter: ReturnType<typeof useTaskViewPagingBridge>['setHistoryStatusFilter']
  setJumpPageValue: ReturnType<typeof useTaskViewPagingBridge>['setJumpPageValue']
  viewTaskHistory: ReturnType<typeof useTaskHistoryRuntime>['viewTaskHistory']
  prevRunsPage: ReturnType<typeof useTaskViewPagingBridge>['prevRunsPage']
  nextRunsPage: ReturnType<typeof useTaskViewPagingBridge>['nextRunsPage']
  backToTasks: ReturnType<typeof useTaskViewPagingBridge>['backToTasks']
  jumpToPage: ReturnType<typeof useTaskHistoryRuntime>['jumpToPage']
  clearRun: ReturnType<typeof useTaskHistoryRuntime>['clearRun']
  clearAllRuns: ReturnType<typeof useTaskHistoryRuntime>['clearAllRuns']
  showDetailModal: ReturnType<typeof useRunDetailRuntime>['showDetailModal']
  runDetail: ReturnType<typeof useRunDetailRuntime>['runDetail']
  openRunDetailModal: ReturnType<typeof useRunDetailRuntime>['openRunDetailModal']
  closeRunDetailModal: ReturnType<typeof useRunDetailRuntime>['closeRunDetailModal']
  runFilesTotal: ReturnType<typeof useRunDetailRuntime>['runFilesTotal']
  runFilesPage: ReturnType<typeof useRunDetailRuntime>['runFilesPage']
  openRunDetailFiles: ReturnType<typeof useRunDetailRuntime>['openRunDetailFiles']
  pagedRunFiles: ReturnType<typeof useRunDetailRuntime>['pagedRunFiles']
  totalRunFilesPages: ReturnType<typeof useRunDetailRuntime>['totalRunFilesPages']
  goPrevFilesPage: ReturnType<typeof useRunDetailRuntime>['goPrevFilesPage']
  goNextFilesPage: ReturnType<typeof useRunDetailRuntime>['goNextFilesPage']
  getFinalSummaryFromComposable: ReturnType<typeof useRunDetailRuntime>['getFinalSummary']
  finalCountAll: ReturnType<typeof useRunDetailRuntime>['finalCountAll']
  finalCountSuccess: ReturnType<typeof useRunDetailRuntime>['finalCountSuccess']
  finalCountFailed: ReturnType<typeof useRunDetailRuntime>['finalCountFailed']
  showRunDetail: ReturnType<typeof useRunDetailEntry>['showRunDetail']
  closeRunDetail: ReturnType<typeof useRunDetailEntry>['closeRunDetail']
  getRunProgressFromSummary: ReturnType<typeof useTaskViewRuntime>['getRunProgressFromSummary']
  getRealtimeProgressByRun: ReturnType<typeof useTaskViewRuntime>['getRealtimeProgressByRun']
  globalStats: ReturnType<typeof useTaskViewRuntimeState>['globalStats']
  showGlobalStatsModal: ReturnType<typeof useTaskViewRuntimeState>['showGlobalStatsModal']
  closeGlobalStatsModal: ReturnType<typeof useTaskViewModalBindings>['closeGlobalStatsModal']
  runningHintVisible: ReturnType<typeof useRunningHintRuntime>['runningHintVisible']
  runningHintRun: ReturnType<typeof useRunningHintRuntime>['runningHintRun']
  runningHintPhaseText: ReturnType<typeof useRunningHintRuntime>['runningHintPhaseText']
  runningHintProgressText: ReturnType<typeof useRunningHintRuntime>['runningHintProgressText']
  openRunningHint: ReturnType<typeof useRunningHintRuntime>['openRunningHint']
  closeRunningHint: ReturnType<typeof useRunningHintRuntime>['closeRunningHint']
  openRunningHintLog: ReturnType<typeof useRunningHintRuntime>['openRunningHintLog']
  showLogModal: ReturnType<typeof useTaskViewAuxRuntime>['showLogModal']
  logModalTitle: ReturnType<typeof useTaskViewAuxRuntime>['logModalTitle']
  logContent: ReturnType<typeof useTaskViewAuxRuntime>['logContent']
  openRunLog: ReturnType<typeof useTaskViewAuxRuntime>['openRunLog']
  closeLogModal: ReturnType<typeof useTaskViewModalBindings>['closeLogModal']
  showWebhookModal: ReturnType<typeof useTaskViewAuxRuntime>['showWebhookModal']
  webhookForm: ReturnType<typeof useTaskViewAuxRuntime>['webhookForm']
  setWebhook: ReturnType<typeof useTaskViewAuxRuntime>['setWebhook']
  saveWebhook: ReturnType<typeof useTaskViewAuxRuntime>['saveWebhook']
  testWebhook: ReturnType<typeof useTaskViewAuxRuntime>['testWebhook']
  regenerateWebhookSecret: ReturnType<typeof useTaskViewAuxRuntime>['regenerateWebhookSecret']
  setWebhookTriggerId: ReturnType<typeof useTaskViewModalBindings>['setWebhookTriggerId']
  setWebhookMatchText: ReturnType<typeof useTaskViewModalBindings>['setWebhookMatchText']
  setWebhookSecret: ReturnType<typeof useTaskViewModalBindings>['setWebhookSecret']
  setWebhookPostUrl: ReturnType<typeof useTaskViewModalBindings>['setWebhookPostUrl']
  setWebhookWecomUrl: ReturnType<typeof useTaskViewModalBindings>['setWebhookWecomUrl']
  setWebhookNotifyManual: ReturnType<typeof useTaskViewModalBindings>['setWebhookNotifyManual']
  setWebhookNotifySchedule: ReturnType<typeof useTaskViewModalBindings>['setWebhookNotifySchedule']
  setWebhookNotifyWebhook: ReturnType<typeof useTaskViewModalBindings>['setWebhookNotifyWebhook']
  setWebhookStatusSuccess: ReturnType<typeof useTaskViewModalBindings>['setWebhookStatusSuccess']
  setWebhookStatusFailed: ReturnType<typeof useTaskViewModalBindings>['setWebhookStatusFailed']
  setWebhookStatusHasTransfer: ReturnType<typeof useTaskViewModalBindings>['setWebhookStatusHasTransfer']
  closeWebhookModal: ReturnType<typeof useTaskViewModalBindings>['closeWebhookModal']
  showSingletonModal: ReturnType<typeof useTaskViewAuxRuntime>['showSingletonModal']
  singletonForm: ReturnType<typeof useTaskViewAuxRuntime>['singletonForm']
  setSingletonMode: ReturnType<typeof useTaskViewAuxRuntime>['setSingletonMode']
  saveSingleton: ReturnType<typeof useTaskViewAuxRuntime>['saveSingleton']
  setSingletonEnabled: ReturnType<typeof useTaskViewModalBindings>['setSingletonEnabled']
  closeSingletonModal: ReturnType<typeof useTaskViewModalBindings>['closeSingletonModal']
  scheduleConfigVisible: ReturnType<typeof useScheduleConfigModal>['scheduleConfigVisible']
  scheduleConfigSaving: ReturnType<typeof useScheduleConfigModal>['scheduleConfigSaving']
  scheduleConfigDraft: ReturnType<typeof useScheduleConfigModal>['scheduleConfigDraft']
  scheduleConfigTitle: ReturnType<typeof useScheduleConfigModal>['scheduleConfigTitle']
  closeScheduleConfig: ReturnType<typeof useScheduleConfigModal>['closeScheduleConfig']
  saveScheduleConfig: ReturnType<typeof useScheduleConfigModal>['saveScheduleConfig']
  taskEditorVisible: ReturnType<typeof computed<boolean>>
  taskEditorTitle: ReturnType<typeof computed<string>>
  commandMode: ReturnType<typeof useTaskFormRuntime>['commandMode']
  commandText: ReturnType<typeof useTaskFormRuntime>['commandText']
  createForm: ReturnType<typeof useTaskFormRuntime>['createForm']
  editingTask: ReturnType<typeof useTaskFormRuntime>['editingTask']
  showAdvancedOptions: ReturnType<typeof useTaskFormRuntime>['showAdvancedOptions']
  creatingState: ReturnType<typeof useTaskFormRuntime>['creatingState']
  createTask: ReturnType<typeof useTaskFormRuntime>['createTask']
  sourcePathOptions: ReturnType<typeof useTaskFormRuntime>['sourcePathOptions']
  targetPathOptions: ReturnType<typeof useTaskFormRuntime>['targetPathOptions']
  showSourcePathInput: ReturnType<typeof useTaskFormRuntime>['showSourcePathInput']
  showTargetPathInput: ReturnType<typeof useTaskFormRuntime>['showTargetPathInput']
  sourceCurrentPath: ReturnType<typeof useTaskFormRuntime>['sourceCurrentPath']
  targetCurrentPath: ReturnType<typeof useTaskFormRuntime>['targetCurrentPath']
  sourceBreadcrumbs: ReturnType<typeof useTaskFormRuntime>['sourceBreadcrumbs']
  targetBreadcrumbs: ReturnType<typeof useTaskFormRuntime>['targetBreadcrumbs']
  setShowSourcePathInput: ReturnType<typeof useTaskFormRuntime>['setShowSourcePathInput']
  setShowTargetPathInput: ReturnType<typeof useTaskFormRuntime>['setShowTargetPathInput']
  resetTaskPathBrowse: ReturnType<typeof useTaskFormRuntime>['resetTaskPathBrowse']
  restoreTaskPathBrowse: ReturnType<typeof useTaskFormRuntime>['restoreTaskPathBrowse']
  onSourceRemoteChange: ReturnType<typeof useTaskFormRuntime>['onSourceRemoteChange']
  onTargetRemoteChange: ReturnType<typeof useTaskFormRuntime>['onTargetRemoteChange']
  onSourceBreadcrumbClick: ReturnType<typeof useTaskFormRuntime>['onSourceBreadcrumbClick']
  onTargetBreadcrumbClick: ReturnType<typeof useTaskFormRuntime>['onTargetBreadcrumbClick']
  onSourceClick: ReturnType<typeof useTaskFormRuntime>['onSourceClick']
  onSourceArrow: ReturnType<typeof useTaskFormRuntime>['onSourceArrow']
  onTargetClick: ReturnType<typeof useTaskFormRuntime>['onTargetClick']
  onTargetArrow: ReturnType<typeof useTaskFormRuntime>['onTargetArrow']
  setCommandMode: ReturnType<typeof useTaskViewModalBindings>['setCommandMode']
  setCommandText: ReturnType<typeof useTaskViewModalBindings>['setCommandText']
  setShowAdvancedOptions: ReturnType<typeof useTaskViewModalBindings>['setShowAdvancedOptions']
  closeTaskEditorModal: () => void
  activeTransferVisible: ReturnType<typeof useActiveTransferDetail>['activeTransferVisible']
  activeTransferTrackingMode: ReturnType<typeof useActiveTransferDetail>['activeTransferTrackingMode']
  activeTransferSummary: ReturnType<typeof useActiveTransferDetail>['activeTransferSummary']
  activeTransferCurrentFile: ReturnType<typeof useActiveTransferDetail>['activeTransferCurrentFile']
  activeTransferCurrentFiles: ReturnType<typeof useActiveTransferDetail>['activeTransferCurrentFiles']
  activeTransferSlots: ReturnType<typeof useActiveTransferDetail>['activeTransferSlots']
  activeTransferCompletedItems: ReturnType<typeof useActiveTransferDetail>['activeTransferCompletedItems']
  activeTransferPendingItems: ReturnType<typeof useActiveTransferDetail>['activeTransferPendingItems']
  activeTransferCompletedTotal: ReturnType<typeof useActiveTransferDetail>['activeTransferCompletedTotal']
  activeTransferPendingTotal: ReturnType<typeof useActiveTransferDetail>['activeTransferPendingTotal']
  activeTransferCompletedPage: ReturnType<typeof useActiveTransferDetail>['activeTransferCompletedPage']
  activeTransferPendingPage: ReturnType<typeof useActiveTransferDetail>['activeTransferPendingPage']
  activeTransferCompletedJumpPage: ReturnType<typeof useActiveTransferDetail>['activeTransferCompletedJumpPage']
  activeTransferPendingJumpPage: ReturnType<typeof useActiveTransferDetail>['activeTransferPendingJumpPage']
  activeTransferCompletedTotalPages: ReturnType<typeof useActiveTransferDetail>['activeTransferCompletedTotalPages']
  activeTransferPendingTotalPages: ReturnType<typeof useActiveTransferDetail>['activeTransferPendingTotalPages']
  activeTransferDegraded: ReturnType<typeof useActiveTransferDetail>['activeTransferDegraded']
  activeTransferLoading: ReturnType<typeof useActiveTransferDetail>['activeTransferLoading']
  activeTransferError: ReturnType<typeof useActiveTransferDetail>['activeTransferError']
  openActiveTransfer: ReturnType<typeof useActiveTransferDetail>['openActiveTransfer']
  closeActiveTransfer: ReturnType<typeof useActiveTransferDetail>['closeActiveTransfer']
  refreshActiveTransfer: ReturnType<typeof useActiveTransferDetail>['refreshActiveTransfer']
  prevActiveTransferCompletedPage: ReturnType<typeof useActiveTransferDetail>['prevActiveTransferCompletedPage']
  nextActiveTransferCompletedPage: ReturnType<typeof useActiveTransferDetail>['nextActiveTransferCompletedPage']
  jumpActiveTransferCompletedPage: ReturnType<typeof useActiveTransferDetail>['jumpActiveTransferCompletedPage']
  prevActiveTransferPendingPage: ReturnType<typeof useActiveTransferDetail>['prevActiveTransferPendingPage']
  nextActiveTransferPendingPage: ReturnType<typeof useActiveTransferDetail>['nextActiveTransferPendingPage']
  jumpActiveTransferPendingPage: ReturnType<typeof useActiveTransferDetail>['jumpActiveTransferPendingPage']
  activeRuns: ReturnType<typeof useTaskViewRuntimeState>['activeRuns']
  activeRunLookup: ReturnType<typeof useTaskViewRuntimeState>['activeRunLookup']
  formatBps: ReturnType<typeof useTaskViewRuntime>['formatBps']
  formatBytes: typeof formatBytes
  formatBytesPerSec: typeof formatBytesPerSec
  formatEta: typeof formatEta
  formatTime: ReturnType<typeof useTaskViewAuxRuntime>['formatTime']
  getStatusClass: ReturnType<typeof useTaskViewAuxRuntime>['getStatusClass']
  getStatusText: ReturnType<typeof useTaskViewAuxRuntime>['getStatusText']
}

function useTaskViewTagManager(tasks: any, reloadTags: () => void) {
  const tagManagerVisible = ref(false)

  function openTagManager() {
    tagManagerVisible.value = true
  }

  function closeTagManager() {
    tagManagerVisible.value = false
  }

  watch(tasks, () => { reloadTags() }, { deep: true })

  return {
    tagManagerVisible,
    openTagManager,
    closeTagManager,
  }
}

function useTaskViewBisyncManager() {
  const bisyncLstManagerVisible = ref(false)
  const bisyncLstManagerTaskId = ref<number | null>(null)
  const bisyncLstManagerOptions = ref<any>({})

  function openBisyncLstManager(task: any) {
    bisyncLstManagerTaskId.value = task.id
    bisyncLstManagerOptions.value = task.bisyncOptions || {}
    bisyncLstManagerVisible.value = true
  }

  function closeBisyncLstManager() {
    bisyncLstManagerVisible.value = false
    bisyncLstManagerTaskId.value = null
    bisyncLstManagerOptions.value = {}
  }

  return {
    bisyncLstManagerVisible,
    bisyncLstManagerTaskId,
    bisyncLstManagerOptions,
    openBisyncLstManager,
    closeBisyncLstManager,
  }
}

function useTaskViewEditor(
  currentModule: any,
  editingTask: any,
  creatingState: any,
  commandMode: any,
  commandText: any,
  createForm: any,
  showSourcePathInput: any,
  showTargetPathInput: any,
  showAdvancedOptions: any,
  showConfirm: any,
) {
  const taskEditorVisible = computed(() => currentModule.value === 'add')
  const taskEditorTitle = computed(() => editingTask.value ? t('taskEditor.editTitle') : t('taskEditor.createTitle'))

  const taskEditorSnapshot = computed(() => JSON.stringify({
    commandMode: commandMode.value,
    commandText: commandText.value,
    createForm: createForm.value,
    showSourcePathInput: showSourcePathInput.value,
    showTargetPathInput: showTargetPathInput.value,
    showAdvancedOptions: showAdvancedOptions.value,
  }))

  const taskEditorBaseline = computed(() => JSON.stringify({
    commandMode: false,
    commandText: '',
    createForm: {
      name: editingTask.value?.name ?? '',
      mode: editingTask.value ? createForm.value.mode : 'copy',
      sourceRemote: editingTask.value?.sourceRemote ?? '',
      sourcePath: editingTask.value?.sourcePath ?? '',
      targetRemote: editingTask.value?.targetRemote ?? '',
      targetPath: editingTask.value?.targetPath ?? '',
      options: editingTask.value ? createForm.value.options : { enableStreaming: true },
      bisyncOptions: editingTask.value?.bisyncOptions ?? {},
    },
    showSourcePathInput: false,
    showTargetPathInput: false,
    showAdvancedOptions: false,
  }))

  const hasTaskEditorChanges = computed(() => taskEditorSnapshot.value !== taskEditorBaseline.value)

  function doCloseTaskEditorModal() {
    creatingState.value = 'idle'
    currentModule.value = 'tasks'
  }

  function closeTaskEditorModal() {
    if (creatingState.value === 'loading') return
    if (!hasTaskEditorChanges.value) {
      doCloseTaskEditorModal()
      return
    }
    showConfirm(
      t('taskEditor.closeConfirmTitle'),
      t('taskEditor.closeConfirmMessage'),
      () => doCloseTaskEditorModal(),
    )
  }

  return {
    taskEditorVisible,
    taskEditorTitle,
    closeTaskEditorModal,
  }
}

export function useTaskView(): UseTaskViewReturn {
  const { toasts, showToast } = useToastCenter()
  const { normalizeTaskOptions } = useTaskFormNormalize()

  const {
    activeTransferVisible,
    activeTransferTrackingMode,
    activeTransferSummary,
    activeTransferCurrentFile,
    activeTransferCurrentFiles,
    activeTransferSlots,
    activeTransferCompletedItems,
    activeTransferPendingItems,
    activeTransferCompletedTotal,
    activeTransferPendingTotal,
    activeTransferCompletedPage,
    activeTransferPendingPage,
    activeTransferCompletedJumpPage,
    activeTransferPendingJumpPage,
    activeTransferCompletedTotalPages,
    activeTransferPendingTotalPages,
    activeTransferDegraded,
    activeTransferLoading,
    activeTransferError,
    openActiveTransfer,
    closeActiveTransfer,
    refreshActiveTransfer,
    prevActiveTransferCompletedPage,
    nextActiveTransferCompletedPage,
    jumpActiveTransferCompletedPage,
    prevActiveTransferPendingPage,
    nextActiveTransferPendingPage,
    jumpActiveTransferPendingPage,
  } = useActiveTransferDetail()

  setErrorHandler((message, type) => {
    showToast(message, type as 'info' | 'success' | 'error')
  })

  const {
    tasks,
    schedules,
    runs,
    runsTotal,
    taskRuns,
    runsPage,
    runsPageSize,
    jumpPage,
    remotes,
    currentModule,
    historyFilterTaskId,
    historyStatusFilter,
    activeRuns,
    globalStats,
    showGlobalStatsModal,
    activeRunLookup,
    lastNonDecreasingTotalsByTask,
    STUCK_MS,
  } = useTaskViewRuntimeState()

  const runningTaskIds = computed(() => {
    const ids = new Set<number>()
    for (const item of activeRuns.value || []) {
      const tid = Number(item?.runRecord?.taskId ?? item?.taskId ?? 0)
      if (tid > 0) ids.add(tid)
    }
    return ids
  })

  const {
    tasksPage,
    tasksPageSize,
    tasksJumpPage,
    taskSearch,
    tasksTotal,
    currentTasksPages,
    filteredTasksRaw,
    filteredTasks,
    jumpToTasksPage,
  } = useTaskListView(tasks, runningTaskIds)

  const {
    actionTags,
    keywordTags,
    selectedKeywordTags,
    suggestedTags,
    reload: reloadTags,
    toggleTag,
    createManualTag,
    deleteManualTag,
  } = useTaskTags(tasks)

  const {
    tagManagerVisible,
    openTagManager,
    closeTagManager,
  } = useTaskViewTagManager(tasks, reloadTags)

  async function handleCreateTag(tag: string) {
    await createManualTag(tag)
  }

  async function handleSelectTag(tag: string) {
    await toggleTag(tag, true)
  }

  async function handleUnselectTag(tag: string) {
    await toggleTag(tag, false)
  }

  async function handleDeleteTag(tag: string) {
    await deleteManualTag(tag)
  }

  const {
    bisyncLstManagerVisible,
    bisyncLstManagerTaskId,
    bisyncLstManagerOptions,
    openBisyncLstManager,
    closeBisyncLstManager,
  } = useTaskViewBisyncManager()

  const {
    showDetailModal,
    runDetail,
    openRunDetailModal,
    closeRunDetailModal,
    runFilesTotal,
    runFilesPage,
    openRunDetailFiles,
    pagedRunFiles,
    totalRunFilesPages,
    goPrevFilesPage,
    goNextFilesPage,
    getFinalSummary: getFinalSummaryFromComposable,
    finalCountAll,
    finalCountSuccess,
    finalCountFailed,
  } = useRunDetailRuntime({ runApi })

  const {
    loadData,
    loadActiveRuns,
    getRunProgressFromSummary,
    getRealtimeProgressByRun,
    getTaskCardProgressByTask,
    formatBps,
  } = useTaskViewRuntime({
    tasks,
    remotes,
    schedules,
    runs,
    runsTotal,
    runsPage,
    runsPageSize,
    activeRuns,
    globalStats,
    showGlobalStatsModal,
    activeRunLookup,
    lastNonDecreasingTotalsByTask,
    currentModule,
    stuckMs: STUCK_MS,
    taskApi,
    remoteApi,
    scheduleApi,
    runApi,
    jobApi,
  })

  const {
    setTaskSearch,
    setTasksJumpPageValue,
    setHistoryStatusFilter,
    setJumpPageValue,
    prevTasksPage,
    nextTasksPage,
    backToTasks,
    prevRunsPage,
    nextRunsPage,
  } = useTaskViewPagingBridge({
    taskSearch,
    tasksJumpPage,
    historyStatusFilter,
    jumpPage,
    tasksPage,
    runsPage,
    currentModule,
    loadData,
  })

  const {
    showWebhookModal,
    webhookForm,
    setWebhook,
    saveWebhook,
    testWebhook,
    regenerateWebhookSecret,
    showSingletonModal,
    singletonForm,
    setSingletonMode,
    saveSingleton,
    showLogModal,
    logModalTitle,
    logContent,
    openRunLog,
    openMenuId,
    showConfirm,
    confirmModal,
    closeConfirm,
    confirmAndClose,
    formatTime,
    getStatusClass,
    getStatusText,
  } = useTaskViewAuxRuntime({
    loadData,
    showToast,
    taskApi,
    getFinalSummary: getFinalSummaryFromComposable,
  })

  const openRunLogFromHint = (run: any) => openRunLog(run)

  const {
    runningHintVisible,
    runningHintRun,
    runningHintPhaseText,
    runningHintProgressText,
    openRunningHint,
    closeRunningHint,
    openRunningHintLog,
  } = useRunningHintRuntime(activeRuns, openRunLogFromHint)

  const {
    showRunDetail,
    closeRunDetail,
  } = useRunDetailEntry({
    openRunningHint,
    openRunDetailModal,
    openRunDetailFiles,
    closeRunDetailModal,
  })

  const {
    createForm,
    commandMode,
    commandText,
    editingTask,
    showAdvancedOptions,
    resetTaskFormForCreate,
    fillTaskFormForEdit,
    getScheduleByTaskId,
    creatingState,
    createTask,
    sourcePathOptions,
    targetPathOptions,
    showSourcePathInput,
    showTargetPathInput,
    sourceCurrentPath,
    targetCurrentPath,
    sourceBreadcrumbs,
    targetBreadcrumbs,
    setShowSourcePathInput,
    setShowTargetPathInput,
    resetTaskPathBrowse,
    restoreTaskPathBrowse,
    onSourceRemoteChange,
    onTargetRemoteChange,
    onSourceBreadcrumbClick,
    onTargetBreadcrumbClick,
    onSourceClick,
    onSourceArrow,
    onTargetClick,
    onTargetArrow,
  } = useTaskFormRuntime({
    schedules,
    currentModule,
    normalizeTaskOptions,
    loadData,
    taskApi,
    scheduleApi,
    showToast,
    parseRcloneCommand,
  })

  const {
    closeWebhookModal,
    closeSingletonModal,
    closeLogModal,
    closeGlobalStatsModal,
    setWebhookTriggerId,
    setWebhookMatchText,
    setWebhookSecret,
    setWebhookPostUrl,
    setWebhookWecomUrl,
    setWebhookNotifyManual,
    setWebhookNotifySchedule,
    setWebhookNotifyWebhook,
    setWebhookStatusSuccess,
    setWebhookStatusFailed,
    setWebhookStatusHasTransfer,
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

  const {
    filteredRuns,
    currentTotal,
    currentTotalPages,
    viewTaskHistory,
    jumpToPage,
    clearRun,
    clearAllRuns,
  } = useTaskHistoryRuntime({
    runs,
    runsTotal,
    taskRuns,
    historyFilterTaskId,
    historyStatusFilter,
    runsPage,
    runsPageSize,
    jumpPage,
    currentModule,
    getFinalSummary: getFinalSummaryFromComposable,
    loadData,
    runApi,
  })

  const {
    deleteTask,
    clearAllRunsWithConfirm,
    runningTaskId,
    stoppedTaskId,
    stopTaskAny,
    runTask,
    goToAddTask,
    editTask,
    saveTaskSortOrders,
  } = useTaskListRuntime({
    openMenuId,
    historyFilterTaskId,
    schedules,
    loadData,
    loadActiveRuns,
    showConfirm,
    showToast,
    clearAllRuns,
    currentModule,
    remotes,
    remoteApi,
    resetTaskFormForCreate,
    resetTaskPathBrowse,
    getScheduleByTaskId,
    fillTaskFormForEdit,
    restoreTaskPathBrowse,
    taskApi,
    scheduleApi,
  })

  const {
    scheduleConfigVisible,
    scheduleConfigSaving,
    scheduleConfigDraft,
    scheduleConfigTitle,
    openScheduleConfigForTask,
    closeScheduleConfig,
    saveScheduleConfig,
  } = useScheduleConfigModal({
    createForm,
    getScheduleByTaskId,
    scheduleApi,
    loadData,
    showToast,
  })

  const {
    taskEditorVisible,
    taskEditorTitle,
    closeTaskEditorModal,
  } = useTaskViewEditor(
    currentModule,
    editingTask,
    creatingState,
    commandMode,
    commandText,
    createForm,
    showSourcePathInput,
    showTargetPathInput,
    showAdvancedOptions,
    showConfirm,
  )

  return {
    toasts,
    showToast,
    currentModule,
    tasks,
    schedules,
    remotes,
    filteredTasks,
    tasksTotal,
    tasksPage,
    tasksPageSize,
    currentTasksPages,
    tasksJumpPage,
    taskSearch,
    filteredTasksRaw,
    setTaskSearch,
    setTasksJumpPageValue,
    prevTasksPage,
    nextTasksPage,
    jumpToTasksPage,
    actionTags,
    keywordTags,
    selectedKeywordTags,
    suggestedTags,
    tagManagerVisible,
    openTagManager,
    closeTagManager,
    handleCreateTag,
    handleSelectTag,
    handleUnselectTag,
    handleDeleteTag,
    bisyncLstManagerVisible,
    bisyncLstManagerTaskId,
    bisyncLstManagerOptions,
    openBisyncLstManager,
    closeBisyncLstManager,
    runningTaskId,
    stoppedTaskId,
    runTask,
    goToAddTask,
    editTask,
    deleteTask,
    stopTaskAny,
    saveTaskSortOrders,
    openScheduleConfigForTask,
    getScheduleByTaskId,
    getTaskCardProgressByTask,
    openMenuId,
    showConfirm,
    confirmModal,
    closeConfirm,
    confirmAndClose,
    clearAllRunsWithConfirm,
    loadData,
    loadActiveRuns,
    runs,
    runsTotal,
    taskRuns,
    runsPage,
    runsPageSize,
    jumpPage,
    filteredRuns,
    currentTotal,
    currentTotalPages,
    historyFilterTaskId,
    historyStatusFilter,
    setHistoryStatusFilter,
    setJumpPageValue,
    viewTaskHistory,
    prevRunsPage,
    nextRunsPage,
    backToTasks,
    jumpToPage,
    clearRun,
    clearAllRuns,
    showDetailModal,
    runDetail,
    openRunDetailModal,
    closeRunDetailModal,
    runFilesTotal,
    runFilesPage,
    openRunDetailFiles,
    pagedRunFiles,
    totalRunFilesPages,
    goPrevFilesPage,
    goNextFilesPage,
    getFinalSummaryFromComposable,
    finalCountAll,
    finalCountSuccess,
    finalCountFailed,
    showRunDetail,
    closeRunDetail,
    getRunProgressFromSummary,
    getRealtimeProgressByRun,
    globalStats,
    showGlobalStatsModal,
    closeGlobalStatsModal,
    runningHintVisible,
    runningHintRun,
    runningHintPhaseText,
    runningHintProgressText,
    openRunningHint,
    closeRunningHint,
    openRunningHintLog,
    showLogModal,
    logModalTitle,
    logContent,
    openRunLog,
    closeLogModal,
    showWebhookModal,
    webhookForm,
    setWebhook,
    saveWebhook,
    testWebhook,
    regenerateWebhookSecret,
    setWebhookTriggerId,
    setWebhookMatchText,
    setWebhookSecret,
    setWebhookPostUrl,
    setWebhookWecomUrl,
    setWebhookNotifyManual,
    setWebhookNotifySchedule,
    setWebhookNotifyWebhook,
    setWebhookStatusSuccess,
    setWebhookStatusFailed,
    setWebhookStatusHasTransfer,
    closeWebhookModal,
    showSingletonModal,
    singletonForm,
    setSingletonMode,
    saveSingleton,
    setSingletonEnabled,
    closeSingletonModal,
    scheduleConfigVisible,
    scheduleConfigSaving,
    scheduleConfigDraft,
    scheduleConfigTitle,
    closeScheduleConfig,
    saveScheduleConfig,
    taskEditorVisible,
    taskEditorTitle,
    commandMode,
    commandText,
    createForm,
    editingTask,
    showAdvancedOptions,
    creatingState,
    createTask,
    sourcePathOptions,
    targetPathOptions,
    showSourcePathInput,
    showTargetPathInput,
    sourceCurrentPath,
    targetCurrentPath,
    sourceBreadcrumbs,
    targetBreadcrumbs,
    setShowSourcePathInput,
    setShowTargetPathInput,
    resetTaskPathBrowse,
    restoreTaskPathBrowse,
    onSourceRemoteChange,
    onTargetRemoteChange,
    onSourceBreadcrumbClick,
    onTargetBreadcrumbClick,
    onSourceClick,
    onSourceArrow,
    onTargetClick,
    onTargetArrow,
    setCommandMode,
    setCommandText,
    setShowAdvancedOptions,
    closeTaskEditorModal,
    activeTransferVisible,
    activeTransferTrackingMode,
    activeTransferSummary,
    activeTransferCurrentFile,
    activeTransferCurrentFiles,
    activeTransferSlots,
    activeTransferCompletedItems,
    activeTransferPendingItems,
    activeTransferCompletedTotal,
    activeTransferPendingTotal,
    activeTransferCompletedPage,
    activeTransferPendingPage,
    activeTransferCompletedJumpPage,
    activeTransferPendingJumpPage,
    activeTransferCompletedTotalPages,
    activeTransferPendingTotalPages,
    activeTransferDegraded,
    activeTransferLoading,
    activeTransferError,
    openActiveTransfer,
    closeActiveTransfer,
    refreshActiveTransfer,
    prevActiveTransferCompletedPage,
    nextActiveTransferCompletedPage,
    jumpActiveTransferCompletedPage,
    prevActiveTransferPendingPage,
    nextActiveTransferPendingPage,
    jumpActiveTransferPendingPage,
    activeRuns,
    activeRunLookup,
    formatBps,
    formatBytes,
    formatBytesPerSec,
    formatEta,
    formatTime,
    getStatusClass,
    getStatusText,
  }
}

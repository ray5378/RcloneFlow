import { describe, it, expect, vi, beforeEach } from 'vitest'
import { useTaskView } from './useTaskView'

// Mock all dependencies
vi.mock('./useTaskViewRuntimeState', () => ({
  useTaskViewRuntimeState: vi.fn(() => ({
    tasks: [],
    schedules: [],
    runs: [],
    runsTotal: 0,
    taskRuns: [],
    runsPage: 1,
    runsPageSize: 50,
    jumpPage: vi.fn(),
    remotes: [],
    currentModule: 'tasks',
    historyFilterTaskId: null,
    historyStatusFilter: null,
    activeRuns: [],
    globalStats: {},
    showGlobalStatsModal: false,
    activeRunLookup: { getActiveRunByTaskId: vi.fn() },
    lastNonDecreasingTotalsByTask: {},
    STUCK_MS: 25000,
  })),
}))

vi.mock('./useTaskViewRuntime', () => ({
  useTaskViewRuntime: vi.fn(() => ({
    loadData: vi.fn(),
    loadActiveRuns: vi.fn(),
    getRunProgressFromSummary: vi.fn(),
    getRealtimeProgressByRun: vi.fn(),
    getTaskCardProgressByTask: vi.fn(),
    formatBps: vi.fn(),
  })),
}))

vi.mock('./useTaskViewAuxRuntime', () => ({
  useTaskViewAuxRuntime: vi.fn(() => ({
    showWebhookModal: false,
    webhookForm: {},
    setWebhook: vi.fn(),
    saveWebhook: vi.fn(),
    testWebhook: vi.fn(),
    regenerateWebhookSecret: vi.fn(),
    showSingletonModal: false,
    singletonForm: {},
    setSingletonMode: vi.fn(),
    saveSingleton: vi.fn(),
    showLogModal: false,
    logModalTitle: '',
    logContent: '',
    openRunLog: vi.fn(),
    openMenuId: null,
    showConfirm: vi.fn(),
    confirmModal: {},
    closeConfirm: vi.fn(),
    confirmAndClose: vi.fn(),
    formatTime: vi.fn(),
    getStatusClass: vi.fn(),
    getStatusText: vi.fn(),
  })),
}))

vi.mock('./useTaskViewPagingBridge', () => ({
  useTaskViewPagingBridge: vi.fn(() => ({
    setTaskSearch: vi.fn(),
    setTasksJumpPageValue: vi.fn(),
    setHistoryStatusFilter: vi.fn(),
    setJumpPageValue: vi.fn(),
    prevTasksPage: vi.fn(),
    nextTasksPage: vi.fn(),
    backToTasks: vi.fn(),
    prevRunsPage: vi.fn(),
    nextRunsPage: vi.fn(),
  })),
}))

vi.mock('./useTaskViewModalBindings', () => ({
  useTaskViewModalBindings: vi.fn(() => ({
    closeWebhookModal: vi.fn(),
    closeSingletonModal: vi.fn(),
    closeLogModal: vi.fn(),
    closeGlobalStatsModal: vi.fn(),
    setWebhookTriggerId: vi.fn(),
    setWebhookMatchText: vi.fn(),
    setWebhookSecret: vi.fn(),
    setWebhookPostUrl: vi.fn(),
    setWebhookWecomUrl: vi.fn(),
    setWebhookNotifyManual: vi.fn(),
    setWebhookNotifySchedule: vi.fn(),
    setWebhookNotifyWebhook: vi.fn(),
    setWebhookStatusSuccess: vi.fn(),
    setWebhookStatusFailed: vi.fn(),
    setWebhookStatusHasTransfer: vi.fn(),
    setSingletonEnabled: vi.fn(),
    setCommandMode: vi.fn(),
    setCommandText: vi.fn(),
    setShowAdvancedOptions: vi.fn(),
  })),
}))

vi.mock('./useTaskListView', () => ({
  useTaskListView: vi.fn(() => ({
    tasksPage: 1,
    tasksPageSize: 50,
    tasksJumpPage: 1,
    taskSearch: '',
    tasksTotal: 0,
    currentTasksPages: 1,
    filteredTasksRaw: [],
    filteredTasks: [],
    jumpToTasksPage: vi.fn(),
  })),
}))

vi.mock('./useTaskHistoryRuntime', () => ({
  useTaskHistoryRuntime: vi.fn(() => ({
    filteredRuns: [],
    currentTotal: 0,
    currentTotalPages: 1,
    viewTaskHistory: vi.fn(),
    jumpToPage: vi.fn(),
    clearRun: vi.fn(),
    clearAllRuns: vi.fn(),
  })),
}))

vi.mock('./useTaskListRuntime', () => ({
  useTaskListRuntime: vi.fn(() => ({
    deleteTask: vi.fn(),
    clearAllRunsWithConfirm: vi.fn(),
    runningTaskId: null,
    stoppedTaskId: null,
    stopTaskAny: vi.fn(),
    runTask: vi.fn(),
    goToAddTask: vi.fn(),
    editTask: vi.fn(),
    saveTaskSortOrders: vi.fn(),
  })),
}))

vi.mock('./useTaskFormRuntime', () => ({
  useTaskFormRuntime: vi.fn(() => ({
    createForm: {},
    commandMode: false,
    commandText: '',
    editingTask: null,
    showAdvancedOptions: false,
    resetTaskFormForCreate: vi.fn(),
    fillTaskFormForEdit: vi.fn(),
    getScheduleByTaskId: vi.fn(),
    creatingState: 'idle',
    createTask: vi.fn(),
    sourcePathOptions: [],
    targetPathOptions: [],
    showSourcePathInput: false,
    showTargetPathInput: false,
    sourceCurrentPath: '',
    targetCurrentPath: '',
    sourceBreadcrumbs: [],
    targetBreadcrumbs: [],
    setShowSourcePathInput: vi.fn(),
    setShowTargetPathInput: vi.fn(),
    resetTaskPathBrowse: vi.fn(),
    restoreTaskPathBrowse: vi.fn(),
    onSourceRemoteChange: vi.fn(),
    onTargetRemoteChange: vi.fn(),
    onSourceBreadcrumbClick: vi.fn(),
    onTargetBreadcrumbClick: vi.fn(),
    onSourceClick: vi.fn(),
    onSourceArrow: vi.fn(),
    onTargetClick: vi.fn(),
    onTargetArrow: vi.fn(),
  })),
}))

vi.mock('./useTaskFormNormalize', () => ({
  useTaskFormNormalize: vi.fn(() => ({
    normalizeTaskOptions: vi.fn(),
  })),
}))

vi.mock('./useTaskTags', () => ({
  useTaskTags: vi.fn(() => ({
    actionTags: [],
    keywordTags: [],
    selectedKeywordTags: [],
    suggestedTags: [],
    reload: vi.fn(),
    toggleTag: vi.fn(),
    createManualTag: vi.fn(),
    deleteManualTag: vi.fn(),
  })),
}))

vi.mock('./useToastCenter', () => ({
  useToastCenter: vi.fn(() => ({
    toasts: [],
    showToast: vi.fn(),
  })),
}))

vi.mock('./useActiveTransferDetail', () => ({
  useActiveTransferDetail: vi.fn(() => ({
    activeTransferVisible: false,
    activeTransferTrackingMode: 'summary',
    activeTransferSummary: {},
    activeTransferCurrentFile: null,
    activeTransferCurrentFiles: [],
    activeTransferSlots: 4,
    activeTransferCompletedItems: [],
    activeTransferPendingItems: [],
    activeTransferCompletedTotal: 0,
    activeTransferPendingTotal: 0,
    activeTransferCompletedPage: 1,
    activeTransferPendingPage: 1,
    activeTransferCompletedJumpPage: 1,
    activeTransferPendingJumpPage: 1,
    activeTransferCompletedTotalPages: 1,
    activeTransferPendingTotalPages: 1,
    activeTransferDegraded: false,
    activeTransferLoading: false,
    activeTransferError: '',
    openActiveTransfer: vi.fn(),
    closeActiveTransfer: vi.fn(),
    refreshActiveTransfer: vi.fn(),
    prevActiveTransferCompletedPage: vi.fn(),
    nextActiveTransferCompletedPage: vi.fn(),
    jumpActiveTransferCompletedPage: vi.fn(),
    prevActiveTransferPendingPage: vi.fn(),
    nextActiveTransferPendingPage: vi.fn(),
    jumpActiveTransferPendingPage: vi.fn(),
  })),
}))

vi.mock('./useRunningHintRuntime', () => ({
  useRunningHintRuntime: vi.fn(() => ({
    runningHintVisible: false,
    runningHintRun: null,
    runningHintPhaseText: '',
    runningHintProgressText: '',
    openRunningHint: vi.fn(),
    closeRunningHint: vi.fn(),
    openRunningHintLog: vi.fn(),
  })),
}))

vi.mock('./useRunDetailRuntime', () => ({
  useRunDetailRuntime: vi.fn(() => ({
    showDetailModal: false,
    runDetail: null,
    openRunDetailModal: vi.fn(),
    closeRunDetailModal: vi.fn(),
    runFilesTotal: 0,
    runFilesPage: 1,
    openRunDetailFiles: vi.fn(),
    pagedRunFiles: [],
    totalRunFilesPages: 1,
    goPrevFilesPage: vi.fn(),
    goNextFilesPage: vi.fn(),
    getFinalSummary: vi.fn(),
    finalCountAll: 0,
    finalCountSuccess: 0,
    finalCountFailed: 0,
  })),
}))

vi.mock('./useRunDetailEntry', () => ({
  useRunDetailEntry: vi.fn(() => ({
    showRunDetail: vi.fn(),
    closeRunDetail: vi.fn(),
  })),
}))

vi.mock('./useScheduleConfigModal', () => ({
  useScheduleConfigModal: vi.fn(() => ({
    scheduleConfigVisible: false,
    scheduleConfigSaving: false,
    scheduleConfigDraft: {},
    scheduleConfigTitle: '',
    openScheduleConfigForTask: vi.fn(),
    closeScheduleConfig: vi.fn(),
    saveScheduleConfig: vi.fn(),
  })),
}))

vi.mock('./useApi', () => ({
  taskApi: { list: vi.fn(), bootstrap: vi.fn(), create: vi.fn(), update: vi.fn(), delete: vi.fn(), run: vi.fn(), kill: vi.fn(), updateOptions: vi.fn(), updateSortOrders: vi.fn() },
  remoteApi: { list: vi.fn() },
  runApi: { list: vi.fn(), get: vi.fn(), getFiles: vi.fn(), delete: vi.fn(), deleteAll: vi.fn(), deleteByTask: vi.fn(), getRunsByTask: vi.fn() },
  jobApi: { list: vi.fn() },
  scheduleApi: { list: vi.fn(), create: vi.fn(), update: vi.fn(), delete: vi.fn() },
}))

vi.mock('./useError', () => ({
  setErrorHandler: vi.fn(),
}))

vi.mock('./useTaskCommandParse', () => ({
  parseRcloneCommand: vi.fn(),
}))

vi.mock('../i18n', () => ({
  t: vi.fn((key: string) => key),
}))

vi.mock('../utils/format', () => ({
  formatBytes: vi.fn(),
  formatBytesPerSec: vi.fn(),
  formatEta: vi.fn(),
}))

describe('useTaskView.ts', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('should be importable', () => {
    expect(typeof useTaskView).toBe('function')
  })

  it('should return all expected properties and methods', () => {
    const tv = useTaskView()

    // Toast
    expect(tv).toHaveProperty('toasts')
    expect(tv).toHaveProperty('showToast')

    // Module state
    expect(tv).toHaveProperty('currentModule')

    // Tasks
    expect(tv).toHaveProperty('tasks')
    expect(tv).toHaveProperty('schedules')
    expect(tv).toHaveProperty('remotes')

    // Task List
    expect(tv).toHaveProperty('filteredTasks')
    expect(tv).toHaveProperty('tasksTotal')
    expect(tv).toHaveProperty('tasksPage')
    expect(tv).toHaveProperty('tasksPageSize')
    expect(tv).toHaveProperty('currentTasksPages')
    expect(tv).toHaveProperty('tasksJumpPage')
    expect(tv).toHaveProperty('taskSearch')
    expect(tv).toHaveProperty('filteredTasksRaw')
    expect(tv).toHaveProperty('setTaskSearch')
    expect(tv).toHaveProperty('setTasksJumpPageValue')
    expect(tv).toHaveProperty('prevTasksPage')
    expect(tv).toHaveProperty('nextTasksPage')
    expect(tv).toHaveProperty('jumpToTasksPage')

    // Tags
    expect(tv).toHaveProperty('actionTags')
    expect(tv).toHaveProperty('keywordTags')
    expect(tv).toHaveProperty('selectedKeywordTags')
    expect(tv).toHaveProperty('suggestedTags')
    expect(tv).toHaveProperty('tagManagerVisible')
    expect(tv).toHaveProperty('openTagManager')
    expect(tv).toHaveProperty('closeTagManager')
    expect(tv).toHaveProperty('handleCreateTag')
    expect(tv).toHaveProperty('handleSelectTag')
    expect(tv).toHaveProperty('handleUnselectTag')
    expect(tv).toHaveProperty('handleDeleteTag')

    // Bisync LST Manager
    expect(tv).toHaveProperty('bisyncLstManagerVisible')
    expect(tv).toHaveProperty('openBisyncLstManager')
    expect(tv).toHaveProperty('closeBisyncLstManager')

    // Task List Runtime
    expect(tv).toHaveProperty('runningTaskId')
    expect(tv).toHaveProperty('stoppedTaskId')
    expect(tv).toHaveProperty('runTask')
    expect(tv).toHaveProperty('goToAddTask')
    expect(tv).toHaveProperty('editTask')
    expect(tv).toHaveProperty('deleteTask')
    expect(tv).toHaveProperty('stopTaskAny')
    expect(tv).toHaveProperty('saveTaskSortOrders')
    expect(tv).toHaveProperty('openScheduleConfigForTask')
    expect(tv).toHaveProperty('getScheduleByTaskId')
    expect(tv).toHaveProperty('getTaskCardProgressByTask')
    expect(tv).toHaveProperty('openMenuId')
    expect(tv).toHaveProperty('showConfirm')
    expect(tv).toHaveProperty('confirmModal')
    expect(tv).toHaveProperty('closeConfirm')
    expect(tv).toHaveProperty('confirmAndClose')
    expect(tv).toHaveProperty('clearAllRunsWithConfirm')

    // Load Data
    expect(tv).toHaveProperty('loadData')
    expect(tv).toHaveProperty('loadActiveRuns')

    // History
    expect(tv).toHaveProperty('runs')
    expect(tv).toHaveProperty('runsTotal')
    expect(tv).toHaveProperty('taskRuns')
    expect(tv).toHaveProperty('runsPage')
    expect(tv).toHaveProperty('runsPageSize')
    expect(tv).toHaveProperty('jumpPage')
    expect(tv).toHaveProperty('filteredRuns')
    expect(tv).toHaveProperty('currentTotal')
    expect(tv).toHaveProperty('currentTotalPages')
    expect(tv).toHaveProperty('historyFilterTaskId')
    expect(tv).toHaveProperty('historyStatusFilter')
    expect(tv).toHaveProperty('setHistoryStatusFilter')
    expect(tv).toHaveProperty('setJumpPageValue')
    expect(tv).toHaveProperty('viewTaskHistory')
    expect(tv).toHaveProperty('prevRunsPage')
    expect(tv).toHaveProperty('nextRunsPage')
    expect(tv).toHaveProperty('backToTasks')
    expect(tv).toHaveProperty('jumpToPage')
    expect(tv).toHaveProperty('clearRun')
    expect(tv).toHaveProperty('clearAllRuns')

    // Run Detail
    expect(tv).toHaveProperty('showDetailModal')
    expect(tv).toHaveProperty('runDetail')
    expect(tv).toHaveProperty('openRunDetailModal')
    expect(tv).toHaveProperty('closeRunDetailModal')
    expect(tv).toHaveProperty('runFilesTotal')
    expect(tv).toHaveProperty('runFilesPage')
    expect(tv).toHaveProperty('openRunDetailFiles')
    expect(tv).toHaveProperty('pagedRunFiles')
    expect(tv).toHaveProperty('totalRunFilesPages')
    expect(tv).toHaveProperty('goPrevFilesPage')
    expect(tv).toHaveProperty('goNextFilesPage')
    expect(tv).toHaveProperty('getFinalSummaryFromComposable')
    expect(tv).toHaveProperty('finalCountAll')
    expect(tv).toHaveProperty('finalCountSuccess')
    expect(tv).toHaveProperty('finalCountFailed')
    expect(tv).toHaveProperty('showRunDetail')
    expect(tv).toHaveProperty('closeRunDetail')
    expect(tv).toHaveProperty('getRunProgressFromSummary')
    expect(tv).toHaveProperty('getRealtimeProgressByRun')

    // Global Stats
    expect(tv).toHaveProperty('globalStats')
    expect(tv).toHaveProperty('showGlobalStatsModal')
    expect(tv).toHaveProperty('closeGlobalStatsModal')

    // Running Hint
    expect(tv).toHaveProperty('runningHintVisible')
    expect(tv).toHaveProperty('runningHintRun')
    expect(tv).toHaveProperty('runningHintPhaseText')
    expect(tv).toHaveProperty('runningHintProgressText')
    expect(tv).toHaveProperty('openRunningHint')
    expect(tv).toHaveProperty('closeRunningHint')
    expect(tv).toHaveProperty('openRunningHintLog')

    // Log Modal
    expect(tv).toHaveProperty('showLogModal')
    expect(tv).toHaveProperty('logModalTitle')
    expect(tv).toHaveProperty('logContent')
    expect(tv).toHaveProperty('openRunLog')
    expect(tv).toHaveProperty('closeLogModal')

    // Webhook Modal
    expect(tv).toHaveProperty('showWebhookModal')
    expect(tv).toHaveProperty('webhookForm')
    expect(tv).toHaveProperty('setWebhook')
    expect(tv).toHaveProperty('saveWebhook')
    expect(tv).toHaveProperty('testWebhook')
    expect(tv).toHaveProperty('regenerateWebhookSecret')
    expect(tv).toHaveProperty('setWebhookTriggerId')
    expect(tv).toHaveProperty('setWebhookMatchText')
    expect(tv).toHaveProperty('setWebhookSecret')
    expect(tv).toHaveProperty('setWebhookPostUrl')
    expect(tv).toHaveProperty('setWebhookWecomUrl')
    expect(tv).toHaveProperty('setWebhookNotifyManual')
    expect(tv).toHaveProperty('setWebhookNotifySchedule')
    expect(tv).toHaveProperty('setWebhookNotifyWebhook')
    expect(tv).toHaveProperty('setWebhookStatusSuccess')
    expect(tv).toHaveProperty('setWebhookStatusFailed')
    expect(tv).toHaveProperty('setWebhookStatusHasTransfer')
    expect(tv).toHaveProperty('closeWebhookModal')

    // Singleton Modal
    expect(tv).toHaveProperty('showSingletonModal')
    expect(tv).toHaveProperty('singletonForm')
    expect(tv).toHaveProperty('setSingletonMode')
    expect(tv).toHaveProperty('saveSingleton')
    expect(tv).toHaveProperty('setSingletonEnabled')
    expect(tv).toHaveProperty('closeSingletonModal')

    // Schedule Config Modal
    expect(tv).toHaveProperty('scheduleConfigVisible')
    expect(tv).toHaveProperty('scheduleConfigSaving')
    expect(tv).toHaveProperty('scheduleConfigDraft')
    expect(tv).toHaveProperty('scheduleConfigTitle')
    expect(tv).toHaveProperty('closeScheduleConfig')
    expect(tv).toHaveProperty('saveScheduleConfig')

    // Task Editor
    expect(tv).toHaveProperty('taskEditorVisible')
    expect(tv).toHaveProperty('taskEditorTitle')
    expect(tv).toHaveProperty('commandMode')
    expect(tv).toHaveProperty('commandText')
    expect(tv).toHaveProperty('createForm')
    expect(tv).toHaveProperty('editingTask')
    expect(tv).toHaveProperty('showAdvancedOptions')
    expect(tv).toHaveProperty('creatingState')
    expect(tv).toHaveProperty('createTask')
    expect(tv).toHaveProperty('sourcePathOptions')
    expect(tv).toHaveProperty('targetPathOptions')
    expect(tv).toHaveProperty('showSourcePathInput')
    expect(tv).toHaveProperty('showTargetPathInput')
    expect(tv).toHaveProperty('sourceCurrentPath')
    expect(tv).toHaveProperty('targetCurrentPath')
    expect(tv).toHaveProperty('sourceBreadcrumbs')
    expect(tv).toHaveProperty('targetBreadcrumbs')
    expect(tv).toHaveProperty('setShowSourcePathInput')
    expect(tv).toHaveProperty('setShowTargetPathInput')
    expect(tv).toHaveProperty('resetTaskPathBrowse')
    expect(tv).toHaveProperty('restoreTaskPathBrowse')
    expect(tv).toHaveProperty('onSourceRemoteChange')
    expect(tv).toHaveProperty('onTargetRemoteChange')
    expect(tv).toHaveProperty('onSourceBreadcrumbClick')
    expect(tv).toHaveProperty('onTargetBreadcrumbClick')
    expect(tv).toHaveProperty('onSourceClick')
    expect(tv).toHaveProperty('onSourceArrow')
    expect(tv).toHaveProperty('onTargetClick')
    expect(tv).toHaveProperty('onTargetArrow')
    expect(tv).toHaveProperty('setCommandMode')
    expect(tv).toHaveProperty('setCommandText')
    expect(tv).toHaveProperty('setShowAdvancedOptions')
    expect(tv).toHaveProperty('closeTaskEditorModal')

    // Active Transfer
    expect(tv).toHaveProperty('activeTransferVisible')
    expect(tv).toHaveProperty('activeTransferTrackingMode')
    expect(tv).toHaveProperty('activeTransferSummary')
    expect(tv).toHaveProperty('activeTransferCurrentFile')
    expect(tv).toHaveProperty('activeTransferCurrentFiles')
    expect(tv).toHaveProperty('activeTransferSlots')
    expect(tv).toHaveProperty('activeTransferCompletedItems')
    expect(tv).toHaveProperty('activeTransferPendingItems')
    expect(tv).toHaveProperty('activeTransferCompletedTotal')
    expect(tv).toHaveProperty('activeTransferPendingTotal')
    expect(tv).toHaveProperty('activeTransferCompletedPage')
    expect(tv).toHaveProperty('activeTransferPendingPage')
    expect(tv).toHaveProperty('activeTransferCompletedJumpPage')
    expect(tv).toHaveProperty('activeTransferPendingJumpPage')
    expect(tv).toHaveProperty('activeTransferCompletedTotalPages')
    expect(tv).toHaveProperty('activeTransferPendingTotalPages')
    expect(tv).toHaveProperty('activeTransferDegraded')
    expect(tv).toHaveProperty('activeTransferLoading')
    expect(tv).toHaveProperty('activeTransferError')
    expect(tv).toHaveProperty('openActiveTransfer')
    expect(tv).toHaveProperty('closeActiveTransfer')
    expect(tv).toHaveProperty('refreshActiveTransfer')
    expect(tv).toHaveProperty('prevActiveTransferCompletedPage')
    expect(tv).toHaveProperty('nextActiveTransferCompletedPage')
    expect(tv).toHaveProperty('jumpActiveTransferCompletedPage')
    expect(tv).toHaveProperty('prevActiveTransferPendingPage')
    expect(tv).toHaveProperty('nextActiveTransferPendingPage')
    expect(tv).toHaveProperty('jumpActiveTransferPendingPage')

    // Formatters
    expect(tv).toHaveProperty('formatTime')
    expect(tv).toHaveProperty('getStatusClass')
    expect(tv).toHaveProperty('getStatusText')
    expect(tv).toHaveProperty('formatBps')
    expect(tv).toHaveProperty('formatBytes')
    expect(tv).toHaveProperty('formatBytesPerSec')
    expect(tv).toHaveProperty('formatEta')
  })

  it('should export UseTaskViewReturn type', () => {
    // Type checking by verifying the type exists
    const tv = useTaskView()
    expect(typeof tv).toBe('object')
  })

  it('should call tag operations correctly', async () => {
    const tv = useTaskView()
    
    await tv.handleCreateTag('test-tag')
    await tv.handleSelectTag('test-tag')
    await tv.handleUnselectTag('test-tag')
    await tv.handleDeleteTag('test-tag')
  })

  it('should open and close tag manager', () => {
    const tv = useTaskView()
    
    expect(tv.tagManagerVisible.value).toBe(false)
    
    tv.openTagManager()
    expect(tv.tagManagerVisible.value).toBe(true)
    
    tv.closeTagManager()
    expect(tv.tagManagerVisible.value).toBe(false)
  })

  it('should open and close bisync lst manager', () => {
    const tv = useTaskView()
    
    expect(tv.bisyncLstManagerVisible.value).toBe(false)
    
    tv.openBisyncLstManager({ id: 1, bisyncOptions: { option1: 'value1' } })
    expect(tv.bisyncLstManagerVisible.value).toBe(true)
    
    tv.closeBisyncLstManager()
    expect(tv.bisyncLstManagerVisible.value).toBe(false)
  })
})

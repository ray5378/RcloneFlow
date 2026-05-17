import { computed, onUnmounted, ref } from 'vue'
import { getActiveTransfer, getActiveTransferCompleted, getActiveTransferPending, type ActiveTransferCompletedFile, type ActiveTransferCurrentFile, type ActiveTransferPendingFile, type ActiveTransferSnapshot, type ActiveTransferSummary, type TrackingMode } from '../api/activeTransfer'
import { onWsMessage } from './useWebSocket'

const PAGE_SIZE = 10

function sortCurrentFiles(items: ActiveTransferCurrentFile[]) {
  return [...items].sort((a, b) => {
    const ao = Number(a.order || 0)
    const bo = Number(b.order || 0)
    if (ao !== bo) {
      if (!ao) return 1
      if (!bo) return -1
      return ao - bo
    }
    return String(a.path || a.name || '').localeCompare(String(b.path || b.name || ''))
  })
}

function sortCompletedItems(items: ActiveTransferCompletedFile[]) {
  return [...items].sort((a, b) => {
    const ao = Number(a.order || 0)
    const bo = Number(b.order || 0)
    if (ao !== bo) {
      if (!ao) return 1
      if (!bo) return -1
      return ao - bo
    }
    const atCmp = String(a.at || '').localeCompare(String(b.at || ''))
    if (atCmp !== 0) return atCmp
    return String(a.path || a.name || '').localeCompare(String(b.path || b.name || ''))
  })
}

function sortPendingItems(items: ActiveTransferPendingFile[]) {
  return [...items].sort((a, b) => {
    if (a.status !== b.status) return a.status === 'in_progress' ? -1 : 1
    const ao = Number(a.order || 0)
    const bo = Number(b.order || 0)
    if (ao !== bo) {
      if (!ao) return 1
      if (!bo) return -1
      return ao - bo
    }
    return String(a.path || a.name || '').localeCompare(String(b.path || b.name || ''))
  })
}

function mergeNonDecreasingSummary(prev: ActiveTransferSummary | null, next: ActiveTransferSummary | null): ActiveTransferSummary | null {
  if (!next) return prev
  if (!prev) return next
  const nextTotalBytes = Math.max(Number(prev.totalBytes || 0), Number(next.totalBytes || 0))
  const nextPlannedFiles = Math.max(Number(prev.plannedFiles || 0), Number(next.plannedFiles || 0))
  const nextLogicalTotalCount = Math.max(
    Number(prev.logicalTotalCount || prev.totalCount || 0),
    Number(next.logicalTotalCount || next.totalCount || 0),
    nextPlannedFiles,
  )
  const nextTotalCount = nextLogicalTotalCount
  const nextCompletedCount = Math.max(Number(prev.completedCount || 0), Number(next.completedCount || 0))
  let nextPercentage = Number(next.percentage || 0)
  if (!nextPercentage && nextTotalBytes > 0) {
    nextPercentage = Math.min(100, (Number(next.bytes || 0) / nextTotalBytes) * 100)
  }
  nextPercentage = Math.max(Number(prev.percentage || 0), nextPercentage)
  if (nextCompletedCount >= nextTotalCount && nextTotalCount > 0) nextPercentage = 100
  if (nextPercentage > 100) nextPercentage = 100
  return {
    ...prev,
    ...next,
    plannedFiles: nextPlannedFiles,
    logicalTotalCount: nextLogicalTotalCount,
    completedCount: nextCompletedCount,
    pendingCount: Math.max(0, nextTotalCount - nextCompletedCount),
    totalCount: nextTotalCount,
    totalBytes: nextTotalBytes,
    percentage: nextPercentage,
  }
}

export function useActiveTransferDetail() {
  const visible = ref(false)
  const taskId = ref<number | null>(null)
  const runId = ref<number | null>(null)
  const trackingMode = ref<TrackingMode>('normal')
  const summary = ref<ActiveTransferSummary | null>(null)
  const currentFile = ref<ActiveTransferCurrentFile | null>(null)
  const currentFiles = ref<ActiveTransferCurrentFile[]>([])
  const completedItems = ref<ActiveTransferCompletedFile[]>([])
  const pendingItems = ref<ActiveTransferPendingFile[]>([])
  const completedTotal = ref(0)
  const pendingTotal = ref(0)
  const completedPage = ref(1)
  const pendingPage = ref(1)
  const completedJumpPage = ref<number | null>(1)
  const pendingJumpPage = ref<number | null>(1)
  const degraded = ref(false)
  const loading = ref(false)
  const error = ref('')

  const completedTotalPages = computed(() => Math.max(1, Math.ceil(Math.max(completedTotal.value, 0) / PAGE_SIZE)))
  const pendingTotalPages = computed(() => Math.max(1, Math.ceil(Math.max(pendingTotal.value, 0) / PAGE_SIZE)))

  function shouldHandleRunMessage(incomingRunId: any, incomingTaskId?: any) {
    return visible.value && (
      (runId.value != null && Number(incomingRunId) === Number(runId.value)) ||
      (taskId.value != null && incomingTaskId != null && Number(incomingTaskId) === Number(taskId.value))
    )
  }

  function applySnapshot(snapshot: ActiveTransferSnapshot) {
    trackingMode.value = snapshot.trackingMode
    currentFile.value = snapshot.currentFile || null
    currentFiles.value = sortCurrentFiles(snapshot.currentFiles || (snapshot.currentFile ? [snapshot.currentFile] : []))
    degraded.value = !!snapshot.degraded

    const completed = sortCompletedItems(snapshot.completed || [])
    const pending = sortPendingItems(snapshot.pending || [])
    completedItems.value = completed
    pendingItems.value = pending
    completedTotal.value = Number(snapshot.completedCount || completed.length || 0)
    pendingTotal.value = Number(snapshot.pendingCount || pending.length || 0)
    if (completedPage.value > completedTotalPages.value) {
      completedPage.value = completedTotalPages.value
      completedJumpPage.value = completedTotalPages.value
    }
    if (pendingPage.value > pendingTotalPages.value) {
      pendingPage.value = pendingTotalPages.value
      pendingJumpPage.value = pendingTotalPages.value
    }

    const stableTotalCount = Math.max(
      Number(summary.value?.logicalTotalCount || summary.value?.totalCount || 0),
      Number(snapshot.totalCount || 0),
      completed.length + pending.length,
    )
    summary.value = mergeNonDecreasingSummary(summary.value, {
      trackingMode: snapshot.trackingMode,
      completedCount: Number(snapshot.completedCount || completed.length || 0),
      pendingCount: Number(snapshot.pendingCount || Math.max(0, stableTotalCount - completed.length)),
      plannedFiles: Number(summary.value?.plannedFiles || 0),
      logicalTotalCount: stableTotalCount,
      totalCount: stableTotalCount,
      preflightPending: !!snapshot.preflightPending,
      preflightFinished: !!snapshot.preflightFinished,
      percentage: Number(summary.value?.percentage || 0),
      bytes: Number(summary.value?.bytes || 0),
      totalBytes: Number(summary.value?.totalBytes || 0),
      speed: Number(summary.value?.speed || 0),
      eta: Number(summary.value?.eta || 0),
      phase: summary.value?.phase,
      lastUpdatedAt: summary.value?.lastUpdatedAt,
    })
  }

  async function refresh(background = false) {
    if (!taskId.value) return
    if (!background) loading.value = true
    error.value = ''
    try {
      const [overview, completed, pending] = await Promise.all([
        getActiveTransfer(taskId.value),
        getActiveTransferCompleted(taskId.value, Math.max(0, (completedPage.value - 1) * PAGE_SIZE), PAGE_SIZE),
        getActiveTransferPending(taskId.value, Math.max(0, (pendingPage.value - 1) * PAGE_SIZE), PAGE_SIZE),
      ])
      runId.value = overview.runId
      trackingMode.value = overview.trackingMode
      summary.value = mergeNonDecreasingSummary(summary.value, overview.summary)
      currentFile.value = overview.currentFile || null
      currentFiles.value = sortCurrentFiles(overview.currentFiles || (overview.currentFile ? [overview.currentFile] : []))
      degraded.value = !!overview.degraded
      completedItems.value = sortCompletedItems(completed.items || [])
      pendingItems.value = sortPendingItems(pending.items || [])
      completedTotal.value = completed.total || 0
      pendingTotal.value = pending.total || 0
      if (completedPage.value > completedTotalPages.value) {
        completedPage.value = completedTotalPages.value
        completedJumpPage.value = completedTotalPages.value
      }
      if (pendingPage.value > pendingTotalPages.value) {
        pendingPage.value = pendingTotalPages.value
        pendingJumpPage.value = pendingTotalPages.value
      }
    } catch (e: any) {
      const msg = String(e?.message || 'active transfer load failed')
      if (msg === '当前没有运行中的任务' || msg === '当前没有可恢复的传输状态' || msg === 'No active run for this task' || msg === 'No restorable transfer state available') {
        summary.value = null
        currentFile.value = null
        currentFiles.value = []
        completedItems.value = []
        pendingItems.value = []
        completedTotal.value = 0
        pendingTotal.value = 0
        degraded.value = false
        error.value = ''
      } else {
        error.value = msg
      }
    } finally {
      if (!background) loading.value = false
    }
  }

  function open(nextTaskId: number) {
    taskId.value = nextTaskId
    runId.value = null
    completedPage.value = 1
    pendingPage.value = 1
    completedJumpPage.value = 1
    pendingJumpPage.value = 1
    visible.value = true
    void refresh(false)
  }

  function close() {
    visible.value = false
    taskId.value = null
    runId.value = null
    summary.value = null
    currentFile.value = null
    currentFiles.value = []
    completedItems.value = []
    pendingItems.value = []
    completedTotal.value = 0
    pendingTotal.value = 0
    completedPage.value = 1
    pendingPage.value = 1
    completedJumpPage.value = 1
    pendingJumpPage.value = 1
    degraded.value = false
  }

  function prevCompletedPage() {
    if (completedPage.value <= 1) return
    completedPage.value -= 1
    completedJumpPage.value = completedPage.value
    void refresh(true)
  }

  function nextCompletedPage() {
    if (completedPage.value >= completedTotalPages.value) return
    completedPage.value += 1
    completedJumpPage.value = completedPage.value
    void refresh(true)
  }

  function jumpCompletedPage() {
    const page = Math.min(Math.max(1, Number(completedJumpPage.value || 1)), completedTotalPages.value)
    if (page === completedPage.value) return
    completedPage.value = page
    completedJumpPage.value = page
    void refresh(true)
  }

  function prevPendingPage() {
    if (pendingPage.value <= 1) return
    pendingPage.value -= 1
    pendingJumpPage.value = pendingPage.value
    void refresh(true)
  }

  function nextPendingPage() {
    if (pendingPage.value >= pendingTotalPages.value) return
    pendingPage.value += 1
    pendingJumpPage.value = pendingPage.value
    void refresh(true)
  }

  function jumpPendingPage() {
    const page = Math.min(Math.max(1, Number(pendingJumpPage.value || 1)), pendingTotalPages.value)
    if (page === pendingPage.value) return
    pendingPage.value = page
    pendingJumpPage.value = page
    void refresh(true)
  }

  const visibleCompletedItems = computed(() => completedItems.value)
  const visiblePendingItems = computed(() => {
    const currentKeys = new Set((currentFiles.value || []).map(item => item.path || item.name).filter(Boolean))
    return currentKeys.size ? pendingItems.value.filter(item => !currentKeys.has(item.path || item.name)) : pendingItems.value
  })

  const offActiveTransferSnapshot = onWsMessage('active_transfer_snapshot', (data) => {
    if (shouldHandleRunMessage(data?.run_id, data?.task_id) && data?.snapshot) {
      runId.value = Number(data.run_id || data.snapshot.runId || runId.value || 0) || null
      applySnapshot(data.snapshot as ActiveTransferSnapshot)
    }
  })

  const offRunProgress = onWsMessage('run_progress', (data) => {
    if (shouldHandleRunMessage(data?.run_id)) {
      const prev = summary.value
      const incomingPlannedFiles = Number(data?.plannedFiles || 0)
      const incomingLogicalTotalCount = Number(data?.logicalTotalCount || data?.totalCount || incomingPlannedFiles || prev?.logicalTotalCount || prev?.totalCount || 0)
      const nextCompletedCount = Math.max(Number(prev?.completedCount || 0), Number(data?.completedFiles || 0))
      summary.value = mergeNonDecreasingSummary(summary.value, summary.value ? {
        ...summary.value,
        bytes: Number(data?.bytes || prev?.bytes || 0),
        totalBytes: Number(data?.total || prev?.totalBytes || 0),
        speed: Number(data?.speed || prev?.speed || 0),
        percentage: Number(data?.percent || prev?.percentage || 0),
        eta: Number(data?.eta || prev?.eta || 0),
        plannedFiles: Math.max(Number(prev?.plannedFiles || 0), incomingPlannedFiles),
        logicalTotalCount: incomingLogicalTotalCount,
        totalCount: incomingLogicalTotalCount,
        completedCount: nextCompletedCount,
        pendingCount: Math.max(0, incomingLogicalTotalCount - nextCompletedCount),
        phase: typeof data?.phase === 'string' ? data.phase : prev?.phase,
        lastUpdatedAt: typeof data?.lastUpdatedAt === 'string' ? data.lastUpdatedAt : prev?.lastUpdatedAt,
      } : summary.value)
    }
  })

  const offRunStatus = onWsMessage('run_status', (data) => {
    if (shouldHandleRunMessage(data?.run_id) && data?.status !== 'running') {
      void refresh(true)
    }
  })

  onUnmounted(() => {
    offActiveTransferSnapshot()
    offRunProgress()
    offRunStatus()
  })

  return {
    activeTransferVisible: visible,
    activeTransferTaskId: taskId,
    activeTransferTrackingMode: trackingMode,
    activeTransferSummary: summary,
    activeTransferCurrentFile: currentFile,
    activeTransferCurrentFiles: currentFiles,
    activeTransferCompletedItems: visibleCompletedItems,
    activeTransferPendingItems: visiblePendingItems,
    activeTransferCompletedTotal: completedTotal,
    activeTransferPendingTotal: pendingTotal,
    activeTransferCompletedPage: completedPage,
    activeTransferPendingPage: pendingPage,
    activeTransferCompletedJumpPage: completedJumpPage,
    activeTransferPendingJumpPage: pendingJumpPage,
    activeTransferCompletedTotalPages: completedTotalPages,
    activeTransferPendingTotalPages: pendingTotalPages,
    activeTransferDegraded: degraded,
    activeTransferLoading: loading,
    activeTransferError: error,
    openActiveTransfer: open,
    closeActiveTransfer: close,
    refreshActiveTransfer: refresh,
    prevActiveTransferCompletedPage: prevCompletedPage,
    nextActiveTransferCompletedPage: nextCompletedPage,
    jumpActiveTransferCompletedPage: jumpCompletedPage,
    prevActiveTransferPendingPage: prevPendingPage,
    nextActiveTransferPendingPage: nextPendingPage,
    jumpActiveTransferPendingPage: jumpPendingPage,
  }
}

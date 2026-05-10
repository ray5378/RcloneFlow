import { computed, onUnmounted, ref } from 'vue'
import { getActiveTransfer, getActiveTransferCompleted, getActiveTransferPending, type ActiveTransferCompletedFile, type ActiveTransferCurrentFile, type ActiveTransferPendingFile, type ActiveTransferSnapshot, type ActiveTransferSummary, type TrackingMode } from '../api/activeTransfer'
import { onWsMessage } from './useWebSocket'

const PAGE_SIZE = 100

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
    if (ao !== bo) return bo - ao
    const atCmp = String(b.at || '').localeCompare(String(a.at || ''))
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
  const nextTotalCount = Math.max(Number(prev.totalCount || 0), Number(next.totalCount || 0))
  const nextCompletedCount = Math.max(Number(prev.completedCount || 0), Number(next.completedCount || 0))
  let nextPercentage = Number(next.percentage || 0)
  if (nextTotalBytes > 0) {
    nextPercentage = Math.max(nextPercentage, Math.min(100, (Number(next.bytes || 0) / nextTotalBytes) * 100))
  }
  nextPercentage = Math.max(Number(prev.percentage || 0), nextPercentage)
  if (nextCompletedCount >= nextTotalCount && nextTotalCount > 0) nextPercentage = 100
  if (nextPercentage > 100) nextPercentage = 100
  return {
    ...prev,
    ...next,
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
  const completedOffset = ref(0)
  const pendingOffset = ref(0)
  const degraded = ref(false)
  const loading = ref(false)
  const error = ref('')

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
    completedItems.value = sortCompletedItems(snapshot.completed || [])
    pendingItems.value = sortPendingItems(snapshot.pending || [])
    completedTotal.value = (snapshot.completed || []).length
    pendingTotal.value = (snapshot.pending || []).length
    summary.value = mergeNonDecreasingSummary(summary.value, {
      trackingMode: snapshot.trackingMode,
      completedCount: (snapshot.completed || []).length,
      pendingCount: (snapshot.pending || []).length,
      totalCount: snapshot.totalCount || ((snapshot.completed || []).length + (snapshot.pending || []).length),
      preflightPending: !!snapshot.preflightPending,
      preflightFinished: !!snapshot.preflightFinished,
      percentage: (snapshot.totalCount || 0) > 0 ? (((snapshot.completed || []).length / (snapshot.totalCount || 1)) * 100) : 0,
      bytes: summary.value?.bytes || 0,
      totalBytes: summary.value?.totalBytes || 0,
      speed: summary.value?.speed || 0,
      eta: summary.value?.eta,
    })
  }

  async function refresh(background = false) {
    if (!taskId.value) return
    if (!background) {
      loading.value = true
    }
    error.value = ''
    try {
      const [overview, completed, pending] = await Promise.all([
        getActiveTransfer(taskId.value),
        getActiveTransferCompleted(taskId.value, 0, Math.max(completedOffset.value + PAGE_SIZE, PAGE_SIZE)),
        getActiveTransferPending(taskId.value, 0, Math.max(pendingOffset.value + PAGE_SIZE, PAGE_SIZE)),
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
      if (!background) {
        loading.value = false
      }
    }
  }

  function open(nextTaskId: number) {
    taskId.value = nextTaskId
    runId.value = null
    completedOffset.value = 0
    pendingOffset.value = 0
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
    completedOffset.value = 0
    pendingOffset.value = 0
    degraded.value = false
  }

  async function loadMoreCompleted() {
    completedOffset.value = Math.min(completedOffset.value + PAGE_SIZE, Math.max(completedTotal.value - PAGE_SIZE, 0))
  }

  async function loadMorePending() {
    pendingOffset.value = Math.min(pendingOffset.value + PAGE_SIZE, Math.max(pendingTotal.value - PAGE_SIZE, 0))
  }

  const visibleCompletedItems = computed(() => completedItems.value.slice(0, completedOffset.value + PAGE_SIZE))

  const visiblePendingItems = computed(() => {
    const currentKeys = new Set((currentFiles.value || []).map(item => item.path || item.name).filter(Boolean))
    const items = currentKeys.size ? pendingItems.value.filter(item => !currentKeys.has(item.path || item.name)) : pendingItems.value
    return items.slice(0, pendingOffset.value + PAGE_SIZE)
  })

  const offActiveTransferSnapshot = onWsMessage('active_transfer_snapshot', (data) => {
    if (shouldHandleRunMessage(data?.run_id, data?.task_id) && data?.snapshot) {
      runId.value = Number(data.run_id || data.snapshot.runId || runId.value || 0) || null
      applySnapshot(data.snapshot as ActiveTransferSnapshot)
    }
  })

  const offRunProgress = onWsMessage('run_progress', (data) => {
    if (shouldHandleRunMessage(data?.run_id)) {
      summary.value = mergeNonDecreasingSummary(summary.value, summary.value ? {
        ...summary.value,
        bytes: Number(data?.bytes || summary.value.bytes || 0),
        totalBytes: Number(data?.total || summary.value.totalBytes || 0),
        speed: Number(data?.speed || summary.value.speed || 0),
        percentage: Number(data?.percent || summary.value.percentage || 0),
        eta: Number(data?.eta || summary.value.eta || 0),
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
    activeTransferDegraded: degraded,
    activeTransferLoading: loading,
    activeTransferError: error,
    openActiveTransfer: open,
    closeActiveTransfer: close,
    refreshActiveTransfer: refresh,
    loadMoreActiveTransferCompleted: loadMoreCompleted,
    loadMoreActiveTransferPending: loadMorePending,
  }
}

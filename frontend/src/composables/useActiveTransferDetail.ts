import { computed, onUnmounted, ref } from 'vue'
import { getActiveTransfer, getActiveTransferCompleted, getActiveTransferPending, type ActiveTransferCompletedFile, type ActiveTransferCurrentFile, type ActiveTransferPendingFile, type ActiveTransferSnapshot, type ActiveTransferSummary, type TrackingMode } from '../api/activeTransfer'
import { onWsMessage } from './useWebSocket'

const PAGE_SIZE = 100

export function useActiveTransferDetail() {
  const visible = ref(false)
  const taskId = ref<number | null>(null)
  const runId = ref<number | null>(null)
  const trackingMode = ref<TrackingMode>('normal')
  const summary = ref<ActiveTransferSummary | null>(null)
  const currentFile = ref<ActiveTransferCurrentFile | null>(null)
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
    degraded.value = !!snapshot.degraded
    completedItems.value = snapshot.completed || []
    pendingItems.value = snapshot.pending || []
    completedTotal.value = (snapshot.completed || []).length
    pendingTotal.value = (snapshot.pending || []).length
    summary.value = {
      trackingMode: snapshot.trackingMode,
      completedCount: (snapshot.completed || []).length,
      pendingCount: (snapshot.pending || []).length,
      totalCount: snapshot.totalCount || ((snapshot.completed || []).length + (snapshot.pending || []).length),
      percentage: (snapshot.totalCount || 0) > 0 ? (((snapshot.completed || []).length / (snapshot.totalCount || 1)) * 100) : 0,
      bytes: summary.value?.bytes || 0,
      totalBytes: summary.value?.totalBytes || 0,
      speed: summary.value?.speed || 0,
      eta: summary.value?.eta,
    }
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
      summary.value = overview.summary
      currentFile.value = overview.currentFile || null
      degraded.value = !!overview.degraded
      completedItems.value = completed.items || []
      pendingItems.value = pending.items || []
      completedTotal.value = completed.total || 0
      pendingTotal.value = pending.total || 0
    } catch (e: any) {
      const msg = String(e?.message || 'active transfer load failed')
      if (msg === '当前没有运行中的任务' || msg === '当前没有可恢复的传输状态' || msg === 'No active run for this task' || msg === 'No restorable transfer state available') {
        summary.value = null
        currentFile.value = null
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
    const cur = currentFile.value?.path || currentFile.value?.name
    const items = cur ? pendingItems.value.filter(item => (item.path || item.name) !== cur) : pendingItems.value
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
      summary.value = summary.value ? {
        ...summary.value,
        bytes: Number(data?.bytes || summary.value.bytes || 0),
        totalBytes: Number(data?.total || summary.value.totalBytes || 0),
        speed: Number(data?.speed || summary.value.speed || 0),
        percentage: Number(data?.percent || summary.value.percentage || 0),
        eta: Number(data?.eta || summary.value.eta || 0),
      } : summary.value
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

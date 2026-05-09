import { computed, onUnmounted, ref } from 'vue'
import { getActiveTransfer, getActiveTransferCompleted, getActiveTransferPending, type ActiveTransferCompletedFile, type ActiveTransferCurrentFile, type ActiveTransferPendingFile, type ActiveTransferSummary, type TrackingMode } from '../api/activeTransfer'

const PAGE_SIZE = 100

export function useActiveTransferDetail() {
  const visible = ref(false)
  const taskId = ref<number | null>(null)
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
  let timer: number | null = null

  function stopTimer() {
    if (timer != null) {
      clearInterval(timer)
      timer = null
    }
  }

  async function refresh() {
    if (!taskId.value) return
    loading.value = true
    error.value = ''
    try {
      const [overview, completed, pending] = await Promise.all([
        getActiveTransfer(taskId.value),
        getActiveTransferCompleted(taskId.value, completedOffset.value, PAGE_SIZE),
        getActiveTransferPending(taskId.value, pendingOffset.value, PAGE_SIZE),
      ])
      trackingMode.value = overview.trackingMode
      summary.value = overview.summary
      currentFile.value = overview.currentFile || null
      degraded.value = !!overview.degraded
      completedTotal.value = completed.total || 0
      pendingTotal.value = pending.total || 0
      completedItems.value = completed.items || []
      pendingItems.value = pending.items || []
    } catch (e: any) {
      error.value = e?.message || 'active transfer load failed'
    } finally {
      loading.value = false
    }
  }

  function open(nextTaskId: number) {
    taskId.value = nextTaskId
    completedOffset.value = 0
    pendingOffset.value = 0
    visible.value = true
    void refresh()
    stopTimer()
    timer = window.setInterval(() => void refresh(), 3000)
  }

  function close() {
    visible.value = false
    taskId.value = null
    summary.value = null
    currentFile.value = null
    completedItems.value = []
    pendingItems.value = []
    completedTotal.value = 0
    pendingTotal.value = 0
    completedOffset.value = 0
    pendingOffset.value = 0
    degraded.value = false
    stopTimer()
  }

  async function loadMoreCompleted() {
    if (!taskId.value) return
    const nextOffset = completedOffset.value + PAGE_SIZE
    if (nextOffset >= completedTotal.value) return
    const res = await getActiveTransferCompleted(taskId.value, nextOffset, PAGE_SIZE)
    completedOffset.value = nextOffset
    completedTotal.value = res.total || completedTotal.value
    completedItems.value = [...completedItems.value, ...(res.items || [])]
  }

  async function loadMorePending() {
    if (!taskId.value) return
    const nextOffset = pendingOffset.value + PAGE_SIZE
    if (nextOffset >= pendingTotal.value) return
    const res = await getActiveTransferPending(taskId.value, nextOffset, PAGE_SIZE)
    pendingOffset.value = nextOffset
    pendingTotal.value = res.total || pendingTotal.value
    pendingItems.value = [...pendingItems.value, ...(res.items || [])]
  }

  const visiblePendingItems = computed(() => {
    const cur = currentFile.value?.path || currentFile.value?.name
    if (!cur) return pendingItems.value
    return pendingItems.value.filter(item => (item.path || item.name) !== cur)
  })

  onUnmounted(() => stopTimer())

  return {
    activeTransferVisible: visible,
    activeTransferTaskId: taskId,
    activeTransferTrackingMode: trackingMode,
    activeTransferSummary: summary,
    activeTransferCurrentFile: currentFile,
    activeTransferCompletedItems: completedItems,
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

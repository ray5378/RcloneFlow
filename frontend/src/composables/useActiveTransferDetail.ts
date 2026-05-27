import { computed, onUnmounted, ref } from 'vue'
import { getActiveTransfer, getActiveTransferCompleted, getActiveTransferPending, type ActiveTransferCompletedFile, type ActiveTransferCurrentFile, type ActiveTransferPendingFile, type ActiveTransferSnapshot, type ActiveTransferSummary, type TrackingMode } from '../api/activeTransfer'
import { onWsMessage } from './useWebSocket'

const PAGE_SIZE = 10
const MIN_TRANSFER_SLOTS = 1
const TIME_CONSTANTS = {
  FINISH_WINDOW_MS: 15000,
}

interface SortableItem {
  order?: number | string
  path?: string
  name?: string
  at?: string
  status?: string
}

function getOrderValue(item: SortableItem): number {
  return Number(item.order || 0)
}

function sortByOrderAndName<T extends SortableItem>(items: T[], customSort?: (a: T, b: T) => number): T[] {
  return [...items].sort((a, b) => {
    const ao = getOrderValue(a)
    const bo = getOrderValue(b)
    if (ao !== bo) {
      if (!ao) return 1
      if (!bo) return -1
      return ao - bo
    }
    if (customSort) {
      const customResult = customSort(a, b)
      if (customResult !== 0) return customResult
    }
    return String(a.path || a.name || '').localeCompare(String(b.path || b.name || ''))
  })
}

function sortCurrentFiles(items: ActiveTransferCurrentFile[]): ActiveTransferCurrentFile[] {
  return sortByOrderAndName(items)
}

function sortCompletedItems(items: ActiveTransferCompletedFile[]): ActiveTransferCompletedFile[] {
  return sortByOrderAndName(items, (a, b) => {
    return String(a.at || '').localeCompare(String(b.at || ''))
  })
}

function sortPendingItems(items: ActiveTransferPendingFile[]): ActiveTransferPendingFile[] {
  return sortByOrderAndName(items, (a, b) => {
    if (a.status !== b.status) return a.status === 'in_progress' ? -1 : 1
    return 0
  })
}

function completedItemKey(item: ActiveTransferCompletedFile): string {
  return String(item.path || item.name || '')
}

function isCompletedItemNewerThanPage(item: ActiveTransferCompletedFile, pageItems: ActiveTransferCompletedFile[]): boolean {
  if (!pageItems.length) return true
  const itemOrder = getOrderValue(item)
  const maxOrder = Math.max(...pageItems.map(existing => getOrderValue(existing)))
  if (itemOrder > 0 || maxOrder > 0) return itemOrder > maxOrder
  const itemAt = String(item.at || '')
  const maxAt = pageItems.reduce((max, existing) => {
    const at = String(existing.at || '')
    return at > max ? at : max
  }, '')
  return itemAt > maxAt
}

function appendNewCompletedItemsForLastPage(current: ActiveTransferCompletedFile[], incoming: ActiveTransferCompletedFile[]): ActiveTransferCompletedFile[] {
  const remainingSlots = Math.max(0, PAGE_SIZE - current.length)
  if (remainingSlots <= 0) return current
  const existingKeys = new Set(current.map(completedItemKey).filter(Boolean))
  const additions = sortCompletedItems(incoming.filter(item => {
    const key = completedItemKey(item)
    return key && !existingKeys.has(key) && isCompletedItemNewerThanPage(item, current)
  })).slice(0, remainingSlots)
  return additions.length ? sortCompletedItems([...current, ...additions]) : current
}

function freezeCompletedProgress(p: ActiveTransferSummary | null): ActiveTransferSummary | null {
  if (!p) return null
  
  const frozen = { ...p }
  const percentage = Number(frozen.percentage || 0)
  
  if (percentage >= 99.999) {
    frozen.percentage = 100
    const totalBytes = Number(frozen.totalBytes || 0)
    if (totalBytes > 0) frozen.bytes = totalBytes
    const totalCount = Number(frozen.totalCount || 0)
    if (totalCount > 0) frozen.completedCount = totalCount
    frozen.speed = 0
    frozen.eta = 0
    frozen.phase = 'completed'
  }
  
  return frozen
}

function mergeNonDecreasingSummary(prev: ActiveTransferSummary | null, next: ActiveTransferSummary | null): ActiveTransferSummary | null {
  if (!next) return prev
  if (!prev) {
    const frozen = freezeCompletedProgress(next)
    return frozen || next
  }

  const nextTotalBytes = Math.max(Number(prev.totalBytes || 0), Number(next.totalBytes || 0))
  const nextPlannedFiles = Math.max(Number(prev.plannedFiles || 0), Number(next.plannedFiles || 0))
  const nextLogicalTotalCount = Math.max(
    Number(prev.logicalTotalCount || prev.totalCount || 0),
    Number(next.logicalTotalCount || next.totalCount || 0),
    nextPlannedFiles,
  )
  const nextTotalCount = nextLogicalTotalCount
  const nextCompletedCount = Math.max(Number(prev.completedCount || 0), Number(next.completedCount || 0))
  const nextBytes = Math.max(Number(prev.bytes || 0), Number(next.bytes || 0))
  
  let nextPercentage = Number(next.percentage || 0)
  if (!nextPercentage && nextTotalBytes > 0) {
    nextPercentage = Math.min(100, (nextBytes / nextTotalBytes) * 100)
  }
  nextPercentage = Math.max(Number(prev.percentage || 0), nextPercentage)
  if (nextCompletedCount >= nextTotalCount && nextTotalCount > 0) nextPercentage = 100
  if (nextPercentage > 100) nextPercentage = 100

  let nextSpeed = Number(next.speed || prev.speed || 0)
  const nextEta = Number(next.eta || prev.eta || 0)

  const merged = {
    ...prev,
    ...next,
    plannedFiles: nextPlannedFiles,
    logicalTotalCount: nextLogicalTotalCount,
    completedCount: nextCompletedCount,
    pendingCount: Math.max(0, nextTotalCount - nextCompletedCount),
    totalCount: nextTotalCount,
    totalBytes: nextTotalBytes,
    bytes: nextBytes,
    percentage: nextPercentage,
    speed: nextSpeed,
    eta: nextEta,
  }

  const final = freezeCompletedProgress(merged)
  return final || merged
}

const NO_ACTIVE_TRANSFER_ERRORS = [
  '当前没有运行中的任务',
  '当前没有可恢复的传输状态',
  'No active run for this task',
  'No restorable transfer state available',
]

function isNoActiveTransferError(message: string): boolean {
  return NO_ACTIVE_TRANSFER_ERRORS.includes(message)
}

function createEmptySnapshot(): Partial<ActiveTransferSummary> {
  return {
    completedCount: 0,
    pendingCount: 0,
    plannedFiles: 0,
    logicalTotalCount: 0,
    totalCount: 0,
    preflightPending: false,
    preflightFinished: false,
    percentage: 0,
    bytes: 0,
    totalBytes: 0,
    speed: 0,
    transferSlots: 1,
    eta: 0,
  }
}

export function useActiveTransferDetail() {
  const visible = ref(false)
  const taskId = ref<number | null>(null)
  const runId = ref<number | null>(null)
  const trackingMode = ref<TrackingMode>('normal')
  const summary = ref<ActiveTransferSummary | null>(null)
  const completedFreeze = ref<ActiveTransferSummary & { __frozenAt: number } | null>(null)
  const currentFile = ref<ActiveTransferCurrentFile | null>(null)
  const currentFiles = ref<ActiveTransferCurrentFile[]>([])
  const transferSlots = ref(MIN_TRANSFER_SLOTS)
  const completedItems = ref<ActiveTransferCompletedFile[]>([])
  const pendingItems = ref<ActiveTransferPendingFile[]>([])
  const rawPendingItems = ref<ActiveTransferPendingFile[]>([])
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
  const pendingTotalPages = computed(() => Math.max(1, Math.ceil(pendingTotal.value / PAGE_SIZE)))

  const stableSummary = computed(() => {
    if (completedFreeze.value) {
      return completedFreeze.value
    }
    return summary.value
  })

  function shouldHandleRunMessage(incomingRunId: any, incomingTaskId?: any): boolean {
    return visible.value && (
      (runId.value != null && Number(incomingRunId) === Number(runId.value)) ||
      (taskId.value != null && incomingTaskId != null && Number(incomingTaskId) === Number(taskId.value))
    )
  }

  function resetTransferState(): void {
    summary.value = null
    completedFreeze.value = null
    currentFile.value = null
    currentFiles.value = []
    transferSlots.value = MIN_TRANSFER_SLOTS
    completedItems.value = []
    pendingItems.value = []
    rawPendingItems.value = []
    completedTotal.value = 0
    pendingTotal.value = 0
    degraded.value = false
    error.value = ''
  }

  function applySnapshot(snapshot: ActiveTransferSnapshot): void {
    trackingMode.value = snapshot.trackingMode
    currentFile.value = snapshot.currentFile || null
    currentFiles.value = sortCurrentFiles(snapshot.currentFiles || (snapshot.currentFile ? [snapshot.currentFile] : []))
    transferSlots.value = Math.max(MIN_TRANSFER_SLOTS, Number(snapshot.transferSlots || transferSlots.value || MIN_TRANSFER_SLOTS))
    degraded.value = !!snapshot.degraded

    const completed = sortCompletedItems(snapshot.completed || [])
    const pending = sortPendingItems(snapshot.pending || [])
    rawPendingItems.value = pending
    
    const wasBrowsingCompletedLastPage = completedPage.value > 1 && completedPage.value === completedTotalPages.value
    const nextCompletedTotal = Number(snapshot.completedCount || completed.length || 0)
    const nextPendingTotal = Number(snapshot.pendingCount || pending.length || 0)
    
    completedTotal.value = nextCompletedTotal
    pendingTotal.value = nextPendingTotal
    
    clampPageValues()

    if (completedPage.value <= 1) {
      completedItems.value = completed.slice(0, PAGE_SIZE)
    } else if (wasBrowsingCompletedLastPage) {
      completedItems.value = appendNewCompletedItemsForLastPage(completedItems.value, completed)
    }
    
    const currentKeys = new Set((currentFiles.value || []).map(item => item.path || item.name).filter(Boolean))
    const filteredPending = currentKeys.size ? pending.filter(item => !currentKeys.has(item.path || item.name)) : pending
    
    if (pendingPage.value <= 1) {
      pendingItems.value = filteredPending.slice(0, PAGE_SIZE)
    }

    const stableTotalCount = Math.max(
      Number(summary.value?.logicalTotalCount || summary.value?.totalCount || 0),
      Number(snapshot.totalCount || 0),
      completed.length + pending.length,
    )
    
    const merged = mergeNonDecreasingSummary(summary.value, {
      ...summary.value,
      ...createEmptySnapshot(),
      trackingMode: snapshot.trackingMode,
      completedCount: Number(snapshot.completedCount || completed.length || 0),
      pendingCount: Number(snapshot.pendingCount || Math.max(0, stableTotalCount - completed.length)),
      logicalTotalCount: stableTotalCount,
      totalCount: stableTotalCount,
      preflightPending: !!snapshot.preflightPending,
      preflightFinished: !!snapshot.preflightFinished,
    })

    if (merged && Number(merged.percentage || 0) >= 99.999) {
      if (!completedFreeze.value) {
        completedFreeze.value = {
          ...merged,
          __frozenAt: Date.now(),
        }
      }
      summary.value = completedFreeze.value
    } else {
      completedFreeze.value = null
      summary.value = merged
    }
  }

  function clampPageValues(): void {
    if (completedPage.value > completedTotalPages.value) {
      completedPage.value = completedTotalPages.value
      completedJumpPage.value = completedTotalPages.value
    }
    if (pendingPage.value > pendingTotalPages.value) {
      pendingPage.value = pendingTotalPages.value
      pendingJumpPage.value = pendingTotalPages.value
    }
  }

  async function refresh(background = false): Promise<void> {
    if (!taskId.value) return
    if (!background) loading.value = true
    error.value = ''
    
    try {
      const [overview, completed, pending] = await Promise.all([
        getActiveTransfer(taskId.value),
        getActiveTransferCompleted(taskId.value, Math.max(0, (completedPage.value - 1) * PAGE_SIZE), PAGE_SIZE),
        getActiveTransferPending(taskId.value, Math.max(0, (pendingPage.value - 1) * PAGE_SIZE), PAGE_SIZE * 2),
      ])
      
      runId.value = overview.runId
      trackingMode.value = overview.trackingMode
      
      const merged = mergeNonDecreasingSummary(summary.value, overview.summary)
      
      if (merged && Number(merged.percentage || 0) >= 99.999) {
        if (!completedFreeze.value) {
          completedFreeze.value = {
            ...merged,
            __frozenAt: Date.now(),
          }
        }
        summary.value = completedFreeze.value
      } else {
        completedFreeze.value = null
        summary.value = merged
      }
      
      currentFile.value = overview.currentFile || null
      currentFiles.value = sortCurrentFiles(overview.currentFiles || (overview.currentFile ? [overview.currentFile] : []))
      transferSlots.value = Math.max(MIN_TRANSFER_SLOTS, Number(overview.transferSlots || overview.summary?.transferSlots || transferSlots.value || MIN_TRANSFER_SLOTS))
      degraded.value = !!overview.degraded
      
      completedItems.value = sortCompletedItems(completed.items || [])
      rawPendingItems.value = sortPendingItems(pending.items || [])
      
      const currentKeys = new Set((currentFiles.value || []).map(item => item.path || item.name).filter(Boolean))
      const filteredPending = currentKeys.size ? rawPendingItems.value.filter(item => !currentKeys.has(item.path || item.name)) : rawPendingItems.value
      pendingItems.value = filteredPending.slice(0, PAGE_SIZE)
      
      completedTotal.value = completed.total || 0
      pendingTotal.value = pending.total || 0
      
      clampPageValues()
    } catch (e: any) {
      const msg = String(e?.message || 'active transfer load failed')
      if (isNoActiveTransferError(msg)) {
        resetTransferState()
      } else {
        error.value = msg
        console.error('[useActiveTransferDetail] Refresh error:', e)
      }
    } finally {
      if (!background) loading.value = false
    }
  }

  function open(nextTaskId: number): void {
    taskId.value = nextTaskId
    runId.value = null
    completedPage.value = 1
    pendingPage.value = 1
    completedJumpPage.value = 1
    pendingJumpPage.value = 1
    visible.value = true
    void refresh(false)
  }

  function close(): void {
    visible.value = false
    resetTransferState()
    taskId.value = null
    runId.value = null
    completedPage.value = 1
    pendingPage.value = 1
    completedJumpPage.value = 1
    pendingJumpPage.value = 1
  }

  function navigatePage(pageRef: Ref<number>, jumpPageRef: Ref<number | null>, totalPages: number, direction: 'prev' | 'next'): void {
    if (direction === 'prev' && pageRef.value > 1) {
      pageRef.value -= 1
      jumpPageRef.value = pageRef.value
      void refresh(true)
    } else if (direction === 'next' && pageRef.value < totalPages) {
      pageRef.value += 1
      jumpPageRef.value = pageRef.value
      void refresh(true)
    }
  }

  function jumpToPage(pageRef: Ref<number>, jumpPageRef: Ref<number | null>, totalPages: number): void {
    const page = Math.min(Math.max(1, Number(jumpPageRef.value || 1)), totalPages)
    if (page === pageRef.value) return
    pageRef.value = page
    jumpPageRef.value = page
    void refresh(true)
  }

  function prevCompletedPage(): void {
    navigatePage(completedPage, completedJumpPage, completedTotalPages.value, 'prev')
  }

  function nextCompletedPage(): void {
    navigatePage(completedPage, completedJumpPage, completedTotalPages.value, 'next')
  }

  function jumpCompletedPage(): void {
    jumpToPage(completedPage, completedJumpPage, completedTotalPages.value)
  }

  function prevPendingPage(): void {
    navigatePage(pendingPage, pendingJumpPage, pendingTotalPages.value, 'prev')
  }

  function nextPendingPage(): void {
    navigatePage(pendingPage, pendingJumpPage, pendingTotalPages.value, 'next')
  }

  function jumpPendingPage(): void {
    jumpToPage(pendingPage, pendingJumpPage, pendingTotalPages.value)
  }

  const visibleCompletedItems = computed(() => completedItems.value)
  const visiblePendingItems = computed(() => pendingItems.value)

  const offActiveTransferSnapshot = onWsMessage('active_transfer_snapshot', (data) => {
    if (shouldHandleRunMessage(data?.run_id, data?.task_id) && data?.snapshot) {
      runId.value = Number(data.run_id || data.snapshot.runId || runId.value || 0) || null
      applySnapshot(data.snapshot as ActiveTransferSnapshot)
    }
  })

  const offRunProgress = onWsMessage('run_progress', (data) => {
    if (shouldHandleRunMessage(data?.run_id)) {
      const prev = summary.value
      if (!prev) return
      
      const incomingPlannedFiles = Number(data?.plannedFiles || 0)
      const incomingLogicalTotalCount = Number(data?.logicalTotalCount || data?.totalCount || incomingPlannedFiles || prev.logicalTotalCount || prev.totalCount || 0)
      const nextCompletedCount = Math.max(Number(prev.completedCount || 0), Number(data?.completedFiles || 0))
      const nextBytes = Math.max(Number(prev.bytes || 0), Number(data?.bytes || 0))
      
      const merged = mergeNonDecreasingSummary(summary.value, {
        ...prev,
        bytes: nextBytes,
        totalBytes: Math.max(Number(prev.totalBytes || 0), Number(data?.total || prev.totalBytes || 0)),
        speed: Number(data?.speed || prev.speed || 0),
        percentage: Number(data?.percent || prev.percentage || 0),
        eta: Number(data?.eta || prev.eta || 0),
        plannedFiles: Math.max(Number(prev.plannedFiles || 0), incomingPlannedFiles),
        logicalTotalCount: incomingLogicalTotalCount,
        totalCount: incomingLogicalTotalCount,
        completedCount: nextCompletedCount,
        pendingCount: Math.max(0, incomingLogicalTotalCount - nextCompletedCount),
        phase: typeof data?.phase === 'string' ? data.phase : prev.phase,
        lastUpdatedAt: typeof data?.lastUpdatedAt === 'string' ? data.lastUpdatedAt : prev.lastUpdatedAt,
      })

      if (merged && Number(merged.percentage || 0) >= 99.999) {
        if (!completedFreeze.value) {
          completedFreeze.value = {
            ...merged,
            __frozenAt: Date.now(),
          }
        }
        summary.value = completedFreeze.value
      } else {
        completedFreeze.value = null
        summary.value = merged
      }
    }
  })

  const offRunStatus = onWsMessage('run_status', (data) => {
    if (shouldHandleRunMessage(data?.run_id) && data?.status !== 'running') {
      resetTransferState()
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
    activeTransferSummary: stableSummary,
    activeTransferCurrentFile: currentFile,
    activeTransferCurrentFiles: currentFiles,
    activeTransferSlots: transferSlots,
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

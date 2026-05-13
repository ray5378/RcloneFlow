import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { nextTick } from 'vue'

const mocks = vi.hoisted(() => {
  const wsHandlers = new Map<string, (data: any) => void>()
  return {
    getActiveTransfer: vi.fn(),
    getActiveTransferCompleted: vi.fn(),
    getActiveTransferPending: vi.fn(),
    wsHandlers,
    onWsMessage: vi.fn((event: string, handler: (data: any) => void) => {
      wsHandlers.set(event, handler)
      return vi.fn(() => {
        if (wsHandlers.get(event) === handler) wsHandlers.delete(event)
      })
    }),
  }
})

vi.mock('../api/activeTransfer', () => ({
  getActiveTransfer: mocks.getActiveTransfer,
  getActiveTransferCompleted: mocks.getActiveTransferCompleted,
  getActiveTransferPending: mocks.getActiveTransferPending,
}))

vi.mock('./useWebSocket', () => ({
  onWsMessage: mocks.onWsMessage,
}))

import { useActiveTransferDetail } from './useActiveTransferDetail'

describe('useActiveTransferDetail', () => {
  const originalSetTimeout = globalThis.setTimeout

  beforeEach(() => {
    vi.clearAllMocks()
    mocks.wsHandlers.clear()
    ;(globalThis as any).setTimeout = ((fn: any) => { fn(); return 1 as any }) as any
  })

  afterEach(() => {
    ;(globalThis as any).setTimeout = originalSetTimeout
  })

  it('opens, refreshes, sorts data, paginates, and closes cleanly', async () => {
    mocks.getActiveTransfer.mockResolvedValue({
      runId: 88,
      trackingMode: 'cas',
      summary: {
        trackingMode: 'cas',
        completedCount: 1,
        pendingCount: 2,
        totalCount: 3,
        plannedFiles: 3,
        logicalTotalCount: 3,
        percentage: 10,
        bytes: 100,
        totalBytes: 1000,
        speed: 20,
        eta: 30,
        phase: 'copying',
      },
      currentFile: { name: 'z-last', path: '/z', status: 'in_progress', order: 2 },
      currentFiles: [
        { name: 'z-last', path: '/z', status: 'in_progress', order: 2 },
        { name: 'a-first', path: '/a', status: 'in_progress', order: 1 },
      ],
      degraded: true,
    })
    mocks.getActiveTransferCompleted.mockResolvedValue({
      total: 12,
      items: [
        { name: 'older', path: '/older', status: 'copied', order: 1, at: '2024-01-01T00:00:00Z' },
        { name: 'newer', path: '/newer', status: 'copied', order: 3, at: '2024-01-02T00:00:00Z' },
      ],
    })
    mocks.getActiveTransferPending.mockResolvedValue({
      total: 12,
      items: [
        { name: 'b', path: '/b', status: 'pending', order: 2 },
        { name: 'a-first', path: '/a', status: 'in_progress', order: 1 },
        { name: 'c', path: '/c', status: 'pending', order: 0 },
      ],
    })

    const api = useActiveTransferDetail()
    api.openActiveTransfer(77)
    await Promise.resolve()
    await Promise.resolve()
    await nextTick()

    expect(api.activeTransferVisible.value).toBe(true)
    expect(api.activeTransferTaskId.value).toBe(77)
    expect(api.activeTransferTrackingMode.value).toBe('cas')
    expect(api.activeTransferDegraded.value).toBe(true)
    expect(api.activeTransferLoading.value).toBe(false)
    expect(api.activeTransferSummary.value?.totalCount).toBe(3)
    expect(api.activeTransferCurrentFiles.value.map(it => it.name)).toEqual(['a-first', 'z-last'])
    expect(api.activeTransferCompletedItems.value.map(it => it.name)).toEqual(['newer', 'older'])
    expect(api.activeTransferPendingItems.value.map(it => it.name)).toEqual(['b', 'c'])
    expect(api.activeTransferCompletedTotal.value).toBe(12)
    expect(api.activeTransferPendingTotal.value).toBe(12)
    expect(api.activeTransferCompletedTotalPages.value).toBe(2)
    expect(api.activeTransferPendingTotalPages.value).toBe(2)

    await api.nextActiveTransferCompletedPage()
    expect(mocks.getActiveTransferCompleted).toHaveBeenLastCalledWith(77, 10, 10)
    await api.nextActiveTransferPendingPage()
    expect(mocks.getActiveTransferPending).toHaveBeenLastCalledWith(77, 10, 10)

    api.activeTransferCompletedJumpPage.value = 99
    await api.jumpActiveTransferCompletedPage()
    expect(api.activeTransferCompletedPage.value).toBe(2)

    api.activeTransferPendingJumpPage.value = 0
    await api.jumpActiveTransferPendingPage()
    expect(api.activeTransferPendingPage.value).toBe(1)

    api.closeActiveTransfer()
    expect(api.activeTransferVisible.value).toBe(false)
    expect(api.activeTransferTaskId.value).toBeNull()
    expect(api.activeTransferSummary.value).toBeNull()
    expect(api.activeTransferCompletedItems.value).toEqual([])
    expect(api.activeTransferPendingItems.value).toEqual([])
  })

  it('applies websocket snapshot/progress updates and refreshes on terminal run status', async () => {
    mocks.getActiveTransfer.mockResolvedValue({
      runId: 88,
      trackingMode: 'cas',
      summary: {
        trackingMode: 'cas',
        completedCount: 1,
        pendingCount: 2,
        totalCount: 3,
        plannedFiles: 3,
        logicalTotalCount: 3,
        percentage: 10,
        bytes: 100,
        totalBytes: 1000,
        speed: 20,
        eta: 30,
        phase: 'copying',
      },
      currentFile: { name: 'seed', path: '/seed', status: 'in_progress', order: 2 },
      currentFiles: [{ name: 'seed', path: '/seed', status: 'in_progress', order: 2 }],
      degraded: false,
    })
    mocks.getActiveTransferCompleted.mockResolvedValue({ total: 1, items: [{ name: 'done-1', path: '/done-1', order: 1, at: '2024-01-01T00:00:00Z' }] })
    mocks.getActiveTransferPending.mockResolvedValue({ total: 2, items: [{ name: 'pending-1', path: '/pending-1', status: 'pending', order: 2 }] })

    const api = useActiveTransferDetail()
    api.openActiveTransfer(77)
    await Promise.resolve()
    await Promise.resolve()
    await nextTick()

    mocks.wsHandlers.get('active_transfer_snapshot')?.({
      run_id: 88,
      task_id: 77,
      snapshot: {
        runId: 88,
        trackingMode: 'cas',
        currentFile: { name: 'live-b', path: '/b', status: 'in_progress', order: 2 },
        currentFiles: [
          { name: 'live-b', path: '/b', status: 'in_progress', order: 2 },
          { name: 'live-a', path: '/a', status: 'in_progress', order: 1 },
        ],
        completed: [
          { name: 'older', path: '/older', order: 1, at: '2024-01-01T00:00:00Z' },
          { name: 'newer', path: '/newer', order: 3, at: '2024-01-02T00:00:00Z' },
        ],
        pending: [
          { name: 'visible-pending', path: '/pending-visible', status: 'pending', order: 3 },
          { name: 'live-a', path: '/a', status: 'in_progress', order: 1 },
        ],
        totalCount: 4,
        degraded: true,
        preflightPending: true,
        preflightFinished: false,
      },
    })
    await nextTick()

    expect(api.activeTransferCurrentFiles.value.map(it => it.name)).toEqual(['live-a', 'live-b'])
    expect(api.activeTransferCompletedItems.value.map(it => it.name)).toEqual(['newer', 'older'])
    expect(api.activeTransferPendingItems.value.map(it => it.name)).toEqual(['visible-pending'])
    expect(api.activeTransferSummary.value?.totalCount).toBe(4)
    expect(api.activeTransferSummary.value?.completedCount).toBe(2)
    expect(api.activeTransferSummary.value?.pendingCount).toBe(2)
    expect(api.activeTransferSummary.value?.preflightPending).toBe(true)
    expect(api.activeTransferDegraded.value).toBe(true)

    mocks.wsHandlers.get('run_progress')?.({
      run_id: 88,
      bytes: 800,
      total: 1000,
      speed: 40,
      percent: 20,
      eta: 12,
      plannedFiles: 5,
      logicalTotalCount: 5,
      completedFiles: 3,
      phase: 'copying',
      lastUpdatedAt: '2024-01-02T00:00:00Z',
    })
    await nextTick()

    expect(api.activeTransferSummary.value?.bytes).toBe(800)
    expect(api.activeTransferSummary.value?.totalBytes).toBe(1000)
    expect(api.activeTransferSummary.value?.speed).toBe(40)
    expect(api.activeTransferSummary.value?.plannedFiles).toBe(5)
    expect(api.activeTransferSummary.value?.logicalTotalCount).toBe(5)
    expect(api.activeTransferSummary.value?.completedCount).toBe(3)
    expect(api.activeTransferSummary.value?.pendingCount).toBe(2)
    expect(api.activeTransferSummary.value?.phase).toBe('copying')
    expect(api.activeTransferSummary.value?.lastUpdatedAt).toBe('2024-01-02T00:00:00Z')
    expect(api.activeTransferSummary.value?.percentage).toBeGreaterThanOrEqual(80)

    mocks.wsHandlers.get('run_status')?.({ run_id: 88, status: 'success' })
    await Promise.resolve()
    await Promise.resolve()
    expect(mocks.getActiveTransfer).toHaveBeenCalledTimes(2)

    mocks.wsHandlers.get('run_status')?.({ run_id: 88, status: 'running' })
    await Promise.resolve()
    expect(mocks.getActiveTransfer).toHaveBeenCalledTimes(2)
  })

  it('treats not-found style errors as empty state but keeps real errors', async () => {
    mocks.getActiveTransfer.mockRejectedValueOnce(new Error('当前没有运行中的任务'))
    mocks.getActiveTransferCompleted.mockResolvedValue({ total: 0, items: [] })
    mocks.getActiveTransferPending.mockResolvedValue({ total: 0, items: [] })

    const api = useActiveTransferDetail()
    api.openActiveTransfer(55)
    await Promise.resolve()
    await Promise.resolve()
    await nextTick()

    expect(api.activeTransferError.value).toBe('')
    expect(api.activeTransferSummary.value).toBeNull()
    expect(api.activeTransferCompletedItems.value).toEqual([])

    mocks.getActiveTransfer.mockRejectedValueOnce(new Error('boom'))
    await api.refreshActiveTransfer()
    await Promise.resolve()
    await nextTick()
    expect(api.activeTransferError.value).toBe('boom')
  })
})

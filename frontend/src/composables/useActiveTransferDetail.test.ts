import { describe, it, expect, vi, beforeEach } from 'vitest'
import { createApp, defineComponent } from 'vue'
import { useActiveTransferDetail } from './useActiveTransferDetail'

const getActiveTransfer = vi.fn()
const getActiveTransferCompleted = vi.fn()
const getActiveTransferPending = vi.fn()
const listeners = new Map<string, (data: any) => void>()

vi.mock('../api/activeTransfer', () => ({
  getActiveTransfer: (...args: any[]) => getActiveTransfer(...args),
  getActiveTransferCompleted: (...args: any[]) => getActiveTransferCompleted(...args),
  getActiveTransferPending: (...args: any[]) => getActiveTransferPending(...args),
}))

vi.mock('./useWebSocket', () => ({
  onWsMessage: (type: string, cb: (data: any) => void) => {
    listeners.set(type, cb)
    return () => listeners.delete(type)
  },
}))

describe('useActiveTransferDetail', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    listeners.clear()
  })

  it('orders completed transfer items oldest first so new files append to the bottom', async () => {
    getActiveTransfer.mockResolvedValueOnce({
      runId: 41,
      taskId: 7,
      trackingMode: 'normal',
      summary: {
        trackingMode: 'normal',
        completedCount: 3,
        pendingCount: 0,
        totalCount: 3,
        percentage: 100,
        bytes: 300,
        totalBytes: 300,
        speed: 0,
        eta: 0,
      },
      currentFile: null,
      currentFiles: [],
      degraded: false,
    })
    getActiveTransferCompleted.mockResolvedValueOnce({
      total: 3,
      items: [
        { name: 'newest', path: 'newest', status: 'copied', order: 3, at: '2026-05-17T03:00:00Z' },
        { name: 'oldest', path: 'oldest', status: 'copied', order: 1, at: '2026-05-17T01:00:00Z' },
        { name: 'middle', path: 'middle', status: 'copied', order: 2, at: '2026-05-17T02:00:00Z' },
      ],
    })
    getActiveTransferPending.mockResolvedValueOnce({ total: 0, items: [] })

    let api!: ReturnType<typeof useActiveTransferDetail>
    const Host = defineComponent({
      setup() {
        api = useActiveTransferDetail()
        return () => null
      },
    })

    const el = document.createElement('div')
    document.body.appendChild(el)
    const app = createApp(Host)
    app.mount(el)

    api.openActiveTransfer(7)
    await Promise.resolve()
    await Promise.resolve()

    expect(api.activeTransferCompletedItems.value.map(item => item.name)).toEqual(['oldest', 'middle', 'newest'])

    listeners.get('active_transfer_snapshot')?.({
      run_id: 41,
      task_id: 7,
      snapshot: {
        runId: 41,
        taskId: 7,
        trackingMode: 'normal',
        totalCount: 4,
        completedCount: 4,
        pendingCount: 0,
        completed: [
          { name: 'latest', path: 'latest', status: 'copied', order: 4, at: '2026-05-17T04:00:00Z' },
          { name: 'oldest', path: 'oldest', status: 'copied', order: 1, at: '2026-05-17T01:00:00Z' },
        ],
        pending: [],
        degraded: true,
      },
    })
    await Promise.resolve()

    expect(api.activeTransferCompletedItems.value.map(item => item.name)).toEqual(['oldest', 'latest'])

    app.unmount()
    el.remove()
  })

  it('prefers snapshot count fields over retained item lengths in degraded mode', async () => {
    getActiveTransfer.mockResolvedValueOnce({
      runId: 31,
      taskId: 7,
      trackingMode: 'normal',
      summary: {
        trackingMode: 'normal',
        completedCount: 1,
        pendingCount: 1,
        totalCount: 2,
        percentage: 10,
        bytes: 10,
        totalBytes: 100,
        speed: 1,
        eta: 90,
      },
      currentFile: null,
      currentFiles: [],
      degraded: false,
    })
    getActiveTransferCompleted.mockResolvedValueOnce({
      total: 1,
      items: [{ name: 'done-a', path: 'done-a', status: 'copied', order: 1 }],
    })
    getActiveTransferPending.mockResolvedValueOnce({
      total: 1,
      items: [{ name: 'pending-a', path: 'pending-a', status: 'pending', order: 2 }],
    })

    let api!: ReturnType<typeof useActiveTransferDetail>
    const Host = defineComponent({
      setup() {
        api = useActiveTransferDetail()
        return () => null
      },
    })

    const el = document.createElement('div')
    document.body.appendChild(el)
    const app = createApp(Host)
    app.mount(el)

    api.openActiveTransfer(7)
    await Promise.resolve()
    await Promise.resolve()

    listeners.get('active_transfer_snapshot')?.({
      run_id: 31,
      task_id: 7,
      snapshot: {
        runId: 31,
        taskId: 7,
        trackingMode: 'normal',
        totalCount: 2050,
        completedCount: 2050,
        pendingCount: 0,
        completed: [{ name: 'done-z', path: 'done-z', status: 'copied', order: 2050 }],
        pending: [],
        degraded: true,
      },
    })
    await Promise.resolve()

    expect(api.activeTransferCompletedTotal.value).toBe(2050)
    expect(api.activeTransferPendingTotal.value).toBe(0)
    expect(api.activeTransferCompletedItems.value).toHaveLength(1)
    expect(api.activeTransferCompletedItems.value[0].name).toBe('done-z')
    expect(api.activeTransferDegraded.value).toBe(true)
    expect(api.activeTransferSummary.value?.completedCount).toBe(2050)
    expect(api.activeTransferSummary.value?.pendingCount).toBe(0)

    app.unmount()
    el.remove()
  })
})

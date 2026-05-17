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

async function mountActiveTransferDetail() {
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

  return {
    api,
    unmount: () => {
      app.unmount()
      el.remove()
    },
  }
}

async function flushPromises() {
  await Promise.resolve()
  await Promise.resolve()
}

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

    const { api, unmount } = await mountActiveTransferDetail()

    api.openActiveTransfer(7)
    await flushPromises()

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

    unmount()
  })

  it('limits websocket snapshot completed and pending lists to one page while browsing page one', async () => {
    getActiveTransfer.mockResolvedValueOnce({
      runId: 81,
      taskId: 11,
      trackingMode: 'normal',
      summary: {
        trackingMode: 'normal',
        completedCount: 0,
        pendingCount: 0,
        totalCount: 0,
        percentage: 0,
        bytes: 0,
        totalBytes: 0,
        speed: 0,
        eta: 0,
      },
      currentFile: null,
      currentFiles: [],
      degraded: false,
    })
    getActiveTransferCompleted.mockResolvedValueOnce({ total: 0, items: [] })
    getActiveTransferPending.mockResolvedValueOnce({ total: 0, items: [] })

    const { api, unmount } = await mountActiveTransferDetail()

    api.openActiveTransfer(11)
    await flushPromises()

    listeners.get('active_transfer_snapshot')?.({
      run_id: 81,
      task_id: 11,
      snapshot: {
        runId: 81,
        taskId: 11,
        trackingMode: 'normal',
        totalCount: 25,
        completedCount: 25,
        pendingCount: 25,
        completed: Array.from({ length: 25 }, (_, idx) => ({
          name: `completed-${idx + 1}`,
          path: `completed-${idx + 1}`,
          status: 'copied',
          order: idx + 1,
        })),
        pending: Array.from({ length: 25 }, (_, idx) => ({
          name: `pending-${idx + 1}`,
          path: `pending-${idx + 1}`,
          status: 'pending',
          order: idx + 1,
        })),
        degraded: false,
      },
    })
    await Promise.resolve()

    expect(api.activeTransferCompletedTotal.value).toBe(25)
    expect(api.activeTransferPendingTotal.value).toBe(25)
    expect(api.activeTransferCompletedItems.value).toHaveLength(10)
    expect(api.activeTransferPendingItems.value).toHaveLength(10)
    expect(api.activeTransferCompletedItems.value.map(item => item.name)).toEqual(
      Array.from({ length: 10 }, (_, idx) => `completed-${idx + 1}`),
    )
    expect(api.activeTransferPendingItems.value.map(item => item.name)).toEqual(
      Array.from({ length: 10 }, (_, idx) => `pending-${idx + 1}`),
    )

    unmount()
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

    const { api, unmount } = await mountActiveTransferDetail()

    api.openActiveTransfer(7)
    await flushPromises()

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

    unmount()
  })

  it('does not let websocket snapshots replace completed items while the user is browsing another completed page', async () => {
    const overview = {
      runId: 51,
      taskId: 8,
      trackingMode: 'normal',
      summary: {
        trackingMode: 'normal',
        completedCount: 12,
        pendingCount: 0,
        totalCount: 12,
        percentage: 80,
        bytes: 800,
        totalBytes: 1000,
        speed: 10,
        eta: 20,
      },
      currentFile: null,
      currentFiles: [],
      degraded: false,
    }
    getActiveTransfer.mockResolvedValueOnce(overview).mockResolvedValueOnce(overview)
    getActiveTransferCompleted
      .mockResolvedValueOnce({
        total: 12,
        items: Array.from({ length: 10 }, (_, idx) => ({
          name: `page-1-${idx}`,
          path: `page-1-${idx}`,
          status: 'copied',
          order: idx + 1,
        })),
      })
      .mockResolvedValueOnce({
        total: 12,
        items: [
          { name: 'page-2-a', path: 'page-2-a', status: 'copied', order: 11 },
          { name: 'page-2-b', path: 'page-2-b', status: 'copied', order: 12 },
        ],
      })
    getActiveTransferPending
      .mockResolvedValueOnce({ total: 0, items: [] })
      .mockResolvedValueOnce({ total: 0, items: [] })

    const { api, unmount } = await mountActiveTransferDetail()

    api.openActiveTransfer(8)
    await flushPromises()

    api.nextActiveTransferCompletedPage()
    await flushPromises()

    expect(api.activeTransferCompletedPage.value).toBe(2)
    expect(api.activeTransferCompletedItems.value.map(item => item.name)).toEqual(['page-2-a', 'page-2-b'])

    listeners.get('active_transfer_snapshot')?.({
      run_id: 51,
      task_id: 8,
      snapshot: {
        runId: 51,
        taskId: 8,
        trackingMode: 'normal',
        totalCount: 13,
        completedCount: 13,
        pendingCount: 0,
        currentFile: { name: 'current', path: 'current', status: 'in_progress', order: 13 },
        currentFiles: [{ name: 'current', path: 'current', status: 'in_progress', order: 13 }],
        completed: [
          { name: 'socket-first-page-a', path: 'socket-first-page-a', status: 'copied', order: 1 },
          { name: 'socket-first-page-b', path: 'socket-first-page-b', status: 'copied', order: 2 },
        ],
        pending: [],
        degraded: false,
      },
    })
    await Promise.resolve()

    expect(api.activeTransferCompletedTotal.value).toBe(13)
    expect(api.activeTransferCurrentFiles.value.map(item => item.name)).toEqual(['current'])
    expect(api.activeTransferCompletedPage.value).toBe(2)
    expect(api.activeTransferCompletedItems.value.map(item => item.name)).toEqual(['page-2-a', 'page-2-b'])

    unmount()
  })

  it('appends new websocket completed items while the user is browsing the last completed page', async () => {
    const overview = {
      runId: 61,
      taskId: 9,
      trackingMode: 'normal',
      summary: {
        trackingMode: 'normal',
        completedCount: 12,
        pendingCount: 0,
        totalCount: 12,
        percentage: 80,
        bytes: 800,
        totalBytes: 1000,
        speed: 10,
        eta: 20,
      },
      currentFile: null,
      currentFiles: [],
      degraded: false,
    }
    getActiveTransfer.mockResolvedValueOnce(overview).mockResolvedValueOnce(overview)
    getActiveTransferCompleted
      .mockResolvedValueOnce({
        total: 12,
        items: Array.from({ length: 10 }, (_, idx) => ({
          name: `page-1-${idx}`,
          path: `page-1-${idx}`,
          status: 'copied',
          order: idx + 1,
        })),
      })
      .mockResolvedValueOnce({
        total: 12,
        items: [
          { name: 'page-2-a', path: 'page-2-a', status: 'copied', order: 11 },
          { name: 'page-2-b', path: 'page-2-b', status: 'copied', order: 12 },
        ],
      })
    getActiveTransferPending
      .mockResolvedValueOnce({ total: 0, items: [] })
      .mockResolvedValueOnce({ total: 0, items: [] })

    const { api, unmount } = await mountActiveTransferDetail()

    api.openActiveTransfer(9)
    await flushPromises()

    api.nextActiveTransferCompletedPage()
    await flushPromises()

    expect(api.activeTransferCompletedPage.value).toBe(2)
    expect(api.activeTransferCompletedTotalPages.value).toBe(2)
    expect(api.activeTransferCompletedItems.value.map(item => item.name)).toEqual(['page-2-a', 'page-2-b'])

    listeners.get('active_transfer_snapshot')?.({
      run_id: 61,
      task_id: 9,
      snapshot: {
        runId: 61,
        taskId: 9,
        trackingMode: 'normal',
        totalCount: 23,
        completedCount: 23,
        pendingCount: 0,
        completed: [
          { name: 'retained-from-other-page-a', path: 'retained-from-other-page-a', status: 'copied', order: 3 },
          { name: 'retained-from-other-page-b', path: 'retained-from-other-page-b', status: 'copied', order: 4 },
          { name: 'page-2-a', path: 'page-2-a', status: 'copied', order: 11 },
          { name: 'page-2-b', path: 'page-2-b', status: 'copied', order: 12 },
          { name: 'newly-completed-13', path: 'newly-completed-13', status: 'copied', order: 13 },
          { name: 'newly-completed-14', path: 'newly-completed-14', status: 'copied', order: 14 },
          { name: 'newly-completed-15', path: 'newly-completed-15', status: 'copied', order: 15 },
          { name: 'newly-completed-16', path: 'newly-completed-16', status: 'copied', order: 16 },
          { name: 'newly-completed-17', path: 'newly-completed-17', status: 'copied', order: 17 },
          { name: 'newly-completed-18', path: 'newly-completed-18', status: 'copied', order: 18 },
          { name: 'newly-completed-19', path: 'newly-completed-19', status: 'copied', order: 19 },
          { name: 'newly-completed-20', path: 'newly-completed-20', status: 'copied', order: 20 },
          { name: 'belongs-page-3-21', path: 'belongs-page-3-21', status: 'copied', order: 21 },
          { name: 'belongs-page-3-22', path: 'belongs-page-3-22', status: 'copied', order: 22 },
          { name: 'belongs-page-3-23', path: 'belongs-page-3-23', status: 'copied', order: 23 },
        ],
        pending: [],
        degraded: false,
      },
    })
    await Promise.resolve()

    expect(api.activeTransferCompletedTotal.value).toBe(23)
    expect(api.activeTransferCompletedPage.value).toBe(2)
    expect(api.activeTransferCompletedTotalPages.value).toBe(3)
    expect(api.activeTransferCompletedItems.value.map(item => item.name)).toEqual([
      'page-2-a',
      'page-2-b',
      'newly-completed-13',
      'newly-completed-14',
      'newly-completed-15',
      'newly-completed-16',
      'newly-completed-17',
      'newly-completed-18',
      'newly-completed-19',
      'newly-completed-20',
    ])

    unmount()
  })

  it('does not append websocket completed items that belong to pages after the currently browsed page', async () => {
    const overview = {
      runId: 71,
      taskId: 10,
      trackingMode: 'normal',
      summary: {
        trackingMode: 'normal',
        completedCount: 20,
        pendingCount: 0,
        totalCount: 20,
        percentage: 80,
        bytes: 800,
        totalBytes: 1000,
        speed: 10,
        eta: 20,
      },
      currentFile: null,
      currentFiles: [],
      degraded: false,
    }
    getActiveTransfer.mockResolvedValueOnce(overview).mockResolvedValueOnce(overview)
    getActiveTransferCompleted
      .mockResolvedValueOnce({
        total: 20,
        items: Array.from({ length: 10 }, (_, idx) => ({
          name: `page-1-${idx + 1}`,
          path: `page-1-${idx + 1}`,
          status: 'copied',
          order: idx + 1,
        })),
      })
      .mockResolvedValueOnce({
        total: 20,
        items: Array.from({ length: 10 }, (_, idx) => ({
          name: `page-2-${idx + 11}`,
          path: `page-2-${idx + 11}`,
          status: 'copied',
          order: idx + 11,
        })),
      })
    getActiveTransferPending
      .mockResolvedValueOnce({ total: 0, items: [] })
      .mockResolvedValueOnce({ total: 0, items: [] })

    const { api, unmount } = await mountActiveTransferDetail()

    api.openActiveTransfer(10)
    await flushPromises()

    api.nextActiveTransferCompletedPage()
    await flushPromises()

    expect(api.activeTransferCompletedPage.value).toBe(2)
    expect(api.activeTransferCompletedTotalPages.value).toBe(2)
    expect(api.activeTransferCompletedItems.value.map(item => item.name)).toEqual([
      'page-2-11',
      'page-2-12',
      'page-2-13',
      'page-2-14',
      'page-2-15',
      'page-2-16',
      'page-2-17',
      'page-2-18',
      'page-2-19',
      'page-2-20',
    ])

    listeners.get('active_transfer_snapshot')?.({
      run_id: 71,
      task_id: 10,
      snapshot: {
        runId: 71,
        taskId: 10,
        trackingMode: 'normal',
        totalCount: 23,
        completedCount: 23,
        pendingCount: 0,
        completed: [
          { name: 'page-2-11', path: 'page-2-11', status: 'copied', order: 11 },
          { name: 'page-2-12', path: 'page-2-12', status: 'copied', order: 12 },
          { name: 'belongs-page-3-21', path: 'belongs-page-3-21', status: 'copied', order: 21 },
          { name: 'belongs-page-3-22', path: 'belongs-page-3-22', status: 'copied', order: 22 },
          { name: 'belongs-page-3-23', path: 'belongs-page-3-23', status: 'copied', order: 23 },
        ],
        pending: [],
        degraded: false,
      },
    })
    await Promise.resolve()

    expect(api.activeTransferCompletedTotal.value).toBe(23)
    expect(api.activeTransferCompletedTotalPages.value).toBe(3)
    expect(api.activeTransferCompletedPage.value).toBe(2)
    expect(api.activeTransferCompletedItems.value.map(item => item.name)).toEqual([
      'page-2-11',
      'page-2-12',
      'page-2-13',
      'page-2-14',
      'page-2-15',
      'page-2-16',
      'page-2-17',
      'page-2-18',
      'page-2-19',
      'page-2-20',
    ])

    unmount()
  })
})

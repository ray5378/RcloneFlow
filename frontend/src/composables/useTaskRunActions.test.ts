import { describe, it, expect, vi } from 'vitest'
import { useTaskRunActions } from './useTaskRunActions'

describe('useTaskRunActions', () => {
  it('runs task successfully and schedules refreshes', async () => {
    const loadData = vi.fn(async () => {})
    const loadActiveRuns = vi.fn(async () => {})
    const showToast = vi.fn()
    const originalSetTimeout = globalThis.setTimeout
    const timeoutSpy = vi.fn((fn: any) => { fn(); return 1 as any })
    ;(globalThis as any).setTimeout = timeoutSpy

    try {
      const api = useTaskRunActions({
        loadData,
        loadActiveRuns,
        showToast,
        taskApi: { run: vi.fn(async () => ({ started: true, taskId: 1 })), kill: vi.fn(async () => {}) },
      })

      const result = await api.runTask(1)

      expect(result).toEqual({ started: true, taskId: 1 })
      expect(loadData).toHaveBeenCalled()
      expect(loadActiveRuns).toHaveBeenCalled()
      expect(api.runningTaskId.value).toBeNull()
      expect(showToast).not.toHaveBeenCalled()
    } finally {
      ;(globalThis as any).setTimeout = originalSetTimeout
    }
  })

  it('blocks concurrent run requests and reports backend singleton block', async () => {
    const showToast = vi.fn()
    const api = useTaskRunActions({
      loadData: vi.fn(async () => {}),
      loadActiveRuns: vi.fn(async () => {}),
      showToast,
      taskApi: { run: vi.fn(async () => ({ started: true })), kill: vi.fn(async () => {}) },
    })

    api.runningTaskId.value = 5
    await api.runTask(6)
    expect(showToast).toHaveBeenCalled()

    api.runningTaskId.value = null
    const blocked = useTaskRunActions({
      loadData: vi.fn(async () => {}),
      loadActiveRuns: vi.fn(async () => {}),
      showToast,
      taskApi: { run: vi.fn(async () => ({ started: false, message: 'busy' })), kill: vi.fn(async () => {}) },
    })
    await expect(blocked.runTask(6)).resolves.toEqual({ started: false, message: 'busy' })
    expect(showToast).toHaveBeenCalledWith('busy', 'error')
  })

  it('handles run and stop failures', async () => {
    const showToast = vi.fn()
    const originalSetTimeout = globalThis.setTimeout
    const timeoutSpy = vi.fn((fn: any) => { fn(); return 1 as any })
    ;(globalThis as any).setTimeout = timeoutSpy

    try {
      const api = useTaskRunActions({
        loadData: vi.fn(async () => {}),
        loadActiveRuns: vi.fn(async () => {}),
        showToast,
        taskApi: {
          run: vi.fn(async () => { throw new Error('run fail') }),
          kill: vi.fn(async () => { throw new Error('stop fail') }),
        },
      })

      await expect(api.runTask(1)).resolves.toBeUndefined()
      expect(api.runningTaskId.value).toBeNull()
      expect(showToast).toHaveBeenCalledWith('run fail', 'error')

      await api.stopTaskAny(2)
      expect(api.stoppedTaskId.value).toBeNull()
      expect(showToast).toHaveBeenCalledWith('stop fail', 'error')
    } finally {
      ;(globalThis as any).setTimeout = originalSetTimeout
    }
  })

  it('stops task successfully and clears stop marker after delay', async () => {
    const loadData = vi.fn(async () => {})
    const loadActiveRuns = vi.fn(async () => {})
    const showToast = vi.fn()
    const originalSetTimeout = globalThis.setTimeout
    const timeoutSpy = vi.fn((fn: any) => { fn(); return 1 as any })
    ;(globalThis as any).setTimeout = timeoutSpy

    try {
      const api = useTaskRunActions({
        loadData,
        loadActiveRuns,
        showToast,
        taskApi: { run: vi.fn(async () => ({ started: true })), kill: vi.fn(async () => {}) },
      })

      await api.stopTaskAny(8)

      expect(loadData).toHaveBeenCalled()
      expect(loadActiveRuns).toHaveBeenCalled()
      expect(api.stoppedTaskId.value).toBeNull()
    } finally {
      ;(globalThis as any).setTimeout = originalSetTimeout
    }
  })
})

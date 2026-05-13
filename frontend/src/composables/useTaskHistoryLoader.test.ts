import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { ref, nextTick } from 'vue'
import { useTaskHistoryLoader } from './useTaskHistoryLoader'

describe('useTaskHistoryLoader', () => {
  const originalSetInterval = globalThis.setInterval
  const originalClearInterval = globalThis.clearInterval
  const originalVisibility = Object.getOwnPropertyDescriptor(document, 'visibilityState')

  beforeEach(() => {
    vi.restoreAllMocks()
  })

  afterEach(() => {
    ;(globalThis as any).setInterval = originalSetInterval
    ;(globalThis as any).clearInterval = originalClearInterval
    if (originalVisibility) {
      Object.defineProperty(document, 'visibilityState', originalVisibility)
    }
  })

  it('refreshes task runs when task filter exists', async () => {
    const taskRuns = ref<any[]>([])
    const runApi = { getRunsByTask: vi.fn(async () => [{ id: 1 }, { id: 2 }]) }
    const api = useTaskHistoryLoader({
      taskRuns,
      historyFilterTaskId: ref<number | null>(9),
      runsPage: ref(1),
      jumpPage: ref(1),
      currentModule: ref<'history' | 'add' | 'tasks'>('tasks'),
      runApi,
    })

    await api.refreshTaskHistoryRuns()
    expect(runApi.getRunsByTask).toHaveBeenCalledWith(9)
    expect(taskRuns.value).toEqual([{ id: 1 }, { id: 2 }])
  })

  it('viewTaskHistory resets paging, switches module, refreshes and starts loop', async () => {
    const setIntervalSpy = vi.fn((fn: any) => { return 123 as any })
    const clearIntervalSpy = vi.fn()
    ;(globalThis as any).setInterval = setIntervalSpy
    ;(globalThis as any).clearInterval = clearIntervalSpy

    const taskRuns = ref<any[]>([])
    const historyFilterTaskId = ref<number | null>(null)
    const runsPage = ref(3)
    const jumpPage = ref(5)
    const currentModule = ref<'history' | 'add' | 'tasks'>('tasks')
    const runApi = { getRunsByTask: vi.fn(async () => [{ id: 7 }]) }
    const api = useTaskHistoryLoader({ taskRuns, historyFilterTaskId, runsPage, jumpPage, currentModule, runApi })

    api.viewTaskHistory(7)
    await Promise.resolve()

    expect(runsPage.value).toBe(1)
    expect(jumpPage.value).toBe(1)
    expect(historyFilterTaskId.value).toBe(7)
    expect(currentModule.value).toBe('history')
    expect(runApi.getRunsByTask).toHaveBeenCalledWith(7)
    expect(setIntervalSpy).toHaveBeenCalled()
  })

  it('interval callback refreshes only when page is visible', async () => {
    let intervalFn: (() => void) | undefined
    ;(globalThis as any).setInterval = vi.fn((fn: any) => { intervalFn = fn; return 456 as any })
    ;(globalThis as any).clearInterval = vi.fn()
    Object.defineProperty(document, 'visibilityState', { configurable: true, get: () => 'visible' })

    const taskRuns = ref<any[]>([])
    const runApi = { getRunsByTask: vi.fn(async () => [{ id: 8 }]) }
    const api = useTaskHistoryLoader({
      taskRuns,
      historyFilterTaskId: ref<number | null>(null),
      runsPage: ref(1),
      jumpPage: ref(1),
      currentModule: ref<'history' | 'add' | 'tasks'>('tasks'),
      runApi,
    })

    api.viewTaskHistory(8)
    await Promise.resolve()
    expect(runApi.getRunsByTask).toHaveBeenCalledTimes(2)

    intervalFn?.()
    await Promise.resolve()
    expect(runApi.getRunsByTask).toHaveBeenCalledTimes(3)
    expect(runApi.getRunsByTask).toHaveBeenLastCalledWith(8)

    Object.defineProperty(document, 'visibilityState', { configurable: true, get: () => 'hidden' })
    intervalFn?.()
    await Promise.resolve()
    expect(runApi.getRunsByTask).toHaveBeenCalledTimes(3)
  })

  it('watch stops refresh loop when leaving history or clearing task filter', async () => {
    ;(globalThis as any).setInterval = vi.fn(() => 789 as any)
    const clearIntervalSpy = vi.fn()
    ;(globalThis as any).clearInterval = clearIntervalSpy

    const historyFilterTaskId = ref<number | null>(null)
    const currentModule = ref<'history' | 'add' | 'tasks'>('tasks')
    const api = useTaskHistoryLoader({
      taskRuns: ref<any[]>([]),
      historyFilterTaskId,
      runsPage: ref(1),
      jumpPage: ref(1),
      currentModule,
      runApi: { getRunsByTask: vi.fn(async () => []) },
    })

    api.viewTaskHistory(5)
    await Promise.resolve()

    currentModule.value = 'tasks'
    await nextTick()
    expect(clearIntervalSpy).toHaveBeenCalled()

    historyFilterTaskId.value = null
    currentModule.value = 'history'
    await nextTick()
    expect(clearIntervalSpy).toHaveBeenCalled()
  })
})

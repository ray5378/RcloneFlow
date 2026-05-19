import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { ref, nextTick } from 'vue'
import { useTaskHistoryLoader } from './useTaskHistoryLoader'

describe('useTaskHistoryLoader', () => {
  beforeEach(() => {
    vi.useFakeTimers()
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  const createOptions = () => ({
    taskRuns: ref<any[]>([]),
    historyFilterTaskId: ref<number | null>(null),
    runsPage: ref(1),
    jumpPage: ref(1),
    currentModule: ref<'history' | 'add' | 'tasks'>('history'),
    runApi: {
      getRunsByTask: vi.fn().mockResolvedValue([]),
    },
  })

  it('should return expected properties', () => {
    const options = createOptions()
    const loader = useTaskHistoryLoader(options)

    expect(loader).toHaveProperty('refreshTaskHistoryRuns')
    expect(loader).toHaveProperty('viewTaskHistory')
  })

  it('refreshTaskHistoryRuns should be a function', () => {
    const options = createOptions()
    const { refreshTaskHistoryRuns } = useTaskHistoryLoader(options)

    expect(typeof refreshTaskHistoryRuns).toBe('function')
  })

  it('viewTaskHistory should be a function', () => {
    const options = createOptions()
    const { viewTaskHistory } = useTaskHistoryLoader(options)

    expect(typeof viewTaskHistory).toBe('function')
  })

  it('viewTaskHistory should update state and trigger refresh', async () => {
    const options = createOptions()
    const mockRuns = [{ id: 1, status: 'success' }]
    options.runApi.getRunsByTask.mockResolvedValue(mockRuns)

    const { viewTaskHistory } = useTaskHistoryLoader(options)
    viewTaskHistory(42)

    expect(options.historyFilterTaskId.value).toBe(42)
    expect(options.currentModule.value).toBe('history')
    expect(options.runsPage.value).toBe(1)
    expect(options.jumpPage.value).toBe(1)
    expect(options.runApi.getRunsByTask).toHaveBeenCalledWith(42)

    await nextTick()
    expect(options.taskRuns.value).toEqual(mockRuns)
  })

  it('refreshTaskHistoryRuns should not call API when taskId is null', async () => {
    const options = createOptions()
    const { refreshTaskHistoryRuns } = useTaskHistoryLoader(options)

    await refreshTaskHistoryRuns()

    expect(options.runApi.getRunsByTask).not.toHaveBeenCalled()
  })

  it('should reconcile runs without unnecessary updates', async () => {
    const options = createOptions()
    const mockRuns = [{ id: 1, status: 'success' }]
    options.runApi.getRunsByTask.mockResolvedValue(mockRuns)

    const { refreshTaskHistoryRuns } = useTaskHistoryLoader(options)
    options.historyFilterTaskId.value = 1

    await refreshTaskHistoryRuns()
    const firstRef = options.taskRuns.value[0]

    // Call again with same data
    await refreshTaskHistoryRuns()
    const secondRef = options.taskRuns.value[0]

    // Should reuse same reference if data unchanged
    expect(firstRef).toBe(secondRef)
  })

  it('stableStringify should handle circular references', async () => {
    const options = createOptions()
    const circular: any = { a: 1 }
    circular.self = circular
    options.runApi.getRunsByTask.mockResolvedValue([circular])
    options.historyFilterTaskId.value = 1

    const { refreshTaskHistoryRuns } = useTaskHistoryLoader(options)
    await expect(refreshTaskHistoryRuns()).resolves.not.toThrow()
  })
})

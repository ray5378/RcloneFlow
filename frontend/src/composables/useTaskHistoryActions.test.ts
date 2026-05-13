import { describe, it, expect, vi } from 'vitest'
import { ref } from 'vue'
import { useTaskHistoryActions } from './useTaskHistoryActions'

describe('useTaskHistoryActions', () => {
  it('optimistically clears one run and refreshes on success', async () => {
    const runs = ref<any[]>([{ id: 1, taskId: 9 }, { id: 2, taskId: 9 }])
    const taskRuns = ref<any[]>([{ id: 1, taskId: 9 }, { id: 2, taskId: 9 }])
    const loadData = vi.fn(async () => {})
    const refreshTaskHistoryRuns = vi.fn(async () => {})
    const api = useTaskHistoryActions({
      runs,
      taskRuns,
      historyFilterTaskId: ref(9),
      runsPage: ref(2),
      jumpPage: ref(2),
      filteredRuns: ref([]),
      loadData,
      refreshTaskHistoryRuns,
      runApi: { delete: vi.fn(async () => true), deleteByTask: vi.fn(async () => true) },
    })

    await api.clearRun(1)

    expect(runs.value.map(r => r.id)).toEqual([2])
    expect(taskRuns.value.map(r => r.id)).toEqual([2])
    expect(loadData).toHaveBeenCalled()
    expect(refreshTaskHistoryRuns).toHaveBeenCalled()
  })

  it('restores lists when single delete fails', async () => {
    const runs = ref<any[]>([{ id: 1, taskId: 9 }, { id: 2, taskId: 9 }])
    const taskRuns = ref<any[]>([{ id: 1, taskId: 9 }, { id: 2, taskId: 9 }])
    const api = useTaskHistoryActions({
      runs,
      taskRuns,
      historyFilterTaskId: ref(null),
      runsPage: ref(1),
      jumpPage: ref(1),
      filteredRuns: ref([{ id: 1 } as any]),
      loadData: vi.fn(async () => {}),
      refreshTaskHistoryRuns: vi.fn(async () => {}),
      runApi: { delete: vi.fn(async () => false), deleteByTask: vi.fn(async () => true) },
    })

    await api.clearRun(1)

    expect(runs.value.map(r => r.id)).toEqual([1, 2])
    expect(taskRuns.value.map(r => r.id)).toEqual([1, 2])
  })

  it('clears all task runs and resets paging on success', async () => {
    const runs = ref<any[]>([{ id: 1, taskId: 9 }, { id: 2, taskId: 10 }])
    const taskRuns = ref<any[]>([{ id: 1, taskId: 9 }])
    const loadData = vi.fn(async () => {})
    const refreshTaskHistoryRuns = vi.fn(async () => {})
    const api = useTaskHistoryActions({
      runs,
      taskRuns,
      historyFilterTaskId: ref(9),
      runsPage: ref(3),
      jumpPage: ref(3),
      filteredRuns: ref([]),
      loadData,
      refreshTaskHistoryRuns,
      runApi: { delete: vi.fn(async () => true), deleteByTask: vi.fn(async () => true) },
    })

    await expect(api.clearAllRuns()).resolves.toBe(true)
    expect(taskRuns.value).toEqual([])
    expect(runs.value).toEqual([{ id: 2, taskId: 10 }])
    expect(loadData).toHaveBeenCalled()
    expect(refreshTaskHistoryRuns).toHaveBeenCalled()
  })

  it('returns false and restores state when clearAll fails or no task selected', async () => {
    const runs = ref<any[]>([{ id: 1, taskId: 9 }])
    const taskRuns = ref<any[]>([{ id: 1, taskId: 9 }])

    const noTask = useTaskHistoryActions({
      runs,
      taskRuns,
      historyFilterTaskId: ref(null),
      runsPage: ref(1),
      jumpPage: ref(1),
      filteredRuns: ref([]),
      loadData: vi.fn(async () => {}),
      refreshTaskHistoryRuns: vi.fn(async () => {}),
      runApi: { delete: vi.fn(async () => true), deleteByTask: vi.fn(async () => true) },
    })
    await expect(noTask.clearAllRuns()).resolves.toBe(false)

    const fail = useTaskHistoryActions({
      runs,
      taskRuns,
      historyFilterTaskId: ref(9),
      runsPage: ref(2),
      jumpPage: ref(2),
      filteredRuns: ref([]),
      loadData: vi.fn(async () => {}),
      refreshTaskHistoryRuns: vi.fn(async () => {}),
      runApi: { delete: vi.fn(async () => true), deleteByTask: vi.fn(async () => false) },
    })
    await expect(fail.clearAllRuns()).resolves.toBe(false)
    expect(runs.value).toEqual([{ id: 1, taskId: 9 }])
    expect(taskRuns.value).toEqual([{ id: 1, taskId: 9 }])
  })
})

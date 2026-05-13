import { describe, it, expect, vi } from 'vitest'
import { ref } from 'vue'

const computedApi = {
  filteredRuns: ref([{ id: 1 }]),
  filteredRunsTotal: ref(1),
  currentTotal: ref(5),
  currentTotalPages: ref(3),
}
const loaderApi = {
  refreshTaskHistoryRuns: vi.fn(async () => {}),
  viewTaskHistory: vi.fn(),
}
const pagingApi = {
  jumpToPage: vi.fn(),
}
const actionsApi = {
  clearRun: vi.fn(async () => {}),
  clearAllRuns: vi.fn(async () => true),
}

vi.mock('./useTaskHistoryComputed', () => ({ useTaskHistoryComputed: () => computedApi }))
vi.mock('./useTaskHistoryLoader', () => ({ useTaskHistoryLoader: () => loaderApi }))
vi.mock('./useTaskHistoryPagingEntry', () => ({ useTaskHistoryPagingEntry: () => pagingApi }))
vi.mock('./useTaskHistoryActions', () => ({ useTaskHistoryActions: () => actionsApi }))

import { useTaskHistoryRuntime } from './useTaskHistoryRuntime'

describe('useTaskHistoryRuntime', () => {
  it('wires child composables and exposes their state', async () => {
    const api = useTaskHistoryRuntime({
      runs: ref([]),
      runsTotal: ref(0),
      taskRuns: ref([]),
      historyFilterTaskId: ref(null),
      historyStatusFilter: ref('all'),
      runsPage: ref(1),
      runsPageSize: 50,
      jumpPage: ref(1),
      currentModule: ref('tasks'),
      getFinalSummary: vi.fn(),
      loadData: vi.fn(async () => {}),
      runApi: { getRunsByTask: vi.fn(), delete: vi.fn(), deleteByTask: vi.fn() },
    })

    expect(api.filteredRuns.value).toEqual([{ id: 1 }])
    expect(api.filteredRunsTotal.value).toBe(1)
    expect(api.currentTotal.value).toBe(5)
    expect(api.currentTotalPages.value).toBe(3)

    await api.refreshTaskHistoryRuns()
    api.viewTaskHistory(10)
    api.jumpToPage()
    await api.clearRun(1)
    await api.clearAllRuns()

    expect(loaderApi.refreshTaskHistoryRuns).toHaveBeenCalled()
    expect(loaderApi.viewTaskHistory).toHaveBeenCalledWith(10)
    expect(pagingApi.jumpToPage).toHaveBeenCalled()
    expect(actionsApi.clearRun).toHaveBeenCalledWith(1)
    expect(actionsApi.clearAllRuns).toHaveBeenCalled()
  })
})

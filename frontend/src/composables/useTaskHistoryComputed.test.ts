import { describe, it, expect, vi } from 'vitest'
import { ref } from 'vue'
import { useTaskHistoryComputed } from './useTaskHistoryComputed'

describe('useTaskHistoryComputed', () => {
  it('filters by status and paginates from global runs when no task filter is selected', () => {
    const getFinalSummary = vi.fn((run: any) => run.finalSummary)
    const api = useTaskHistoryComputed({
      runs: ref<any[]>([
        { id: 1, status: 'finished', finalSummary: { totalCount: 2, transferredBytes: 100 } },
        { id: 2, status: 'failed', finalSummary: { totalCount: 0, transferredBytes: 0 } },
        { id: 3, status: 'finished', finalSummary: { totalCount: 1, transferredBytes: 0 } },
      ]),
      runsTotal: ref(99),
      taskRuns: ref<any[]>([]),
      historyFilterTaskId: ref<number | null>(null),
      historyStatusFilter: ref('finished'),
      runsPage: ref(1),
      runsPageSize: 1,
      getFinalSummary,
    })

    expect(api.filteredRuns.value.map(r => r.id)).toEqual([1])
    expect(api.filteredRunsTotal.value).toBe(2)
    expect(api.currentTotal.value).toBe(99)
    expect(api.currentTotalPages.value).toBe(99)
  })

  it('uses task-scoped runs and hasTransfer filter', () => {
    const getFinalSummary = vi.fn((run: any) => run.finalSummary)
    const api = useTaskHistoryComputed({
      runs: ref<any[]>([]),
      runsTotal: ref(0),
      taskRuns: ref<any[]>([
        { id: 1, status: 'finished', finalSummary: { totalCount: 2, transferredBytes: 0 } },
        { id: 2, status: 'finished', finalSummary: { totalCount: 1, transferredBytes: 0 } },
        { id: 3, status: 'finished', finalSummary: { totalCount: 0, transferredBytes: 12 } },
      ]),
      historyFilterTaskId: ref<number | null>(7),
      historyStatusFilter: ref('hasTransfer'),
      runsPage: ref(1),
      runsPageSize: 10,
      getFinalSummary,
    })

    expect(api.filteredRuns.value.map(r => r.id)).toEqual([1, 3])
    expect(api.filteredRunsTotal.value).toBe(2)
    expect(api.currentTotal.value).toBe(2)
    expect(api.currentTotalPages.value).toBe(1)
  })

  it('caches final summary by run id and summary payload', () => {
    const getFinalSummary = vi.fn((run: any) => run.finalSummary)
    const runs = ref<any[]>([
      { id: 1, status: 'finished', summary: '{"a":1}', finalSummary: { totalCount: 2, transferredBytes: 0 } },
      { id: 2, status: 'finished', summary: '{"a":2}', finalSummary: { totalCount: 0, transferredBytes: 0 } },
    ])
    const historyStatusFilter = ref('hasTransfer')
    const api = useTaskHistoryComputed({
      runs,
      runsTotal: ref(2),
      taskRuns: ref<any[]>([]),
      historyFilterTaskId: ref<number | null>(null),
      historyStatusFilter,
      runsPage: ref(1),
      runsPageSize: 10,
      getFinalSummary,
    })

    expect(api.filteredRuns.value.map(r => r.id)).toEqual([1])
    expect(getFinalSummary).toHaveBeenCalledTimes(2)

    historyStatusFilter.value = 'all'
    expect(api.filteredRuns.value.map(r => r.id)).toEqual([1, 2])

    historyStatusFilter.value = 'hasTransfer'
    expect(api.filteredRuns.value.map(r => r.id)).toEqual([1])
    expect(getFinalSummary).toHaveBeenCalledTimes(2)

    runs.value = [
      { id: 1, status: 'finished', summary: '{"a":99}', finalSummary: { totalCount: 0, transferredBytes: 0 } },
      { id: 2, status: 'finished', summary: '{"a":2}', finalSummary: { totalCount: 0, transferredBytes: 0 } },
    ]
    expect(api.filteredRuns.value).toEqual([])
    expect(getFinalSummary).toHaveBeenCalledTimes(3)
  })
})

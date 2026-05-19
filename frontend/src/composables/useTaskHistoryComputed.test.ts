import { describe, it, expect } from 'vitest'
import { ref } from 'vue'
import { useTaskHistoryComputed } from './useTaskHistoryComputed'

describe('useTaskHistoryComputed', () => {
  function makeOptions() {
    return {
      runs: ref([
        { id: 1, taskId: 10, status: 'finished', summary: null },
        { id: 2, taskId: 20, status: 'failed', summary: null },
        { id: 3, taskId: 10, status: 'running', summary: null },
      ] as any[]),
      runsTotal: ref(3),
      taskRuns: ref([
        { id: 1, taskId: 10, status: 'finished', summary: null },
        { id: 3, taskId: 10, status: 'running', summary: null },
      ] as any[]),
      historyFilterTaskId: ref<number | null>(null),
      historyStatusFilter: ref('all'),
      runsPage: ref(1),
      runsPageSize: 10,
      getFinalSummary: () => null,
    }
  }

  it('should filter runs by status', () => {
    const opts = makeOptions()
    opts.historyStatusFilter.value = 'finished'
    const { filteredRuns } = useTaskHistoryComputed(opts)
    expect(filteredRuns.value).toHaveLength(1)
    expect(filteredRuns.value[0].status).toBe('finished')
  })

  it('should filter runs with transfer', () => {
    const opts = makeOptions()
    opts.historyStatusFilter.value = 'hasTransfer'
    opts.getFinalSummary = (run: any) => run.id === 1 ? { totalCount: 5, transferredBytes: 100 } : null
    const { filteredRuns } = useTaskHistoryComputed(opts)
    expect(filteredRuns.value).toHaveLength(1)
    expect(filteredRuns.value[0].id).toBe(1)
  })

  it('should paginate results', () => {
    const opts = makeOptions()
    opts.runsPageSize = 2
    const { filteredRuns } = useTaskHistoryComputed(opts)
    expect(filteredRuns.value).toHaveLength(2)
  })

  it('should compute total and pages', () => {
    const opts = makeOptions()
    opts.runsPageSize = 2
    const { currentTotal, currentTotalPages } = useTaskHistoryComputed(opts)
    expect(currentTotal.value).toBe(3)
    expect(currentTotalPages.value).toBe(2)
  })

  it('should use taskRuns when filter is set', () => {
    const opts = makeOptions()
    opts.historyFilterTaskId.value = 10
    const { filteredRunsTotal } = useTaskHistoryComputed(opts)
    expect(filteredRunsTotal.value).toBe(2)
  })

  it('should cache final summary', () => {
    const opts = makeOptions()
    opts.historyStatusFilter.value = 'hasTransfer'
    const getFinalSummary = vi.fn(() => ({ totalCount: 1 }))
    opts.getFinalSummary = getFinalSummary
    const { filteredRuns } = useTaskHistoryComputed(opts)
    // Access twice to verify caching
    void filteredRuns.value
    void filteredRuns.value
    expect(getFinalSummary).toHaveBeenCalledTimes(3) // Once per run
  })
})

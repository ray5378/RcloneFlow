import { describe, it, expect, vi } from 'vitest'
import { ref } from 'vue'
import { useTaskHistoryRuntime } from './useTaskHistoryRuntime'

describe('useTaskHistoryRuntime', () => {
  function makeOptions() {
    return {
      runs: ref([]),
      runsTotal: ref(0),
      taskRuns: ref([]),
      historyFilterTaskId: ref<number | null>(null),
      historyStatusFilter: ref('all'),
      runsPage: ref(1),
      runsPageSize: 50,
      jumpPage: ref(1),
      currentModule: ref<'history' | 'add' | 'tasks'>('tasks'),
      getFinalSummary: () => null,
      loadData: vi.fn().mockResolvedValue(undefined),
      runApi: {
        getRunsByTask: vi.fn().mockResolvedValue([]),
        delete: vi.fn().mockResolvedValue(true),
        deleteByTask: vi.fn().mockResolvedValue(true),
      },
    }
  }

  it('should expose all methods', () => {
    const opts = makeOptions()
    const runtime = useTaskHistoryRuntime(opts)
    expect(runtime.filteredRuns).toBeDefined()
    expect(runtime.filteredRunsTotal).toBeDefined()
    expect(runtime.currentTotal).toBeDefined()
    expect(runtime.currentTotalPages).toBeDefined()
    expect(typeof runtime.refreshTaskHistoryRuns).toBe('function')
    expect(typeof runtime.viewTaskHistory).toBe('function')
    expect(typeof runtime.jumpToPage).toBe('function')
    expect(typeof runtime.clearRun).toBe('function')
    expect(typeof runtime.clearAllRuns).toBe('function')
  })
})

import { describe, it, expect, vi } from 'vitest'
import { ref } from 'vue'
import { useTaskHistoryActions } from './useTaskHistoryActions'

describe('useTaskHistoryActions', () => {
  function makeOptions() {
    return {
      runs: ref([{ id: 1, taskId: 10 }, { id: 2, taskId: 20 }] as any[]),
      taskRuns: ref([{ id: 1, taskId: 10 }] as any[]),
      historyFilterTaskId: ref<number | null>(10),
      runsPage: ref(1),
      jumpPage: ref(1),
      filteredRuns: ref([{ id: 1, taskId: 10 }] as any[]),
      loadData: vi.fn().mockResolvedValue(undefined),
      refreshTaskHistoryRuns: vi.fn().mockResolvedValue(undefined),
      runApi: {
        delete: vi.fn().mockResolvedValue(true),
        deleteByTask: vi.fn().mockResolvedValue(true),
      },
    }
  }

  it('should clear a single run', async () => {
    const opts = makeOptions()
    const { clearRun } = useTaskHistoryActions(opts)
    await clearRun(1)
    expect(opts.runApi.delete).toHaveBeenCalledWith(1)
    expect(opts.loadData).toHaveBeenCalled()
  })

  it('should rollback on delete failure', async () => {
    const opts = makeOptions()
    opts.runApi.delete.mockResolvedValueOnce(false)
    const { clearRun } = useTaskHistoryActions(opts)
    await clearRun(1)
    expect(opts.runs.value).toHaveLength(2)
  })

  it('should clear all runs for a task', async () => {
    const opts = makeOptions()
    const { clearAllRuns } = useTaskHistoryActions(opts)
    const result = await clearAllRuns()
    expect(result).toBe(true)
    expect(opts.runApi.deleteByTask).toHaveBeenCalledWith(10)
    expect(opts.runsPage.value).toBe(1)
  })

  it('should return false when no task selected', async () => {
    const opts = makeOptions()
    opts.historyFilterTaskId.value = null
    const { clearAllRuns } = useTaskHistoryActions(opts)
    const result = await clearAllRuns()
    expect(result).toBe(false)
  })

  it('should rollback clearAllRuns on failure', async () => {
    const opts = makeOptions()
    opts.runApi.deleteByTask.mockResolvedValueOnce(false)
    const { clearAllRuns } = useTaskHistoryActions(opts)
    const result = await clearAllRuns()
    expect(result).toBe(false)
    expect(opts.runs.value).toHaveLength(2)
  })
})

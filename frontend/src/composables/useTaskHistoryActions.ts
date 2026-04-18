import type { Ref } from 'vue'
import type { Run } from '../types'

interface UseTaskHistoryActionsOptions {
  runs: Ref<Run[]>
  taskRuns: Ref<Run[]>
  historyFilterTaskId: Ref<number | null>
  runsPage: Ref<number>
  jumpPage: Ref<number>
  filteredRuns: Ref<Run[]>
  loadData: () => Promise<void>
  refreshTaskHistoryRuns: () => Promise<void>
  runApi: {
    delete: (id: number) => Promise<boolean>
    deleteByTask: (taskId: number) => Promise<boolean>
  }
}

export function useTaskHistoryActions(options: UseTaskHistoryActionsOptions) {
  async function clearRun(id: number) {
    const prevRuns = options.runs.value
    const prevTaskRuns = options.taskRuns.value

    options.runs.value = options.runs.value.filter(r => r.id !== id)
    options.taskRuns.value = options.taskRuns.value.filter(r => r.id !== id)
    if (options.historyFilterTaskId.value !== null && options.runsPage.value > 1 && options.filteredRuns.value.length === 0) {
      options.runsPage.value -= 1
      options.jumpPage.value = options.runsPage.value
    }

    const ok = await options.runApi.delete(id)
    if (!ok) {
      options.runs.value = prevRuns
      options.taskRuns.value = prevTaskRuns
      return
    }

    await options.loadData()
    if (options.historyFilterTaskId.value !== null) {
      await options.refreshTaskHistoryRuns()
    }
  }

  async function clearAllRuns() {
    if (options.historyFilterTaskId.value === null) {
      return false
    }

    const prevRuns = options.runs.value
    const prevTaskRuns = options.taskRuns.value
    const taskId = options.historyFilterTaskId.value

    options.taskRuns.value = []
    options.runs.value = options.runs.value.filter(r => r.taskId !== taskId)
    options.runsPage.value = 1
    options.jumpPage.value = 1

    const ok = await options.runApi.deleteByTask(taskId)
    if (!ok) {
      options.runs.value = prevRuns
      options.taskRuns.value = prevTaskRuns
      return false
    }

    await options.loadData()
    await options.refreshTaskHistoryRuns()
    return true
  }

  return {
    clearRun,
    clearAllRuns,
  }
}

import type { Ref } from 'vue'
import type { Run } from '../types'

interface UseTaskHistoryLoaderOptions {
  taskRuns: Ref<Run[]>
  historyFilterTaskId: Ref<number | null>
  runsPage: Ref<number>
  jumpPage: Ref<number>
  currentModule: Ref<'history' | 'add' | 'tasks'>
  runApi: {
    getRunsByTask: (taskId: number) => Promise<Run[]>
  }
}

export function useTaskHistoryLoader(options: UseTaskHistoryLoaderOptions) {
  async function refreshTaskHistoryRuns() {
    if (options.historyFilterTaskId.value === null) return
    const data = await options.runApi.getRunsByTask(options.historyFilterTaskId.value)
    if (data && Array.isArray(data)) {
      options.taskRuns.value = data
    }
  }

  function viewTaskHistory(taskId: number) {
    options.runsPage.value = 1
    options.jumpPage.value = 1
    options.historyFilterTaskId.value = taskId
    options.currentModule.value = 'history'
    refreshTaskHistoryRuns().catch(console.error)
  }

  return {
    refreshTaskHistoryRuns,
    viewTaskHistory,
  }
}

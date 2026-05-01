import { onMounted, onUnmounted, watch, type Ref } from 'vue'
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
  let historyRefreshTimer: number | null = null

  async function refreshTaskHistoryRuns() {
    if (options.historyFilterTaskId.value === null) return
    const data = await options.runApi.getRunsByTask(options.historyFilterTaskId.value)
    if (data && Array.isArray(data)) {
      options.taskRuns.value = data
    }
  }

  function stopHistoryRefreshLoop() {
    if (historyRefreshTimer) {
      clearInterval(historyRefreshTimer)
      historyRefreshTimer = null
    }
  }

  function startHistoryRefreshLoop() {
    stopHistoryRefreshLoop()
    if (options.currentModule.value !== 'history' || options.historyFilterTaskId.value === null) return
    historyRefreshTimer = window.setInterval(() => {
      if (document.visibilityState === 'visible') {
        refreshTaskHistoryRuns().catch(console.error)
      }
    }, 3000)
  }

  function viewTaskHistory(taskId: number) {
    options.runsPage.value = 1
    options.jumpPage.value = 1
    options.historyFilterTaskId.value = taskId
    options.currentModule.value = 'history'
    refreshTaskHistoryRuns().catch(console.error)
    startHistoryRefreshLoop()
  }

  watch([options.currentModule, options.historyFilterTaskId], ([module, taskId]) => {
    if (module === 'history' && taskId !== null) {
      refreshTaskHistoryRuns().catch(console.error)
      startHistoryRefreshLoop()
      return
    }
    stopHistoryRefreshLoop()
  })

  onMounted(() => {
    if (options.currentModule.value === 'history' && options.historyFilterTaskId.value !== null) {
      startHistoryRefreshLoop()
    }
  })

  onUnmounted(() => {
    stopHistoryRefreshLoop()
  })

  return {
    refreshTaskHistoryRuns,
    viewTaskHistory,
  }
}

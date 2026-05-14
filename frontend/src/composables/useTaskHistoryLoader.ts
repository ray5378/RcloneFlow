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
  const HISTORY_POLL_FAST_MS = 3000
  const HISTORY_POLL_IDLE_MS = 15000

  let historyRefreshTimer: number | null = null

  function stableStringify(value: any) {
    try {
      return JSON.stringify(value)
    } catch {
      return ''
    }
  }

  function reconcileRuns(current: Run[], incoming: Run[]) {
    const prev = Array.isArray(current) ? current : []
    const next = Array.isArray(incoming) ? incoming : []
    const prevById = new Map<number, Run>()
    for (const run of prev) {
      if (typeof run?.id === 'number') prevById.set(run.id, run)
    }

    let changed = prev.length !== next.length
    const merged = next.map((run, idx) => {
      const prevRun = typeof run?.id === 'number' ? prevById.get(run.id) : undefined
      if (!prevRun) {
        changed = true
        return run
      }
      const reused = stableStringify(prevRun) === stableStringify(run) ? prevRun : run
      if (!changed && prev[idx] !== reused) changed = true
      return reused
    })

    return changed ? merged : prev
  }

  async function refreshTaskHistoryRuns() {
    if (options.historyFilterTaskId.value === null) return
    const data = await options.runApi.getRunsByTask(options.historyFilterTaskId.value)
    if (data && Array.isArray(data)) {
      const merged = reconcileRuns(options.taskRuns.value || [], data)
      if (merged !== options.taskRuns.value) {
        options.taskRuns.value = merged
      }
    }
  }

  function stopHistoryRefreshLoop() {
    if (historyRefreshTimer) {
      clearTimeout(historyRefreshTimer)
      historyRefreshTimer = null
    }
  }

  function getHistoryPollDelay() {
    const hasRunning = (options.taskRuns.value || []).some(run => String(run?.status || '').toLowerCase() === 'running')
    return hasRunning ? HISTORY_POLL_FAST_MS : HISTORY_POLL_IDLE_MS
  }

  function scheduleNextHistoryRefresh(delay?: number) {
    stopHistoryRefreshLoop()
    if (options.currentModule.value !== 'history' || options.historyFilterTaskId.value === null) return
    const nextDelay = typeof delay === 'number' ? delay : getHistoryPollDelay()
    historyRefreshTimer = window.setTimeout(async () => {
      historyRefreshTimer = null
      try {
        if (document.visibilityState === 'visible') {
          await refreshTaskHistoryRuns()
        }
      } catch (err) {
        console.error(err)
      } finally {
        scheduleNextHistoryRefresh()
      }
    }, nextDelay)
  }

  function startHistoryRefreshLoop(delay?: number) {
    scheduleNextHistoryRefresh(delay)
  }

  function viewTaskHistory(taskId: number) {
    options.runsPage.value = 1
    options.jumpPage.value = 1
    options.historyFilterTaskId.value = taskId
    options.currentModule.value = 'history'
    refreshTaskHistoryRuns().catch(console.error)
    startHistoryRefreshLoop(1500)
  }

  watch([options.currentModule, options.historyFilterTaskId], ([module, taskId]) => {
    if (module === 'history' && taskId !== null) {
      refreshTaskHistoryRuns().catch(console.error)
      startHistoryRefreshLoop(1500)
      return
    }
    stopHistoryRefreshLoop()
  })

  watch(() => (options.taskRuns.value || []).some(run => String(run?.status || '').toLowerCase() === 'running'), () => {
    if (options.currentModule.value !== 'history' || options.historyFilterTaskId.value === null) return
    startHistoryRefreshLoop()
  })

  onMounted(() => {
    if (options.currentModule.value === 'history' && options.historyFilterTaskId.value !== null) {
      startHistoryRefreshLoop(1500)
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

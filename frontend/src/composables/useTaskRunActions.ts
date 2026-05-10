import { ref } from 'vue'
import { t } from '../i18n'

interface UseTaskRunActionsOptions {
  loadData: () => Promise<void>
  loadActiveRuns?: () => Promise<void>
  showToast: (message: string, type?: 'info' | 'success' | 'error') => void
  taskApi: {
    run: (taskId: number) => Promise<any>
    kill: (taskId: number) => Promise<void>
  }
}

export function useTaskRunActions(options: UseTaskRunActionsOptions) {
  const runningTaskId = ref<number | null>(null)
  const stoppedTaskId = ref<number | null>(null)

  async function stopTaskAny(taskId: number) {
    try {
      await options.taskApi.kill(taskId)
      stoppedTaskId.value = taskId
      setTimeout(() => {
        if (stoppedTaskId.value === taskId) {
          stoppedTaskId.value = null
        }
      }, 1000)
      await Promise.all([
        options.loadData(),
        options.loadActiveRuns?.() ?? Promise.resolve(),
      ])
    } catch (e) {
      stoppedTaskId.value = null
      options.showToast((e as Error).message, 'error')
    }
  }

  async function runTask(taskId: number) {
    if (runningTaskId.value !== null) {
      options.showToast(t('runtime.singletonBlocked'), 'error')
      return
    }
    runningTaskId.value = taskId
    try {
      const result = await options.taskApi.run(taskId)
      if (!result) {
        runningTaskId.value = null
        return
      }

      if (result.started === false) {
        runningTaskId.value = null
        options.showToast(result.message || t('runtime.singletonBlocked'), 'error')
        return result
      }

      try {
        await options.loadData()
        await options.loadActiveRuns?.()
        setTimeout(() => { options.loadActiveRuns?.().catch(console.error) }, 300)
        setTimeout(() => { options.loadActiveRuns?.().catch(console.error) }, 1200)
      } catch (e) {
        console.error(e)
      }
      setTimeout(() => {
        if (runningTaskId.value === taskId) {
          runningTaskId.value = null
        }
      }, 1000)
      return result
    } catch (e) {
      runningTaskId.value = null
      options.showToast((e as Error).message, 'error')
      return
    }
  }

  return {
    runningTaskId,
    stoppedTaskId,
    stopTaskAny,
    runTask,
  }
}

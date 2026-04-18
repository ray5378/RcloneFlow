import { ref } from 'vue'

interface UseTaskRunActionsOptions {
  loadData: () => Promise<void>
  showToast: (message: string, type?: 'info' | 'success' | 'error') => void
  taskApi: {
    run: (taskId: number) => Promise<any>
    kill: (taskId: number) => Promise<void>
  }
  jobApi: {
    list: () => Promise<any[]>
    stop: (jobId: number | string) => Promise<void>
  }
}

export function useTaskRunActions(options: UseTaskRunActionsOptions) {
  const runningTaskId = ref<number | null>(null)
  const stoppedTaskId = ref<number | null>(null)

  async function stopTaskAny(taskId: number) {
    try {
      await options.taskApi.kill(taskId)
      const active = await options.jobApi.list()
      const current = Array.isArray(active)
        ? active.find(x => x?.runRecord?.taskId === taskId && x?.runRecord?.rcJobId)
        : null
      if (current?.runRecord?.rcJobId) {
        await options.jobApi.stop(current.runRecord.rcJobId)
      }
      stoppedTaskId.value = taskId
      setTimeout(() => {
        if (stoppedTaskId.value === taskId) {
          stoppedTaskId.value = null
        }
      }, 10000)
      await options.loadData()
    } catch (e) {
      stoppedTaskId.value = null
      options.showToast((e as Error).message, 'error')
    }
  }

  async function runTask(taskId: number) {
    if (runningTaskId.value !== null) {
      options.showToast('单例模式：已有任务正在运行，跳过本次执行', 'error')
      return
    }
    runningTaskId.value = taskId
    const result = await options.taskApi.run(taskId)
    if (!result) {
      runningTaskId.value = null
      return
    }
    setTimeout(() => {
      if (runningTaskId.value === taskId) {
        runningTaskId.value = null
      }
    }, 5000)
    return result
  }

  return {
    runningTaskId,
    stoppedTaskId,
    stopTaskAny,
    runTask,
  }
}

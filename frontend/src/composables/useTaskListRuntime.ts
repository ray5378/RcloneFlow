import type { Ref } from 'vue'
import type { Schedule, Task } from '../types'
import { t } from '../i18n'
import { useTaskListActions } from './useTaskListActions'
import { useTaskRunActions } from './useTaskRunActions'
import { useTaskFormEntry } from './useTaskFormEntry'

export function useTaskListRuntime(options: {
  openMenuId: Ref<number | null>
  historyFilterTaskId: Ref<number | null>
  schedules: Ref<Schedule[]>
  loadData: () => Promise<void>
  loadActiveRuns: () => Promise<void>
  showConfirm: (title: string, message: string, onConfirm: () => void) => void
  showToast: (message: string, type?: 'info' | 'success' | 'error') => void
  clearAllRuns: () => Promise<boolean | void>
  currentModule: Ref<'history' | 'add' | 'tasks'>
  remotes: Ref<string[]>
  remoteApi: {
    list: () => Promise<{ remotes?: string[] }>
  }
  resetTaskFormForCreate: () => void
  resetTaskPathBrowse: () => void
  getScheduleByTaskId: (taskId: number) => any
  fillTaskFormForEdit: (task: Task, scheduleSpec?: string) => any
  restoreTaskPathBrowse: (task: Task) => Promise<void>
  taskApi: {
    delete: (id: number) => Promise<boolean>
    run: (taskId: number) => Promise<any>
    kill: (taskId: number) => Promise<void>
    updateSortOrders: (orders: Record<number, number>, priorityTaskId?: number) => Promise<boolean>
  }
  scheduleApi: {
    delete: (id: number) => Promise<void>
    update: (id: number, enabled: boolean) => Promise<void>
  }
}) {
  const {
    deleteTask,
    toggleSchedule,
    deleteSchedule,
    clearAllRunsWithConfirm,
    scheduleToggledTaskId,
  } = useTaskListActions({
    openMenuId: options.openMenuId,
    historyFilterTaskId: options.historyFilterTaskId,
    schedules: options.schedules,
    loadData: options.loadData,
    showConfirm: options.showConfirm,
    showToast: options.showToast,
    clearAllRuns: options.clearAllRuns,
    taskApi: {
      delete: options.taskApi.delete,
    },
    scheduleApi: options.scheduleApi,
  })

  const {
    runningTaskId,
    stoppedTaskId,
    stopTaskAny,
    runTask,
  } = useTaskRunActions({
    loadData: options.loadData,
    loadActiveRuns: options.loadActiveRuns,
    showToast: options.showToast,
    taskApi: {
      run: options.taskApi.run,
      kill: options.taskApi.kill,
    },
  })

  const {
    goToAddTask,
    editTask,
  } = useTaskFormEntry({
    currentModule: options.currentModule,
    openMenuId: options.openMenuId,
    remotes: options.remotes,
    remoteApi: options.remoteApi,
    resetTaskFormForCreate: options.resetTaskFormForCreate,
    resetTaskPathBrowse: options.resetTaskPathBrowse,
    getScheduleByTaskId: options.getScheduleByTaskId,
    fillTaskFormForEdit: options.fillTaskFormForEdit,
    restoreTaskPathBrowse: options.restoreTaskPathBrowse,
  })

  async function saveTaskSortOrders(orders: Record<number, number>, priorityTaskId?: number) {
    const ok = await options.taskApi.updateSortOrders(orders, priorityTaskId)
    if (!ok) return false
    await options.loadData()
    options.showToast(t('runtime.taskSortSave'), 'success')
    return true
  }

  return {
    deleteTask,
    toggleSchedule,
    deleteSchedule,
    clearAllRunsWithConfirm,
    runningTaskId,
    stoppedTaskId,
    scheduleToggledTaskId,
    stopTaskAny,
    runTask,
    goToAddTask,
    editTask,
    saveTaskSortOrders,
  }
}

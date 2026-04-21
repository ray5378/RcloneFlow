import type { Ref } from 'vue'
import type { Schedule } from '../types'
import { t } from '../i18n'

interface UseTaskListActionsOptions {
  openMenuId: Ref<number | null>
  historyFilterTaskId: Ref<number | null>
  schedules: Ref<Schedule[]>
  loadData: () => Promise<void>
  showConfirm: (title: string, message: string, onConfirm: () => void) => void
  showToast: (message: string, type?: 'info' | 'success' | 'error') => void
  clearAllRuns: () => Promise<boolean | void>
  taskApi: {
    delete: (id: number) => Promise<boolean>
  }
  scheduleApi: {
    delete: (id: number) => Promise<void>
    update: (id: number, enabled: boolean) => Promise<void>
  }
}

export function useTaskListActions(options: UseTaskListActionsOptions) {
  function getScheduleByTaskId(taskId: number) {
    return options.schedules.value.find(s => s.taskId === taskId)
  }

  async function deleteTask(id: number) {
    options.showConfirm(t('common.delete'), t('runtime.deleteTaskConfirm'), async () => {
      const success = await options.taskApi.delete(id)
      if (success) {
        options.openMenuId.value = null
        await options.loadData()
      }
    })
  }

  async function toggleSchedule(taskId: number) {
    const schedule = getScheduleByTaskId(taskId)
    if (!schedule) return
    await options.scheduleApi.update(schedule.id, !schedule.enabled)
    await options.loadData()
  }

  async function deleteSchedule(id: number) {
    if (!confirm(t('schedule.deleteConfirm'))) return
    await options.scheduleApi.delete(id)
    await options.loadData()
  }

  function clearAllRunsWithConfirm() {
    if (options.historyFilterTaskId.value === null) {
      options.showToast(t('runtime.chooseTaskFirst'), 'error')
      return
    }
    options.showConfirm(t('runtime.deleteAllHistory'), t('runtime.deleteAllHistoryConfirm'), async () => {
      await options.clearAllRuns()
    })
  }

  return {
    getScheduleByTaskId,
    deleteTask,
    toggleSchedule,
    deleteSchedule,
    clearAllRunsWithConfirm,
  }
}

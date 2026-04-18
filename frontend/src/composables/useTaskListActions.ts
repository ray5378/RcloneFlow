import type { Ref } from 'vue'
import type { Schedule } from '../types'

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
    options.showConfirm('删除任务', '确定删除此任务？此操作不可恢复！', async () => {
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
    if (!confirm('确定删除此定时任务？')) return
    await options.scheduleApi.delete(id)
    await options.loadData()
  }

  function clearAllRunsWithConfirm() {
    if (options.historyFilterTaskId.value === null) {
      options.showToast('请先选择任务', 'error')
      return
    }
    options.showConfirm('删除所有历史', '确定删除该任务所有历史记录？此操作不可恢复！', async () => {
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

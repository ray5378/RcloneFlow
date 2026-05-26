import { computed, ref, type Ref } from 'vue'
import type { Schedule } from '../types'
import { t } from '../i18n'

export interface UseTaskListActionsOptions {
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

export interface UseTaskListActionsReturn {
  getScheduleByTaskId: (taskId: number) => Schedule | undefined
  deleteTask: (id: number) => Promise<void>
  deleteSchedule: (id: number) => Promise<void>
  clearAllRunsWithConfirm: () => void
}

const TOAST_TYPES = {
  ERROR: 'error' as const,
}

export function useTaskListActions(options: UseTaskListActionsOptions): UseTaskListActionsReturn {
  const scheduleByTaskId = computed(() => {
    const index = new Map<number, Schedule>()
    for (const s of options.schedules.value || []) {
      const taskId = Number(s?.taskId)
      if (taskId > 0) {
        index.set(taskId, s)
      }
    }
    return index
  })

  function getScheduleByTaskId(taskId: number): Schedule | undefined {
    return scheduleByTaskId.value.get(Number(taskId))
  }

  async function deleteTask(id: number): Promise<void> {
    options.showConfirm(t('common.delete'), t('runtime.deleteTaskConfirm'), async () => {
      const success = await options.taskApi.delete(id)
      if (success) {
        options.openMenuId.value = null
        await options.loadData()
      }
    })
  }

  async function deleteSchedule(id: number): Promise<void> {
    options.showConfirm(t('common.delete'), t('schedule.deleteConfirm'), async () => {
      await options.scheduleApi.delete(id)
      await options.loadData()
    })
  }

  function clearAllRunsWithConfirm(): void {
    if (options.historyFilterTaskId.value === null) {
      options.showToast(t('runtime.chooseTaskFirst'), TOAST_TYPES.ERROR)
      return
    }
    options.showConfirm(
      t('runtime.deleteAllHistory'),
      t('runtime.deleteAllHistoryConfirm'),
      async () => {
        await options.clearAllRuns()
      }
    )
  }

  return {
    getScheduleByTaskId,
    deleteTask,
    deleteSchedule,
    clearAllRunsWithConfirm,
  }
}

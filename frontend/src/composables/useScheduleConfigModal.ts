import { computed, ref } from 'vue'
import { t } from '../i18n'
import type { Task } from '../types'
import { buildScheduleSpec, createScheduleForm, parseScheduleSpec, type ScheduleFormLike } from '../components/task/scheduleOptions'

export function useScheduleConfigModal(options: {
  getScheduleByTaskId: (taskId: number) => any
  scheduleApi: {
    create: (schedule: any) => Promise<any>
    update: (id: number, enabled: boolean, spec?: string) => Promise<any>
  }
  loadData: () => Promise<void>
  showToast: (message: string, type?: 'info' | 'success' | 'error') => void
}) {
  const scheduleConfigVisible = ref(false)
  const scheduleConfigSaving = ref(false)
  const scheduleConfigTask = ref<Task | null>(null)
  const scheduleConfigDraft = ref<ScheduleFormLike>(createScheduleForm())

  const scheduleConfigTitle = computed(() => {
    return scheduleConfigTask.value
      ? `${t('schedule.configTitle')} · ${scheduleConfigTask.value.name}`
      : t('schedule.configTitle')
  })

  function openScheduleConfigForTask(task: Task) {
    const schedule = options.getScheduleByTaskId(task.id)
    scheduleConfigTask.value = task
    scheduleConfigDraft.value = parseScheduleSpec(schedule?.spec, !!schedule?.enabled)
    scheduleConfigVisible.value = true
  }

  function closeScheduleConfig() {
    if (scheduleConfigSaving.value) return
    scheduleConfigVisible.value = false
  }

  async function saveScheduleConfig(draft: ScheduleFormLike) {
    scheduleConfigSaving.value = true
    try {
      const task = scheduleConfigTask.value
      if (!task?.id) {
        scheduleConfigVisible.value = false
        return
      }
      const oldSchedule = options.getScheduleByTaskId(task.id)
      if (draft.enableSchedule) {
        const spec = buildScheduleSpec(draft)
        if (oldSchedule) {
          await options.scheduleApi.update(oldSchedule.id, true, spec)
        } else {
          await options.scheduleApi.create({ taskId: task.id, spec, enabled: true })
        }
      } else if (oldSchedule) {
        await options.scheduleApi.update(oldSchedule.id, false)
      }
      await options.loadData()
      options.showToast(t('schedule.saved'), 'success')
      scheduleConfigVisible.value = false
    } finally {
      scheduleConfigSaving.value = false
    }
  }

  return {
    scheduleConfigVisible,
    scheduleConfigSaving,
    scheduleConfigDraft,
    scheduleConfigTitle,
    openScheduleConfigForTask,
    closeScheduleConfig,
    saveScheduleConfig,
  }
}

import type { Ref } from 'vue'
import type { Task } from '../types'

interface UseTaskFormSubmitOptions {
  createForm: Ref<any>
  editingTask: Ref<Task | null>
  normalizeTaskOptions: (raw: Record<string, any> | undefined | null) => Record<string, any>
  getScheduleByTaskId: (taskId: number) => any
  taskApi: {
    create: (task: any) => Promise<any>
    update: (id: number, task: any) => Promise<any>
  }
  scheduleApi: {
    create: (schedule: any) => Promise<any>
    update: (id: number, enabled: boolean, spec?: string) => Promise<any>
  }
}

export function useTaskFormSubmit(options: UseTaskFormSubmitOptions) {
  function validateTaskForm() {
    if (!options.createForm.value.name) {
      return '请输入任务名称'
    }
    if (!options.createForm.value.sourceRemote || !options.createForm.value.targetRemote) {
      return '请选择源和目标存储'
    }
    return ''
  }

  function buildTaskPayload() {
    return {
      name: options.createForm.value.name,
      mode: options.createForm.value.mode,
      sourceRemote: options.createForm.value.sourceRemote,
      sourcePath: options.createForm.value.sourcePath,
      targetRemote: options.createForm.value.targetRemote,
      targetPath: options.createForm.value.targetPath,
      options: options.normalizeTaskOptions(options.createForm.value.options),
    }
  }

  function buildScheduleSpec() {
    return [
      options.createForm.value.scheduleMinute || '00',
      options.createForm.value.scheduleHour || '*',
      options.createForm.value.scheduleDay || '*',
      options.createForm.value.scheduleMonth || '*',
      options.createForm.value.scheduleWeek || '*',
    ].join('|')
  }

  async function submitTaskForm() {
    const taskData = buildTaskPayload()

    if (options.editingTask.value) {
      await options.taskApi.update(options.editingTask.value.id, taskData)
      const oldSchedule = options.getScheduleByTaskId(options.editingTask.value.id)
      if (options.createForm.value.enableSchedule) {
        const spec = buildScheduleSpec()
        if (oldSchedule) {
          await options.scheduleApi.update(oldSchedule.id, true, spec)
        } else {
          await options.scheduleApi.create({ taskId: options.editingTask.value.id, spec, enabled: true })
        }
      } else if (oldSchedule) {
        await options.scheduleApi.update(oldSchedule.id, false)
      }
      return { kind: 'update' as const }
    }

    const task = await options.taskApi.create(taskData)
    if (task && options.createForm.value.enableSchedule) {
      const spec = buildScheduleSpec()
      await options.scheduleApi.create({ taskId: task.id, spec, enabled: true })
    }
    return { kind: 'create' as const, task }
  }

  return {
    validateTaskForm,
    buildTaskPayload,
    buildScheduleSpec,
    submitTaskForm,
  }
}

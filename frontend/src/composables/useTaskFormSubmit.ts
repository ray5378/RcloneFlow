import type { Ref } from 'vue'
import type { Task } from '../types'

interface UseTaskFormSubmitOptions {
  createForm: Ref<any>
  editingTask: Ref<Task | null>
  creatingState: Ref<'idle' | 'loading' | 'done'>
  currentModule: Ref<'history' | 'add' | 'tasks'>
  normalizeTaskOptions: (raw: Record<string, any> | undefined | null) => Record<string, any>
  getScheduleByTaskId: (taskId: number) => any
  loadData: () => Promise<any>
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
  function handleTaskFormDoneClick() {
    if (options.creatingState.value !== 'done') return false
    options.creatingState.value = 'idle'
    options.currentModule.value = 'tasks'
    return true
  }

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

  async function completeTaskFormSubmit() {
    options.editingTask.value = null
    await options.loadData()
    options.currentModule.value = 'add'
    options.creatingState.value = 'done'
  }

  function resetTaskFormSubmitState() {
    options.creatingState.value = 'idle'
  }

  async function executeTaskFormSubmit() {
    options.creatingState.value = 'loading'
    try {
      await submitTaskForm()
      await completeTaskFormSubmit()
      return ''
    } catch (e) {
      resetTaskFormSubmitState()
      return (e as Error).message
    }
  }

  return {
    handleTaskFormDoneClick,
    validateTaskForm,
    buildTaskPayload,
    buildScheduleSpec,
    submitTaskForm,
    completeTaskFormSubmit,
    resetTaskFormSubmitState,
    executeTaskFormSubmit,
  }
}

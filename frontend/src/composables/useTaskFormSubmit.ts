import type { Ref } from 'vue'
import type { CreateForm, TaskFormOptions, TaskMode } from '../components/task/types'
import type { Task } from '../types'
import { t } from '../i18n'

interface TaskPayload {
  name: string
  mode: TaskMode
  sourceRemote: string
  sourcePath: string
  targetRemote: string
  targetPath: string
  options: TaskFormOptions
}

interface UseTaskFormSubmitOptions {
  createForm: Ref<CreateForm>
  editingTask: Ref<Task | null>
  creatingState: Ref<'idle' | 'loading' | 'done'>
  currentModule: Ref<'history' | 'add' | 'tasks'>
  normalizeTaskOptions: (raw: TaskFormOptions | undefined | null) => TaskFormOptions
  loadData: () => Promise<void>
  showToast: (message: string, type?: 'info' | 'success' | 'error') => void
  taskApi: {
    create: (task: TaskPayload) => Promise<Task>
    update: (id: number, task: TaskPayload) => Promise<unknown>
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
      return t('runtime.enterTaskName')
    }
    if (!options.createForm.value.sourceRemote || !options.createForm.value.targetRemote) {
      return t('runtime.chooseSourceTarget')
    }
    return ''
  }

  function buildTaskPayload(): TaskPayload {
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

  async function submitTaskForm() {
    const taskData = buildTaskPayload()

    if (options.editingTask.value) {
      await options.taskApi.update(options.editingTask.value.id, taskData)
      return { kind: 'update' as const }
    }

    const task = await options.taskApi.create(taskData)
    return { kind: 'create' as const, task }
  }

  async function completeTaskFormSubmit(kind: 'create' | 'update') {
    options.editingTask.value = null
    await options.loadData()
    options.showToast(
      kind === 'create' ? t('runtime.taskCreateSuccess') : t('runtime.taskUpdateSuccess'),
      'success',
    )
    options.currentModule.value = 'tasks'
    options.creatingState.value = 'done'
  }

  function resetTaskFormSubmitState() {
    options.creatingState.value = 'idle'
  }

  async function executeTaskFormSubmit() {
    options.creatingState.value = 'loading'
    try {
      const result = await submitTaskForm()
      await completeTaskFormSubmit(result.kind)
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
    submitTaskForm,
    completeTaskFormSubmit,
    resetTaskFormSubmitState,
    executeTaskFormSubmit,
  }
}

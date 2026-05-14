import { ref, type Ref } from 'vue'
import type { CreateForm, ParsedRcloneCommand, TaskFormOptions, TaskMode } from '../components/task/types'
import type { Task } from '../types'
import { useTaskFormSubmit } from './useTaskFormSubmit'
import { useTaskFormPrepare } from './useTaskFormPrepare'
import { useTaskFormFlow } from './useTaskFormFlow'
import { useTaskFormEntrySubmit } from './useTaskFormEntrySubmit'

export function useTaskFormOrchestrator(options: {
  createForm: Ref<CreateForm>
  editingTask: Ref<Task | null>
  currentModule: Ref<'history' | 'add' | 'tasks'>
  normalizeTaskOptions: (raw: TaskFormOptions | undefined | null) => TaskFormOptions
  loadData: () => Promise<void>
  taskApi: {
    create: (task: {
      name: string
      mode: TaskMode
      sourceRemote: string
      sourcePath: string
      targetRemote: string
      targetPath: string
      options: TaskFormOptions
    }) => Promise<Task>
    update: (id: number, task: {
      name: string
      mode: TaskMode
      sourceRemote: string
      sourcePath: string
      targetRemote: string
      targetPath: string
      options: TaskFormOptions
    }) => Promise<unknown>
  }
  commandMode: Ref<boolean>
  commandText: Ref<string>
  parseRcloneCommand: (cmd: string) => ParsedRcloneCommand
  showToast: (message: string, type?: 'info' | 'success' | 'error') => void
}) {
  const creatingState = ref<'idle' | 'loading' | 'done'>('idle')

  const {
    handleTaskFormDoneClick,
    validateTaskForm,
    executeTaskFormSubmit,
  } = useTaskFormSubmit({
    createForm: options.createForm,
    editingTask: options.editingTask,
    creatingState,
    currentModule: options.currentModule,
    normalizeTaskOptions: options.normalizeTaskOptions,
    loadData: options.loadData,
    taskApi: options.taskApi,
  })

  const { validateTaskFormBeforeSubmit } = useTaskFormPrepare({
    createForm: options.createForm,
    commandMode: options.commandMode,
    commandText: options.commandText,
    normalizeTaskOptions: options.normalizeTaskOptions,
    parseRcloneCommand: options.parseRcloneCommand,
    validateTaskForm,
  })

  const { runTaskFormFlow } = useTaskFormFlow({
    handleTaskFormDoneClick,
    validateTaskFormBeforeSubmit,
    executeTaskFormSubmit,
  })

  const { createTask } = useTaskFormEntrySubmit({
    runTaskFormFlow,
    showToast: options.showToast,
  })

  return {
    creatingState,
    handleTaskFormDoneClick,
    validateTaskForm,
    executeTaskFormSubmit,
    validateTaskFormBeforeSubmit,
    runTaskFormFlow,
    createTask,
  }
}

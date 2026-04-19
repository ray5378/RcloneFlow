import { ref } from 'vue'
import { useTaskFormSubmit } from './useTaskFormSubmit'
import { useTaskFormPrepare } from './useTaskFormPrepare'
import { useTaskFormFlow } from './useTaskFormFlow'
import { useTaskFormEntrySubmit } from './useTaskFormEntrySubmit'

export function useTaskFormOrchestrator(options: {
  createForm: any
  editingTask: any
  currentModule: any
  normalizeTaskOptions: (raw: Record<string, any> | undefined | null) => Record<string, any>
  getScheduleByTaskId: (taskId: number) => any
  loadData: () => Promise<void>
  taskApi: any
  scheduleApi: any
  commandMode: any
  commandText: any
  parseRcloneCommand: (cmd: string) => any
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
    getScheduleByTaskId: options.getScheduleByTaskId,
    loadData: options.loadData,
    taskApi: options.taskApi,
    scheduleApi: options.scheduleApi,
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

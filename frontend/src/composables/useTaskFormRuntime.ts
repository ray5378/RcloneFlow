import * as api from '../api'
import { useTaskFormState } from './useTaskFormState'
import { useTaskScheduleLookup } from './useTaskScheduleLookup'
import { useTaskFormOrchestrator } from './useTaskFormOrchestrator'
import { useTaskPathBrowse } from './useTaskPathBrowse'

export function useTaskFormRuntime(options: {
  schedules: any
  currentModule: any
  normalizeTaskOptions: (raw: Record<string, any> | undefined | null) => Record<string, any>
  loadData: () => Promise<void>
  taskApi: any
  scheduleApi: any
  showToast: (message: string, type?: 'info' | 'success' | 'error') => void
  parseRcloneCommand: (cmd: string) => any
}) {
  const {
    createForm,
    commandMode,
    commandText,
    editingTask,
    showAdvancedOptions,
    resetTaskFormForCreate,
    fillTaskFormForEdit,
  } = useTaskFormState()

  const { getScheduleByTaskId } = useTaskScheduleLookup(options.schedules)

  const {
    creatingState,
    handleTaskFormDoneClick,
    validateTaskForm,
    executeTaskFormSubmit,
    validateTaskFormBeforeSubmit,
    runTaskFormFlow,
    createTask,
  } = useTaskFormOrchestrator({
    createForm,
    editingTask,
    currentModule: options.currentModule,
    normalizeTaskOptions: options.normalizeTaskOptions,
    getScheduleByTaskId,
    loadData: options.loadData,
    taskApi: options.taskApi,
    scheduleApi: options.scheduleApi,
    commandMode,
    commandText,
    parseRcloneCommand: options.parseRcloneCommand,
    showToast: options.showToast,
  })

  const {
    sourcePathOptions,
    targetPathOptions,
    showSourcePathInput,
    showTargetPathInput,
    sourceCurrentPath,
    targetCurrentPath,
    sourceBreadcrumbs,
    targetBreadcrumbs,
    setShowSourcePathInput,
    setShowTargetPathInput,
    resetTaskPathBrowse,
    restoreTaskPathBrowse,
    onSourceRemoteChange,
    onTargetRemoteChange,
    onSourceBreadcrumbClick,
    onTargetBreadcrumbClick,
    loadSourcePath,
    loadTargetPath,
    onSourceClick,
    onSourceArrow,
    onTargetClick,
    onTargetArrow,
  } = useTaskPathBrowse({
    createForm,
    listPath: api.listPath,
  })

  return {
    createForm,
    commandMode,
    commandText,
    editingTask,
    showAdvancedOptions,
    resetTaskFormForCreate,
    fillTaskFormForEdit,
    getScheduleByTaskId,
    creatingState,
    handleTaskFormDoneClick,
    validateTaskForm,
    executeTaskFormSubmit,
    validateTaskFormBeforeSubmit,
    runTaskFormFlow,
    createTask,
    sourcePathOptions,
    targetPathOptions,
    showSourcePathInput,
    showTargetPathInput,
    sourceCurrentPath,
    targetCurrentPath,
    sourceBreadcrumbs,
    targetBreadcrumbs,
    setShowSourcePathInput,
    setShowTargetPathInput,
    resetTaskPathBrowse,
    restoreTaskPathBrowse,
    onSourceRemoteChange,
    onTargetRemoteChange,
    onSourceBreadcrumbClick,
    onTargetBreadcrumbClick,
    loadSourcePath,
    loadTargetPath,
    onSourceClick,
    onSourceArrow,
    onTargetClick,
    onTargetArrow,
  }
}

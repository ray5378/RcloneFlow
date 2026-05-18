import * as api from '../api'
import type { Ref } from 'vue'
import type { ParsedRcloneCommand, TaskFormOptions, TaskMode } from '../components/task/types'
import type { Schedule, Task } from '../types'
import { useTaskFormState } from './useTaskFormState'
import { useTaskScheduleLookup } from './useTaskScheduleLookup'
import { useTaskFormOrchestrator } from './useTaskFormOrchestrator'
import { useTaskPathBrowse } from './useTaskPathBrowse'

export function useTaskFormRuntime(options: {
  schedules: Ref<Schedule[]>
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
  showToast: (message: string, type?: 'info' | 'success' | 'error') => void
  parseRcloneCommand: (cmd: string) => ParsedRcloneCommand
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
    loadData: options.loadData,
    taskApi: options.taskApi,
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

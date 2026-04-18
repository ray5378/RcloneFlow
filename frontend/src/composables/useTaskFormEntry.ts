import type { Ref } from 'vue'
import type { Task } from '../types'

interface UseTaskFormEntryOptions {
  currentModule: Ref<'history' | 'add' | 'tasks'>
  openMenuId: Ref<number | null>
  remotes: Ref<string[]>
  remoteApi: {
    list: () => Promise<{ remotes?: string[] }>
  }
  resetTaskFormForCreate: () => void
  resetTaskPathBrowse: () => void
  getScheduleByTaskId: (taskId: number) => any
  fillTaskFormForEdit: (task: Task, scheduleSpec?: string) => any
  restoreTaskPathBrowse: (task: Task) => Promise<void>
}

export function useTaskFormEntry(options: UseTaskFormEntryOptions) {
  function goToTaskFormModule() {
    options.currentModule.value = 'add'
    options.openMenuId.value = null
  }

  async function goToAddTask() {
    const remoteData = await options.remoteApi.list()
    options.remotes.value = remoteData?.remotes || []
    options.resetTaskFormForCreate()
    options.resetTaskPathBrowse()
    goToTaskFormModule()
  }

  function editTask(task: Task) {
    const schedule = options.getScheduleByTaskId(task.id)
    options.fillTaskFormForEdit(task, schedule?.spec)
    options.restoreTaskPathBrowse(task)
    goToTaskFormModule()
  }

  return {
    goToAddTask,
    editTask,
  }
}

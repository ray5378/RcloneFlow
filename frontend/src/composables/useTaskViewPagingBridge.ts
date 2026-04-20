import type { Ref } from 'vue'

export function useTaskViewPagingBridge(options: {
  taskSearch: Ref<string>
  tasksJumpPage: Ref<number | null>
  historyStatusFilter: Ref<string>
  jumpPage: Ref<number>
  finalFilesJump: Ref<number | null>
  tasksPage: Ref<number>
  runsPage: Ref<number>
  currentModule: Ref<string>
  loadData: () => Promise<void>
}) {
  const setTaskSearch = (value: string) => { options.taskSearch.value = value }
  const setTasksJumpPageValue = (value: number | null) => { options.tasksJumpPage.value = value }
  const setHistoryStatusFilter = (value: string) => { options.historyStatusFilter.value = value }
  const setJumpPageValue = (value: number) => { options.jumpPage.value = value }
  const setFinalFilesJumpValue = (value: number | null) => { options.finalFilesJump.value = value }

  const prevTasksPage = () => { options.tasksPage.value-- }
  const nextTasksPage = () => { options.tasksPage.value++ }
  const backToTasks = () => { options.currentModule.value = 'tasks' }
  const prevRunsPage = async () => { options.runsPage.value--; await options.loadData() }
  const nextRunsPage = async () => { options.runsPage.value++; await options.loadData() }

  return {
    setTaskSearch,
    setTasksJumpPageValue,
    setHistoryStatusFilter,
    setJumpPageValue,
    setFinalFilesJumpValue,
    prevTasksPage,
    nextTasksPage,
    backToTasks,
    prevRunsPage,
    nextRunsPage,
  }
}

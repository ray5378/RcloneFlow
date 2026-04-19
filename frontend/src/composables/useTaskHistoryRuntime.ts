import type { Ref } from 'vue'
import type { Run } from '../types'
import { useTaskHistoryComputed } from './useTaskHistoryComputed'
import { useTaskHistoryLoader } from './useTaskHistoryLoader'
import { useTaskHistoryPagingEntry } from './useTaskHistoryPagingEntry'
import { useTaskHistoryActions } from './useTaskHistoryActions'

export function useTaskHistoryRuntime(options: {
  runs: Ref<Run[]>
  runsTotal: Ref<number>
  taskRuns: Ref<Run[]>
  historyFilterTaskId: Ref<number | null>
  historyStatusFilter: Ref<string>
  runsPage: Ref<number>
  runsPageSize: number
  jumpPage: Ref<number>
  currentModule: Ref<'history' | 'add' | 'tasks'>
  getFinalSummary: (run: Run) => any
  loadData: () => Promise<void>
  runApi: {
    getRunsByTask: (taskId: number) => Promise<Run[]>
    delete: (id: number) => Promise<boolean>
    deleteByTask: (taskId: number) => Promise<boolean>
  }
}) {
  const {
    filteredRuns,
    filteredRunsTotal,
    currentTotal,
    currentTotalPages,
  } = useTaskHistoryComputed({
    runs: options.runs,
    runsTotal: options.runsTotal,
    taskRuns: options.taskRuns,
    historyFilterTaskId: options.historyFilterTaskId,
    historyStatusFilter: options.historyStatusFilter,
    runsPage: options.runsPage,
    runsPageSize: options.runsPageSize,
    getFinalSummary: options.getFinalSummary,
  })

  const {
    refreshTaskHistoryRuns,
    viewTaskHistory,
  } = useTaskHistoryLoader({
    taskRuns: options.taskRuns,
    historyFilterTaskId: options.historyFilterTaskId,
    runsPage: options.runsPage,
    jumpPage: options.jumpPage,
    currentModule: options.currentModule,
    runApi: options.runApi,
  })

  const { jumpToPage } = useTaskHistoryPagingEntry({
    jumpPage: options.jumpPage,
    runsPage: options.runsPage,
    currentTotalPages,
    loadData: options.loadData,
  })

  const {
    clearRun,
    clearAllRuns,
  } = useTaskHistoryActions({
    runs: options.runs,
    taskRuns: options.taskRuns,
    historyFilterTaskId: options.historyFilterTaskId,
    runsPage: options.runsPage,
    jumpPage: options.jumpPage,
    filteredRuns,
    loadData: options.loadData,
    refreshTaskHistoryRuns,
    runApi: options.runApi,
  })

  return {
    filteredRuns,
    filteredRunsTotal,
    currentTotal,
    currentTotalPages,
    refreshTaskHistoryRuns,
    viewTaskHistory,
    jumpToPage,
    clearRun,
    clearAllRuns,
  }
}

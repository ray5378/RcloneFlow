import type { Ref } from 'vue'
import type { Run, Schedule, Task } from '../types'
import { useTaskViewDataSync } from './useTaskViewDataSync'
import { useTaskProgressSync } from './useTaskProgressSync'
import { useTaskViewRefreshLifecycle } from './useTaskViewRefreshLifecycle'

export function useTaskViewRuntime(options: {
  tasks: Ref<Task[]>
  remotes: Ref<string[]>
  schedules: Ref<Schedule[]>
  runs: Ref<Run[]>
  runsTotal: Ref<number>
  runsPage: Ref<number>
  runsPageSize: number
  activeRuns: Ref<any[]>
  globalStats: Ref<any>
  showGlobalStatsModal: Ref<boolean>
  activeRunLookup: { getActiveRunByTaskId: (taskId: number) => any }
  lastRunningProgressByTask: Ref<Record<number, { sp: any; at: number }>>
  lastNonDecreasingTotalsByTask: Ref<Record<number, { totalBytes: number; totalCount: number }>>
  currentModule: Ref<'history' | 'add' | 'tasks'>
  stuckMs: number
  taskApi: { list: () => Promise<Task[]> }
  remoteApi: { list: () => Promise<{ remotes?: string[] }> }
  scheduleApi: { list: () => Promise<Schedule[]> }
  runApi: { list: (page: number, pageSize: number) => Promise<{ runs?: Run[]; total?: number }> }
  jobApi: { list: () => Promise<any[]> }
}) {
  const {
    loadData,
    loadActiveRuns,
    loadGlobalStats,
    openGlobalStats,
    setupRealtimeSync,
  } = useTaskViewDataSync({
    tasks: options.tasks,
    remotes: options.remotes,
    schedules: options.schedules,
    runs: options.runs,
    runsTotal: options.runsTotal,
    runsPage: options.runsPage,
    runsPageSize: options.runsPageSize,
    activeRuns: options.activeRuns,
    globalStats: options.globalStats,
    showGlobalStatsModal: options.showGlobalStatsModal,
    lastNonDecreasingTotalsByTask: options.lastNonDecreasingTotalsByTask,
    taskApi: options.taskApi,
    remoteApi: options.remoteApi,
    scheduleApi: options.scheduleApi,
    runApi: options.runApi,
    jobApi: options.jobApi,
  })

  const {
    getRunProgressFromSummary,
    getRealtimeProgressByRun,
    getRunningProgressByRun,
    getTaskCardProgressByTask,
    getRunningProgressByTask,
    formatBps,
    calcEtaFromAvg,
    triggerAutoRefresh,
  } = useTaskProgressSync({
    runs: options.runs,
    activeRuns: options.activeRuns,
    activeRunLookup: options.activeRunLookup,
    loadData,
    loadActiveRuns,
  })

  useTaskViewRefreshLifecycle({
    tasks: options.tasks,
    currentModule: options.currentModule,
    getRunningProgressByTask,
    loadData,
    loadActiveRuns,
    setupRealtimeSync,
    stuckMs: options.stuckMs,
  })

  return {
    loadData,
    loadActiveRuns,
    loadGlobalStats,
    openGlobalStats,
    setupRealtimeSync,
    getRunProgressFromSummary,
    getRealtimeProgressByRun,
    getRunningProgressByRun,
    getTaskCardProgressByTask,
    getRunningProgressByTask,
    formatBps,
    calcEtaFromAvg,
    triggerAutoRefresh,
  }
}

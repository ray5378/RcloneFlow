import { onUnmounted, type Ref } from 'vue'
import { useWebSocket, onWsMessage } from './useWebSocket'
import * as api from '../api'
import type { Run, Schedule, Task } from '../types'
import type { ActiveRun, GlobalStats } from '../api/run'
import type { TaskBootstrapPayload } from '../api/task'

interface UseTaskViewDataSyncOptions {
  tasks: Ref<Task[]>
  remotes: Ref<string[]>
  schedules: Ref<Schedule[]>
  runs: Ref<Run[]>
  runsTotal: Ref<number>
  runsPage: Ref<number>
  runsPageSize: number
  activeRuns: Ref<ActiveRun[]>
  globalStats: Ref<GlobalStats | null>
  showGlobalStatsModal: Ref<boolean>
  currentModule: Ref<'history' | 'add' | 'tasks'>
  lastNonDecreasingTotalsByTask: Ref<Record<number, { runId?: number; totalBytes: number; totalCount: number }>>
  taskApi: { list: () => Promise<Task[]>; bootstrap?: (page?: number, pageSize?: number) => Promise<TaskBootstrapPayload> }
  remoteApi: { list: () => Promise<{ remotes?: string[] }> }
  scheduleApi: { list: () => Promise<Schedule[]> }
  runApi: { list: (page: number, pageSize: number) => Promise<{ runs?: Run[]; total?: number }> }
  jobApi: { list: () => Promise<ActiveRun[]> }
}

const TASKS_SNAPSHOT_KEY = 'lastTasksSnapshot'
const TASKS_SNAPSHOT_VERSION = 2
const TASKS_SNAPSHOT_WRITE_DELAY_MS = 1200
const ACTIVE_RUNS_RELOAD_DELAY_MS = 150
const DATA_RELOAD_DELAY_MS = 300
const STAGED_RELOAD_DELAYS_MS = [500, 2000]
const ZERO_DELAY_MS = 0

export function useTaskViewDataSync(options: UseTaskViewDataSyncOptions) {
  let loadSeq = 0
  let activeRunsReloadTimer: number | null = null
  let dataReloadTimer: number | null = null
  let tasksSnapshotWriteTimer: number | null = null
  let realtimeInitialized = false
  let cleanupRealtime: (() => void) | null = null

  function stableStringify(value: any) {
    try {
      return JSON.stringify(value)
    } catch (e) {
      console.warn('[useTaskViewDataSync] Failed to stringify value:', e)
      return ''
    }
  }

  function isSameShape(a: any, b: any) {
    return stableStringify(a) === stableStringify(b)
  }

  function getActiveRunKey(item: any) {
    const keyFields = [
      () => item?.runRecord?.id,
      () => item?.runId,
      () => item?.id,
      () => item?.runRecord?.taskId,
      () => item?.taskId,
      () => item?.taskID,
      () => item?.task_id
    ]
    for (const getter of keyFields) {
      const val = getter()
      if (val != null) return String(val)
    }
    return ''
  }

  function reconcileListByKey<T>(current: T[], incoming: T[], getKey: (item: T) => string, isSame: (a: T, b: T) => boolean) {
    const prev = Array.isArray(current) ? current : []
    const next = Array.isArray(incoming) ? incoming : []
    const prevByKey = new Map<string, T>()
    for (const item of prev) {
      const key = getKey(item)
      if (key) prevByKey.set(key, item)
    }

    let changed = prev.length !== next.length
    const merged = next.map((item, idx) => {
      const key = getKey(item)
      const prevItem = key ? prevByKey.get(key) : undefined
      if (!prevItem) {
        changed = true
        return item
      }
      const reused = isSame(prevItem, item) ? prevItem : item
      if (!changed && prev[idx] !== reused) changed = true
      return reused
    })

    return changed ? merged : prev
  }

  function replaceActiveRuns(nextList: ActiveRun[]) {
    const merged = reconcileListByKey<ActiveRun>(
      options.activeRuns.value || [],
      nextList || [],
      getActiveRunKey,
      isSameShape,
    )
    if (merged !== options.activeRuns.value) {
      options.activeRuns.value = merged
    }
  }

  function compactTaskSnapshot(tasks: Task[]) {
    return (tasks || []).map((task) => ({
      id: Number(task?.id || 0),
      name: String(task?.name || ''),
      sourcePath: String(task?.sourcePath || task?.src || ''),
      destPath: String(task?.destPath || task?.dst || ''),
      scheduleId: task?.scheduleId ?? null,
      cron: task?.cron ?? null,
      autoRun: !!task?.autoRun,
      enabled: task?.enabled !== false,
      command: task?.command ?? task?.cmd ?? '',
      updatedAt: task?.updatedAt ?? task?.updated_at ?? null,
      sortIndex: task?.sortIndex ?? task?.sort_index ?? null,
    }))
  }

  function scheduleTasksSnapshotWrite() {
    if (tasksSnapshotWriteTimer) return
    tasksSnapshotWriteTimer = window.setTimeout(() => {
      tasksSnapshotWriteTimer = null
      try {
        const payload = {
          version: TASKS_SNAPSHOT_VERSION,
          savedAt: new Date().toISOString(),
          tasks: compactTaskSnapshot(options.tasks.value || []),
        }
        localStorage.setItem(TASKS_SNAPSHOT_KEY, JSON.stringify(payload))
      } catch (e) {
        console.error('[useTaskViewDataSync] Failed to write tasks snapshot:', e)
      }
    }, TASKS_SNAPSHOT_WRITE_DELAY_MS)
  }

  function restoreTasksSnapshot(): Task[] | null {
    try {
      const raw = localStorage.getItem(TASKS_SNAPSHOT_KEY)
      if (!raw) return null
      const parsed = JSON.parse(raw)
      if (Array.isArray(parsed)) return parsed as Task[]
      if (parsed && Number(parsed.version) >= 2 && Array.isArray(parsed.tasks)) {
        return parsed.tasks as Task[]
      }
    } catch (e) {
      console.warn('[useTaskViewDataSync] Failed to restore tasks snapshot:', e)
    }
    return null
  }

  async function loadData() {
    const seq = ++loadSeq
    try {
      if (options.taskApi.bootstrap) {
        const boot = await options.taskApi.bootstrap(options.runsPage.value, options.runsPageSize)
        if (seq !== loadSeq || !boot) return
        if (Array.isArray(boot.tasks)) options.tasks.value = boot.tasks
        if (Array.isArray(boot.activeRuns)) {
          replaceActiveRuns(boot.activeRuns)
        }
      }

      const [remoteData, scheduleData, runResult] = await Promise.all([
        options.remoteApi.list(),
        options.scheduleApi.list(),
        options.runApi.list(options.runsPage.value, options.runsPageSize),
      ])
      if (seq !== loadSeq) return
      if (Array.isArray(remoteData?.remotes) && remoteData.remotes.length > 0) options.remotes.value = remoteData.remotes
      if (Array.isArray(scheduleData) && scheduleData.length > 0) options.schedules.value = scheduleData
      if (runResult?.runs) {
        options.runs.value = runResult.runs
        options.runsTotal.value = typeof runResult.total === 'number' ? runResult.total : (runResult.runs?.length || 0)
      }
      scheduleTasksSnapshotWrite()
    } catch (e) {
      console.error('[useTaskViewDataSync] Failed to load data:', e)
      if (!options.tasks.value || options.tasks.value.length === 0) {
        const snap = restoreTasksSnapshot()
        if (Array.isArray(snap)) options.tasks.value = snap
      }
    }
  }

  async function loadActiveRuns() {
    const data = await options.jobApi.list()
    const list: any[] = (data || []).map((it: any) => {
      const raw: any = (it && typeof it.progress === 'object' && it.progress)
          ? { ...it.progress }
          : null
        if (!raw) return it
        raw.bytes = Number(raw.bytes || 0)
        raw.totalBytes = Number(raw.totalBytes || 0)
        raw.speed = Number(raw.speed || 0)
        raw.percentage = Number(raw.percentage || 0)
        raw.completedFiles = Number(raw.completedFiles || 0)
        raw.totalCount = Number(raw.totalCount || 0)
        raw.eta = Number(raw.eta || 0)
        if (raw.percentage < 0) raw.percentage = 0
        if (raw.percentage > 100) raw.percentage = 100
        const tid = it.runRecord?.taskId
        const runId = it.runRecord?.id
        if (tid) {
          const prevTotals = options.lastNonDecreasingTotalsByTask.value[tid] as any
          const prevRunId = prevTotals?.runId
          if (prevTotals && prevRunId === runId) {
            if (prevTotals.totalBytes > 0 && raw.totalBytes > 0 && raw.totalBytes < prevTotals.totalBytes) {
              raw.totalBytes = prevTotals.totalBytes
            }
          }
          const nextTotals = {
            runId,
            totalBytes: Math.max(prevRunId === runId ? (prevTotals?.totalBytes || 0) : 0, raw.totalBytes || 0),
            totalCount: raw.totalCount || 0,
          }
          options.lastNonDecreasingTotalsByTask.value[tid] = nextTotals
        }
        return {
          ...it,
          progress: raw,
        }
      })
      replaceActiveRuns(list)
    }

  async function loadGlobalStats() {
    const stats = await api.getGlobalStats()
    options.globalStats.value = stats || {}
  }

  function openGlobalStats() {
    options.showGlobalStatsModal.value = true
    loadGlobalStats()
  }

  function scheduleActiveRunsReload(delay = ACTIVE_RUNS_RELOAD_DELAY_MS) {
    if (activeRunsReloadTimer) return
    activeRunsReloadTimer = window.setTimeout(() => {
      activeRunsReloadTimer = null
      loadActiveRuns().catch((e) => {
        console.error('[useTaskViewDataSync] Failed to reload active runs:', e)
      })
    }, delay)
  }

  function scheduleDataReload(delay = DATA_RELOAD_DELAY_MS) {
    if (dataReloadTimer) return
    dataReloadTimer = window.setTimeout(() => {
      dataReloadTimer = null
      loadData().catch((e) => {
        console.error('[useTaskViewDataSync] Failed to reload data:', e)
      })
    }, delay)
  }

  function scheduleActiveRunsReloadStaged() {
    STAGED_RELOAD_DELAYS_MS.forEach(delay => {
      window.setTimeout(() => {
        if (activeRunsReloadTimer) return
        loadActiveRuns().catch((e) => {
          console.error('[useTaskViewDataSync] Failed to reload active runs (staged):', e)
        })
      }, delay)
    })
  }

  function setupRealtimeSync() {
    if (realtimeInitialized) return
    realtimeInitialized = true

    const wsClient = useWebSocket({
      onMessage: (msg) => {
        if (msg.type === 'run_status' && msg.data) {
          const status = msg.data.status
          if (status && status !== 'running') {
            options.activeRuns.value = options.activeRuns.value.filter(
              r => r.runRecord?.id !== msg.data.run_id
            )
          }
          if (options.currentModule?.value === 'history') {
            const idx = options.runs.value.findIndex(r => r.id === msg.data.run_id)
            if (idx !== -1) {
              options.runs.value[idx] = { ...options.runs.value[idx], status: msg.data.status }
            }
          }
          scheduleActiveRunsReload(ZERO_DELAY_MS)
          scheduleActiveRunsReloadStaged()
          scheduleDataReload(ZERO_DELAY_MS)
        } else if (msg.type === 'run_progress' && msg.data) {
          const idx = options.activeRuns.value.findIndex(r => r.runRecord?.id === msg.data.run_id)
          if (idx !== -1) {
            const cur = options.activeRuns.value[idx] || {}
            const prev = cur.progress || {}
            const tid = cur.runRecord?.taskId
            const runId = cur.runRecord?.id
            const incomingTotalBytes = Number(msg.data.total || prev.totalBytes || 0)
            const incomingPlannedFiles = Number(msg.data.plannedFiles || 0)
            const incomingLogicalTotalCount = Number(msg.data.logicalTotalCount || msg.data.totalCount || incomingPlannedFiles || prev.logicalTotalCount || prev.totalCount || 0)
            const prevTotals = tid ? (options.lastNonDecreasingTotalsByTask.value[tid] as any) : undefined
            const prevRunId = prevTotals?.runId
            const nextTotalBytes = Math.max(prevRunId === runId ? (prevTotals?.totalBytes || 0) : 0, incomingTotalBytes || 0)
            const nextLogicalTotalCount = incomingLogicalTotalCount > 0 ? incomingLogicalTotalCount : (prevRunId === runId ? (prevTotals?.totalCount || 0) : 0)
            const nextCompletedFiles = Math.max(
              Number(prev.completedFiles || 0),
              Number(msg.data.completedFiles || 0),
            )
            const nextProgress = {
              ...prev,
              bytes: Number(msg.data.bytes || 0),
              totalBytes: nextTotalBytes,
              speed: Number(msg.data.speed || 0),
              percentage: Number(msg.data.percent || prev.percentage || 0),
              completedFiles: nextCompletedFiles,
              plannedFiles: Math.max(Number(prev.plannedFiles || 0), incomingPlannedFiles),
              logicalTotalCount: nextLogicalTotalCount,
              totalCount: nextLogicalTotalCount,
              eta: Number(msg.data.eta || prev.eta || 0),
            }
            if (tid) {
              options.lastNonDecreasingTotalsByTask.value[tid] = {
                runId,
                totalBytes: nextTotalBytes,
                totalCount: nextLogicalTotalCount,
              }
            }
            options.activeRuns.value[idx] = {
              ...cur,
              progress: nextProgress,
            }
            if (!nextProgress.completedFiles || !nextProgress.totalCount) {
              scheduleActiveRunsReload()
            }
          } else {
            scheduleActiveRunsReload(ZERO_DELAY_MS)
          }
          if (options.currentModule?.value === 'history') {
            const runIdx = options.runs.value.findIndex(r => r.id === msg.data.run_id)
            if (runIdx !== -1) {
              const curRun = options.runs.value[runIdx] || {}
              let sum: any = {}
              try {
                if (typeof curRun.summary === 'string') {
                  sum = JSON.parse(curRun.summary)
                } else {
                  sum = curRun.summary || {}
                }
              } catch (e) {
                console.warn('[useTaskViewDataSync] Failed to parse run summary:', e)
              }
              const prevProgress = sum.progress || {}
              const nextPlannedFiles = Math.max(Number(prevProgress.plannedFiles || 0), Number(msg.data.plannedFiles || 0))
              const nextLogicalTotalCount = Math.max(
                Number(prevProgress.logicalTotalCount || prevProgress.totalCount || 0),
                Number(msg.data.logicalTotalCount || msg.data.totalCount || msg.data.plannedFiles || 0),
              )
              sum.progress = {
                ...prevProgress,
                bytes: Number(msg.data.bytes || 0),
                totalBytes: Number(msg.data.total || prevProgress.totalBytes || 0),
                speed: Number(msg.data.speed || 0),
                percentage: Number(msg.data.percent || prevProgress.percentage || 0),
                completedFiles: Math.max(Number(prevProgress.completedFiles || 0), Number(msg.data.completedFiles || 0)),
                plannedFiles: nextPlannedFiles,
                logicalTotalCount: nextLogicalTotalCount,
                totalCount: nextLogicalTotalCount,
                eta: Number(msg.data.eta || prevProgress.eta || 0),
              }
              options.runs.value[runIdx] = {
                ...curRun,
                summary: sum,
              }
            }
          }
        }
      }
    })
    wsClient.connect()

    const offRunStatus = onWsMessage('run_status', () => {
      Promise.all([
        loadActiveRuns().catch((e) => {
          console.error('[useTaskViewDataSync] Failed to load active runs on status change:', e)
        }),
        loadData().catch((e) => {
          console.error('[useTaskViewDataSync] Failed to load data on status change:', e)
        }),
      ]).catch((e) => {
        console.error('[useTaskViewDataSync] Failed to handle status change:', e)
      })
    })

    cleanupRealtime = () => {
      offRunStatus()
      wsClient.cleanup()
      cleanupRealtime = null
      realtimeInitialized = false
    }
  }

  onUnmounted(() => {
    if (activeRunsReloadTimer) {
      clearTimeout(activeRunsReloadTimer)
      activeRunsReloadTimer = null
    }
    if (dataReloadTimer) {
      clearTimeout(dataReloadTimer)
      dataReloadTimer = null
    }
    if (tasksSnapshotWriteTimer) {
      clearTimeout(tasksSnapshotWriteTimer)
      tasksSnapshotWriteTimer = null
      try {
        const payload = {
          version: TASKS_SNAPSHOT_VERSION,
          savedAt: new Date().toISOString(),
          tasks: compactTaskSnapshot(options.tasks.value || []),
        }
        localStorage.setItem(TASKS_SNAPSHOT_KEY, JSON.stringify(payload))
      } catch (e) {
        console.error('[useTaskViewDataSync] Failed to write tasks snapshot on unmount:', e)
      }
    }
    cleanupRealtime?.()
  })

  return {
    loadData,
    loadActiveRuns,
    loadGlobalStats,
    openGlobalStats,
    setupRealtimeSync,
  }
}

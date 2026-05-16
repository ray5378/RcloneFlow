import { onUnmounted, type Ref } from 'vue'
import { useWebSocket, onWsMessage } from './useWebSocket'
import * as api from '../api'
import type { Run, Schedule, Task } from '../types'

interface UseTaskViewDataSyncOptions {
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
  currentModule: Ref<'history' | 'add' | 'tasks'>
  lastNonDecreasingTotalsByTask: Ref<Record<number, { runId?: number; totalBytes: number; totalCount: number }>>
  taskApi: { list: () => Promise<Task[]>; bootstrap?: (page?: number, pageSize?: number) => Promise<any> }
  remoteApi: { list: () => Promise<{ remotes?: string[] }> }
  scheduleApi: { list: () => Promise<Schedule[]> }
  runApi: { list: (page: number, pageSize: number) => Promise<{ runs?: Run[]; total?: number }> }
  jobApi: { list: () => Promise<any[]> }
}

const TASKS_SNAPSHOT_KEY = 'lastTasksSnapshot'
const TASKS_SNAPSHOT_VERSION = 2
const TASKS_SNAPSHOT_WRITE_DELAY_MS = 1200

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
    } catch {
      return ''
    }
  }

  function isSameShape(a: any, b: any) {
    return stableStringify(a) === stableStringify(b)
  }

  function getActiveRunKey(item: any) {
    return String(item?.runRecord?.id ?? item?.runId ?? item?.id ?? item?.runRecord?.taskId ?? item?.taskId ?? item?.taskID ?? item?.task_id ?? '')
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

  function replaceActiveRuns(nextList: any[]) {
    const merged = reconcileListByKey<any>(
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
    return (tasks || []).map((task: any) => ({
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
      } catch {}
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
    } catch {}
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
      console.error(e)
      if (!options.tasks.value || options.tasks.value.length === 0) {
        const snap = restoreTasksSnapshot()
        if (Array.isArray(snap)) options.tasks.value = snap
      }
    }
  }

  async function loadActiveRuns() {
    try {
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
    } catch (e) {
      console.error(e)
    }
  }

  async function loadGlobalStats() {
    try {
      const stats = await api.getGlobalStats()
      options.globalStats.value = stats || {}
    } catch (e) {
      console.error(e)
    }
  }

  function openGlobalStats() {
    options.showGlobalStatsModal.value = true
    loadGlobalStats()
  }

  function scheduleActiveRunsReload(delay = 150) {
    if (activeRunsReloadTimer) return
    activeRunsReloadTimer = window.setTimeout(() => {
      activeRunsReloadTimer = null
      loadActiveRuns().catch(console.error)
    }, delay)
  }

  function scheduleDataReload(delay = 300) {
    if (dataReloadTimer) return
    dataReloadTimer = window.setTimeout(() => {
      dataReloadTimer = null
      loadData().catch(console.error)
    }, delay)
  }

  function setupRealtimeSync() {
    if (realtimeInitialized) return
    realtimeInitialized = true

    const wsClient = useWebSocket({
      onMessage: (msg) => {
        if (msg.type === 'run_status' && msg.data) {
          if (options.currentModule?.value === 'history') {
            const idx = options.runs.value.findIndex(r => r.id === msg.data.run_id)
            if (idx !== -1) {
              options.runs.value[idx] = { ...options.runs.value[idx], status: msg.data.status }
            }
          }
          scheduleActiveRunsReload(0)
          scheduleDataReload(0)
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
            scheduleActiveRunsReload(0)
          }
          if (options.currentModule?.value === 'history') {
            const runIdx = options.runs.value.findIndex(r => r.id === msg.data.run_id)
            if (runIdx !== -1) {
              const curRun = options.runs.value[runIdx] || {}
              const sum = typeof curRun.summary === 'string'
                ? (() => { try { return JSON.parse(curRun.summary) } catch { return {} } })()
                : (curRun.summary || {})
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
        loadActiveRuns().catch(console.error),
        loadData().catch(console.error),
      ]).catch(console.error)
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
      } catch {}
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

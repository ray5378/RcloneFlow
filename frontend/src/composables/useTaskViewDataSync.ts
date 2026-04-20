import { type Ref } from 'vue'
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
  lastStableByTask: Ref<Record<number, { sp: any; at: number }>>
  lastNonDecreasingTotalsByTask: Ref<Record<number, { totalBytes: number; totalCount: number }>>
  taskApi: { list: () => Promise<Task[]> }
  remoteApi: { list: () => Promise<{ remotes?: string[] }> }
  scheduleApi: { list: () => Promise<Schedule[]> }
  runApi: { list: (page: number, pageSize: number) => Promise<{ runs?: Run[]; total?: number }> }
  jobApi: { list: () => Promise<any[]> }
}

export function useTaskViewDataSync(options: UseTaskViewDataSyncOptions) {
  let loadSeq = 0
  let activeRunsReloadTimer: number | null = null

  async function loadData() {
    const seq = ++loadSeq
    try {
      const [taskData, remoteData, scheduleData, runResult] = await Promise.all([
        options.taskApi.list(),
        options.remoteApi.list(),
        options.scheduleApi.list(),
        options.runApi.list(options.runsPage.value, options.runsPageSize),
      ])
      if (seq !== loadSeq) return
      if (Array.isArray(taskData)) options.tasks.value = taskData
      if (Array.isArray(remoteData?.remotes) && remoteData.remotes.length > 0) options.remotes.value = remoteData.remotes
      if (Array.isArray(scheduleData) && scheduleData.length > 0) options.schedules.value = scheduleData
      if (runResult?.runs) {
        options.runs.value = runResult.runs
        options.runsTotal.value = typeof runResult.total === 'number' ? runResult.total : (runResult.runs?.length || 0)
      }
      try {
        localStorage.setItem('lastTasksSnapshot', JSON.stringify(options.tasks.value || []))
      } catch {}
    } catch (e) {
      console.error(e)
      if (!options.tasks.value || options.tasks.value.length === 0) {
        try {
          const snap = JSON.parse(localStorage.getItem('lastTasksSnapshot') || '[]')
          if (Array.isArray(snap)) options.tasks.value = snap
        } catch {}
      }
    }
  }

  async function loadActiveRuns() {
    try {
      const data = await options.jobApi.list()
      const now = Date.now()
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
        if (tid) {
          const prevTotals = options.lastNonDecreasingTotalsByTask.value[tid]
          if (prevTotals) {
            if (prevTotals.totalBytes > 0 && raw.totalBytes > 0 && raw.totalBytes < prevTotals.totalBytes) {
              raw.totalBytes = prevTotals.totalBytes
            }
            if (prevTotals.totalCount > 0 && raw.totalCount > 0 && raw.totalCount < prevTotals.totalCount) {
              raw.totalCount = prevTotals.totalCount
            }
          }
          const nextTotals = {
            totalBytes: Math.max(prevTotals?.totalBytes || 0, raw.totalBytes || 0),
            totalCount: Math.max(prevTotals?.totalCount || 0, raw.totalCount || 0),
          }
          options.lastNonDecreasingTotalsByTask.value[tid] = nextTotals
        }
        it.progress = raw
        if (!it.stableProgress) it.stableProgress = raw
        if (tid) options.lastStableByTask.value[tid] = { sp: raw, at: now }
        return it
      })
      if (list.length === 0) {
        options.activeRuns.value = []
        return
      }
      options.activeRuns.value = list
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

  function setupRealtimeSync() {
    const wsClient = useWebSocket({
      onMessage: (msg) => {
        if (msg.type === 'run_status' && msg.data) {
          const idx = options.runs.value.findIndex(r => r.id === msg.data.run_id)
          if (idx !== -1) {
            options.runs.value[idx] = { ...options.runs.value[idx], status: msg.data.status }
          }
          scheduleActiveRunsReload(0)
        } else if (msg.type === 'run_progress' && msg.data) {
          const idx = options.activeRuns.value.findIndex(r => r.runRecord?.id === msg.data.run_id)
          if (idx !== -1) {
            const cur = options.activeRuns.value[idx] || {}
            const prev = cur.progress || cur.stableProgress || {}
            const tid = cur.runRecord?.taskId
            const incomingTotalBytes = Number(msg.data.total || prev.totalBytes || 0)
            const incomingTotalCount = Number(msg.data.totalCount || msg.data.plannedFiles || prev.totalCount || 0)
            const prevTotals = tid ? options.lastNonDecreasingTotalsByTask.value[tid] : undefined
            const nextTotalBytes = Math.max(prevTotals?.totalBytes || 0, incomingTotalBytes || 0)
            const nextTotalCount = Math.max(prevTotals?.totalCount || 0, incomingTotalCount || 0)
            const nextProgress = {
              ...prev,
              bytes: Number(msg.data.bytes || 0),
              totalBytes: nextTotalBytes,
              speed: Number(msg.data.speed || 0),
              percentage: Number(msg.data.percent || prev.percentage || 0),
              completedFiles: Number(msg.data.completedFiles || prev.completedFiles || 0),
              totalCount: nextTotalCount,
              eta: Number(msg.data.eta || prev.eta || 0),
            }
            if (tid) {
              options.lastNonDecreasingTotalsByTask.value[tid] = {
                totalBytes: nextTotalBytes,
                totalCount: nextTotalCount,
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
          const runIdx = options.runs.value.findIndex(r => r.id === msg.data.run_id)
          if (runIdx !== -1) {
            const curRun = options.runs.value[runIdx] || {}
            const sum = typeof curRun.summary === 'string'
              ? (() => { try { return JSON.parse(curRun.summary) } catch { return {} } })()
              : (curRun.summary || {})
            const prevProgress = sum.progress || {}
            sum.progress = {
              ...prevProgress,
              bytes: Number(msg.data.bytes || 0),
              totalBytes: Number(msg.data.total || prevProgress.totalBytes || 0),
              speed: Number(msg.data.speed || 0),
              percentage: Number(msg.data.percent || prevProgress.percentage || 0),
              completedFiles: Number(msg.data.completedFiles || prevProgress.completedFiles || 0),
              plannedFiles: Number(msg.data.totalCount || msg.data.plannedFiles || prevProgress.plannedFiles || 0),
              eta: Number(msg.data.eta || prevProgress.eta || 0),
            }
            options.runs.value[runIdx] = {
              ...curRun,
              summary: sum,
            }
          }
        }
      }
    })
    wsClient.connect()

    onWsMessage('run_status', () => {
      Promise.all([
        loadActiveRuns().catch(console.error),
        loadData().catch(console.error),
      ]).catch(console.error)
    })
  }

  return {
    loadData,
    loadActiveRuns,
    loadGlobalStats,
    openGlobalStats,
    setupRealtimeSync,
  }
}

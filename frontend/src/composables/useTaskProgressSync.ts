import { computed, type Ref } from 'vue'
import type { ActiveRunProgress } from '../api/run'
import type { Run, RunSummaryPayload } from '../types'

const TIME_CONSTANTS = {
  FINISH_WINDOW_MS: 15000,
  AUTO_REFRESH_DELAY_MS: 20000,
  POST_REFRESH_DELAY_MS: 1000,
  LOCK_RELEASE_DELAY_MS: 5000,
}

const BYTE_UNITS = ['B', 'KB', 'MB', 'GB', 'TB']

export interface UseTaskProgressSyncOptions {
  runs: Ref<Run[]>
  activeRuns: Ref<any[]>
  activeRunLookup: { getActiveRunByTaskId: (taskId: number) => any }
  loadData: () => Promise<void>
  loadActiveRuns: () => Promise<void>
}

export interface UseTaskProgressSyncReturn {
  getRunProgressFromSummary: (run: Run) => ActiveRunProgress | null
  getRealtimeProgressByRun: (run: any) => ActiveRunProgress | null
  getRunningProgressByRun: (run: Run) => ActiveRunProgress | null
  getTaskCardProgressByTask: (taskId: number) => (ActiveRunProgress & { __frozenAt?: number }) | null
  getRunningProgressByTask: (taskId: number) => ActiveRunProgress | null
  formatBps: (bps: number) => string
  calcEtaFromAvg: (run: any, live: ActiveRunProgress | null) => number | null
  triggerAutoRefresh: (taskId: number) => Promise<string>
}

export function useTaskProgressSync(options: UseTaskProgressSyncOptions): UseTaskProgressSyncReturn {
  const lastDbFrameByRunId: Record<number, ActiveRunProgress> = {}
  const lastNonZeroSpeedByTask: Record<number, number> = {}
  const completedFreezeByTask: Record<number, ActiveRunProgress & { __frozenAt: number }> = {}
  const refreshLocks: Record<number, boolean> = {}

  const runningRunByTaskId = computed(() => {
    const index = new Map<number, Run>()
    for (const item of options.runs.value || []) {
      const candidateTaskId = extractTaskId(item)
      if (candidateTaskId > 0 && item?.status === 'running') {
        index.set(candidateTaskId, item)
      }
    }
    return index
  })

  function extractTaskId(item: any): number {
    return Number(item?.taskId ?? item?.taskID ?? item?.task_id ?? 0)
  }

  function normalizeSummaryProgress(p: unknown): ActiveRunProgress | null {
    if (!p || typeof p !== 'object') return null
    
    const raw = p as Record<string, unknown>
    const bytes = Number(raw.bytes || 0)
    const totalBytes = Number(raw.totalBytes || 0)
    const speed = Number(raw.speed || 0)
    const eta = Number(raw.eta || 0)
    const plannedFiles = Number(raw.plannedFiles || 0)
    const logicalTotalCount = Number(raw.logicalTotalCount || raw.totalCount || plannedFiles || 0)
    const totalCount = logicalTotalCount
    
    let percentage = Number(raw.percentage || 0)
    if ((!percentage || Number.isNaN(percentage)) && totalBytes > 0) {
      percentage = (bytes / totalBytes) * 100
    }
    
    const completedFiles = Number(raw.completedFiles || 0)
    
    return {
      bytes,
      totalBytes,
      speed,
      eta,
      totalCount,
      percentage,
      completedFiles,
      phase: typeof raw.phase === 'string' ? raw.phase : undefined,
      lastUpdatedAt: typeof raw.lastUpdatedAt === 'string' ? raw.lastUpdatedAt : undefined,
    }
  }

  function freezeCompletedProgress(p: ActiveRunProgress | null): ActiveRunProgress | null {
    if (!p) return null
    
    const frozen = { ...p }
    const percentage = Number(frozen.percentage || 0)
    
    if (percentage >= 99.999) {
      frozen.percentage = 100
      const totalBytes = Number(frozen.totalBytes || 0)
      if (totalBytes > 0) frozen.bytes = totalBytes
      const totalCount = Number(frozen.totalCount || 0)
      if (totalCount > 0) frozen.completedFiles = totalCount
      frozen.speed = 0
      frozen.eta = 0
      frozen.phase = 'completed'
    }
    
    return frozen
  }

  function getLiveSummaryFromDB(run: Run): ActiveRunProgress | null {
    try {
      const sum = typeof run?.summary === 'string'
        ? JSON.parse(run.summary)
        : (run?.summary as RunSummaryPayload | undefined)
      return normalizeSummaryProgress(sum?.progress)
    } catch (e) {
      console.warn('[useTaskProgressSync] Failed to parse summary:', e)
      return null
    }
  }

  function getRunProgressFromSummary(run: Run): ActiveRunProgress | null {
    const db = getLiveSummaryFromDB(run)
    const id = run?.id
    
    if (db && id) {
      lastDbFrameByRunId[id] = db
      return db
    }
    
    if (id && lastDbFrameByRunId[id]) {
      return lastDbFrameByRunId[id]
    }
    
    return db || null
  }

  function getRealtimeProgressByRun(run: any): ActiveRunProgress | null {
    try {
      const tid = extractTaskId(run)
      if (tid > 0) {
        const active = options.activeRunLookup.getActiveRunByTaskId(tid)
        if (active?.progress) {
          return active.progress
        }
      }
    } catch (e) {
      console.warn('[useTaskProgressSync] Failed to get realtime progress:', e)
    }
    return getRunProgressFromSummary(run)
  }

  function getRunningProgressByRun(run: Run): ActiveRunProgress | null {
    return getRealtimeProgressByRun(run)
  }

  function getTaskCardProgressByTask(taskId: number): (ActiveRunProgress & { __frozenAt?: number }) | null {
    const active = options.activeRunLookup.getActiveRunByTaskId(taskId)
    
    if (active?.progress) {
      const normalizedActive = normalizeSummaryProgress(active.progress)
      const frozenActive = freezeCompletedProgress(normalizedActive || active.progress)
      
      if (frozenActive && Number(frozenActive.percentage || 0) >= 99.999) {
        if (!completedFreezeByTask[taskId]) {
          completedFreezeByTask[taskId] = {
            ...frozenActive,
            __frozenAt: Date.now(),
          }
        }
        return completedFreezeByTask[taskId]
      }
      
      delete completedFreezeByTask[taskId]
      return frozenActive
    }

    const frozen = completedFreezeByTask[taskId]
    if (frozen) {
      if (!runningRunByTaskId.value.has(Number(taskId))) {
        delete completedFreezeByTask[taskId]
        return null
      }
      
      const frozenAt = Number(frozen.__frozenAt || 0)
      if (frozenAt > 0 && Date.now() - frozenAt <= TIME_CONSTANTS.FINISH_WINDOW_MS) {
        return frozen
      }
      
      delete completedFreezeByTask[taskId]
    }

    const running = runningRunByTaskId.value.get(Number(taskId))
    if (!running) return null
    
    return getRunProgressFromSummary(running)
  }

  function getRunningProgressByTask(taskId: number): ActiveRunProgress | null {
    const raw = getTaskCardProgressByTask(taskId)
    if (!raw) return null
    
    const st: ActiveRunProgress = { ...raw }
    st.bytes = Number(st.bytes || 0)
    st.totalBytes = Number(st.totalBytes || 0)
    st.speed = Number(st.speed || 0)
    st.percentage = Number(st.percentage || 0)
    st.completedFiles = Number(st.completedFiles || 0)
    st.totalCount = Number(st.totalCount || 0)
    st.eta = Number(st.eta || 0)
    
    if (st.percentage < 0) st.percentage = 0
    if (st.percentage > 100) st.percentage = 100
    
    return st
  }

  function formatBytes(bytes: number): string {
    if (!bytes || bytes <= 0) return '-'
    
    let size = bytes
    let idx = 0
    
    while (size >= 1024 && idx < BYTE_UNITS.length - 1) {
      size /= 1024
      idx++
    }
    
    const decimals = size >= 10 || idx === 0 ? 0 : 1
    return `${size.toFixed(decimals)} ${BYTE_UNITS[idx]}`
  }

  function formatBytesPerSec(n: number): string {
    const formatted = formatBytes(n)
    return formatted === '-' ? '-' : `${formatted}/s`
  }

  function formatBps(bps: number): string {
    if (!bps || bps <= 0) return '-'
    return `${formatBytes(bps)}/s`
  }

  function calcEtaFromAvg(run: any, live: ActiveRunProgress | null): number | null {
    try {
      if (!run?.startedAt || !live) return null
      
      const tid = extractTaskId(run)
      const total = Number(live.totalBytes || 0)
      if (!total) return null
      
      const bytes = Number(live.bytes || 0)
      if (bytes <= 0) return null
      
      const remaining = Math.max(0, total - bytes)
      const speed = Number(live.speed || 0)
      
      if (formatBytesPerSec(speed) === '-') return null
      
      if (tid && speed > 0) {
        lastNonZeroSpeedByTask[tid] = speed
      }
      
      const sp = tid ? (lastNonZeroSpeedByTask[tid] || 0) : speed
      if (!sp || sp <= 0) return null
      
      const etaSec = Math.floor(remaining / sp)
      if (etaSec > 99 * 3600) return null
      
      return etaSec
    } catch {
      return null
    }
  }

  async function triggerAutoRefresh(taskId: number): Promise<string> {
    if (refreshLocks[taskId]) return ''
    
    refreshLocks[taskId] = true
    
    try {
      await new Promise(r => setTimeout(r, TIME_CONSTANTS.AUTO_REFRESH_DELAY_MS))
      await Promise.all([options.loadActiveRuns(), options.loadData()])
      await new Promise(r => setTimeout(r, TIME_CONSTANTS.POST_REFRESH_DELAY_MS))
      await Promise.all([options.loadActiveRuns(), options.loadData()])
    } finally {
      setTimeout(() => {
        delete refreshLocks[taskId]
      }, TIME_CONSTANTS.LOCK_RELEASE_DELAY_MS)
    }
    
    return ''
  }

  return {
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

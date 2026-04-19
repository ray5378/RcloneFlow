import type { Ref } from 'vue'

export function useTaskProgressSync(options: {
  activeRuns: Ref<any[]>
  activeRunLookup: { getActiveRunByTaskId: (taskId: number) => any }
  lastStableByTask: Ref<Record<number, { sp: any; at: number }>>
  loadData: () => Promise<void>
  loadActiveRuns: () => Promise<void>
  lingerMs: number
}) {
  const lastDbFrameByRunId: Record<number, any> = {}
  const lastNonZeroSpeedByTask: Record<number, number> = {}
  const refreshLocks: Record<number, boolean> = {}

  function getDbProgressStable(run: any) {
    const db = getLiveSummaryFromDB(run)
    const id = run?.id
    if (db && id) {
      lastDbFrameByRunId[id] = db
      return db
    }
    if (id && lastDbFrameByRunId[id]) return lastDbFrameByRunId[id]
    return db || null
  }

  function getDeNoisedStableByRun(run: any) {
    try {
      const tid = run?.taskId as number
      if (tid) {
        const active = options.activeRunLookup.getActiveRunByTaskId(tid)
        if (active?.progress) return active.progress
        if (active?.stableProgress) return active.stableProgress
      }
    } catch {}
    return getDbProgressStable(run)
  }

  function getDeNoisedStableByTask(taskId: number) {
    const active = options.activeRunLookup.getActiveRunByTaskId(taskId)
    const raw = active?.progress || active?.stableProgress
    if (!raw) return null
    const st: any = { ...raw }
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

  function formatBps(bps: number) {
    if (!bps || bps <= 0) return '-'
    return formatBytes(bps) + '/s'
  }

  function getLiveSummaryFromDB(run: any) {
    try {
      const sum = typeof run?.summary === 'string' ? JSON.parse(run.summary) : run?.summary
      const p = sum?.progress
      if (p && typeof p === 'object') {
        const bytes = Number(p.bytes || 0)
        const totalBytes = Number(p.totalBytes || 0)
        const speed = Number(p.speed || 0)
        const eta = Number(p.eta || 0)
        const totalCount = Number(p.plannedFiles || 0)
        let percentage = Number(p.percentage || 0)
        if ((!percentage || Number.isNaN(percentage)) && totalBytes > 0) percentage = (bytes / totalBytes) * 100
        const completedFiles = Number(p.completedFiles || 0)
        return { bytes, totalBytes, speed, eta, totalCount, percentage, completedFiles }
      }
      const sp = sum?.stableProgress
      if (sp && typeof sp === 'object') {
        const bytes = Number(sp.bytes || 0)
        const totalBytes = Number(sp.totalBytes || 0)
        const speed = Number(sp.speed || 0)
        const eta = Number(sp.eta || 0)
        const totalCount = Number(sp.totalCount || 0)
        let percentage = Number(sp.percentage || 0)
        if ((!percentage || Number.isNaN(percentage)) && totalBytes > 0) percentage = (bytes / totalBytes) * 100
        const completedFiles = Number(sp.completedFiles || 0)
        return { bytes, totalBytes, speed, eta, totalCount, percentage, completedFiles }
      }
    } catch {}
    return null
  }

  function calcEtaFromAvg(run: any, live: any) {
    try {
      if (!run?.startedAt || !live) return null
      const tid = (run.taskId || run.taskID || run.task_id || run.runRecord?.taskId) as number
      const total = Number(live.totalBytes || 0)
      if (!total) return null
      const bytes = Number(live.bytes || 0)
      if (bytes <= 0) return null
      const remaining = Math.max(0, total - bytes)
      let speed = Number(live.speed || 0)
      if (formatBytesPerSec(speed) === '-') return null
      if (tid && speed > 0) lastNonZeroSpeedByTask[tid] = speed
      const sp = tid ? (lastNonZeroSpeedByTask[tid] || 0) : speed
      if (!sp || sp <= 0) return null
      const etaSec = Math.floor(remaining / sp)
      if (etaSec > 99 * 3600) return null
      return etaSec
    } catch {
      return null
    }
  }

  async function triggerAutoRefresh(taskId: number) {
    if (refreshLocks[taskId]) return ''
    refreshLocks[taskId] = true
    try {
      await new Promise(r => setTimeout(r, 20000))
      await Promise.all([options.loadActiveRuns(), options.loadData()])
      await new Promise(r => setTimeout(r, 1000))
      await Promise.all([options.loadActiveRuns(), options.loadData()])
      const st = options.lastStableByTask.value?.[taskId]
      if (st && Date.now() - st.at > options.lingerMs) {
        delete options.lastStableByTask.value[taskId]
      }
    } finally {
      setTimeout(() => {
        delete refreshLocks[taskId]
      }, 5000)
    }
    return ''
  }

  return {
    getDbProgressStable,
    getDeNoisedStableByRun,
    getDeNoisedStableByTask,
    formatBps,
    calcEtaFromAvg,
    triggerAutoRefresh,
  }
}

function formatBytesPerSec(n: number) {
  return formatBytes(n) === '-' ? '-' : `${formatBytes(n)}/s`
}

function formatBytes(bytes: number) {
  if (!bytes || bytes <= 0) return '-'
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  let size = bytes
  let idx = 0
  while (size >= 1024 && idx < units.length - 1) {
    size /= 1024
    idx++
  }
  return `${size.toFixed(size >= 10 || idx === 0 ? 0 : 1)} ${units[idx]}`
}

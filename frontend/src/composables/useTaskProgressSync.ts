import type { Ref } from 'vue'

export function useTaskProgressSync(options: {
  runs: Ref<any[]>
  activeRuns: Ref<any[]>
  activeRunLookup: { getActiveRunByTaskId: (taskId: number) => any }
  loadData: () => Promise<void>
  loadActiveRuns: () => Promise<void>
}) {
  const lastDbFrameByRunId: Record<number, any> = {}
  const lastNonZeroSpeedByTask: Record<number, number> = {}
  // 任务卡片完成态只保留一份冻结帧：
  // 1) active.progress 到 100% 时立即冻结；
  // 2) active 消失后继续沿用这同一帧；
  // 3) 超过完成短窗口后清掉，不再继续显示进度条；
  // 4) 不再引入第二份完成态摘要参与 handoff，避免二次抖动。
  const completedFreezeByTask: Record<number, any> = {}
  const refreshLocks: Record<number, boolean> = {}
  const FINISH_WINDOW_MS = 15000

  function normalizeSummaryProgress(p: any) {
    if (!p || typeof p !== 'object') return null
    const bytes = Number(p.bytes || 0)
    const totalBytes = Number(p.totalBytes || 0)
    const speed = Number(p.speed || 0)
    const eta = Number(p.eta || 0)
    const totalCount = Number(p.totalCount || p.plannedFiles || 0)
    let percentage = Number(p.percentage || 0)
    if ((!percentage || Number.isNaN(percentage)) && totalBytes > 0) percentage = (bytes / totalBytes) * 100
    const completedFiles = Number(p.completedFiles || 0)
    return { bytes, totalBytes, speed, eta, totalCount, percentage, completedFiles, phase: p.phase }
  }

  function freezeCompletedProgress(p: any) {
    if (!p) return null
    const frozen = { ...p }
    if (Number(frozen.percentage || 0) >= 99.999) {
      frozen.percentage = 100
      if (Number(frozen.totalBytes || 0) > 0) frozen.bytes = Number(frozen.totalBytes || 0)
      if (Number(frozen.totalCount || 0) > 0) frozen.completedFiles = Number(frozen.totalCount || 0)
      frozen.speed = 0
      frozen.eta = 0
      frozen.phase = 'completed'
    }
    return frozen
  }

  function getLiveSummaryFromDB(run: any) {
    try {
      const sum = typeof run?.summary === 'string' ? JSON.parse(run.summary) : run?.summary
      return normalizeSummaryProgress(sum?.progress)
    } catch {}
    return null
  }

  function getRunProgressFromSummary(run: any) {
    const db = getLiveSummaryFromDB(run)
    const id = run?.id
    if (db && id) {
      lastDbFrameByRunId[id] = db
      return db
    }
    if (id && lastDbFrameByRunId[id]) return lastDbFrameByRunId[id]
    return db || null
  }

  function getRealtimeProgressByRun(run: any) {
    try {
      const tid = Number(run?.taskId ?? run?.taskID ?? run?.task_id ?? run?.runRecord?.taskId)
      if (tid > 0) {
        const active = options.activeRunLookup.getActiveRunByTaskId(tid)
        if (active?.progress) return active.progress
      }
    } catch {}
    return getRunProgressFromSummary(run)
  }

  function getRunningProgressByRun(run: any) {
    return getRealtimeProgressByRun(run)
  }

  // 任务卡片完成态规则：
  // - 运行中：优先读 active.progress
  // - 完成后短窗口：只复用前端单份冻结帧 completedFreezeByTask
  // - 超过 FINISH_WINDOW_MS：清掉冻结帧，不再继续显示进度条
  // - 不允许再引入第二份完成态摘要 handoff，否则会把刚修好的完成态抖动带回来
  function getTaskCardProgressByTask(taskId: number) {
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

    // 任务卡片不再在 finished 短窗口切到第二份完成态摘要。
    // 这样完成态只保留一份冻结帧，避免 active 消失后再次 handoff 造成二次抖动。
    const frozen = completedFreezeByTask[taskId]
    if (frozen) {
      const frozenAt = Number(frozen.__frozenAt || 0)
      if (frozenAt > 0 && Date.now() - frozenAt <= FINISH_WINDOW_MS) return frozen
      delete completedFreezeByTask[taskId]
    }

    const running = (options.runs.value || []).find((item: any) => {
      const candidateTaskId = Number(item?.taskId ?? item?.taskID ?? item?.task_id)
      return candidateTaskId > 0 && candidateTaskId === Number(taskId) && item?.status === 'running'
    })
    if (!running) return null
    return getRunProgressFromSummary(running)
  }

  function getRunningProgressByTask(taskId: number) {
    const raw = getTaskCardProgressByTask(taskId)
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
    } finally {
      setTimeout(() => {
        delete refreshLocks[taskId]
      }, 5000)
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

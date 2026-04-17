export function getDeNoisedStableByRun(run: any, getActiveRunByTaskId: (taskId: number) => any, getDbProgressStable: (run: any) => any) {
  try {
    const tid = run?.taskId as number
    if (tid) {
      const active = getActiveRunByTaskId(tid)
      if (active?.progress) return active.progress
      if (active?.stableProgress) return active.stableProgress
    }
  } catch {}
  return getDbProgressStable(run)
}

export function getDeNoisedStableByTask(taskId: number, getActiveRunByTaskId: (taskId: number) => any) {
  const active = getActiveRunByTaskId(taskId)
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

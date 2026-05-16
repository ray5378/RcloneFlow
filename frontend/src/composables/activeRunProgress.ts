export function getRunningProgressByRun(run: any, getActiveRunByTaskId: (taskId: number) => any, getRunProgressFromSummary: (run: any) => any) {
  try {
    const tid = run?.taskId as number
    if (tid) {
      const active = getActiveRunByTaskId(tid)
      if (active?.progress) return active.progress
    }
  } catch {}
  return getRunProgressFromSummary(run)
}

export function getRunningProgressByTask(taskId: number, getActiveRunByTaskId: (taskId: number) => any) {
  const active = getActiveRunByTaskId(taskId)
  const raw = active?.progress
  if (!raw) return null
  const st: any = { ...raw }
  st.bytes = Number(st.bytes || 0)
  st.totalBytes = Number(st.totalBytes || 0)
  st.speed = Number(st.speed || 0)
  st.percentage = (st.totalBytes > 0) ? (st.bytes / st.totalBytes) * 100 : Number(raw.percentage || 0)
  st.completedFiles = Number(st.completedFiles || 0)
  st.totalCount = Number(st.logicalTotalCount || st.totalCount || 0)
  st.eta = Number(st.eta || 0)
  if (st.percentage < 0) st.percentage = 0
  if (st.percentage > 100) st.percentage = 100
  return st
}

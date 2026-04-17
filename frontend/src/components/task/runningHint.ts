import { formatBytes, formatBytesPerSec, formatEta } from '../../utils/format'

export function getActiveProgress(active: any) {
  return active?.progress || active?.stableProgress || null
}

export function getActiveProgressText(active: any) {
  const p: any = getActiveProgress(active)
  if (!p) return '-'
  if (p.phase === 'preparing') {
    return `准备中 · 已传 ${formatBytes(p.bytes || 0)} · 速度 ${formatBytesPerSec(p.speed || 0)}`
  }
  let etaStr = ''
  if (Number(p.eta || 0) > 0) etaStr = ` · 预计完成 ${formatEta(Number(p.eta || 0))}`
  return `${Number(p.percentage || 0).toFixed(2)}% · ${formatBytes(Number(p.bytes || 0))} / ${formatBytes(Number(p.totalBytes || 0))} · ${formatBytesPerSec(Number(p.speed || 0))} · 总数量 ${Number(p.totalCount || 0)} ／ 已传输 ${Number(p.completedFiles || 0)}${etaStr}`
}

export function getActiveProgressLine(active: any) {
  return active?.progressLine || '-'
}

export function getActiveProgressCheck(active: any) {
  return active?.progressCheck || null
}

export function getActiveProgressCheckText(active: any) {
  const check: any = getActiveProgressCheck(active)
  if (!check) return '-'
  if (check.ok) return `OK · calcPct=${Number(check.calcPct || 0).toFixed(2)}%`
  const parts: string[] = []
  if (check.pctMismatch) parts.push('百分比异常')
  if (check.countMismatch) parts.push('数量异常')
  if (check.etaMismatch) parts.push('ETA异常')
  return `${parts.join(' / ') || '异常'} · calcPct=${Number(check.calcPct || 0).toFixed(2)}%`
}

export function getActiveProgressJson(active: any) {
  const p = getActiveProgress(active)
  try {
    return p ? JSON.stringify(p, null, 2) : '-'
  } catch {
    return '-'
  }
}

export function getRunningHintDebug(active: any) {
  return {
    checkText: getActiveProgressCheckText(active),
    progressLine: getActiveProgressLine(active),
    progressJson: getActiveProgressJson(active),
  }
}

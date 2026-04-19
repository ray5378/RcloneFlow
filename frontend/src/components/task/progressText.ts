import { formatBytes, formatBytesPerSec, formatEta } from '../../utils/format'

export interface TaskProgressLike {
  percentage?: number
  bytes?: number
  totalBytes?: number
  speed?: number
  eta?: number
  completedFiles?: number
  totalCount?: number
  phase?: string
}

export function getUnifiedProgressText(progress?: TaskProgressLike | null) {
  const p = progress || null
  if (!p) return '-'
  if (p.phase === 'preparing') {
    return `准备中 · 已传 ${formatBytes(p.bytes || 0)} · 速度 ${formatBytesPerSec(p.speed || 0)}`
  }
  let etaStr = ''
  if (p.eta && p.eta > 0) etaStr = ` · 预计完成 ${formatEta(p.eta)}`
  return `${Number(p.percentage || 0).toFixed(2)}% · ${formatBytes(p.bytes || 0)} / ${formatBytes(p.totalBytes || 0)} · ${formatBytesPerSec(p.speed || 0)} · 总数量 ${Number(p.totalCount || 0)} ／ 已传输 ${Number(p.completedFiles || 0)}${etaStr}`
}

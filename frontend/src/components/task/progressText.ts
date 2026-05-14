import { formatBytes, formatBytesPerSec, formatEta } from '../../utils/format'
import { t } from '../../i18n'

export interface TaskProgressLike {
  percentage?: number
  bytes?: number
  totalBytes?: number
  speed?: number
  eta?: number
  completedFiles?: number
  plannedFiles?: number
  logicalTotalCount?: number
  totalCount?: number
  phase?: string
}

export function getResolvedTotalCount(progress?: TaskProgressLike | null) {
  const p = progress || null
  if (!p) return 0
  return Math.max(
    Number((p as any).logicalTotalCount || 0),
    Number(p.totalCount || 0),
    Number((p as any).plannedFiles || 0),
    Number((p as any).completedFiles || 0),
  )
}

export function getUnifiedProgressText(progress?: TaskProgressLike | null) {
  const p = progress || null
  if (!p) return '-'
  const totalTextValue = getResolvedTotalCount(p)
  const totalText = totalTextValue > 0 ? ` · ${t('runtime.totalCount')} ${totalTextValue}` : ''
  if (p.phase === 'preparing') {
    return `${t('runtime.preparing')} · ${t('runtime.transferred')} ${formatBytes(p.bytes || 0)}${totalText} · ${t('runtime.speed')} ${formatBytesPerSec(p.speed || 0)}`
  }
  let etaStr = ''
  if (p.eta && p.eta > 0) etaStr = ` · ${t('runtime.etaDone')} ${formatEta(p.eta)}`
  return `${Number(p.percentage || 0).toFixed(2)}% · ${formatBytes(p.bytes || 0)} / ${formatBytes(p.totalBytes || 0)} · ${formatBytesPerSec(p.speed || 0)}${totalText} · ${t('runtime.completedCount')} ${Number(p.completedFiles || 0)}${etaStr}`
}

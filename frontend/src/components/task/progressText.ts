import { formatBytes, formatBytesPerSec, formatEta } from '../../utils/format'
import { t } from '../../i18n'

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
    return `${t('runtime.preparing')} · ${t('runtime.transferred')} ${formatBytes(p.bytes || 0)} · ${t('runtime.speed')} ${formatBytesPerSec(p.speed || 0)}`
  }
  let etaStr = ''
  if (p.eta && p.eta > 0) etaStr = ` · ${t('runtime.etaDone')} ${formatEta(p.eta)}`
  const totalText = Number(p.totalCount || 0) > 0 ? ` · ${t('runtime.totalCount')} ${Number(p.totalCount || 0)}` : ''
  return `${Number(p.percentage || 0).toFixed(2)}% · ${formatBytes(p.bytes || 0)} / ${formatBytes(p.totalBytes || 0)} · ${formatBytesPerSec(p.speed || 0)}${totalText} · ${t('runtime.completedCount')} ${Number(p.completedFiles || 0)}${etaStr}`
}

import type { ActiveRun, ActiveRunProgress } from '../../api/run'
import { formatBytes, formatBytesPerSec, formatEta } from '../../utils/format'
import { t } from '../../i18n'

export function getActiveProgress(active: ActiveRun | null | undefined): ActiveRunProgress | null {
  return active?.progress || null
}

export function getActiveProgressText(active: any) {
  const p: any = getActiveProgress(active)
  if (!p) return '-'
  const totalCount = Number(p.logicalTotalCount || p.totalCount || p.plannedFiles || 0)
  const totalText = totalCount > 0 ? ` · ${t('runtime.totalCount')} ${totalCount}` : ''
  if (p.phase === 'preparing') {
    return `${t('runtime.preparing')} · ${t('runtime.transferred')} ${formatBytes(p.bytes || 0)}${totalText} · ${t('runtime.speed')} ${formatBytesPerSec(p.speed || 0)}`
  }
  let etaStr = ''
  if (Number(p.eta || 0) > 0) etaStr = ` · ${t('runtime.etaDone')} ${formatEta(Number(p.eta || 0))}`
  return `${Number(p.percentage || 0).toFixed(2)}% · ${formatBytes(Number(p.bytes || 0))} / ${formatBytes(Number(p.totalBytes || 0))} · ${formatBytesPerSec(Number(p.speed || 0))} · ${t('runtime.totalCount')} ${totalCount} ／ ${t('runtime.completedCount')} ${Number(p.completedFiles || 0)}${etaStr}`
}


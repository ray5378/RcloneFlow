import type { ActiveRun, ActiveRunProgress } from '../../api/run'
import { formatBytes, formatBytesPerSec, formatEta } from '../../utils/format'
import { t } from '../../i18n'

export function getActiveProgress(active: ActiveRun | null | undefined): ActiveRunProgress | null {
  return active?.progress || null
}

export function getActiveProgressText(active: any) {
  const p: any = getActiveProgress(active)
  if (!p) return '-'
  if (p.phase === 'preparing') {
    return `${t('runtime.preparing')} · ${t('runtime.transferred')} ${formatBytes(p.bytes || 0)} · ${t('runtime.speed')} ${formatBytesPerSec(p.speed || 0)}`
  }
  let etaStr = ''
  if (Number(p.eta || 0) > 0) etaStr = ` · ${t('runtime.etaDone')} ${formatEta(Number(p.eta || 0))}`
  const totalCount = Number(p.logicalTotalCount || p.totalCount || 0)
  return `${Number(p.percentage || 0).toFixed(2)}% · ${formatBytes(Number(p.bytes || 0))} / ${formatBytes(Number(p.totalBytes || 0))} · ${formatBytesPerSec(Number(p.speed || 0))} · ${t('runtime.totalCount')} ${totalCount} ／ ${t('runtime.completedCount')} ${Number(p.completedFiles || 0)}${etaStr}`
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
  if (check.pctMismatch) parts.push(t('runtime.pctMismatch'))
  if (check.countMismatch) parts.push(t('runtime.countMismatch'))
  if (check.etaMismatch) parts.push(t('runtime.etaMismatch'))
  return `${parts.join(' / ') || t('runtime.abnormal')} · calcPct=${Number(check.calcPct || 0).toFixed(2)}%`
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

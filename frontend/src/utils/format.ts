import { t } from '../i18n'

export function formatBytes(bytes: number): string {
  if (bytes === 0 || bytes === undefined || bytes === null) return '-'
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  let unitIndex = 0
  let size = bytes
  while (size >= 1024 && unitIndex < units.length - 1) {
    size /= 1024
    unitIndex++
  }
  return `${size.toFixed(unitIndex > 0 ? 1 : 0)} ${units[unitIndex]}`
}

export function formatBytesPerSec(bps: number): string {
  if (!bps || bps === 0) return '-'
  return formatBytes(bps) + '/s'
}

export function formatDuration(startTime: string | undefined, endTime: string | undefined): string {
  if (!startTime) return '-'
  const start = new Date(startTime).getTime()
  const end = endTime ? new Date(endTime).getTime() : Date.now()
  const diff = Math.max(0, end - start)
  const seconds = Math.floor(diff / 1000)
  const minutes = Math.floor(seconds / 60)
  const hours = Math.floor(minutes / 60)
  const days = Math.floor(hours / 24)

  if (days > 0) return `${days}${t('runtime.daySuffix')}${hours % 24}${t('runtime.hourSuffix')}`
  if (hours > 0) return `${hours}${t('runtime.hourSuffix')}${minutes % 60}${t('runtime.minuteSuffix')}`
  if (minutes > 0) return `${minutes}${t('runtime.minuteSuffix')}${seconds % 60}${t('runtime.secondSuffix')}`
  return `${seconds}${t('runtime.secondSuffix')}`
}

export function formatEta(seconds: number): string {
  if (!seconds || seconds <= 0) return '-'
  const hours = Math.floor(seconds / 3600)
  const minutes = Math.floor((seconds % 3600) / 60)
  if (hours > 0) return `${t('runtime.approx')}${hours}${t('runtime.hourSuffix')}${minutes}${t('runtime.minuteSuffix')}`
  if (minutes > 0) return `${t('runtime.approx')}${minutes}${t('runtime.minuteSuffix')}`
  return `${t('runtime.approx')}${seconds}${t('runtime.secondSuffix')}`
}

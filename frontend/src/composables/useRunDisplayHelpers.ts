import { locale, t } from '../i18n'

export function useRunDisplayHelpers(options: {
  getFinalSummary: (run: any) => any
}) {
  const durationCache = new Map<number, string>()
  function formatTime(time: string | undefined) {
    if (!time) return '-'
    try {
      return new Date(time).toLocaleString(locale.value === 'zh' ? 'zh-CN' : 'en-US', {
        year: 'numeric', month: '2-digit', day: '2-digit',
        hour: '2-digit', minute: '2-digit'
      })
    } catch {
      return time || '-'
    }
  }

  function formatDuration(startTime: string | undefined, endTime: string | undefined) {
    if (!startTime) return '-'
    const start = new Date(startTime).getTime()
    const end = endTime ? new Date(endTime).getTime() : Date.now()
    const diff = Math.max(0, end - start)
    const seconds = Math.floor(diff / 1000) % 60
    const minutes = Math.floor(diff / 60000) % 60
    const hours = Math.floor(diff / 3600000)
    if (hours > 0) return `${hours}h ${minutes}m ${seconds}s`
    if (minutes > 0) return `${minutes}m ${seconds}s`
    return `${seconds}s`
  }

  function getRunDurationText(run: any) {
    if (run?.id && durationCache.has(run.id)) return durationCache.get(run.id) || '-'
    const fs = options.getFinalSummary(run)
    const text = fs && fs.durationText ? fs.durationText : formatDuration(run.startedAt, run.finishedAt)
    if (run?.id) durationCache.set(run.id, text)
    return text
  }

  function getStatusClass(status: string) {
    switch (status) {
      case 'running': return 'running'
      case 'finished': return 'success'
      case 'failed': return 'failed'
      case 'skipped': return 'skipped'
      default: return ''
    }
  }

  function getStatusText(status: string) {
    switch (status) {
      case 'running': return t('runtime.statusRunning')
      case 'finished': return t('runtime.statusFinished')
      case 'failed': return t('runtime.statusFailed')
      case 'skipped': return t('runtime.statusSkipped')
      default: return status
    }
  }

  return {
    formatTime,
    formatDuration,
    getRunDurationText,
    getStatusClass,
    getStatusText,
  }
}

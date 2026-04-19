export function useRunDisplayHelpers(options: {
  getFinalSummary: (run: any) => any
}) {
  function formatTime(time: string | undefined) {
    if (!time) return '-'
    try {
      return new Date(time).toLocaleString('zh-CN', {
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
    const fs = options.getFinalSummary(run)
    if (fs && fs.durationText) return fs.durationText
    return formatDuration(run.startedAt, run.finishedAt)
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
      case 'running': return '运行中'
      case 'finished': return '已完成'
      case 'failed': return '失败'
      case 'skipped': return '已跳过'
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

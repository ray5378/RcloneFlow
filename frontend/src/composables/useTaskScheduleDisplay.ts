export function useTaskScheduleDisplay() {
  function formatScheduleSpec(spec: string): string {
    if (!spec) return ''
    const parts = spec.split('|')
    if (parts.length !== 5) return spec

    const [minute, hour, day, month, week] = parts
    return `${minute} ${hour} ${day} ${month} ${week}`
  }

  return {
    formatScheduleSpec,
  }
}

import { describe, it, expect } from 'vitest'
import { useTaskScheduleDisplay } from './useTaskScheduleDisplay'

describe('useTaskScheduleDisplay', () => {
  it('formats pipe-delimited cron fragments', () => {
    const api = useTaskScheduleDisplay()
    expect(api.formatScheduleSpec('*/5|*|*|*|1')).toBe('*/5 * * * 1')
  })

  it('returns raw spec when empty or malformed', () => {
    const api = useTaskScheduleDisplay()
    expect(api.formatScheduleSpec('')).toBe('')
    expect(api.formatScheduleSpec('* * * * *')).toBe('* * * * *')
    expect(api.formatScheduleSpec('a|b|c')).toBe('a|b|c')
  })
})

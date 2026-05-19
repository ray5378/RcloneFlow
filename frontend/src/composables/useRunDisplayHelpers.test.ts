import { describe, it, expect, vi } from 'vitest'
import { useRunDisplayHelpers } from './useRunDisplayHelpers'

vi.mock('../i18n', () => ({
  locale: { value: 'en' },
  t: (key: string) => key,
}))

describe('useRunDisplayHelpers', () => {
  const getFinalSummary = () => null

  it('should format time correctly', () => {
    const { formatTime } = useRunDisplayHelpers({ getFinalSummary })
    expect(formatTime(undefined)).toBe('-')
    expect(formatTime('')).toBe('-')
    expect(formatTime('2024-01-15T10:30:00Z')).toContain('2024')
  })

  it('should format duration correctly', () => {
    const { formatDuration } = useRunDisplayHelpers({ getFinalSummary })
    expect(formatDuration(undefined, undefined)).toBe('-')
    expect(formatDuration('2024-01-15T10:00:00Z', '2024-01-15T10:00:45Z')).toBe('45s')
    expect(formatDuration('2024-01-15T10:00:00Z', '2024-01-15T10:05:30Z')).toBe('5m 30s')
    expect(formatDuration('2024-01-15T10:00:00Z', '2024-01-15T12:15:45Z')).toBe('2h 15m 45s')
  })

  it('should use current time when endTime is undefined', () => {
    const { formatDuration } = useRunDisplayHelpers({ getFinalSummary })
    const result = formatDuration('2024-01-15T10:00:00Z', undefined)
    expect(result).not.toBe('-')
  })

  it('should get status class correctly', () => {
    const { getStatusClass } = useRunDisplayHelpers({ getFinalSummary })
    expect(getStatusClass('running')).toBe('running')
    expect(getStatusClass('finished')).toBe('success')
    expect(getStatusClass('failed')).toBe('failed')
    expect(getStatusClass('skipped')).toBe('skipped')
    expect(getStatusClass('unknown')).toBe('')
  })

  it('should get status text correctly', () => {
    const { getStatusText } = useRunDisplayHelpers({ getFinalSummary })
    expect(getStatusText('running')).toBe('runtime.statusRunning')
    expect(getStatusText('finished')).toBe('runtime.statusFinished')
    expect(getStatusText('failed')).toBe('runtime.statusFailed')
    expect(getStatusText('skipped')).toBe('runtime.statusSkipped')
    expect(getStatusText('unknown')).toBe('unknown')
  })

  it('should cache run duration text', () => {
    const { getRunDurationText } = useRunDisplayHelpers({ getFinalSummary })
    const run = { id: 1, startedAt: '2024-01-15T10:00:00Z', finishedAt: '2024-01-15T10:00:30Z' }
    const first = getRunDurationText(run)
    const second = getRunDurationText(run)
    expect(first).toBe(second)
  })

  it('should use finalSummary durationText when available', () => {
    const getFinalSummaryWithDuration = () => ({ durationText: 'cached duration' })
    const { getRunDurationText } = useRunDisplayHelpers({ getFinalSummary: getFinalSummaryWithDuration })
    const run = { id: 2, startedAt: '2024-01-15T10:00:00Z' }
    expect(getRunDurationText(run)).toBe('cached duration')
  })
})

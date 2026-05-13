import { describe, it, expect, vi, beforeEach } from 'vitest'
import { useRunDisplayHelpers } from './useRunDisplayHelpers'

vi.mock('../i18n', () => ({
  locale: { value: 'zh' },
  t: (key: string) => ({
    'runtime.statusRunning': '运行中',
    'runtime.statusFinished': '已完成',
    'runtime.statusFailed': '失败',
    'runtime.statusSkipped': '已跳过',
  }[key] || key),
}))

describe('useRunDisplayHelpers', () => {
  beforeEach(() => {
    vi.useRealTimers()
  })

  it('formats time and duration', () => {
    const api = useRunDisplayHelpers({ getFinalSummary: () => null })
    expect(api.formatTime(undefined)).toBe('-')
    expect(api.formatTime('2024-01-01T00:00:00Z')).toMatch(/2024/)
    expect(api.formatDuration(undefined, undefined)).toBe('-')
    expect(api.formatDuration('2024-01-01T00:00:00Z', '2024-01-01T00:01:05Z')).toBe('1m 5s')
    expect(api.formatDuration('2024-01-01T00:00:00Z', '2024-01-01T02:03:04Z')).toBe('2h 3m 4s')
  })

  it('uses final summary durationText and caches by run id', () => {
    const getFinalSummary = vi.fn((run: any) => run.summary)
    const api = useRunDisplayHelpers({ getFinalSummary })
    const run = { id: 9, summary: { durationText: '已跑 9 分' }, startedAt: '2024-01-01T00:00:00Z', finishedAt: '2024-01-01T00:10:00Z' }

    expect(api.getRunDurationText(run)).toBe('已跑 9 分')
    run.summary = { durationText: '应被缓存屏蔽' }
    expect(api.getRunDurationText(run)).toBe('已跑 9 分')
  })

  it('falls back to computed duration when final summary lacks durationText', () => {
    vi.useFakeTimers()
    vi.setSystemTime(new Date('2024-01-01T00:00:30Z'))
    const api = useRunDisplayHelpers({ getFinalSummary: () => ({}) })

    expect(api.getRunDurationText({ startedAt: '2024-01-01T00:00:00Z' })).toBe('30s')
  })

  it('maps status class and text', () => {
    const api = useRunDisplayHelpers({ getFinalSummary: () => null })
    expect(api.getStatusClass('running')).toBe('running')
    expect(api.getStatusClass('finished')).toBe('success')
    expect(api.getStatusClass('failed')).toBe('failed')
    expect(api.getStatusClass('skipped')).toBe('skipped')
    expect(api.getStatusClass('mystery')).toBe('')

    expect(api.getStatusText('running')).toBe('运行中')
    expect(api.getStatusText('finished')).toBe('已完成')
    expect(api.getStatusText('failed')).toBe('失败')
    expect(api.getStatusText('skipped')).toBe('已跳过')
    expect(api.getStatusText('mystery')).toBe('mystery')
  })
})

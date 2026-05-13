import { describe, it, expect, vi } from 'vitest'
import { getUnifiedProgressText } from './progressText'

vi.mock('../../i18n', () => ({
  t: (key: string) => ({
    'runtime.totalCount': '总数量',
    'runtime.preparing': '准备中',
    'runtime.transferred': '已传',
    'runtime.speed': '速度',
    'runtime.completedCount': '已传输',
    'runtime.etaDone': '预计还需',
    'runtime.approx': '约',
    'runtime.daySuffix': '天',
    'runtime.hourSuffix': '小时',
    'runtime.minuteSuffix': '分',
    'runtime.secondSuffix': '秒',
  }[key] || key),
}))

describe('progressText', () => {
  it('returns - when progress is empty', () => {
    expect(getUnifiedProgressText(null)).toBe('-')
    expect(getUnifiedProgressText(undefined)).toBe('-')
  })

  it('renders preparing phase without percentage tail', () => {
    const text = getUnifiedProgressText({
      phase: 'preparing',
      bytes: 1024,
      speed: 2048,
      logicalTotalCount: 5,
    })

    expect(text).toContain('准备中')
    expect(text).toContain('已传 1.0 KB')
    expect(text).toContain('总数量 5')
    expect(text).toContain('速度 2.0 KB/s')
    expect(text).not.toContain('预计还需')
  })

  it('renders active transfer progress with eta and completed count', () => {
    const text = getUnifiedProgressText({
      percentage: 12.345,
      bytes: 1024,
      totalBytes: 2048,
      speed: 1024,
      eta: 120,
      completedFiles: 3,
      totalCount: 7,
    })

    expect(text).toContain('12.35%')
    expect(text).toContain('1.0 KB / 2.0 KB')
    expect(text).toContain('1.0 KB/s')
    expect(text).toContain('总数量 7')
    expect(text).toContain('已传输 3')
    expect(text).toContain('预计还需 约2分')
  })

  it('prefers logicalTotalCount then totalCount then plannedFiles', () => {
    const logical = getUnifiedProgressText({ bytes: 0, totalBytes: 0, speed: 0, completedFiles: 0, logicalTotalCount: 9, totalCount: 8, plannedFiles: 7 })
    const total = getUnifiedProgressText({ bytes: 0, totalBytes: 0, speed: 0, completedFiles: 0, totalCount: 8, plannedFiles: 7 })
    const planned = getUnifiedProgressText({ bytes: 0, totalBytes: 0, speed: 0, completedFiles: 0, plannedFiles: 7 })

    expect(logical).toContain('总数量 9')
    expect(total).toContain('总数量 8')
    expect(planned).toContain('总数量 7')
  })
})

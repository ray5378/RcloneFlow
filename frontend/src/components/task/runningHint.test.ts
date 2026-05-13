import { describe, it, expect, vi } from 'vitest'
import {
  getActiveProgress,
  getActiveProgressText,
  getActiveProgressLine,
  getActiveProgressCheck,
  getActiveProgressCheckText,
  getActiveProgressJson,
  getRunningHintDebug,
} from './runningHint'

vi.mock('../../i18n', () => ({
  t: (key: string) => ({
    'runtime.totalCount': '总数量',
    'runtime.preparing': '准备中',
    'runtime.transferred': '已传',
    'runtime.speed': '速度',
    'runtime.completedCount': '已传输',
    'runtime.etaDone': '预计还需',
    'runtime.pctMismatch': '百分比异常',
    'runtime.countMismatch': '数量异常',
    'runtime.etaMismatch': 'ETA异常',
    'runtime.abnormal': '异常',
    'runtime.approx': '约',
    'runtime.daySuffix': '天',
    'runtime.hourSuffix': '小时',
    'runtime.minuteSuffix': '分',
    'runtime.secondSuffix': '秒',
  }[key] || key),
}))

describe('runningHint helpers', () => {
  it('reads progress safely', () => {
    expect(getActiveProgress(undefined)).toBeNull()
    expect(getActiveProgress({ progress: { percentage: 1 } } as any)).toEqual({ percentage: 1 })
  })

  it('renders preparing progress text', () => {
    const text = getActiveProgressText({
      progress: { phase: 'preparing', bytes: 1024, speed: 2048, totalCount: 2 },
    })
    expect(text).toContain('准备中')
    expect(text).toContain('已传 1.0 KB')
    expect(text).toContain('总数量 2')
    expect(text).toContain('速度 2.0 KB/s')
  })

  it('renders normal progress text with eta', () => {
    const text = getActiveProgressText({
      progress: { percentage: 50, bytes: 1024, totalBytes: 2048, speed: 1024, logicalTotalCount: 4, completedFiles: 1, eta: 30 },
    })
    expect(text).toContain('50.00%')
    expect(text).toContain('1.0 KB / 2.0 KB')
    expect(text).toContain('总数量 4')
    expect(text).toContain('已传输 1')
    expect(text).toContain('预计还需 约30秒')
  })

  it('returns fallback line/check/json values', () => {
    expect(getActiveProgressLine({})).toBe('-')
    expect(getActiveProgressCheck({})).toBeNull()
    expect(getActiveProgressJson({})).toBe('-')
  })

  it('renders ok check text', () => {
    const text = getActiveProgressCheckText({ progressCheck: { ok: true, calcPct: 12.345 } })
    expect(text).toBe('OK · calcPct=12.35%')
  })

  it('renders mismatch check text', () => {
    const text = getActiveProgressCheckText({ progressCheck: { ok: false, pctMismatch: true, countMismatch: true, etaMismatch: false, calcPct: 88.888 } })
    expect(text).toContain('百分比异常')
    expect(text).toContain('数量异常')
    expect(text).toContain('calcPct=88.89%')
  })

  it('returns debug bundle', () => {
    const active = {
      progress: { percentage: 1 },
      progressLine: 'INFO : line',
      progressCheck: { ok: true, calcPct: 1 },
    }
    const debug = getRunningHintDebug(active)
    expect(debug.progressLine).toBe('INFO : line')
    expect(debug.checkText).toContain('OK')
    expect(debug.progressJson).toContain('percentage')
  })
})

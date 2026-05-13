import { describe, it, expect } from 'vitest'
import { getRunningProgressByRun, getRunningProgressByTask } from './activeRunProgress'

describe('activeRunProgress', () => {
  it('prefers live active run progress by run task id', () => {
    const result = getRunningProgressByRun(
      { taskId: 7 },
      (taskId: number) => taskId === 7 ? { progress: { percentage: 50 } } : null,
      () => ({ percentage: 10 })
    )
    expect(result).toEqual({ percentage: 50 })
  })

  it('falls back to summary progress when lookup fails or throws', () => {
    expect(getRunningProgressByRun({ taskId: 7 }, () => null, () => ({ percentage: 10 }))).toEqual({ percentage: 10 })
    expect(getRunningProgressByRun({ taskId: 7 }, () => { throw new Error('boom') }, () => ({ percentage: 11 }))).toEqual({ percentage: 11 })
  })

  it('normalizes active task progress fields and clamps percentage', () => {
    const result = getRunningProgressByTask(8, () => ({
      progress: {
        bytes: '12',
        totalBytes: '34',
        speed: '56',
        percentage: '123',
        completedFiles: '7',
        logicalTotalCount: '9',
        eta: '10',
      },
    }))

    expect(result).toEqual({
      bytes: 12,
      totalBytes: 34,
      speed: 56,
      percentage: 100,
      completedFiles: 7,
      logicalTotalCount: '9',
      totalCount: 9,
      eta: 10,
    })
  })

  it('returns null when no active progress exists and clamps low percentage', () => {
    expect(getRunningProgressByTask(1, () => null)).toBeNull()
    expect(getRunningProgressByTask(2, () => ({ progress: { percentage: -5 } }))?.percentage).toBe(0)
  })
})

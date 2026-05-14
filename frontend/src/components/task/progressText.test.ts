import { describe, it, expect } from 'vitest'
import { getUnifiedProgressText, getResolvedTotalCount } from './progressText'

describe('progressText', () => {
  it('keeps total count when the total is exactly 1', () => {
    const progress = {
      percentage: 100,
      bytes: 1024,
      totalBytes: 1024,
      speed: 0,
      completedFiles: 1,
      logicalTotalCount: 1,
      totalCount: 1,
      plannedFiles: 1,
    }

    const text = getUnifiedProgressText(progress)
    expect(getResolvedTotalCount(progress)).toBe(1)
    expect(text).toContain('总数量 1')
    expect(text).toContain('1')
  })
})

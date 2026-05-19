import { describe, it, expect } from 'vitest'
import { useTaskScheduleDisplay } from './useTaskScheduleDisplay'

describe('useTaskScheduleDisplay', () => {
  const { formatScheduleSpec } = useTaskScheduleDisplay()

  it('should handle empty spec', () => {
    expect(formatScheduleSpec('')).toBe('')
  })

  it('should return original spec if not 5 parts', () => {
    expect(formatScheduleSpec('invalid')).toBe('invalid')
    expect(formatScheduleSpec('a|b')).toBe('a|b')
  })

  it('should convert pipe-separated to space-separated', () => {
    const result = formatScheduleSpec('0|12|*|*|1-5')
    expect(result).toBe('0 12 * * 1-5')
  })
})

import { describe, it, expect } from 'vitest'
import { getRunningProgressByRun, getRunningProgressByTask } from './activeRunProgress'

describe('activeRunProgress', () => {
  describe('getRunningProgressByRun', () => {
    it('should return active progress when available', () => {
      const getActiveRunByTaskId = () => ({ progress: { bytes: 100 } })
      const getRunProgressFromSummary = () => ({ bytes: 0 })
      const result = getRunningProgressByRun({ taskId: 1 }, getActiveRunByTaskId, getRunProgressFromSummary)
      expect(result.bytes).toBe(100)
    })

    it('should fallback to summary progress', () => {
      const getActiveRunByTaskId = () => null
      const getRunProgressFromSummary = () => ({ bytes: 50 })
      const result = getRunningProgressByRun({ taskId: 1 }, getActiveRunByTaskId, getRunProgressFromSummary)
      expect(result.bytes).toBe(50)
    })

    it('should handle missing taskId', () => {
      const getActiveRunByTaskId = () => ({ progress: { bytes: 100 } })
      const getRunProgressFromSummary = () => ({ bytes: 50 })
      const result = getRunningProgressByRun({}, getActiveRunByTaskId, getRunProgressFromSummary)
      expect(result.bytes).toBe(50)
    })
  })

  describe('getRunningProgressByTask', () => {
    it('should return normalized progress', () => {
      const getActiveRunByTaskId = () => ({
        progress: {
          bytes: '100',
          totalBytes: '500',
          speed: '10',
          percentage: '20',
          completedFiles: '5',
          totalCount: '10',
          eta: '60',
        },
      })
      const result = getRunningProgressByTask(1, getActiveRunByTaskId)
      expect(result?.bytes).toBe(100)
      expect(result?.totalBytes).toBe(500)
      expect(result?.speed).toBe(10)
      expect(result?.percentage).toBe(20)
      expect(result?.completedFiles).toBe(5)
      expect(result?.totalCount).toBe(10)
      expect(result?.eta).toBe(60)
    })

    it('should clamp percentage to 0-100', () => {
      const getActiveRunByTaskId = () => ({ progress: { percentage: -5 } })
      const result = getRunningProgressByTask(1, getActiveRunByTaskId)
      expect(result?.percentage).toBe(0)

      const getActiveRunByTaskId2 = () => ({ progress: { percentage: 150 } })
      const result2 = getRunningProgressByTask(1, getActiveRunByTaskId2)
      expect(result2?.percentage).toBe(100)
    })

    it('should return null when no active run', () => {
      const getActiveRunByTaskId = () => null
      const result = getRunningProgressByTask(1, getActiveRunByTaskId)
      expect(result).toBeNull()
    })

    it('should return null when no progress', () => {
      const getActiveRunByTaskId = () => ({})
      const result = getRunningProgressByTask(1, getActiveRunByTaskId)
      expect(result).toBeNull()
    })
  })
})

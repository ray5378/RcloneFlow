import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { getVersion } from './version'

const mockFetch = vi.fn()
globalThis.fetch = mockFetch as any

describe('version.ts', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    localStorage.setItem('authToken', 'test-token')
  })

  afterEach(() => {
    vi.restoreAllMocks()
  })

  describe('getVersion', () => {
    it('should fetch version info', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve({ commitHash: 'abc1234', rcloneVersion: 'v1.65.0' })
      })

      const result = await getVersion()

      expect(mockFetch).toHaveBeenCalledWith('/api/version', expect.any(Object))
      expect(result.commitHash).toBe('abc1234')
      expect(result.rcloneVersion).toBe('v1.65.0')
    })

    it('should handle empty rclone version', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve({ commitHash: 'abc1234', rcloneVersion: '' })
      })

      const result = await getVersion()

      expect(result.commitHash).toBe('abc1234')
      expect(result.rcloneVersion).toBe('')
    })
  })
})

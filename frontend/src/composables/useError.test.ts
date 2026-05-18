import { describe, it, expect, vi } from 'vitest'
import { handleError, withErrorHandling } from './useError'

describe('useError', () => {
  describe('handleError', () => {
    it('should format error message correctly', async () => {
      const err = new Error('Test error')
      // Should not throw
      await handleError(err, { module: 'Test', operation: 'test operation' })
    })

    it('should handle string errors', async () => {
      await handleError('String error', { module: 'Test', operation: 'test' })
    })

    it('should handle null/undefined errors', async () => {
      await handleError(null)
      await handleError(undefined)
    })
  })

  describe('withErrorHandling', () => {
    it('should return fallback value on error', async () => {
      const fn = vi.fn().mockRejectedValue(new Error('API Error'))
      const result = await withErrorHandling(fn, { 
        module: 'Test', 
        operation: 'test',
        fallbackValue: 'fallback' 
      })
      expect(result).toBe('fallback')
    })

    it('should return result on success', async () => {
      const fn = vi.fn().mockResolvedValue('success')
      const result = await withErrorHandling(fn, { 
        module: 'Test', 
        operation: 'test',
        fallbackValue: 'fallback' 
      })
      expect(result).toBe('success')
    })
  })
})

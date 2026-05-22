import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import {
  handleError,
  showToastMessage,
  withErrorHandling,
  showSuccess,
  showInfo,
  showWarning,
  setErrorHandler
} from './useError'

describe('useError.ts', () => {
  beforeEach(() => {
    vi.restoreAllMocks()
    setErrorHandler(null as any)
  })

  describe('handleError', () => {
    it('should not throw when no handler is set', () => {
      const err = new Error('test error')
      expect(() => handleError(err, { module: 'Test', operation: 'testOp' })).not.toThrow()
    })

    it('should call global error handler when set', () => {
      const handler = vi.fn()
      setErrorHandler(handler)
      handleError(new Error('fail'))
      expect(handler).toHaveBeenCalledWith(expect.stringContaining('fail'), 'error')
    })

    it('should format message with operation context', () => {
      const handler = vi.fn()
      setErrorHandler(handler)
      handleError(new Error('network error'), { operation: 'fetch' })
      expect(handler).toHaveBeenCalledWith(expect.stringContaining('fetch'), 'error')
    })
  })

  describe('showToastMessage', () => {
    it('should call global handler with type', () => {
      const handler = vi.fn()
      setErrorHandler(handler)
      showToastMessage('hello', 'success')
      expect(handler).toHaveBeenCalledWith('hello', 'success')
    })

    it('should not throw when no handler is set', () => {
      expect(() => showToastMessage('hello', 'info')).not.toThrow()
    })
  })

  describe('showSuccess', () => {
    it('should call showToastMessage with success type', () => {
      const handler = vi.fn()
      setErrorHandler(handler)
      showSuccess('done')
      expect(handler).toHaveBeenCalledWith('done', 'success')
    })
  })

  describe('showInfo', () => {
    it('should call showToastMessage with info type', () => {
      const handler = vi.fn()
      setErrorHandler(handler)
      showInfo('info msg')
      expect(handler).toHaveBeenCalledWith('info msg', 'info')
    })
  })

  describe('showWarning', () => {
    it('should call showToastMessage with warning type', () => {
      const handler = vi.fn()
      setErrorHandler(handler)
      showWarning('warn msg')
      expect(handler).toHaveBeenCalledWith('warn msg', 'warning')
    })
  })

  describe('withErrorHandling', () => {
    it('should return result on success', async () => {
      const result = await withErrorHandling(() => Promise.resolve(42))
      expect(result).toBe(42)
    })

    it('should return fallbackValue on error', async () => {
      const result = await withErrorHandling(
        () => Promise.reject(new Error('fail')),
        { fallbackValue: 'default' }
      )
      expect(result).toBe('default')
    })

    it('should return undefined on error without fallback', async () => {
      const result = await withErrorHandling(() => Promise.reject(new Error('fail')))
      expect(result).toBeUndefined()
    })
  })
})

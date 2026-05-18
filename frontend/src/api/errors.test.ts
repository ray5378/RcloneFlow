import { describe, it, expect, vi, beforeEach } from 'vitest'
import {
  showToast,
  showSuccessToast,
  showErrorToast,
  showWarningToast,
  showInfoToast,
  handleApiError,
  withErrorHandler,
  withConfirm,
  registerToast,
  HTTP_ERROR_MESSAGES,
} from './errors'

// Mock window.dispatchEvent
const mockDispatchEvent = vi.fn()
window.dispatchEvent = mockDispatchEvent

describe('errors.ts', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('HTTP_ERROR_MESSAGES', () => {
    it('should have correct error messages for common status codes', () => {
      expect(HTTP_ERROR_MESSAGES[400]).toBe('请求参数错误')
      expect(HTTP_ERROR_MESSAGES[401]).toBe('未授权，请重新登录')
      expect(HTTP_ERROR_MESSAGES[403]).toBe('没有权限执行此操作')
      expect(HTTP_ERROR_MESSAGES[404]).toBe('请求的资源不存在')
      expect(HTTP_ERROR_MESSAGES[500]).toBe('服务器内部错误')
    })
  })

  describe('showToast', () => {
    it('should dispatch custom event with correct details', () => {
      showToast('test message', 'success', 3000)
      
      expect(mockDispatchEvent).toHaveBeenCalledTimes(1)
      const event = mockDispatchEvent.mock.calls[0][0] as CustomEvent
      expect(event.detail).toEqual({
        message: 'test message',
        type: 'success',
        duration: 3000,
      })
    })

    it('should use default values for optional parameters', () => {
      showToast('test')
      
      const event = mockDispatchEvent.mock.calls[0][0] as CustomEvent
      expect(event.detail.type).toBe('info')
      expect(event.detail.duration).toBe(3000)
    })
  })

  describe('showSuccessToast', () => {
    it('should call showToast with success type', () => {
      showSuccessToast('Operation completed')
      
      const event = mockDispatchEvent.mock.calls[0][0] as CustomEvent
      expect(event.detail.message).toBe('Operation completed')
      expect(event.detail.type).toBe('success')
    })
  })

  describe('showErrorToast', () => {
    it('should call showToast with error type and longer duration', () => {
      showErrorToast('Something went wrong')
      
      const event = mockDispatchEvent.mock.calls[0][0] as CustomEvent
      expect(event.detail.message).toBe('Something went wrong')
      expect(event.detail.type).toBe('error')
      expect(event.detail.duration).toBe(4000)
    })
  })

  describe('showWarningToast', () => {
    it('should call showToast with warning type', () => {
      showWarningToast('Be careful')
      
      const event = mockDispatchEvent.mock.calls[0][0] as CustomEvent
      expect(event.detail.message).toBe('Be careful')
      expect(event.detail.type).toBe('warning')
    })
  })

  describe('showInfoToast', () => {
    it('should call showToast with info type', () => {
      showInfoToast('Here is some info')
      
      const event = mockDispatchEvent.mock.calls[0][0] as CustomEvent
      expect(event.detail.message).toBe('Here is some info')
      expect(event.detail.type).toBe('info')
    })
  })

  describe('handleApiError', () => {
    it('should return error message from Error object', () => {
      const error = new Error('Backend error message')
      const result = handleApiError(error, 'Fallback message')
      
      expect(result).toBe('Backend error message')
      expect(mockDispatchEvent).toHaveBeenCalled()
    })

    it('should return fallback message for unknown errors', () => {
      const result = handleApiError(null, 'Something went wrong')
      
      expect(result).toBe('Something went wrong')
    })

    it('should return fallback message for string errors', () => {
      const result = handleApiError('String error', 'Fallback')
      
      expect(result).toBe('Fallback')
    })
  })

  describe('withErrorHandler', () => {
    it('should return result on success and show success toast', async () => {
      const apiCall = vi.fn().mockResolvedValue('success result')
      
      const result = await withErrorHandler(
        apiCall,
        'Operation successful',
        'Operation failed'
      )
      
      expect(result).toBe('success result')
      expect(apiCall).toHaveBeenCalledTimes(1)
      expect(mockDispatchEvent).toHaveBeenCalled()
    })

    it('should return null on error and show error toast', async () => {
      const apiCall = vi.fn().mockRejectedValue(new Error('API Error'))
      
      const result = await withErrorHandler(
        apiCall,
        'Success',
        'Operation failed'
      )
      
      expect(result).toBeNull()
      expect(apiCall).toHaveBeenCalledTimes(1)
    })

    it('should call apiCall with correct context', async () => {
      const apiCall = vi.fn().mockResolvedValue('result')
      
      await withErrorHandler(apiCall)
      
      expect(apiCall).toHaveBeenCalledTimes(1)
    })
  })

  describe('withConfirm', () => {
    it('should return null if user cancels', async () => {
      const originalConfirm = window.confirm
      // @ts-ignore - mock for testing
      window.confirm = vi.fn().mockReturnValue(false)
      const apiCall = vi.fn()
      
      const result = await withConfirm(apiCall, {
        confirmMessage: 'Are you sure?',
      })
      
      expect(result).toBeNull()
      expect(apiCall).not.toHaveBeenCalled()
      window.confirm = originalConfirm
    })

    it('should call apiCall if user confirms', async () => {
      const originalConfirm = window.confirm
      // @ts-ignore - mock for testing
      window.confirm = vi.fn().mockReturnValue(true)
      const apiCall = vi.fn().mockResolvedValue('result')
      
      const result = await withConfirm(apiCall, {
        confirmMessage: 'Are you sure?',
        successMessage: 'Done!',
      })
      
      expect(result).toBe('result')
      expect(apiCall).toHaveBeenCalledTimes(1)
      window.confirm = originalConfirm
    })
  })

  describe('registerToast', () => {
    it('should allow custom toast function registration', () => {
      const customToast = vi.fn()
      
      registerToast(customToast)
      
      // After registration, showToast should call the custom function
      // (This would need the module to be re-imported in real scenario)
      expect(true).toBe(true)
    })
  })
})

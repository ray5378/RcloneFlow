/**
 * API 统一错误处理模块
 * 统一错误处理和Toast提示
 */

export type ToastType = 'success' | 'error' | 'warning' | 'info'

// Toast显示函数类型
type ShowToastFn = (message: string, type: ToastType, duration?: number) => void

let showToastFn: ShowToastFn | null = null

/**
 * 注册Toast函数
 */
export function registerToast(fn: ShowToastFn) {
  showToastFn = fn
}

/**
 * 显示Toast（通过事件触发，由Toast组件处理）
 */
export function showToast(message: string, type: ToastType = 'info', duration = 3000) {
  if (showToastFn) {
    showToastFn(message, type, duration)
  } else {
    // Fallback: 使用自定义事件
    window.dispatchEvent(new CustomEvent('show-toast', {
      detail: { message, type, duration }
    }))
  }
}

/**
 * 显示成功提示
 */
export function showSuccessToast(message: string, duration = 3000) {
  showToast(message, 'success', duration)
}

/**
 * 显示错误提示
 */
export function showErrorToast(message: string, duration = 4000) {
  showToast(message, 'error', duration)
}

/**
 * 显示警告提示
 */
export function showWarningToast(message: string, duration = 3500) {
  showToast(message, 'warning', duration)
}

/**
 * 显示信息提示
 */
export function showInfoToast(message: string, duration = 3000) {
  showToast(message, 'info', duration)
}

// HTTP状态码对应的错误信息
const HTTP_ERROR_MESSAGES: Record<number, string> = {
  400: '请求参数错误',
  401: '未授权，请重新登录',
  403: '没有权限执行此操作',
  404: '请求的资源不存在',
  408: '请求超时，请重试',
  500: '服务器内部错误',
  502: '网关错误',
  503: '服务暂不可用',
  504: '网关超时',
}

/**
 * 获取HTTP状态码对应的中文错误信息
 */
function getHttpErrorMessage(status: number): string {
  return HTTP_ERROR_MESSAGES[status] || `请求失败 (${status})`
}

/**
 * 解析错误消息
 */
function parseErrorMessage(error: unknown): string {
  if (error instanceof Error) {
    return error.message
  }
  if (typeof error === 'string') {
    return error
  }
  return '操作失败'
}

/**
 * API请求错误处理
 */
export function handleApiError(error: unknown, fallbackMsg = '请求失败'): string {
  let message = fallbackMsg
  
  if (error instanceof Error) {
    // 检查是否是HTTP错误
    const httpStatusMatch = error.message.match(/status.*?(\d{3})/i)
    if (httpStatusMatch) {
      const status = parseInt(httpStatusMatch[1], 10)
      message = getHttpErrorMessage(status)
    } else if (error.message && !error.message.includes('fetch')) {
      // 后端返回的错误信息
      message = error.message
    } else {
      message = fallbackMsg
    }
  }
  
  showErrorToast(message)
  return message
}

/**
 * 带错误处理的API调用包装器
 */
export async function withErrorHandler<T>(
  apiCall: () => Promise<T>,
  successMessage?: string,
  errorMessage = '操作失败'
): Promise<T | null> {
  try {
    const result = await apiCall()
    if (successMessage) {
      showSuccessToast(successMessage)
    }
    return result
  } catch (error) {
    handleApiError(error, errorMessage)
    return null
  }
}

/**
 * 带确认提示的API调用
 */
export async function withConfirm<T>(
  apiCall: () => Promise<T>,
  options: {
    confirmTitle?: string
    confirmMessage: string
    successMessage?: string
    errorMessage?: string
  }
): Promise<T | null> {
  if (!confirm(options.confirmMessage)) {
    return null
  }
  
  return withErrorHandler(
    apiCall,
    options.successMessage,
    options.errorMessage || '操作失败'
  )
}

/**
 * 创建错误边界（用于捕获组件错误）
 */
export function createErrorBoundary(
  errorHandler: (error: Error, errorInfo: string) => void
) {
  return {
    onError(error: Error, errorInfo: string) {
      console.error('[Error Boundary]', error, errorInfo)
      errorHandler(error, errorInfo)
      showErrorToast(parseErrorMessage(error))
    }
  }
}

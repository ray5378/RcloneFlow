/**
 * API 错误处理模块
 * 统一错误处理和Toast提示
 */

import { ref } from 'vue'

// 错误状态
const errorMessage = ref<string | null>(null)
const showError = ref(false)

// Toast相关（如果项目使用element-plus或其他UI框架，可以替换这里）
let toastFn: ((msg: string, type: 'success' | 'error' | 'warning' | 'info') => void) | null = null

/**
 * 注册Toast函数
 */
export function registerToast(fn: typeof toastFn) {
  toastFn = fn
}

/**
 * 显示错误Toast
 */
export function showErrorToast(message: string) {
  if (toastFn) {
    toastFn(message, 'error')
  } else {
    // 默认使用console.error
    console.error('[API Error]', message)
    errorMessage.value = message
    showError.value = true
    
    // 3秒后自动隐藏
    setTimeout(() => {
      showError.value = false
    }, 3000)
  }
}

/**
 * 显示成功Toast
 */
export function showSuccessToast(message: string) {
  if (toastFn) {
    toastFn(message, 'success')
  } else {
    console.log('[API Success]', message)
  }
}

/**
 * API请求错误处理
 */
export function handleApiError(error: unknown, fallbackMsg = '请求失败'): string {
  if (error instanceof Error) {
    // 后端返回的错误信息
    if (error.message && !error.message.includes('fetch')) {
      showErrorToast(error.message)
      return error.message
    }
  }
  
  // 网络错误或其他
  const msg = fallbackMsg
  showErrorToast(msg)
  return msg
}

/**
 * 带错误处理的API调用包装器
 */
export async function withErrorHandler<T>(
  apiCall: () => Promise<T>,
  fallbackMsg = '请求失败'
): Promise<T | null> {
  try {
    return await apiCall()
  } catch (error) {
    handleApiError(error, fallbackMsg)
    return null
  }
}

// 导出错误状态供组件使用
export { errorMessage, showError }

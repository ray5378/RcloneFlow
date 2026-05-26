export type ToastType = 'success' | 'error' | 'warning' | 'info'

const TOAST_EVENTS = {
  SHOW_TOAST: 'show-toast',
}

const TOAST_DURATIONS = {
  SUCCESS: 3000,
  ERROR: 4000,
  WARNING: 3500,
  INFO: 3000,
}

export const HTTP_ERROR_MESSAGES: Record<number, string> = {
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

const DEFAULT_MESSAGES = {
  OPERATION_FAILED: '操作失败',
  REQUEST_FAILED: '请求失败',
  CONFIRM_TITLE: '确认',
}

type ShowToastFn = (message: string, type: ToastType, duration?: number) => void

let showToastFn: ShowToastFn | null = null

export function registerToast(fn: ShowToastFn): void {
  showToastFn = fn
}

function dispatchToastEvent(message: string, type: ToastType, duration: number): void {
  window.dispatchEvent(new CustomEvent(TOAST_EVENTS.SHOW_TOAST, {
    detail: { message, type, duration }
  }))
}

export function showToast(message: string, type: ToastType = 'info', duration = TOAST_DURATIONS.INFO): void {
  if (showToastFn) {
    showToastFn(message, type, duration)
  } else {
    dispatchToastEvent(message, type, duration)
  }
}

export function showSuccessToast(message: string, duration = TOAST_DURATIONS.SUCCESS): void {
  showToast(message, 'success', duration)
}

export function showErrorToast(message: string, duration = TOAST_DURATIONS.ERROR): void {
  showToast(message, 'error', duration)
}

export function showWarningToast(message: string, duration = TOAST_DURATIONS.WARNING): void {
  showToast(message, 'warning', duration)
}

export function showInfoToast(message: string, duration = TOAST_DURATIONS.INFO): void {
  showToast(message, 'info', duration)
}

function getHttpErrorMessage(status: number): string {
  return HTTP_ERROR_MESSAGES[status] || `请求失败 (${status})`
}

export function parseErrorMessage(error: unknown): string {
  if (error instanceof Error) {
    return error.message
  }
  if (typeof error === 'string') {
    return error
  }
  return DEFAULT_MESSAGES.OPERATION_FAILED
}

function extractHttpStatusFromError(error: Error): number | null {
  const httpStatusMatch = error.message.match(/status.*?(\d{3})/i)
  if (httpStatusMatch) {
    const status = parseInt(httpStatusMatch[1], 10)
    if (status >= 400 && status < 600) {
      return status
    }
  }
  return null
}

export function handleApiError(error: unknown, fallbackMsg = DEFAULT_MESSAGES.REQUEST_FAILED): string {
  let message = fallbackMsg
  
  if (error instanceof Error) {
    const status = extractHttpStatusFromError(error)
    if (status !== null) {
      message = getHttpErrorMessage(status)
    } else if (error.message && !error.message.includes('fetch')) {
      message = error.message
    }
  }
  
  showErrorToast(message)
  return message
}

export interface WithErrorHandlerOptions<T> {
  apiCall: () => Promise<T>
  successMessage?: string
  errorMessage?: string
}

export async function withErrorHandler<T>(
  apiCall: () => Promise<T>,
  successMessage?: string,
  errorMessage = DEFAULT_MESSAGES.OPERATION_FAILED
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

type ConfirmCallback = (title: string, message: string) => Promise<boolean>

let confirmCallback: ConfirmCallback | null = null

export function registerConfirmCallback(fn: ConfirmCallback): void {
  confirmCallback = fn
}

export interface WithConfirmOptions<T> {
  apiCall: () => Promise<T>
  confirmTitle?: string
  confirmMessage: string
  successMessage?: string
  errorMessage?: string
}

export async function withConfirm<T>(
  apiCall: () => Promise<T>,
  options: {
    confirmTitle?: string
    confirmMessage: string
    successMessage?: string
    errorMessage?: string
  }
): Promise<T | null> {
  const confirmed = confirmCallback
    ? await confirmCallback(
        options.confirmTitle || DEFAULT_MESSAGES.CONFIRM_TITLE,
        options.confirmMessage
      )
    : window.confirm(options.confirmMessage)
  
  if (!confirmed) {
    return null
  }
  
  return withErrorHandler(
    apiCall,
    options.successMessage,
    options.errorMessage || DEFAULT_MESSAGES.OPERATION_FAILED
  )
}

export interface ErrorBoundaryOptions {
  errorHandler: (error: Error, errorInfo: string) => void
}

export function createErrorBoundary(
  errorHandler: (error: Error, errorInfo: string) => void
) {
  return {
    onError(error: Error, errorInfo: string): void {
      errorHandler(error, errorInfo)
      showErrorToast(parseErrorMessage(error))
    }
  }
}

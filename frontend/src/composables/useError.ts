import { t } from '../i18n'

export interface ErrorContext {
  module?: string
  operation?: string
  fallbackValue?: any
  onError?: (message: string) => void
}

let globalErrorHandler: ((message: string, type: 'error' | 'success' | 'info' | 'warning') => void) | null = null

export function setErrorHandler(handler: (message: string, type: 'error' | 'success' | 'info' | 'warning') => void) {
  globalErrorHandler = handler
}

export function handleError(err: any, context: ErrorContext = {}): void {
  const message = context.operation
    ? t('runtime.operationFailed').replace('{operation}', context.operation).replace('{message}', err?.message || err)
    : err?.message || String(err)

  console.error(`[${context.module || 'Unknown'}] ${context.operation}:`, err)

  if (globalErrorHandler) {
    globalErrorHandler(message, 'error')
  } else {
    console.error(message)
  }
}

export function showToastMessage(message: string, type: 'error' | 'success' | 'info' | 'warning' = 'error'): void {
  if (globalErrorHandler) {
    globalErrorHandler(message, type)
  } else {
    console.log(`[${type}] ${message}`)
  }
}

export async function withErrorHandling<T>(
  fn: () => Promise<T>,
  context: ErrorContext = {}
): Promise<T | undefined> {
  try {
    return await fn()
  } catch (err) {
    handleError(err, context)
    return context.fallbackValue
  }
}

export function showSuccess(message: string): void {
  showToastMessage(message, 'success')
}

export function showInfo(message: string): void {
  showToastMessage(message, 'info')
}

export function showWarning(message: string): void {
  showToastMessage(message, 'warning')
}

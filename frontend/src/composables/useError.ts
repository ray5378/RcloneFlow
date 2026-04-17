// Unified error handling utilities
// Provides consistent error handling across the application

export interface ErrorContext {
  module?: string
  operation?: string
  fallbackValue?: any
  onError?: (message: string) => void
}

// Global error handler callback
let globalErrorHandler: ((message: string, type: 'error' | 'success' | 'info' | 'warning') => void) | null = null

export function setErrorHandler(handler: (message: string, type: 'error' | 'success' | 'info' | 'warning') => void) {
  globalErrorHandler = handler
}

/**
 * Handle error - logs to console and calls global error handler if set
 */
export function handleError(err: any, context: ErrorContext = {}): void {
  const message = context.operation 
    ? `${context.operation}失败: ${err?.message || err}`
    : err?.message || String(err)
  
  console.error(`[${context.module || 'Unknown'}] ${context.operation}:`, err)
  
  if (globalErrorHandler) {
    globalErrorHandler(message, 'error')
  } else {
    // Fallback to console.error if no handler set
    console.error(message)
  }
}

/**
 * Show toast message via global handler
 */
export function showToastMessage(message: string, type: 'error' | 'success' | 'info' | 'warning' = 'error'): void {
  if (globalErrorHandler) {
    globalErrorHandler(message, type)
  } else {
    console.log(`[${type}] ${message}`)
  }
}

/**
 * Wrap async function with unified error handling
 */
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

/**
 * Toast success message
 */
export function showSuccess(message: string): void {
  showToastMessage(message, 'success')
}

/**
 * Toast info message
 */
export function showInfo(message: string): void {
  showToastMessage(message, 'info')
}

/**
 * Toast warning message
 */
export function showWarning(message: string): void {
  showToastMessage(message, 'warning')
}

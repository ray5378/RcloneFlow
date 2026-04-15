// Unified error handling utilities
// Provides consistent error handling across the application

export interface ErrorContext {
  module?: string
  operation?: string
  fallbackValue?: any
}

/**
 * Handle error - logs to console and shows toast via ElMessage
 */
export function handleError(err: any, context: ErrorContext = {}): void {
  const message = context.operation 
    ? `${context.operation}失败: ${err?.message || err}`
    : err?.message || String(err)
  
  console.error(`[${context.module || 'Unknown'}] ${context.operation}:`, err)
  
  // Use setTimeout to avoid blocking and allow toast to show
  setTimeout(() => {
    try {
      // Dynamically import element-plus only on client side
      const { ElMessage } = require('element-plus')
      ElMessage.error(message)
    } catch {
      // Fallback - ElMessage not available
    }
  }, 0)
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
  setTimeout(() => {
    try {
      const { ElMessage } = require('element-plus')
      ElMessage.success(message)
    } catch {
      console.log(message)
    }
  }, 0)
}

/**
 * Toast info message
 */
export function showInfo(message: string): void {
  setTimeout(() => {
    try {
      const { ElMessage } = require('element-plus')
      ElMessage.info(message)
    } catch {
      console.log(message)
    }
  }, 0)
}

/**
 * Toast warning message
 */
export function showWarning(message: string): void {
  setTimeout(() => {
    try {
      const { ElMessage } = require('element-plus')
      ElMessage.warning(message)
    } catch {
      console.warn(message)
    }
  }, 0)
}

/**
 * API 统一导出模块
 * 
 * 使用方式:
 * import { getTasks, createTask } from '@/api'
 * import { getSchedules } from '@/api'
 * import { listPath, copyFile } from '@/api'
 */

// ============ API客户端 ============
export { get, post, put, del, patch, addResponseInterceptor } from './client'
export type { ResponseInterceptor } from './client'
export { default as apiRequest } from './client'

// ============ Task API ============
export { getTasks, createTask, updateTask, runTask, deleteTask, stopTaskTransfer, getTaskProgress, killTask } from './task'
export { getTasks as listTasks } from './task'

// ============ Schedule API ============
export { getSchedules, createSchedule, updateSchedule, deleteSchedule } from './schedule'
export { getSchedules as listSchedules } from './schedule'

// ============ Run API ============
export { getRuns, getRun, clearRun, clearAllRuns, clearRunsByTask, getJobStatus, getActiveRuns, getGlobalStats, stopJob } from './run'
export { getRuns as listRuns } from './run'

// ============ Remote API ============
export { getRemotes, createRemote, updateRemote, getRemoteConfig, deleteRemote, testRemote, getProviders } from './remote'
export { getRemotes as listRemotes } from './remote'
export { getProviders as listProviders } from './remote'

// ============ Browser API ============
export type { FileItem } from './browser'
export { listPath, copyFile, moveFile, copyDir, moveDir, deleteFile, purgeDir, mkdir } from './browser'

// ============ 错误处理 ============
export { showToast, showSuccessToast, showErrorToast, showWarningToast, showInfoToast, handleApiError, withErrorHandler, withConfirm, registerToast } from './errors'
export type { ToastType } from './errors'

// ============ 认证 ============
export { login, register, logout, getToken, isLoggedIn, getUser, changePassword } from './auth'

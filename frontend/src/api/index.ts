/**
 * API 统一导出模块
 * 
 * 使用方式:
 * import { getTasks, createTask } from '@/api'
 * import { getSchedules } from '@/api'
 * import { listPath, copyFile } from '@/api'
 */

// ============ API客户端 ============
export { api, get, post, put, del, patch, addResponseInterceptor } from './client'
export type { ResponseInterceptor } from './client'
export { default as apiRequest } from './client'

// ============ Task API ============
export {
  getTasks,
  createTask,
  updateTask,
  runTask,
  deleteTask,
} from './task'

// 兼容性别名 (旧代码迁移)
export const listTasks = getTasks

// ============ Schedule API ============
export {
  getSchedules,
  createSchedule,
  deleteSchedule,
} from './schedule'

// 兼容性别名
export const listSchedules = getSchedules

// ============ Run API ============
export {
  getRuns,
  getRun,
  clearRun,
  getJobStatus,
  getActiveRuns,
} from './run'

// 兼容性别名
export const listRuns = getRuns

// ============ Remote API ============
export {
  getRemotes,
  createRemote,
  updateRemote,
  getRemoteConfig,
  deleteRemote,
  testRemote,
  getProviders,
} from './remote'

// 兼容性别名
export const listRemotes = getRemotes
export const listProviders = getProviders

// ============ Browser API ============
export type { FileItem } from './browser'
export {
  listPath,
  copyFile,
  moveFile,
  copyDir,
  moveDir,
  deleteFile,
  purgeDir,
  mkdir,
} from './browser'

// ============ 错误处理 ============
export {
  showToast,
  showSuccessToast,
  showErrorToast,
  showWarningToast,
  showInfoToast,
  handleApiError,
  withErrorHandler,
  withConfirm,
  registerToast,
} from './errors'
export type { ToastType } from './errors'

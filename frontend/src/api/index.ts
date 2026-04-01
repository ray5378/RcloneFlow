/**
 * API 统一导出模块
 * 
 * 使用方式:
 * import { getTasks, createTask } from '@/api'
 * import { getSchedules } from '@/api'
 * import { listPath, copyFile } from '@/api'
 */

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
export type { ActiveRun } from './run'

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
  showErrorToast,
  showSuccessToast,
  handleApiError,
  withErrorHandler,
  registerToast,
  errorMessage,
  showError,
} from './errors'

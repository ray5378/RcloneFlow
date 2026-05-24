/**
 * Task API 模块
 * 对应后端 TaskService
 */
import { get, post, put, del, patch } from './client'
import type { ActiveRun, Run, Schedule, Task } from '../types'
import type { BisyncOptions } from '../components/task/types'

/** 强制终止任务的当前传输（按最近 run 定位 PID） */
export async function killTask(taskId: number): Promise<void> {
  await post(`/api/tasks/${taskId}/kill`, {})
}

export interface TaskBootstrapPayload {
  tasks: Task[]
  activeRuns: ActiveRun[]
}

/** 首屏聚合加载 */
export async function getTaskBootstrap(page = 1, pageSize = 50): Promise<TaskBootstrapPayload> {
  return get<TaskBootstrapPayload>(`/api/tasks/bootstrap?page=${page}&pageSize=${pageSize}`)
}

/** 获取所有任务 */
export async function getTasks(): Promise<Task[]> {
  return get<Task[]>('/api/tasks')
}

interface TaskWithBisync extends Omit<Task, 'id' | 'createdAt'> {
  bisyncOptions?: BisyncOptions
}

/** 创建任务 */
export async function createTask(task: TaskWithBisync): Promise<Task> {
  return post<Task>('/api/tasks', task)
}

/** 更新任务（主字段） */
export async function updateTask(taskId: number, task: TaskWithBisync): Promise<void> {
  return put('/api/tasks', { id: taskId, task })
}

/** 仅更新任务 Options（后端合并，不会清空未提交项） */
export async function updateTaskOptions(taskId: number, options: Record<string, any>): Promise<void> {
  return patch('/api/tasks', { id: taskId, options })
}

/** 批量保存任务排序 */
export async function updateTaskSortOrders(orders: Record<number, number>, priorityTaskId?: number): Promise<void> {
  return patch('/api/tasks', { orders, priorityTaskId })
}

export interface RunTaskResult {
  started: boolean
  reason?: 'singleton_blocked' | 'already_running' | string
  message?: string
  taskId?: number
}

/** 运行任务 */
export async function runTask(taskId: number): Promise<RunTaskResult> {
  return post<RunTaskResult>(`/api/tasks/${taskId}/run`, {})
}

/** 删除任务 */
export async function deleteTask(taskId: number): Promise<void> {
  return del(`/api/tasks/${taskId}`)
}

/** 导出所有任务和定时任务 */
export async function exportTasks(): Promise<{ tasks: any[], schedules: any[], version: number, exportedAt: string }> {
  return get('/api/tasks/export')
}

/** 导入任务和定时任务 */
export async function importTasks(payload: {
  tasks: any[]
  schedules: any[]
  conflictStrategy: 'skip' | 'overwrite'
  remoteConflictStrategy?: 'skip' | 'overwrite'
  rcloneConfig?: Record<string, Record<string, unknown>> | null
}): Promise<{ imported: number; skipped: number; overwritten: number; remotesAdded?: number; remotesSkipped?: number; remotesOverwritten?: number; remoteErrors?: string[] }> {
  return post('/api/tasks/import', payload)
}

/** 清空所有任务及其关联数据 */
export async function clearAllTasks(): Promise<void> {
  return del('/api/tasks/clear')
}

// Bisync 相关 API
export async function getBisyncLstFiles(taskId: number): Promise<{ versions: any[] }> {
  return get<{ versions: any[] }>(`/api/tasks/${taskId}/bisync/lst-files`)
}

export async function deleteBisyncLstFile(taskId: number, versionId: string): Promise<void> {
  return post(`/api/tasks/${taskId}/bisync/delete-lst`, { versionId })
}

export async function rollbackBisyncLstFile(taskId: number, versionId: string): Promise<void> {
  return post(`/api/tasks/${taskId}/bisync/rollback-lst`, { versionId })
}

export async function resyncBisync(taskId: number): Promise<void> {
  return post(`/api/tasks/${taskId}/bisync/resync`, {})
}

/**
 * Task API 模块
 * 对应后端 TaskService
 */
import { get, post, put, del, patch } from './client'
import type { ActiveRun, Run, Schedule, Task } from '../types'

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

/** 创建任务 */
export async function createTask(task: Omit<Task, 'id' | 'createdAt'>): Promise<Task> {
  return post<Task>('/api/tasks', task)
}

/** 更新任务（主字段） */
export async function updateTask(taskId: number, task: Omit<Task, 'id' | 'createdAt'>): Promise<void> {
  return put('/api/tasks', { id: taskId, task })
}

/** 仅更新任务 Options（后端合并，不会清空未提交项） */
export async function updateTaskOptions(taskId: number, options: Record<string, any>): Promise<void> {
  return patch('/api/tasks', { id: taskId, options })
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

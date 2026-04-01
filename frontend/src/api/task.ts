/**
 * Task API 模块
 * 对应后端 TaskService
 */
import { get, post, put, del } from './client'
import type { Task } from '../types'

/** 获取所有任务 */
export async function getTasks(): Promise<Task[]> {
  return get<Task[]>('/api/tasks')
}

/** 创建任务 */
export async function createTask(task: Omit<Task, 'id' | 'createdAt'>): Promise<Task> {
  return post<Task>('/api/tasks', task)
}

/** 更新任务 */
export async function updateTask(taskId: number, task: Omit<Task, 'id' | 'createdAt'>): Promise<void> {
  return put('/api/tasks', { id: taskId, task })
}

/** 运行任务 */
export async function runTask(taskId: number): Promise<{ jobId: number }> {
  return post<{ jobId: number }>(`/api/tasks/${taskId}/run`, {})
}

/** 删除任务 */
export async function deleteTask(taskId: number): Promise<void> {
  return del(`/api/tasks/${taskId}`)
}

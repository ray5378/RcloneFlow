/**
 * Task API 模块
 * 对应后端 TaskService
 */
import type { Task } from '../types'

const BASE = '/api'

async function api<T>(path: string, options?: RequestInit): Promise<T> {
  const res = await fetch(BASE + path, {
    headers: { 'Content-Type': 'application/json' },
    ...options,
  })
  const data = await res.json().catch(() => ({}))
  if (!res.ok) throw new Error((data as { error?: string }).error || res.statusText)
  return data as T
}

/** 获取所有任务 */
export async function getTasks(): Promise<Task[]> {
  return api<Task[]>('/tasks')
}

/** 创建任务 */
export async function createTask(task: Omit<Task, 'id' | 'createdAt'>): Promise<Task> {
  return api<Task>('/tasks', {
    method: 'POST',
    body: JSON.stringify(task),
  })
}

/** 更新任务 */
export async function updateTask(taskId: number, task: Omit<Task, 'id' | 'createdAt'>): Promise<void> {
  return api('/tasks', {
    method: 'PUT',
    body: JSON.stringify({ id: taskId, task }),
  })
}

/** 运行任务 */
export async function runTask(taskId: number): Promise<{ jobId: number }> {
  return api(`/tasks/${taskId}/run`, { method: 'POST' })
}

/** 删除任务 */
export async function deleteTask(taskId: number): Promise<void> {
  return api(`/tasks/${taskId}`, { method: 'DELETE' })
}

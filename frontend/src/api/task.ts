/**
 * Task API 模块
 * 对应后端 TaskService
 */
import { get, post, put, del, patch } from './client'
import { getActiveRuns, getRuns, getJobStatus, stopJob } from './run'
import type { Task } from '../types'

/** 强制终止任务的当前传输（按最近 run 定位 PID） */
export async function killTask(taskId: number): Promise<void> {
  await post(`/api/tasks/${taskId}/kill`, {})
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

/** 运行任务 */
export async function runTask(taskId: number): Promise<{ jobId: number }> {
  return post<{ jobId: number }>(`/api/tasks/${taskId}/run`, {})
}

/** 删除任务 */
export async function deleteTask(taskId: number): Promise<void> {
  return del(`/api/tasks/${taskId}`)
}

/** 停止指定任务当前传输 */
export async function stopTaskTransfer(taskId: number): Promise<void> {
  const activeRuns = await getActiveRuns()
  const active = activeRuns.find(item => item.runRecord?.taskId === taskId && item.runRecord?.rcJobId)
  if (!active?.runRecord?.rcJobId) {
    throw new Error('当前任务没有正在运行的传输')
  }
  await stopJob(active.runRecord.rcJobId)
}

// 已废弃：历史态改为读取 finalSummary，运行中由 /api/runs/active 提供轻量数据

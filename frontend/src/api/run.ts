/**
 * Run API 模块
 * 对应后端 RunService
 */
import { get, del } from './client'
import type { Run } from '../types'

/** 获取所有运行记录 */
export async function getRuns(): Promise<Run[]> {
  return get<Run[]>('/api/runs')
}

/** 获取单个运行记录 */
export async function getRun(runId: number): Promise<Run> {
  return get<Run>(`/api/runs/${runId}`)
}

/** 清除运行记录 */
export async function clearRun(runId: number): Promise<void> {
  return del(`/api/runs/${runId}`)
}

/** 清除所有历史记录 */
export async function clearAllRuns(): Promise<void> {
  return del('/api/runs')
}

/** 清除指定任务的所有历史记录 */
export async function clearRunsByTask(taskId: number): Promise<void> {
  return del(`/api/runs/task/${taskId}`)
}

/** 获取Job状态 (直接调用rclone job API) */
export async function getJobStatus(jobId: number): Promise<Record<string, unknown>> {
  return get<Record<string, unknown>>(`/api/fs/jobStatus?jobId=${jobId}`)
}

/** 获取所有运行中的任务及其实时状态 */
export async function getActiveRuns(): Promise<ActiveRun[]> {
  return get<ActiveRun[]>('/api/runs/active')
}

/** 运行中任务的实时状态 */
export interface ActiveRun {
  runRecord: {
    id: number
    taskId: number
    rcJobId: number
    status: string
    trigger: string
    startedAt: string
    summary: string
    error: string
  }
  realtimeStatus?: {
    id: number
    status: string
    success?: boolean
    error?: string
    finished?: boolean
    // 进度信息
    bytes?: number
    size?: number
    speed?: number
    speedAvg?: number
    eta?: number
    percentage?: number
  }
}

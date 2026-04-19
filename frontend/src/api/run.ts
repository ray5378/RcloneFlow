/**
 * Run API 模块
 * 对应后端 RunService
 */
import { get, del, post } from './client'
import type { Run } from '../types'

export interface RunFileRow {
  name: string
  status: 'success' | 'failed' | 'skipped' | 'deleted'
  action: string
  at: string
  sizeBytes?: number
  message?: string
}

export async function getRunFiles(runId: number, offset=0, limit=50): Promise<{ total: number; items: RunFileRow[] }> {
  return get<{ total:number; items: RunFileRow[] }>(`/api/runs/${runId}/files?offset=${offset}&limit=${limit}`)
}

/** 获取所有运行记录（分页） */
export async function getRuns(page = 1, pageSize = 50): Promise<{ runs: Run[], total: number, page: number, pageSize: number }> {
  return get<{ runs: Run[], total: number, page: number, pageSize: number }>(`/api/runs?page=${page}&pageSize=${pageSize}`)
}

/** 获取指定任务的历史记录 */
export async function getRunsByTask(taskId: number): Promise<Run[]> {
  return get<Run[]>(`/api/runs/task/${taskId}`)
}

/** 获取单个运行记录 */
export async function getRun(runId: number): Promise<Run> {
  return get<Run>(`/api/runs/${runId}`)
}

/** 清除运行记录 */
export async function clearRun(runId: number): Promise<void> {
  return del(`/api/runs/${runId}`)
}

/** 强制终止指定 run（无论是否正在传输） */
export async function killRun(runId: number): Promise<{ killed: boolean; pid?: number }> {
  return post<{ killed: boolean; pid?: number }>(`/api/runs/${runId}/kill`, {})
}

/** 清除所有历史记录 */
export async function clearAllRuns(): Promise<void> {
  return del('/api/runs')
}

/** 清除指定任务的所有历史记录 */
export async function clearRunsByTask(taskId: number): Promise<void> {
  return del(`/api/runs/task/${taskId}`)
}

/** 获取所有运行中的任务及其实时状态 */
export async function getActiveRuns(): Promise<ActiveRun[]> {
  return get<ActiveRun[]>('/api/runs/active')
}

/** 获取全局实时统计信息 */
export async function getGlobalStats(): Promise<GlobalStats> {
  return get<GlobalStats>('/api/stats/global')
}

/** 获取任务 Job 状态 */
export async function getJobStatus(jobId: number): Promise<JobStatus> {
  return get<JobStatus>(`/api/jobs/${jobId}/status`)
}

/** 停止任务 Job */
export async function stopJob(jobId: number): Promise<void> {
  return get<void>(`/api/jobs/${jobId}/stop`)
}

/** 全局实时统计 */
export interface GlobalStats {
  bytes: number        // 已传输字节
  totalBytes: number   // 总字节
  speed: number        // 当前速度 (bytes/s)
  speedAvg: number     // 平均速度 (bytes/s)
  eta: number | null  // 预计剩余时间 (秒)
  percentage: number  // 进度百分比 (0-100)
}

/** Job 状态 */
export interface JobStatus {
  id: number
  startTime: string
  stopTime?: string
  finished: boolean
  success: boolean
  error?: string
  duration?: string
  progress?: string  // 人类可读的进度字符串
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
}

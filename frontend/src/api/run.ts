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

/** 获取全局实时统计信息 */
export async function getGlobalStats(): Promise<GlobalStats> {
  return get<GlobalStats>('/api/stats/global')
}

/** 获取任务 Job 状态 */
export async function getJobStatus(jobId: number): Promise<JobStatus> {
  return get<JobStatus>(`/api/jobs/${jobId}/status`)
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

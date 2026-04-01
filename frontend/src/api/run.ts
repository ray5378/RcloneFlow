/**
 * Run API 模块
 * 对应后端 RunService
 */
import type { Run } from '../types'

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

/** 获取所有运行记录 */
export async function getRuns(): Promise<Run[]> {
  return api<Run[]>('/runs')
}

/** 获取单个运行记录 */
export async function getRun(runId: number): Promise<Run> {
  return api<Run>(`/runs/${runId}`)
}

/** 清除运行记录 */
export async function clearRun(runId: number): Promise<void> {
  return api(`/runs/${runId}`, { method: 'DELETE' })
}

/** 获取Job状态 (直接调用rclone job API) */
export async function getJobStatus(jobId: number): Promise<Record<string, unknown>> {
  return api<Record<string, unknown>>(`/fs/jobStatus?jobId=${jobId}`)
}

/** 获取所有运行中的任务及其实时状态 */
export async function getActiveRuns(): Promise<ActiveRun[]> {
  return api<ActiveRun[]>('/runs/active')
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

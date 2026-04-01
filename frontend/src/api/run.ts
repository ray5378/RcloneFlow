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

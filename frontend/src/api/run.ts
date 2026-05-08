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

/** 全局实时统计 */
export interface GlobalStats {
  bytes: number        // 已传输字节
  totalBytes: number   // 总字节
  speed: number        // 当前速度 (bytes/s)
  speedAvg: number     // 平均速度 (bytes/s)
  eta: number | null  // 预计剩余时间 (秒)
  percentage: number  // 进度百分比 (0-100)
}

export interface ActiveRunProgress {
  bytes: number
  totalBytes: number
  speed: number
  percentage: number
  completedFiles: number
  plannedFiles?: number
  logicalTotalCount?: number
  totalCount: number
  eta: number
  phase?: string
  lastUpdatedAt?: string
}

/**
 * 历史 summary.progress 仅作为历史记录里的运行中快照使用：
 * - 可用于历史列表里回看 run 当时的实时帧
 * - 不得替代 `/api/runs/active.progress` 成为运行中 UI 主字段
 */
export type RunSummaryProgress = ActiveRunProgress

export interface FinalSummaryCounts {
  success?: number
  failed?: number
  skipped?: number
  deleted?: number
  [key: string]: number | undefined
}

export interface FinalSummaryFile {
  name?: string
  path?: string
  status?: string
  action?: string
  at?: string
  sizeBytes?: number
  message?: string
}

/**
 * finalSummary 只服务于历史详情 / 最终总结展示。
 * 不得重新回流到运行中任务卡片、running hint 或 active runs 主链。
 */
export interface FinalSummary {
  startAt?: string
  finishedAt?: string
  durationSec?: number
  durationText?: string
  result?: string
  transferredBytes?: number
  totalBytes?: number
  avgSpeedBps?: number
  counts?: FinalSummaryCounts
  files?: FinalSummaryFile[]
}

/** 运行中任务的实时状态 */
export interface ActiveRun {
  runRecord: {
    id: number
    taskId: number
    status: string
    trigger?: string
    startedAt?: string
    finishedAt?: string
    bytesTransferred?: number
    summary?: string
    error?: string
    durationSeconds?: number
    durationText?: string
  }
  /** 运行中 UI 主字段 */
  progress?: ActiveRunProgress
  progressLine?: string
  progressSource?: string
  progressMismatch?: boolean
  progressCheck?: {
    ok: boolean
    pctMismatch?: boolean
    countMismatch?: boolean
    etaMismatch?: boolean
    calcPct?: number
  }
}

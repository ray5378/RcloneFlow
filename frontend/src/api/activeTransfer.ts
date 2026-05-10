import { get } from './client'
import { t } from '../i18n'

export type TrackingMode = 'normal' | 'cas'
export type ActiveTransferFileStatus = 'pending' | 'in_progress' | 'copied' | 'cas_matched' | 'skipped' | 'failed' | 'deleted'

export interface ActiveTransferSummary {
  trackingMode: TrackingMode
  completedCount: number
  pendingCount: number
  totalCount: number
  preflightPending?: boolean
  preflightFinished?: boolean
  percentage: number
  bytes: number
  totalBytes: number
  speed: number
  eta?: number
}

export interface ActiveTransferCurrentFile {
  name: string
  path?: string
  bytes?: number
  totalBytes?: number
  percentage?: number
  speed?: number
  status: ActiveTransferFileStatus
}

export interface ActiveTransferCompletedFile {
  name: string
  path?: string
  sizeBytes?: number
  at?: string
  status: ActiveTransferFileStatus
  message?: string
}

export interface ActiveTransferPendingFile {
  name: string
  path?: string
  sizeBytes?: number
  status: ActiveTransferFileStatus
}

export interface ActiveTransferOverview {
  taskId: number
  runId: number
  trackingMode: TrackingMode
  summary: ActiveTransferSummary
  currentFile?: ActiveTransferCurrentFile | null
  currentFiles?: ActiveTransferCurrentFile[]
  degraded?: boolean
  degradeReason?: string
}

export interface ActiveTransferListResponse<T> {
  total: number
  items: T[]
}

export interface ActiveTransferSnapshot {
  runId: number
  taskId: number
  trackingMode: TrackingMode
  totalCount: number
  currentFile?: ActiveTransferCurrentFile | null
  currentFiles?: ActiveTransferCurrentFile[]
  completed?: ActiveTransferCompletedFile[]
  pending?: ActiveTransferPendingFile[]
  degraded?: boolean
  degradeReason?: string
  preflightPending?: boolean
  preflightFinished?: boolean
  startedAt?: string
  updatedAt?: string
}

function mapActiveTransferError(err: any): never {
  const msg = String(err?.message || '')
  if (msg.includes('active run not found')) {
    throw new Error(t('activeTransfer.errActiveRunNotFound'))
  }
  if (msg.includes('active transfer not found')) {
    throw new Error(t('activeTransfer.errActiveTransferNotFound'))
  }
  if (msg.includes('invalid task id')) {
    throw new Error(t('activeTransfer.errInvalidTaskId'))
  }
  throw err instanceof Error ? err : new Error(msg)
}

export async function getActiveTransfer(taskId: number): Promise<ActiveTransferOverview> {
  try {
    return await get<ActiveTransferOverview>(`/api/tasks/${taskId}/active-transfer`)
  } catch (err) {
    mapActiveTransferError(err)
  }
}

export async function getActiveTransferCompleted(taskId: number, offset = 0, limit = 100): Promise<ActiveTransferListResponse<ActiveTransferCompletedFile>> {
  try {
    return await get<ActiveTransferListResponse<ActiveTransferCompletedFile>>(`/api/tasks/${taskId}/active-transfer/completed?offset=${offset}&limit=${limit}`)
  } catch (err) {
    mapActiveTransferError(err)
  }
}

export async function getActiveTransferPending(taskId: number, offset = 0, limit = 100): Promise<ActiveTransferListResponse<ActiveTransferPendingFile>> {
  try {
    return await get<ActiveTransferListResponse<ActiveTransferPendingFile>>(`/api/tasks/${taskId}/active-transfer/pending?offset=${offset}&limit=${limit}`)
  } catch (err) {
    mapActiveTransferError(err)
  }
}

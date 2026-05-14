// Shared types for RcloneFlow

export interface Task {
  id?: number
  name: string
  mode: 'sync' | 'copy' | 'move' | 'bisync'
  sourceRemote: string
  sourcePath: string
  targetRemote: string
  targetPath: string
  schedule?: string
  scheduleEnabled?: boolean
  webhooks?: string[]
  enableStreaming?: boolean
  singleton?: boolean
  options?: Record<string, any>
  createdAt?: string
  updatedAt?: string
}

export interface RunRecord {
  id: number
  taskId: number
  status: 'running' | 'finished' | 'failed' | 'skipped'
  trigger: 'manual' | 'schedule' | 'webhook'
  summary?: string
  error?: string
  createdAt: string
  updatedAt: string
  finishedAt?: string
  taskName?: string
  taskMode?: string
  sourceRemote?: string
  sourcePath?: string
  targetRemote?: string
  targetPath?: string
  bytesTransferred?: number
  speed?: string
}

export interface FinalSummary {
  totalCount?: number
  completedCount?: number
  failedCount?: number
  skippedCount?: number
  totalBytes?: number
  transferredBytes?: number
  avgSpeedBps?: number
  startedAt?: string
  finishedAt?: string
  files?: Array<{
    path?: string
    name?: string
    size?: number
    status?: string
    error?: string
    at?: string
  }>
}

export interface Remote {
  name: string
  type: string
  isCloud: boolean
}

export type TaskMode = 'sync' | 'copy' | 'move'

export type TaskFormOptionValue = string | number | boolean | string[] | Record<string, string> | undefined

export interface TaskFormOptions {
  [key: string]: TaskFormOptionValue
}

export type UpdateTaskOption = (key: string, value: TaskFormOptionValue) => void

export interface ParsedRcloneCommand {
  mode: TaskMode
  src: { remote: string; path: string }
  dst: { remote: string; path: string }
  options: TaskFormOptions
}

export interface CreateForm {
  name: string
  mode: TaskMode
  sourceRemote: string
  sourcePath: string
  targetRemote: string
  targetPath: string
  options: TaskFormOptions
  webhooks?: string
  enableStreaming?: boolean
  singleton?: boolean
}

export interface PathBreadcrumb {
  name: string
  path: string
}

export interface PathBrowseItem {
  Name?: string
  Path: string
  IsDir?: boolean
}

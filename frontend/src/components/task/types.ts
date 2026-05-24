// Shared types for RcloneFlow task form components

export type TaskMode = 'sync' | 'copy' | 'move' | 'bisync'

export type TaskFormOptionValue = string | number | boolean | string[] | Record<string, string> | undefined

export interface TaskFormOptions {
  [key: string]: TaskFormOptionValue
}

export interface BisyncOptions {
  resync?: boolean
  compare?: string
  maxDelete?: string
  checkAccess?: boolean
  checkFilename?: string
  conflictResolve?: string
  conflictLoser?: string
  conflictSuffix?: string
  backupDir1?: string
  backupDir2?: string
  createEmptySrcDirs?: boolean
  removeEmptyDirs?: boolean
  recover?: boolean
  lstBackupCount?: number
}

export interface BisyncLstVersion {
  id: string
  timestamp: string
  path1Lst: string
  path2Lst: string
  type: 'current' | 'backup' | 'conflict'
  conflict1?: string
  conflict2?: string
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
  bisyncOptions?: BisyncOptions
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

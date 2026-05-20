// Shared types for RcloneFlow task form components

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

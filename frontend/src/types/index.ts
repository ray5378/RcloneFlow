export interface Task {
  id: number
  name: string
  mode: string
  sourceRemote: string
  sourcePath: string
  targetRemote: string
  targetPath: string
  createdAt: string
}

export interface Schedule {
  id: number
  taskId: number
  spec: string
  enabled: boolean
  createdAt: string
}

export interface Run {
  id: number
  taskId: number
  rcJobId: number
  status: string
  trigger: string
  summary?: Record<string, unknown>
  error?: string
  createdAt: string
  updatedAt: string
}

export interface FileItem {
  Path: string
  Name: string
  Size: string
  IsDir: boolean
  ModTime: string
  MimeType?: string
}

export interface Provider {
  Name: string
  ArchInPath: boolean
  HashTypes: string[]
  Flags: ProviderFlag[]
  Options: ProviderOption[]
}

export interface ProviderFlag {
  Name: string
  Short: string
  Default: unknown
  Provider: unknown
  Required: boolean
  Password: boolean
  Hide: number
  Advanced: boolean
}

export interface ProviderOption {
  Name: string
  Help: string
  Default: unknown
  Provider: unknown
  Required: boolean
  Password: boolean
  Hide: number
  Advanced: boolean
  Examples?: ProviderExample[]
}

export interface ProviderExample {
  Value: string
  Help: string
}

export type RemoteTestState = 'idle' | 'testing' | 'success' | 'failed'

export interface RemoteDescription {
  [key: string]: string
}

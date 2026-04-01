import type { Task, Schedule, Run, Provider } from '../types'

const BASE = ''

async function api<T>(path: string, options?: RequestInit): Promise<T> {
  const res = await fetch(BASE + path, {
    headers: { 'Content-Type': 'application/json' },
    ...options,
  })
  const data = await res.json().catch(() => ({}))
  if (!res.ok) throw new Error((data as { error?: string }).error || res.statusText)
  return data as T
}

// Remotes
export function listRemotes() {
  return api<{ remotes: string[]; version: string }>('/api/remotes')
}

export function createRemote(name: string, type: string, parameters: Record<string, unknown>) {
  return api('/api/remotes', {
    method: 'POST',
    body: JSON.stringify({ name, type, parameters }),
  })
}

export function updateRemote(name: string, type: string, parameters: Record<string, unknown>) {
  return api('/api/remotes', {
    method: 'PUT',
    body: JSON.stringify({ name, type, parameters }),
  })
}

export function getRemoteConfig(name: string) {
  return api<Record<string, unknown>>(`/api/remotes/config/${encodeURIComponent(name)}`)
}

export function deleteRemote(name: string) {
  return api(`/api/config/${encodeURIComponent(name)}`, { method: 'DELETE' })
}

export function testRemote(name: string) {
  return api<{ ok: boolean; count: number }>('/api/remotes/test', {
    method: 'POST',
    body: JSON.stringify({ name }),
  })
}

// Providers
export function listProviders() {
  return api<{ providers: Provider[] }>('/api/providers')
}

// Browser
export function listPath(remote: string, path: string) {
  return api<{ fs: string; items: FileItem[] }>(
    `/api/browser/list?remote=${encodeURIComponent(remote)}&path=${encodeURIComponent(path)}`
  )
}

export function copyFile(srcRemote: string, srcPath: string, dstRemote: string, dstPath: string) {
  // Trim leading slash from paths as rclone expects paths without leading /
  const cleanSrcPath = srcPath.startsWith('/') ? srcPath.slice(1) : srcPath
  const cleanDstPath = dstPath.startsWith('/') ? dstPath.slice(1) : dstPath
  console.log('API copyFile:', { srcFs: srcRemote + ':', srcRemote: cleanSrcPath, dstFs: dstRemote + ':', dstRemote: cleanDstPath })
  return api('/api/fs/copy', {
    method: 'POST',
    body: JSON.stringify({ srcFs: srcRemote + ':', srcRemote: cleanSrcPath, dstFs: dstRemote + ':', dstRemote: cleanDstPath }),
  })
}

export function moveFile(srcRemote: string, srcPath: string, dstRemote: string, dstPath: string) {
  // Trim leading slash from paths as rclone expects paths without leading /
  const cleanSrcPath = srcPath.startsWith('/') ? srcPath.slice(1) : srcPath
  const cleanDstPath = dstPath.startsWith('/') ? dstPath.slice(1) : dstPath
  return api('/api/fs/move', {
    method: 'POST',
    body: JSON.stringify({ srcFs: srcRemote + ':', srcRemote: cleanSrcPath, dstFs: dstRemote + ':', dstRemote: cleanDstPath }),
  })
}

export function deleteFile(remote: string, path: string) {
  // Trim leading slash from path as rclone expects paths without leading /
  const cleanPath = path.startsWith('/') ? path.slice(1) : path
  return api('/api/fs/delete', {
    method: 'POST',
    body: JSON.stringify({ srcFs: remote + ':', srcRemote: cleanPath }),
  })
}

export interface FileItem {
  Path: string
  Name: string
  Size: string
  IsDir: boolean
  ModTime: string
  MimeType?: string
}

// Tasks
export function listTasks() {
  return api<Task[]>('/api/tasks')
}

export function createTask(task: Omit<Task, 'id' | 'createdAt'>) {
  return api<Task>('/api/tasks', {
    method: 'POST',
    body: JSON.stringify(task),
  })
}

export function runTask(taskId: number) {
  return api(`/api/tasks/${taskId}/run`, { method: 'POST' })
}

export function deleteTask(taskId: number) {
  return api(`/api/tasks/${taskId}`, { method: 'DELETE' })
}

// Schedules
export function listSchedules() {
  return api<Schedule[]>('/api/schedules')
}

export function createSchedule(schedule: Omit<Schedule, 'id' | 'createdAt'>) {
  return api<Schedule>('/api/schedules', {
    method: 'POST',
    body: JSON.stringify(schedule),
  })
}

export function deleteSchedule(scheduleId: number) {
  return api(`/api/schedules/${scheduleId}`, { method: 'DELETE' })
}

// Runs
export function listRuns() {
  return api<Run[]>('/api/runs')
}

export function getRunStatus(runId: number) {
  return api<Run>(`/api/runs/${runId}`)
}

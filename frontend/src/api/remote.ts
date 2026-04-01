/**
 * Remote API 模块
 * 对应后端 RemoteService
 */
import type { Provider } from '../types'

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

/** 获取所有远程存储列表 */
export async function getRemotes(): Promise<{ remotes: string[]; version: string }> {
  return api<{ remotes: string[]; version: string }>('/remotes')
}

/** 创建远程存储 */
export async function createRemote(
  name: string,
  type: string,
  parameters: Record<string, unknown>
): Promise<void> {
  return api('/remotes', {
    method: 'POST',
    body: JSON.stringify({ name, type, parameters }),
  })
}

/** 更新远程存储 */
export async function updateRemote(
  name: string,
  type: string,
  parameters: Record<string, unknown>
): Promise<void> {
  return api('/remotes', {
    method: 'PUT',
    body: JSON.stringify({ name, type, parameters }),
  })
}

/** 获取远程存储配置 */
export async function getRemoteConfig(name: string): Promise<Record<string, unknown>> {
  return api<Record<string, unknown>>(`/remotes/config/${encodeURIComponent(name)}`)
}

/** 删除远程存储 */
export async function deleteRemote(name: string): Promise<void> {
  return api(`/config/${encodeURIComponent(name)}`, { method: 'DELETE' })
}

/** 测试远程存储连接 */
export async function testRemote(name: string): Promise<{ ok: boolean; count: number }> {
  return api<{ ok: boolean; count: number }>('/remotes/test', {
    method: 'POST',
    body: JSON.stringify({ name }),
  })
}

/** 获取所有Provider */
export async function getProviders(): Promise<{ providers: Provider[] }> {
  return api<{ providers: Provider[] }>('/providers')
}

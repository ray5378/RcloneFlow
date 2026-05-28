/**
 * Remote API 模块
 * 对应后端 RemoteService
 */
import { get, post, put, del } from './client'
import type { Provider } from '../types'

/** 获取所有远程存储列表 */
export async function getRemotes(): Promise<{ remotes: string[]; descriptions: Record<string, string>; version: string }> {
  return get<{ remotes: string[]; descriptions: Record<string, string>; version: string }>('/api/remotes')
}

/** 更新远程存储备注 */
export async function updateRemoteDescription(name: string, description: string): Promise<void> {
  return post('/api/remotes/description', { name, description })
}

/** 创建远程存储 */
export async function createRemote(
  name: string,
  type: string,
  parameters: Record<string, unknown>
): Promise<void> {
  return post('/api/remotes', { name, type, parameters })
}

/** 更新远程存储 */
export async function updateRemote(
  name: string,
  type: string,
  parameters: Record<string, unknown>
): Promise<void> {
  return put('/api/remotes', { name, type, parameters })
}

/** 获取远程存储配置 */
export async function getRemoteConfig(name: string): Promise<Record<string, unknown>> {
  return get<Record<string, unknown>>(`/api/remotes/config/${encodeURIComponent(name)}`)
}

/** 删除远程存储 */
export async function deleteRemote(name: string): Promise<void> {
  return del(`/api/config/${encodeURIComponent(name)}`)
}

/** 删除所有远程存储 */
export async function clearAllRemotes(): Promise<void> {
  return del('/api/remotes/clear')
}

/** 测试远程存储连接 */
export async function testRemote(name: string): Promise<{ ok: boolean; count: number }> {
  return post<{ ok: boolean; count: number }>('/api/remotes/test', { name })
}

/** 获取所有Provider */
export async function getProviders(): Promise<{ providers: Provider[] }> {
  return get<{ providers: Provider[] }>('/api/providers')
}

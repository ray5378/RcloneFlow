/**
 * API 客户端封装
 * 统一的fetch封装，支持错误处理和响应拦截
 */

import { showErrorToast } from './errors'
import { getToken as getStoredToken, getRefreshToken, setTokens, logout } from './auth'

const BASE_URL = ''

// 响应拦截器
export type ResponseInterceptor = (response: Response) => Response | Promise<Response>

// 响应拦截器列表
const responseInterceptors: ResponseInterceptor[] = []

/**
 * 添加响应拦截器
 */
export function addResponseInterceptor(interceptor: ResponseInterceptor) {
  responseInterceptors.push(interceptor)
}

// 401 处理与刷新
let redirectCount = 0
let redirectTimer: number | null = null
let refreshPromise: Promise<void> | null = null

function handleUnauthorizedRedirect() {
  if (redirectTimer) return
  redirectCount++
  redirectTimer = window.setTimeout(() => { redirectCount = 0; redirectTimer = null }, 5000)
  if (redirectCount <= 1) {
    logout()
    window.location.href = '/'
  }
}

async function tryRefreshToken(): Promise<void> {
  if (refreshPromise) return refreshPromise
  const rt = getRefreshToken()
  if (!rt) {
    throw new Error('no-refresh-token')
  }
  refreshPromise = (async () => {
    const res = await fetch(`${BASE_URL}/api/auth/refresh`, {
      method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ refreshToken: rt })
    })
    if (!res.ok) { throw new Error(`refresh-${res.status}`) }
    const data = await res.json()
    setTokens(data.accessToken, data.refreshToken)
  })()
  try { await refreshPromise } finally { refreshPromise = null }
}

/**
 * API请求核心函数（含 401 自动刷新 + 重试）
 */
async function apiRequest<T>(
  path: string,
  options: RequestInit = {}
): Promise<T> {
  const url = BASE_URL + path

  const defaultHeaders: HeadersInit = { 'Content-Type': 'application/json' }
  const token = getStoredToken()
  if (token) defaultHeaders['Authorization'] = `Bearer ${token}`

  const mergedOptions: RequestInit = { ...options, headers: { ...defaultHeaders, ...options.headers } }

  try {
    let response = await fetch(url, mergedOptions)
    for (const interceptor of responseInterceptors) { response = await interceptor(response) }

    if (response.status === 401 && !path.includes('/api/auth/')) {
      try { await tryRefreshToken() } catch { handleUnauthorizedRedirect(); throw new Error('未授权') }
      // 重试一次
      const retryHeaders: Record<string, string> = { ...defaultHeaders, ...(options.headers as Record<string, string> | undefined) }
      const retryToken = getStoredToken()
      if (retryToken) retryHeaders['Authorization'] = `Bearer ${retryToken}`
      response = await fetch(url, { ...options, headers: retryHeaders })
    }

    if (response.status === 401) { handleUnauthorizedRedirect(); throw new Error('未授权') }

    if (response.status >= 400) {
      if (response.status >= 500) { showErrorToast('服务器开小差了，请稍后重试', 4000) }
      let errorMessage = `请求失败 (${response.status})`
      try { const errorData = await response.json(); if (errorData.error) errorMessage = errorData.error } catch {}
      throw new Error(errorMessage)
    }

    if (response.status === 204) { return {} as T }
    return await response.json()
  } catch (error) {
    if (error instanceof Error && error.message !== '未授权') {
      console.error(`[API] ${options.method || 'GET'} ${url} failed:`, error)
    }
    throw error
  }
}

/** GET/POST/PUT/DELETE/PATCH **/
export function get<T>(path: string): Promise<T> { return apiRequest<T>(path) }
export function post<T>(path: string, body: unknown): Promise<T> { return apiRequest<T>(path, { method: 'POST', body: JSON.stringify(body) }) }
export function put<T>(path: string, body: unknown): Promise<T> { return apiRequest<T>(path, { method: 'PUT', body: JSON.stringify(body) }) }
export function del<T>(path: string): Promise<T> { return apiRequest<T>(path, { method: 'DELETE' }) }
export function patch<T>(path: string, body: unknown): Promise<T> { return apiRequest<T>(path, { method: 'PATCH', body: JSON.stringify(body) }) }

export const api = { get, post, put, delete: del, patch }
export default apiRequest

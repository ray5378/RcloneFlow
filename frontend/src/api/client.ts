import { showErrorToast } from './errors'
import { getToken as getStoredToken, getRefreshToken, setTokens, logout } from './auth'

const BASE_URL = ''

const API_PATHS = {
  AUTH_REFRESH: '/api/auth/refresh',
}

const HTTP_STATUS = {
  UNAUTHORIZED: 401,
  NO_CONTENT: 204,
  BAD_REQUEST: 400,
  SERVER_ERROR: 500,
}

const TIMEOUTS = {
  REDIRECT_RESET_MS: 5000,
  ERROR_TOAST_MS: 4000,
}

const ERROR_MESSAGES = {
  NO_REFRESH_TOKEN: 'no-refresh-token',
  REFRESH_FAILED_PREFIX: 'refresh-',
  UNAUTHORIZED: '未授权',
  SERVER_ERROR: '服务器开小差了，请稍后重试',
}

export type ResponseInterceptor = (response: Response) => Response | Promise<Response>

const responseInterceptors: ResponseInterceptor[] = []

export function addResponseInterceptor(interceptor: ResponseInterceptor): void {
  responseInterceptors.push(interceptor)
}

let redirectCount = 0
let redirectTimer: number | null = null
let refreshPromise: Promise<void> | null = null

function handleUnauthorizedRedirect(): void {
  if (redirectTimer) return
  redirectCount++
  redirectTimer = window.setTimeout(() => {
    redirectCount = 0
    redirectTimer = null
  }, TIMEOUTS.REDIRECT_RESET_MS)
  if (redirectCount <= 1) {
    logout()
    window.location.href = '/'
  }
}

async function tryRefreshToken(): Promise<void> {
  if (refreshPromise) return refreshPromise
  
  const rt = getRefreshToken()
  if (!rt) {
    throw new Error(ERROR_MESSAGES.NO_REFRESH_TOKEN)
  }
  
  refreshPromise = (async () => {
    const res = await fetch(`${BASE_URL}${API_PATHS.AUTH_REFRESH}`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ refreshToken: rt })
    })
    if (!res.ok) {
      throw new Error(`${ERROR_MESSAGES.REFRESH_FAILED_PREFIX}${res.status}`)
    }
    const data = await res.json()
    setTokens(data.accessToken, data.refreshToken)
  })()
  
  try {
    await refreshPromise
  } finally {
    refreshPromise = null
  }
}

function buildHeaders(options: RequestInit): HeadersInit {
  const defaultHeaders: HeadersInit = { 'Content-Type': 'application/json' }
  const token = getStoredToken()
  if (token) {
    defaultHeaders['Authorization'] = `Bearer ${token}`
  }
  return { ...defaultHeaders, ...options.headers }
}

function isAuthPath(path: string): boolean {
  return path.startsWith('/api/auth/')
}

async function applyInterceptors(response: Response): Promise<Response> {
  for (const interceptor of responseInterceptors) {
    response = await interceptor(response)
  }
  return response
}

async function handleErrorResponse(response: Response): Promise<never> {
  if (response.status >= HTTP_STATUS.SERVER_ERROR) {
    showErrorToast(ERROR_MESSAGES.SERVER_ERROR, TIMEOUTS.ERROR_TOAST_MS)
  }
  
  let errorMessage = `请求失败 (${response.status})`
  try {
    const errorData = await response.json()
    if (errorData.error) {
      errorMessage = errorData.error
    }
  } catch {
    console.warn('[client] Failed to parse error response')
  }
  
  throw new Error(errorMessage)
}

async function apiRequest<T>(
  path: string,
  options: RequestInit = {}
): Promise<T> {
  const url = BASE_URL + path
  const headers = buildHeaders(options)
  const mergedOptions: RequestInit = { ...options, headers }

  let response = await fetch(url, mergedOptions)
  response = await applyInterceptors(response)

  if (response.status === HTTP_STATUS.UNAUTHORIZED && !isAuthPath(path)) {
    try {
      await tryRefreshToken()
    } catch {
      handleUnauthorizedRedirect()
      throw new Error(ERROR_MESSAGES.UNAUTHORIZED)
    }
    
    const retryHeaders = buildHeaders(options)
    response = await fetch(url, { ...options, headers: retryHeaders })
  }

  if (response.status === HTTP_STATUS.UNAUTHORIZED) {
    handleUnauthorizedRedirect()
    throw new Error(ERROR_MESSAGES.UNAUTHORIZED)
  }

  if (response.status >= HTTP_STATUS.BAD_REQUEST) {
    await handleErrorResponse(response)
  }

  if (response.status === HTTP_STATUS.NO_CONTENT) {
    return {} as T
  }
  
  return response.json()
}

export function get<T>(path: string): Promise<T> {
  return apiRequest<T>(path)
}

export function post<T>(path: string, body: unknown): Promise<T> {
  return apiRequest<T>(path, { method: 'POST', body: JSON.stringify(body) })
}

export function put<T>(path: string, body: unknown): Promise<T> {
  return apiRequest<T>(path, { method: 'PUT', body: JSON.stringify(body) })
}

export function del<T>(path: string): Promise<T> {
  return apiRequest<T>(path, { method: 'DELETE' })
}

export function patch<T>(path: string, body: unknown): Promise<T> {
  return apiRequest<T>(path, { method: 'PATCH', body: JSON.stringify(body) })
}

export const api = { get, post, put, delete: del, patch }
export default apiRequest

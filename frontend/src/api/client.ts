/**
 * API 客户端封装
 * 统一的fetch封装，支持错误处理和响应拦截
 */

import { showErrorToast, showInfoToast } from './errors'

const BASE_URL = ''

// 响应拦截器
type ResponseInterceptor = (response: Response) => Response | Promise<Response>

// 响应拦截器列表
const responseInterceptors: ResponseInterceptor[] = []

/**
 * 添加响应拦截器
 */
export function addResponseInterceptor(interceptor: ResponseInterceptor) {
  responseInterceptors.push(interceptor)
}

// 请求计数器（用于防抖）
let redirectCount = 0
let redirectTimer: number | null = null

/**
 * 处理401未授权
 */
function handleUnauthorized() {
  // 防抖：5秒内只处理一次
  if (redirectTimer) return
  
  redirectCount++
  redirectTimer = window.setTimeout(() => {
    redirectCount = 0
    redirectTimer = null
  }, 5000)
  
  if (redirectCount <= 1) {
    // 清除本地存储的认证信息
    localStorage.removeItem('authToken')
    localStorage.removeItem('refreshToken')
    localStorage.removeItem('user')
    
    // 强制跳转到登录页
    window.location.href = '/'
  }
}

/**
 * 处理500服务器错误
 */
function handleServerError(status: number) {
  if (status >= 500) {
    showErrorToast('服务器开小差了，请稍后重试', 4000)
  }
}

/**
 * API请求核心函数
 */
async function apiRequest<T>(
  path: string,
  options: RequestInit = {}
): Promise<T> {
  const url = BASE_URL + path
  
  const defaultHeaders: HeadersInit = {
    'Content-Type': 'application/json',
  }
  
  // 获取认证token（如果有）
  const token = localStorage.getItem('authToken')
  if (token) {
    defaultHeaders['Authorization'] = `Bearer ${token}`
  }
  
  const mergedOptions: RequestInit = {
    ...options,
    headers: {
      ...defaultHeaders,
      ...options.headers,
    },
  }
  
  try {
    let response = await fetch(url, mergedOptions)
    
    // 通过响应拦截器
    for (const interceptor of responseInterceptors) {
      response = await interceptor(response)
    }
    
    // 处理HTTP错误状态码
    if (response.status === 401) {
      handleUnauthorized()
      throw new Error('未授权')
    }
    
    if (response.status >= 400) {
      handleServerError(response.status)
      
      // 尝试解析错误信息
      let errorMessage = `请求失败 (${response.status})`
      try {
        const errorData = await response.json()
        if (errorData.error) {
          errorMessage = errorData.error
        }
      } catch {
        // 忽略解析错误
      }
      
      throw new Error(errorMessage)
    }
    
    // 处理204 No Content
    if (response.status === 204) {
      return {} as T
    }
    
    return await response.json()
  } catch (error) {
    if (error instanceof Error && error.message !== '未授权') {
      // 网络错误等
      console.error(`[API] ${options.method || 'GET'} ${url} failed:`, error)
    }
    throw error
  }
}

/**
 * GET请求
 */
export function get<T>(path: string): Promise<T> {
  return apiRequest<T>(path)
}

/**
 * POST请求
 */
export function post<T>(path: string, body: unknown): Promise<T> {
  return apiRequest<T>(path, {
    method: 'POST',
    body: JSON.stringify(body),
  })
}

/**
 * PUT请求
 */
export function put<T>(path: string, body: unknown): Promise<T> {
  return apiRequest<T>(path, {
    method: 'PUT',
    body: JSON.stringify(body),
  })
}

/**
 * DELETE请求
 */
export function del<T>(path: string): Promise<T> {
  return apiRequest<T>(path, {
    method: 'DELETE',
  })
}

/**
 * PATCH请求
 */
export function patch<T>(path: string, body: unknown): Promise<T> {
  return apiRequest<T>(path, {
    method: 'PATCH',
    body: JSON.stringify(body),
  })
}

// 导出请求方法
export const api = {
  get,
  post,
  put,
  delete: del,
  patch,
}

// 默认导出apiRequest
export default apiRequest

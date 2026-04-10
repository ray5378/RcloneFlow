const API_BASE = ''

interface AuthResponse {
  accessToken: string
  refreshToken: string
  user: {
    id: number
    username: string
  }
}

let isRefreshing = false

// 提前刷新：Access Token 剩余 < 2h 时尝试刷新
const EARLY_REFRESH_SECONDS = 2 * 60 * 60

function parseJwtExp(token: string | null): number | null {
  try {
    if (!token) return null
    const parts = token.split('.')
    if (parts.length !== 3) return null
    const payload = JSON.parse(atob(parts[1]))
    if (typeof payload.exp === 'number') return payload.exp
    return null
  } catch { return null }
}

function shouldEarlyRefresh(): boolean {
  const token = getToken()
  const exp = parseJwtExp(token)
  if (!exp) return false
  const now = Math.floor(Date.now() / 1000)
  return exp - now <= EARLY_REFRESH_SECONDS
}


async function request(url: string, data?: object): Promise<any> {
  const headers: Record<string, string> = {
    'Content-Type': 'application/json'
  }

  const token = getToken()
  if (token) {
    headers['Authorization'] = `Bearer ${token}`
  }

  // 主动提前刷新（不对 /auth/* 触发，避免循环）
  if (!url.includes('/auth/') && shouldEarlyRefresh() && !isRefreshing) {
    try { isRefreshing = true; await refreshToken() } catch { /* 忽略，回退被动刷新 */ } finally { isRefreshing = false }
  }

  const res = await fetch(`${API_BASE}${url}`, {
    method: data ? 'POST' : 'GET',
    headers,
    body: data ? JSON.stringify(data) : undefined
  })

  // 如果是401且不是刷新请求，尝试刷新token
  if (res.status === 401 && !url.includes('/auth/') && !isRefreshing) {
    isRefreshing = true
    try {
      await refreshToken()
      // 重试原请求
      return request(url, data)
    } catch {
      // 刷新失败，登出
      logout()
      window.location.reload()
      throw new Error('登录已过期')
    } finally {
      isRefreshing = false
    }
  }

  if (!res.ok) {
    const error = await res.json().catch(() => ({ error: '请求失败' }))
    throw new Error(error.error || '请求失败')
  }

  return res.json()
}

async function refreshToken(): Promise<void> {
  const refreshToken = getRefreshToken()
  if (!refreshToken) {
    throw new Error('No refresh token')
  }

  const res = await fetch(`${API_BASE}/api/auth/refresh`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({ refreshToken })
  })

  if (!res.ok) {
    logout()
    throw new Error('Refresh failed')
  }

  const data = await res.json()
  setTokens(data.accessToken, data.refreshToken)
}

export function setTokens(accessToken: string, refreshToken: string) {
  localStorage.setItem('authToken', accessToken)
  localStorage.setItem('refreshToken', refreshToken)
}

export async function login(username: string, password: string): Promise<AuthResponse> {
  const res = await fetch(`${API_BASE}/api/auth/login`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({ username, password })
  })

  if (!res.ok) {
    const error = await res.json().catch(() => ({ error: '登录失败' }))
    throw new Error(error.error || '登录失败')
  }

  const data = await res.json()
  setTokens(data.accessToken, data.refreshToken)
  localStorage.setItem('user', JSON.stringify(data.user))
  return data
}

export async function register(username: string, password: string): Promise<AuthResponse> {
  const res = await fetch(`${API_BASE}/api/auth/register`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({ username, password })
  })

  if (!res.ok) {
    const error = await res.json().catch(() => ({ error: '注册失败' }))
    throw new Error(error.error || '注册失败')
  }

  const data = await res.json()
  setTokens(data.accessToken, data.refreshToken)
  localStorage.setItem('user', JSON.stringify(data.user))
  return data
}

export function logout() {
  localStorage.removeItem('authToken')
  localStorage.removeItem('refreshToken')
  localStorage.removeItem('user')
}

export function getToken(): string | null {
  return localStorage.getItem('authToken')
}

export function getRefreshToken(): string | null {
  return localStorage.getItem('refreshToken')
}

export function isLoggedIn(): boolean {
  return !!getToken()
}

export function getUser(): { id: number; username: string } | null {
  const user = localStorage.getItem('user')
  return user ? JSON.parse(user) : null
}

export async function changePassword(oldPassword: string, newPassword: string, username?: string): Promise<void> {
  const res = await fetch('/api/auth/change-password', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${getToken()}`
    },
    body: JSON.stringify({ oldPassword, newPassword, username })
  })
  
  if (!res.ok) {
    const error = await res.json().catch(() => ({ error: '修改失败' }))
    throw new Error(error.error || '修改失败')
  }
}

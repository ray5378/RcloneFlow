const API_BASE = ''

const STORAGE_KEYS = {
  AUTH_TOKEN: 'authToken',
  REFRESH_TOKEN: 'refreshToken',
  USER: 'user',
}

const API_PATHS = {
  HAS_USERS: '/api/auth/has-users',
  LOGIN: '/api/auth/login',
  REGISTER: '/api/auth/register',
  ME: '/api/auth/me',
  CHANGE_PASSWORD: '/api/auth/change-password',
}

export interface AuthResponse {
  accessToken: string
  refreshToken: string
  user: {
    id: number
    username: string
  }
}

export interface User {
  id: number
  username: string
}

function buildUrl(path: string): string {
  return `${API_BASE}${path}`
}

function getAuthHeader(): Record<string, string> {
  const token = localStorage.getItem(STORAGE_KEYS.AUTH_TOKEN)
  return token ? { Authorization: `Bearer ${token}` } : {}
}

async function handleApiResponse<T>(res: Response, defaultError: string): Promise<T> {
  if (!res.ok) {
    const error = await res.json().catch(() => ({ error: defaultError }))
    throw new Error(error.error || defaultError)
  }
  return res.json()
}

export function setTokens(accessToken: string, refreshToken: string): void {
  localStorage.setItem(STORAGE_KEYS.AUTH_TOKEN, accessToken)
  localStorage.setItem(STORAGE_KEYS.REFRESH_TOKEN, refreshToken)
}

export function setUser(user: User): void {
  localStorage.setItem(STORAGE_KEYS.USER, JSON.stringify(user))
}

export async function hasUsers(): Promise<boolean> {
  const res = await fetch(buildUrl(API_PATHS.HAS_USERS))
  if (!res.ok) return false
  const data = await res.json().catch(() => ({ exists: false }))
  return data.exists === true
}

export async function login(username: string, password: string): Promise<AuthResponse> {
  const res = await fetch(buildUrl(API_PATHS.LOGIN), {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ username, password })
  })
  return handleApiResponse<AuthResponse>(res, '登录失败')
}

export async function register(username: string, password: string): Promise<AuthResponse> {
  const res = await fetch(buildUrl(API_PATHS.REGISTER), {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ username, password })
  })
  const data = await handleApiResponse<AuthResponse>(res, '注册失败')
  setTokens(data.accessToken, data.refreshToken)
  setUser(data.user)
  return data
}

export function logout(): void {
  localStorage.removeItem(STORAGE_KEYS.AUTH_TOKEN)
  localStorage.removeItem(STORAGE_KEYS.REFRESH_TOKEN)
  localStorage.removeItem(STORAGE_KEYS.USER)
}

export function getToken(): string | null {
  return localStorage.getItem(STORAGE_KEYS.AUTH_TOKEN)
}

export function getRefreshToken(): string | null {
  return localStorage.getItem(STORAGE_KEYS.REFRESH_TOKEN)
}

export function isLoggedIn(): boolean {
  return !!getToken()
}

export function getUser(): User | null {
  const user = localStorage.getItem(STORAGE_KEYS.USER)
  if (!user) return null
  try {
    return JSON.parse(user) as User
  } catch {
    console.warn('[auth] Failed to parse user from localStorage')
    logout()
    return null
  }
}

export async function me(): Promise<User> {
  const res = await fetch(buildUrl(API_PATHS.ME), {
    headers: getAuthHeader()
  })
  const data = await handleApiResponse<{ user: User }>(res, 'unauthorized')
  return data.user
}

export async function changePassword(oldPassword: string, newPassword: string, username?: string): Promise<{ message: string }> {
  const res = await fetch(buildUrl(API_PATHS.CHANGE_PASSWORD), {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      ...getAuthHeader()
    },
    body: JSON.stringify({ oldPassword, newPassword, username })
  })
  return handleApiResponse<{ message: string }>(res, '修改失败')
}

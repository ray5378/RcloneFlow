const API_BASE = ''

interface AuthResponse {
  accessToken: string
  refreshToken: string
  user: {
    id: number
    username: string
  }
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

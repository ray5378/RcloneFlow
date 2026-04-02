const API_BASE = ''

interface AuthResponse {
  token: string
  user: {
    id: number
    username: string
  }
}

async function request(url: string, data?: object): Promise<any> {
  const headers: Record<string, string> = {
    'Content-Type': 'application/json'
  }

  const res = await fetch(`${API_BASE}${url}`, {
    method: data ? 'POST' : 'GET',
    headers,
    body: data ? JSON.stringify(data) : undefined
  })

  if (!res.ok) {
    const error = await res.json().catch(() => ({ error: '请求失败' }))
    throw new Error(error.error || '请求失败')
  }

  return res.json()
}

export async function login(username: string, password: string): Promise<AuthResponse> {
  return request('/api/auth/login', { username, password })
}

export async function register(username: string, password: string): Promise<AuthResponse> {
  return request('/api/auth/register', { username, password })
}

export function logout() {
  localStorage.removeItem('authToken')
  localStorage.removeItem('user')
}

export function getToken(): string | null {
  return localStorage.getItem('authToken')
}

export function isLoggedIn(): boolean {
  return !!getToken()
}

export function getUser(): { id: number; username: string } | null {
  const user = localStorage.getItem('user')
  return user ? JSON.parse(user) : null
}

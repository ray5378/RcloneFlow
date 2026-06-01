import { getToken } from './auth'

export type WebdavStatus = {
  running: boolean
  enabled: boolean
  port: number
  path: string
}

export async function getWebdavStatus(): Promise<WebdavStatus> {
  const res = await fetch('/api/webdav/status', {
    headers: { 'Authorization': `Bearer ${getToken()}` }
  })
  if (!res.ok) throw new Error(await res.text())
  return res.json()
}

export async function startWebdav(username: string, password: string): Promise<{ ok: boolean; message: string; dav_path: string }> {
  const res = await fetch('/api/webdav/start', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${getToken()}`
    },
    body: JSON.stringify({ username, password })
  })
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: res.statusText }))
    throw new Error((err as any).error || 'Unknown error')
  }
  return res.json()
}

export async function stopWebdav(): Promise<{ ok: boolean; message: string }> {
  const res = await fetch('/api/webdav/stop', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${getToken()}`
    }
  })
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: res.statusText }))
    throw new Error((err as any).error || 'Unknown error')
  }
  return res.json()
}
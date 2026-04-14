import { getToken } from './auth'

export type SettingsResp = {
  auth: Record<string, {effective: string, default: string}>
  log: Record<string, {effective: string, default: string}>
  history: Record<string, {effective: string, default: string}>
  precheck: Record<string, {effective: string, default: string}>
  progress: Record<string, {effective: string, default: string}>
  webdav: Record<string, {effective: string, default: string}>
}

export async function getSettings(): Promise<SettingsResp> {
  const res = await fetch('/api/settings', {
    headers: { 'Authorization': `Bearer ${getToken()}` }
  })
  if (!res.ok) throw new Error(await res.text())
  return res.json()
}

export async function saveSettings(values: Record<string,string>): Promise<void> {
  const res = await fetch('/api/settings', {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json', 'Authorization': `Bearer ${getToken()}` },
    body: JSON.stringify({ values })
  })
  if (!res.ok) throw new Error(await res.text())
}

export async function resetSettings(): Promise<void> {
  const res = await fetch('/api/settings', {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json', 'Authorization': `Bearer ${getToken()}` },
    body: JSON.stringify({ reset: true })
  })
  if (!res.ok) throw new Error(await res.text())
}

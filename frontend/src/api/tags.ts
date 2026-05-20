import { getToken } from './auth'

export interface Tag {
  id: number
  tag: string
  type: 'keyword' | 'action'
  selected: boolean
}

function authHeaders(): Record<string, string> {
  return { 'Authorization': `Bearer ${getToken()}`, 'Content-Type': 'application/json' }
}

export async function fetchTags(): Promise<Tag[]> {
  const res = await fetch('/api/tags', {
    headers: { 'Authorization': `Bearer ${getToken()}` }
  })
  if (!res.ok) return []
  const data = await res.json()
  return data.tags || []
}

export async function selectTag(tag: string, selected: boolean): Promise<void> {
  await fetch('/api/tags/select', {
    method: 'POST',
    headers: authHeaders(),
    body: JSON.stringify({ tag, selected }),
  })
}

export async function createTag(tag: string): Promise<void> {
  await fetch('/api/tags/create', {
    method: 'POST',
    headers: authHeaders(),
    body: JSON.stringify({ tag }),
  })
}

export async function deleteTag(tag: string): Promise<void> {
  await fetch('/api/tags/delete', {
    method: 'POST',
    headers: authHeaders(),
    body: JSON.stringify({ tag }),
  })
}
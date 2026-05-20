import { getToken } from './auth'

export interface Tag {
  id: number
  tag: string
  type: 'keyword' | 'action'
}

export async function fetchTags(): Promise<Tag[]> {
  const res = await fetch('/api/tags', {
    headers: {
      'Authorization': `Bearer ${getToken()}`
    }
  })
  if (!res.ok) return []
  const data = await res.json()
  return data.tags || []
}
import { describe, it, expect, vi } from 'vitest'
import { fetchTags } from './tags'

vi.mock('./auth', () => ({
  getToken: () => 'mock-token'
}))

describe('fetchTags', () => {
  it('should return tags from API', async () => {
    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: true,
      json: () => Promise.resolve({ tags: [{ id: 1, tag: 'sync', type: 'action' }, { id: 2, tag: '备份', type: 'keyword' }] })
    } as Response)

    const tags = await fetchTags()
    expect(tags).toHaveLength(2)
    expect(tags[0].tag).toBe('sync')
    expect(tags[0].type).toBe('action')
    expect(tags[1].tag).toBe('备份')
    expect(tags[1].type).toBe('keyword')
  })

  it('should return empty array on failure', async () => {
    global.fetch = vi.fn().mockResolvedValue({ ok: false } as Response)
    const tags = await fetchTags()
    expect(tags).toEqual([])
  })
})
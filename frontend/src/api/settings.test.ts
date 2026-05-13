import { describe, it, expect, vi, beforeEach } from 'vitest'

vi.mock('./auth', () => ({
  getToken: vi.fn(() => 'TOKEN123'),
}))

import { getSettings, saveSettings, resetSettings } from './settings'

const mockFetch = vi.fn()
globalThis.fetch = mockFetch as any

describe('settings API', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('gets settings with bearer token', async () => {
    const payload = { auth: {}, log: {}, history: {}, precheck: {}, progress: {}, webdav: {} }
    mockFetch.mockResolvedValueOnce({ ok: true, json: () => Promise.resolve(payload) })

    await expect(getSettings()).resolves.toEqual(payload)
    expect(mockFetch).toHaveBeenCalledWith('/api/settings', {
      headers: { Authorization: 'Bearer TOKEN123' },
    })
  })

  it('saves settings values', async () => {
    mockFetch.mockResolvedValueOnce({ ok: true })

    await saveSettings({ a: '1' })

    expect(mockFetch).toHaveBeenCalledWith('/api/settings', {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json', Authorization: 'Bearer TOKEN123' },
      body: JSON.stringify({ values: { a: '1' } }),
    })
  })

  it('resets settings', async () => {
    mockFetch.mockResolvedValueOnce({ ok: true })

    await resetSettings()

    expect(mockFetch).toHaveBeenCalledWith('/api/settings', {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json', Authorization: 'Bearer TOKEN123' },
      body: JSON.stringify({ reset: true }),
    })
  })

  it('throws response text on failure', async () => {
    mockFetch.mockResolvedValueOnce({ ok: false, text: () => Promise.resolve('bad settings') })
    await expect(getSettings()).rejects.toThrow('bad settings')
  })
})

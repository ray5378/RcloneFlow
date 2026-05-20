import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { getSettings, saveSettings, resetSettings, getRemoteOrder, saveRemoteOrder } from './settings'

const mockFetch = vi.fn()
globalThis.fetch = mockFetch as any

describe('settings.ts', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    localStorage.setItem('authToken', 'test-token')
  })

  afterEach(() => {
    vi.restoreAllMocks()
  })

  describe('getSettings', () => {
    it('should fetch settings', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve({
          auth: { ACCESS_TOKEN_TTL: { effective: '24h', default: '24h' } },
          log: { LOG_LEVEL: { effective: 'info', default: 'info' } },
          history: {},
          precheck: {},
          progress: {},
          webdav: {}
        })
      })

      const result = await getSettings()

      expect(mockFetch).toHaveBeenCalledWith('/api/settings', {
        headers: { 'Authorization': 'Bearer test-token' }
      })
      expect(result.auth.ACCESS_TOKEN_TTL.effective).toBe('24h')
    })

    it('should throw on failure', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: false,
        text: () => Promise.resolve('internal error')
      })

      await expect(getSettings()).rejects.toThrow('internal error')
    })
  })

  describe('saveSettings', () => {
    it('should save settings', async () => {
      mockFetch.mockResolvedValueOnce({ ok: true })

      await saveSettings({ LOG_LEVEL: 'debug' })

      expect(mockFetch).toHaveBeenCalledWith('/api/settings', {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': 'Bearer test-token'
        },
        body: JSON.stringify({ values: { LOG_LEVEL: 'debug' } })
      })
    })

    it('should throw on failure', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: false,
        text: () => Promise.resolve('save failed')
      })

      await expect(saveSettings({})).rejects.toThrow('save failed')
    })
  })

  describe('resetSettings', () => {
    it('should reset settings', async () => {
      mockFetch.mockResolvedValueOnce({ ok: true })

      await resetSettings()

      expect(mockFetch).toHaveBeenCalledWith('/api/settings', {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': 'Bearer test-token'
        },
        body: JSON.stringify({ reset: true })
      })
    })

    it('should throw on failure', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: false,
        text: () => Promise.resolve('reset failed')
      })

      await expect(resetSettings()).rejects.toThrow('reset failed')
    })
  })

  describe('getRemoteOrder', () => {
    it('should parse remote order from settings', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve({
          remote: { REMOTE_ORDER: { effective: 'remote1, remote2, remote3', default: '' } },
          auth: {}, log: {}, history: {}, precheck: {}, progress: {}, webdav: {}
        })
      })

      const result = await getRemoteOrder()

      expect(result).toEqual(['remote1', 'remote2', 'remote3'])
    })

    it('should return empty array when no order', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve({
          remote: {},
          auth: {}, log: {}, history: {}, precheck: {}, progress: {}, webdav: {}
        })
      })

      const result = await getRemoteOrder()
      expect(result).toEqual([])
    })

    it('should return empty array when order is empty string', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve({
          remote: { REMOTE_ORDER: { effective: '', default: '' } },
          auth: {}, log: {}, history: {}, precheck: {}, progress: {}, webdav: {}
        })
      })

      const result = await getRemoteOrder()
      expect(result).toEqual([])
    })

    it('should throw on failure', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: false,
        text: () => Promise.resolve('fetch failed')
      })

      await expect(getRemoteOrder()).rejects.toThrow('fetch failed')
    })
  })

  describe('saveRemoteOrder', () => {
    it('should save remote order', async () => {
      mockFetch.mockResolvedValueOnce({ ok: true })

      await saveRemoteOrder(['remote1', 'remote2'])

      expect(mockFetch).toHaveBeenCalledWith('/api/settings', {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': 'Bearer test-token'
        },
        body: JSON.stringify({ values: { REMOTE_ORDER: 'remote1,remote2' } })
      })
    })

    it('should throw on failure', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: false,
        text: () => Promise.resolve('save order failed')
      })

      await expect(saveRemoteOrder(['remote1'])).rejects.toThrow('save order failed')
    })
  })
})

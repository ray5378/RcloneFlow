import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { getRemotes, createRemote, updateRemote, getRemoteConfig, deleteRemote, testRemote, getProviders } from './remote'

const mockFetch = vi.fn()
globalThis.fetch = mockFetch as any

describe('remote.ts', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    localStorage.setItem('authToken', 'test-token')
  })

  afterEach(() => {
    vi.restoreAllMocks()
  })

  describe('getRemotes', () => {
    it('should fetch remotes', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve({ remotes: ['myRemote', 'backup'], version: '1.65.0' })
      })

      const result = await getRemotes()

      expect(mockFetch).toHaveBeenCalledWith('/api/remotes', expect.any(Object))
      expect(result.remotes).toEqual(['myRemote', 'backup'])
      expect(result.version).toBe('1.65.0')
    })
  })

  describe('createRemote', () => {
    it('should create a remote', async () => {
      mockFetch.mockResolvedValueOnce({ ok: true, json: () => Promise.resolve({}) })

      await createRemote('newRemote', 's3', { endpoint: 's3.amazonaws.com' })

      expect(mockFetch).toHaveBeenCalledWith('/api/remotes', expect.objectContaining({
        method: 'POST',
        body: JSON.stringify({ name: 'newRemote', type: 's3', parameters: { endpoint: 's3.amazonaws.com' } })
      }))
    })
  })

  describe('updateRemote', () => {
    it('should update a remote', async () => {
      mockFetch.mockResolvedValueOnce({ ok: true, json: () => Promise.resolve({}) })

      await updateRemote('myRemote', 's3', { endpoint: 'new.endpoint.com' })

      expect(mockFetch).toHaveBeenCalledWith('/api/remotes', expect.objectContaining({
        method: 'PUT',
        body: JSON.stringify({ name: 'myRemote', type: 's3', parameters: { endpoint: 'new.endpoint.com' } })
      }))
    })
  })

  describe('getRemoteConfig', () => {
    it('should fetch remote config', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve({ type: 's3', endpoint: 's3.amazonaws.com' })
      })

      const result = await getRemoteConfig('myRemote')

      expect(mockFetch).toHaveBeenCalledWith('/api/remotes/config/myRemote', expect.any(Object))
      expect(result.type).toBe('s3')
    })
  })

  describe('deleteRemote', () => {
    it('should delete a remote', async () => {
      mockFetch.mockResolvedValueOnce({ ok: true, json: () => Promise.resolve({}) })

      await deleteRemote('oldRemote')

      expect(mockFetch).toHaveBeenCalledWith('/api/config/oldRemote', expect.any(Object))
    })
  })

  describe('testRemote', () => {
    it('should test a remote connection', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve({ ok: true, count: 5 })
      })

      const result = await testRemote('myRemote')

      expect(mockFetch).toHaveBeenCalledWith('/api/remotes/test', expect.objectContaining({
        method: 'POST',
        body: JSON.stringify({ name: 'myRemote' })
      }))
      expect(result.ok).toBe(true)
      expect(result.count).toBe(5)
    })
  })

  describe('getProviders', () => {
    it('should fetch providers', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve({ providers: [{ Name: 's3', Prefix: ['s3'] }] })
      })

      const result = await getProviders()

      expect(mockFetch).toHaveBeenCalledWith('/api/providers', expect.any(Object))
      expect(result.providers).toHaveLength(1)
    })
  })
})

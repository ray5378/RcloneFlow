import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { get, post, put, del, patch, api, addResponseInterceptor } from './client'

// Mock fetch
const mockFetch = vi.fn()
globalThis.fetch = mockFetch as any

describe('client.ts', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  afterEach(() => {
    vi.restoreAllMocks()
  })

  describe('get', () => {
    it('should make GET request and return JSON', async () => {
      const mockData = { tasks: [{ id: 1, name: 'Test' }] }
      mockFetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        json: () => Promise.resolve(mockData),
      })

      const result = await get<typeof mockData>('/api/tasks')
      
      expect(mockFetch).toHaveBeenCalledTimes(1)
      expect(mockFetch).toHaveBeenCalledWith('/api/tasks', expect.any(Object))
      expect(result).toEqual(mockData)
    })

    it('should handle 204 No Content', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        status: 204,
      })

      const result = await get<void>('/api/tasks/1')
      
      expect(result).toEqual({})
    })
  })

  describe('post', () => {
    it('should make POST request with body', async () => {
      const mockData = { id: 1, name: 'New Task' }
      const requestBody = { name: 'New Task', mode: 'copy' }
      
      mockFetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        json: () => Promise.resolve(mockData),
      })

      const result = await post<typeof mockData>('/api/tasks', requestBody)
      
      expect(mockFetch).toHaveBeenCalledWith(
        '/api/tasks',
        expect.objectContaining({
          method: 'POST',
          body: JSON.stringify(requestBody),
        })
      )
      expect(result).toEqual(mockData)
    })
  })

  describe('put', () => {
    it('should make PUT request with body', async () => {
      const requestBody = { id: 1, name: 'Updated' }
      
      mockFetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        json: () => Promise.resolve({}),
      })

      await put('/api/tasks', requestBody)
      
      expect(mockFetch).toHaveBeenCalledWith(
        '/api/tasks',
        expect.objectContaining({
          method: 'PUT',
          body: JSON.stringify(requestBody),
        })
      )
    })
  })

  describe('del', () => {
    it('should make DELETE request', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        json: () => Promise.resolve({}),
      })

      await del('/api/tasks/1')
      
      expect(mockFetch).toHaveBeenCalledWith(
        '/api/tasks/1',
        expect.objectContaining({
          method: 'DELETE',
        })
      )
    })
  })

  describe('patch', () => {
    it('should make PATCH request with body', async () => {
      const requestBody = { name: 'Patched' }
      
      mockFetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        json: () => Promise.resolve({}),
      })

      await patch('/api/tasks/1', requestBody)
      
      expect(mockFetch).toHaveBeenCalledWith(
        '/api/tasks/1',
        expect.objectContaining({
          method: 'PATCH',
          body: JSON.stringify(requestBody),
        })
      )
    })
  })

  describe('api object', () => {
    it('should export get, post, put, delete, patch methods', () => {
      expect(api.get).toBeDefined()
      expect(api.post).toBeDefined()
      expect(api.put).toBeDefined()
      expect(api.delete).toBeDefined()
      expect(api.patch).toBeDefined()
    })
  })

  describe('response interceptors', () => {
    it('should call interceptors on successful response', async () => {
      const interceptor = vi.fn((response) => response)
      addResponseInterceptor(interceptor)
      
      mockFetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        json: () => Promise.resolve({ data: 'test' }),
      })

      await get<{ data: string }>('/api/test')
      
      expect(interceptor).toHaveBeenCalled()
    })
  })

  describe('error handling', () => {
    it('should throw error on 401 unauthorized', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 401,
      })

      await expect(get('/api/tasks')).rejects.toThrow('未授权')
    })

    it('should throw error on 404 not found', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 404,
        json: () => Promise.resolve({ error: 'Not found' }),
      })

      await expect(get('/api/nonexistent')).rejects.toThrow('Not found')
    })

    it('should throw error on 500 server error', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 500,
        json: () => Promise.resolve({ error: 'Internal error' }),
      })

      await expect(get('/api/tasks')).rejects.toThrow()
    })
  })
})

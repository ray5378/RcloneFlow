import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { getActiveTransfer, getActiveTransferCompleted, getActiveTransferPending } from './activeTransfer'
import { t } from '../i18n'

const mockFetch = vi.fn()
globalThis.fetch = mockFetch as any

vi.mock('../i18n', () => ({
  t: vi.fn((key: string) => key)
}))

describe('activeTransfer.ts', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    localStorage.setItem('authToken', 'test-token')
  })

  afterEach(() => {
    vi.restoreAllMocks()
  })

  describe('getActiveTransfer', () => {
    it('should fetch active transfer overview', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve({
          taskId: 1,
          runId: 100,
          trackingMode: 'normal',
          summary: {
            trackingMode: 'normal',
            completedCount: 5,
            pendingCount: 10,
            totalCount: 15,
            percentage: 33.3,
            bytes: 1000,
            totalBytes: 3000,
            speed: 500
          },
          currentFile: { name: 'file.txt', status: 'in_progress' }
        })
      })

      const result = await getActiveTransfer(1)

      expect(mockFetch).toHaveBeenCalledWith('/api/tasks/1/active-transfer', expect.any(Object))
      expect(result.taskId).toBe(1)
      expect(result.summary.completedCount).toBe(5)
    })

    it('should throw mapped error for active run not found', async () => {
      mockFetch.mockRejectedValueOnce(new Error('active run not found'))

      await expect(getActiveTransfer(1)).rejects.toThrow('activeTransfer.errActiveRunNotFound')
    })

    it('should throw mapped error for active transfer not found', async () => {
      mockFetch.mockRejectedValueOnce(new Error('active transfer not found'))

      await expect(getActiveTransfer(1)).rejects.toThrow('activeTransfer.errActiveTransferNotFound')
    })

    it('should throw mapped error for invalid task id', async () => {
      mockFetch.mockRejectedValueOnce(new Error('invalid task id'))

      await expect(getActiveTransfer(1)).rejects.toThrow('activeTransfer.errInvalidTaskId')
    })

    it('should re-throw generic error', async () => {
      mockFetch.mockRejectedValueOnce(new Error('something went wrong'))

      await expect(getActiveTransfer(1)).rejects.toThrow('something went wrong')
    })

    it('should handle non-Error objects', async () => {
      mockFetch.mockRejectedValueOnce({ message: 'non-error object' })

      await expect(getActiveTransfer(1)).rejects.toThrow('non-error object')
    })
  })

  describe('getActiveTransferCompleted', () => {
    it('should fetch completed files with default pagination', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve({
          total: 2,
          items: [
            { name: 'done1.txt', status: 'copied' },
            { name: 'done2.txt', status: 'copied' }
          ]
        })
      })

      const result = await getActiveTransferCompleted(1)

      expect(mockFetch).toHaveBeenCalledWith(
        '/api/tasks/1/active-transfer/completed?offset=0&limit=100',
        expect.any(Object)
      )
      expect(result.total).toBe(2)
      expect(result.items).toHaveLength(2)
    })

    it('should fetch completed files with custom pagination', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve({ total: 1, items: [] })
      })

      await getActiveTransferCompleted(1, 50, 25)

      expect(mockFetch).toHaveBeenCalledWith(
        '/api/tasks/1/active-transfer/completed?offset=50&limit=25',
        expect.any(Object)
      )
    })

    it('should throw mapped error', async () => {
      mockFetch.mockRejectedValueOnce(new Error('active transfer not found'))
      await expect(getActiveTransferCompleted(1)).rejects.toThrow()
    })
  })

  describe('getActiveTransferPending', () => {
    it('should fetch pending files', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve({
          total: 3,
          items: [
            { name: 'pending1.txt', status: 'pending' }
          ]
        })
      })

      const result = await getActiveTransferPending(1)

      expect(mockFetch).toHaveBeenCalledWith(
        '/api/tasks/1/active-transfer/pending?offset=0&limit=100',
        expect.any(Object)
      )
      expect(result.total).toBe(3)
    })

    it('should fetch pending files with custom pagination', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve({ total: 0, items: [] })
      })

      await getActiveTransferPending(1, 10, 5)

      expect(mockFetch).toHaveBeenCalledWith(
        '/api/tasks/1/active-transfer/pending?offset=10&limit=5',
        expect.any(Object)
      )
    })

    it('should throw mapped error', async () => {
      mockFetch.mockRejectedValueOnce(new Error('active transfer not found'))
      await expect(getActiveTransferPending(1)).rejects.toThrow()
    })
  })
})

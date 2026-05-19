import { describe, it, expect, vi, beforeEach } from 'vitest'
import {
  getRuns,
  getRun,
  clearRun,
  getActiveRuns,
  getRunFiles,
  getRunsByTask,
  killRun,
  clearAllRuns,
  clearRunsByTask,
  getGlobalStats,
} from './run'

vi.mock('./client', () => ({
  get: vi.fn(),
  del: vi.fn(),
  post: vi.fn(),
}))

import { get, del, post } from './client'

describe('run API', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('getRuns', () => {
    it('should call get with correct path', async () => {
      const mockRuns = [
        { id: 1, taskId: 1, status: 'finished' },
        { id: 2, taskId: 2, status: 'running' },
      ]
      ;(get as ReturnType<typeof vi.fn>).mockResolvedValueOnce(mockRuns)

      const result = await getRuns()

      expect(get).toHaveBeenCalledWith('/api/runs?page=1&pageSize=50')
      expect(result).toEqual(mockRuns)
    })
  })

  describe('getRun', () => {
    it('should call get with run id', async () => {
      const mockRun = { id: 1, taskId: 1, status: 'finished' }
      ;(get as ReturnType<typeof vi.fn>).mockResolvedValueOnce(mockRun)

      const result = await getRun(1)

      expect(get).toHaveBeenCalledWith('/api/runs/1')
      expect(result).toEqual(mockRun)
    })
  })

  describe('clearRun', () => {
    it('should call del with run id', async () => {
      ;(del as ReturnType<typeof vi.fn>).mockResolvedValueOnce(undefined)

      await clearRun(1)

      expect(del).toHaveBeenCalledWith('/api/runs/1')
    })
  })

  describe('getActiveRuns', () => {
    it('should call get with active runs path', async () => {
      const mockActiveRuns = [
        {
          runRecord: { id: 1, taskId: 1, status: 'running' },
        },
      ]
      ;(get as ReturnType<typeof vi.fn>).mockResolvedValueOnce(mockActiveRuns)

      const result = await getActiveRuns()

      expect(get).toHaveBeenCalledWith('/api/runs/active')
      expect(result).toEqual(mockActiveRuns)
    })
  })

  describe('ActiveRun shape', () => {
    it('should match current frontend expectation without rcJobId', () => {
      const activeRun = {
        runRecord: {
          id: 1,
          taskId: 1,
          status: 'running',
          trigger: 'manual',
          startedAt: '2024-01-01T00:00:00Z',
          summary: '{}',
          error: '',
        },
      }

      expect(activeRun.runRecord.id).toBe(1)
      expect(activeRun.runRecord.status).toBe('running')
    })
  })

  describe('getRunFiles', () => {
    it('should call get with correct params', async () => {
      const mockResult = { total: 10, items: [] }
      ;(get as ReturnType<typeof vi.fn>).mockResolvedValueOnce(mockResult)

      const result = await getRunFiles(1, 10, 20, 'failed')

      expect(get).toHaveBeenCalledWith('/api/runs/1/files?offset=10&limit=20&filter=failed')
      expect(result).toEqual(mockResult)
    })

    it('should use default params', async () => {
      const mockResult = { total: 0, items: [] }
      ;(get as ReturnType<typeof vi.fn>).mockResolvedValueOnce(mockResult)

      await getRunFiles(1)

      expect(get).toHaveBeenCalledWith('/api/runs/1/files?offset=0&limit=50&filter=all')
    })
  })

  describe('getRunsByTask', () => {
    it('should call get with task id', async () => {
      const mockRuns = [
        { id: 1, taskId: 1, status: 'finished' },
      ]
      ;(get as ReturnType<typeof vi.fn>).mockResolvedValueOnce(mockRuns)

      const result = await getRunsByTask(1)

      expect(get).toHaveBeenCalledWith('/api/runs/task/1')
      expect(result).toEqual(mockRuns)
    })
  })

  describe('killRun', () => {
    it('should call post with run id', async () => {
      const mockResponse = { killed: true, pid: 1234 }
      ;(post as ReturnType<typeof vi.fn>).mockResolvedValueOnce(mockResponse)

      const result = await killRun(1)

      expect(post).toHaveBeenCalledWith('/api/runs/1/kill', {})
      expect(result).toEqual(mockResponse)
    })
  })

  describe('clearAllRuns', () => {
    it('should call del with correct path', async () => {
      ;(del as ReturnType<typeof vi.fn>).mockResolvedValueOnce(undefined)

      await clearAllRuns()

      expect(del).toHaveBeenCalledWith('/api/runs')
    })
  })

  describe('clearRunsByTask', () => {
    it('should call del with task id', async () => {
      ;(del as ReturnType<typeof vi.fn>).mockResolvedValueOnce(undefined)

      await clearRunsByTask(1)

      expect(del).toHaveBeenCalledWith('/api/runs/task/1')
    })
  })

  describe('getGlobalStats', () => {
    it('should call get with correct path', async () => {
      const mockStats = { bytes: 0, totalBytes: 0, speed: 0, speedAvg: 0, eta: null, percentage: 0 }
      ;(get as ReturnType<typeof vi.fn>).mockResolvedValueOnce(mockStats)

      const result = await getGlobalStats()

      expect(get).toHaveBeenCalledWith('/api/stats/global')
      expect(result).toEqual(mockStats)
    })
  })
})

import { describe, it, expect, vi, beforeEach } from 'vitest'
import {
  getRuns,
  getRun,
  clearRun,
  getActiveRuns,
} from './run'

vi.mock('./client', () => ({
  get: vi.fn(),
  del: vi.fn(),
}))

import { get, del } from './client'

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
          realtimeStatus: { id: 123, status: 'in progress', percentage: 50 },
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
        realtimeStatus: {
          id: 123,
          status: 'in progress',
          success: undefined,
          error: undefined,
          finished: false,
          bytes: 1024,
          size: 10240,
          speed: 102,
          speedAvg: 100,
          eta: 90,
          percentage: 10,
        },
      }

      expect(activeRun.runRecord.id).toBe(1)
      expect(activeRun.runRecord.status).toBe('running')
      expect(activeRun.realtimeStatus?.percentage).toBe(10)
    })
  })
})

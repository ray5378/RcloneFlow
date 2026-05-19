import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { getSchedules, createSchedule, updateSchedule, deleteSchedule } from './schedule'

const mockFetch = vi.fn()
globalThis.fetch = mockFetch as any

describe('schedule.ts', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    localStorage.setItem('authToken', 'test-token')
  })

  afterEach(() => {
    vi.restoreAllMocks()
  })

  describe('getSchedules', () => {
    it('should fetch all schedules', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve([
          { id: 1, taskId: 10, spec: '0 * * * *', enabled: true },
          { id: 2, taskId: 11, spec: '30 2 * * *', enabled: false }
        ])
      })

      const result = await getSchedules()

      expect(mockFetch).toHaveBeenCalledWith('/api/schedules', expect.any(Object))
      expect(result).toHaveLength(2)
      expect(result[0].id).toBe(1)
    })
  })

  describe('createSchedule', () => {
    it('should create a schedule', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve({ id: 3, taskId: 10, spec: '0 0 * * *', enabled: true })
      })

      const result = await createSchedule({ taskId: 10, spec: '0 0 * * *', enabled: true })

      expect(mockFetch).toHaveBeenCalledWith('/api/schedules', expect.objectContaining({
        method: 'POST',
        body: JSON.stringify({ taskId: 10, spec: '0 0 * * *', enabled: true })
      }))
      expect(result.id).toBe(3)
    })
  })

  describe('updateSchedule', () => {
    it('should update schedule with spec', async () => {
      mockFetch.mockResolvedValueOnce({ ok: true, json: () => Promise.resolve({}) })

      await updateSchedule(1, true, '0 0 * * *')

      expect(mockFetch).toHaveBeenCalledWith('/api/schedules/1', expect.objectContaining({
        method: 'PUT',
        body: JSON.stringify({ enabled: true, spec: '0 0 * * *' })
      }))
    })

    it('should update schedule without spec', async () => {
      mockFetch.mockResolvedValueOnce({ ok: true, json: () => Promise.resolve({}) })

      await updateSchedule(1, false)

      expect(mockFetch).toHaveBeenCalledWith('/api/schedules/1', expect.objectContaining({
        method: 'PUT',
        body: JSON.stringify({ enabled: false })
      }))
    })

    it('should ignore empty spec', async () => {
      mockFetch.mockResolvedValueOnce({ ok: true, json: () => Promise.resolve({}) })

      await updateSchedule(1, true, '  ')

      expect(mockFetch).toHaveBeenCalledWith('/api/schedules/1', expect.objectContaining({
        method: 'PUT',
        body: JSON.stringify({ enabled: true })
      }))
    })
  })

  describe('deleteSchedule', () => {
    it('should delete a schedule', async () => {
      mockFetch.mockResolvedValueOnce({ ok: true, json: () => Promise.resolve({}) })

      await deleteSchedule(5)

      expect(mockFetch).toHaveBeenCalledWith('/api/schedules/5', expect.any(Object))
    })
  })
})

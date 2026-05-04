import { describe, it, expect, vi, beforeEach } from 'vitest'
import * as api from '../api'
import { taskApi, remoteApi, scheduleApi, runApi } from './useApi'

vi.mock('../api', () => ({
  listTasks: vi.fn(),
  createTask: vi.fn(),
  updateTask: vi.fn(),
  deleteTask: vi.fn(),
  runTask: vi.fn(),
  killTask: vi.fn(),
  updateTaskOptions: vi.fn(),
  listRemotes: vi.fn(),
  listSchedules: vi.fn(),
  listRuns: vi.fn(),
  getRun: vi.fn(),
  getRunFiles: vi.fn(),
  clearRun: vi.fn(),
  clearAllRuns: vi.fn(),
  clearRunsByTask: vi.fn(),
  getRunsByTask: vi.fn(),
}))

describe('useApi - taskApi', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('list', () => {
    it('should return tasks on success', async () => {
      const mockTasks = [{ id: 1, name: 'Test Task' }]
      vi.mocked(api.listTasks).mockResolvedValue(mockTasks as any)

      const result = await taskApi.list()
      expect(result).toEqual(mockTasks)
    })

    it('should return empty array on error', async () => {
      vi.mocked(api.listTasks).mockRejectedValue(new Error('API Error'))

      const result = await taskApi.list()
      expect(result).toEqual([])
    })
  })

  describe('run', () => {
    it('should run task successfully', async () => {
      vi.mocked(api.runTask).mockResolvedValue({ jobId: 123 } as any)

      const result = await taskApi.run(1)
      expect(result).toEqual({ jobId: 123 })
    })

    it('should return null on error', async () => {
      vi.mocked(api.runTask).mockRejectedValue(new Error('API Error'))

      const result = await taskApi.run(1)
      expect(result).toBeNull()
    })
  })

  describe('kill', () => {
    it('should return true on success', async () => {
      vi.mocked(api.killTask).mockResolvedValue(undefined as any)

      const result = await taskApi.kill(1)
      expect(result).toBe(true)
    })

    it('should return false on error', async () => {
      vi.mocked(api.killTask).mockRejectedValue(new Error('API Error'))

      const result = await taskApi.kill(1)
      expect(result).toBe(false)
    })
  })
})

describe('useApi - remoteApi', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('should return remotes list on success', async () => {
    const mockRemotes = { remotes: [{ name: 'gdrive', type: 'drive' }] }
    vi.mocked(api.listRemotes).mockResolvedValue(mockRemotes as any)

    const result = await remoteApi.list()
    expect(result).toEqual(mockRemotes)
  })

  it('should return empty remotes on error', async () => {
    vi.mocked(api.listRemotes).mockRejectedValue(new Error('API Error'))

    const result = await remoteApi.list()
    expect(result).toEqual({ remotes: [] })
  })
})

describe('useApi - scheduleApi', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('should return schedules list on success', async () => {
    const mockSchedules = [{ id: 1, enabled: true, spec: '0 * * * *' }]
    vi.mocked(api.listSchedules).mockResolvedValue(mockSchedules as any)

    const result = await scheduleApi.list()
    expect(result).toEqual(mockSchedules)
  })

  it('should return empty array on error', async () => {
    vi.mocked(api.listSchedules).mockRejectedValue(new Error('API Error'))

    const result = await scheduleApi.list()
    expect(result).toEqual([])
  })
})

describe('useApi - runApi', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('list', () => {
    it('should return runs with pagination', async () => {
      const mockRuns = { runs: [{ id: 1, status: 'finished' }], total: 1, page: 1, pageSize: 50 }
      vi.mocked(api.listRuns).mockResolvedValue(mockRuns as any)

      const result = await runApi.list(1, 50)
      expect(result).toEqual(mockRuns)
      expect(api.listRuns).toHaveBeenCalledWith(1, 50)
    })
  })

  describe('getFiles', () => {
    it('should return files with default pagination', async () => {
      const mockFiles = { files: [{ path: '/test' }], total: 1 }
      vi.mocked(api.getRunFiles).mockResolvedValue(mockFiles as any)

      const result = await runApi.getFiles(1, 0, 100)
      expect(result).toEqual(mockFiles)
    })
  })
})

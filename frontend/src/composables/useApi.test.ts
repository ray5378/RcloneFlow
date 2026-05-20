import { describe, it, expect, vi, beforeEach } from 'vitest'
import * as api from '../api'
import { taskApi, remoteApi, scheduleApi, runApi, jobApi, activeTransferApi } from './useApi'

vi.mock('../api', () => ({
  listTasks: vi.fn(),
  getTaskBootstrap: vi.fn(),
  createTask: vi.fn(),
  updateTask: vi.fn(),
  deleteTask: vi.fn(),
  runTask: vi.fn(),
  killTask: vi.fn(),
  updateTaskOptions: vi.fn(),
  updateTaskSortOrders: vi.fn(),
  listRemotes: vi.fn(),
  listSchedules: vi.fn(),
  createSchedule: vi.fn(),
  updateSchedule: vi.fn(),
  deleteSchedule: vi.fn(),
  listRuns: vi.fn(),
  getRun: vi.fn(),
  getRunFiles: vi.fn(),
  clearRun: vi.fn(),
  clearAllRuns: vi.fn(),
  clearRunsByTask: vi.fn(),
  getRunsByTask: vi.fn(),
  getActiveRuns: vi.fn(),
  getActiveTransfer: vi.fn(),
  getActiveTransferCompleted: vi.fn(),
  getActiveTransferPending: vi.fn(),
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

  describe('bootstrap', () => {
    it('should return bootstrap data on success', async () => {
      const mockBootstrap = { tasks: [{ id: 1 }] }
      vi.mocked(api.getTaskBootstrap).mockResolvedValue(mockBootstrap as any)

      const result = await taskApi.bootstrap()
      expect(result).toEqual(mockBootstrap)
    })

    it('should return null on error', async () => {
      vi.mocked(api.getTaskBootstrap).mockRejectedValue(new Error('API Error'))

      const result = await taskApi.bootstrap()
      expect(result).toBeNull()
    })
  })

  describe('create', () => {
    it('should create task successfully', async () => {
      const mockTask = { id: 1, name: 'New Task' }
      vi.mocked(api.createTask).mockResolvedValue(mockTask as any)

      const result = await taskApi.create({ name: 'New Task' })
      expect(result).toEqual(mockTask)
    })

    it('should return null on error', async () => {
      vi.mocked(api.createTask).mockRejectedValue(new Error('API Error'))

      const result = await taskApi.create({ name: 'New Task' })
      expect(result).toBeNull()
    })
  })

  describe('update', () => {
    it('should update task successfully', async () => {
      const mockTask = { id: 1, name: 'Updated Task' }
      vi.mocked(api.updateTask).mockResolvedValue(mockTask as any)

      const result = await taskApi.update(1, { name: 'Updated Task' })
      expect(result).toEqual(mockTask)
    })

    it('should return null on error', async () => {
      vi.mocked(api.updateTask).mockRejectedValue(new Error('API Error'))

      const result = await taskApi.update(1, { name: 'Updated Task' })
      expect(result).toBeNull()
    })
  })

  describe('delete', () => {
    it('should delete task successfully', async () => {
      vi.mocked(api.deleteTask).mockResolvedValue(true as any)

      const result = await taskApi.delete(1)
      expect(result).toBe(true)
    })

    it('should return false on error', async () => {
      vi.mocked(api.deleteTask).mockRejectedValue(new Error('API Error'))

      const result = await taskApi.delete(1)
      expect(result).toBe(false)
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

  describe('updateOptions', () => {
    it('should return true on success', async () => {
      vi.mocked(api.updateTaskOptions).mockResolvedValue(undefined as any)

      const result = await taskApi.updateOptions(1, { transfers: 4 })
      expect(result).toBe(true)
    })

    it('should return false on error', async () => {
      vi.mocked(api.updateTaskOptions).mockRejectedValue(new Error('API Error'))

      const result = await taskApi.updateOptions(1, { transfers: 4 })
      expect(result).toBe(false)
    })
  })

  describe('updateSortOrders', () => {
    it('should return true on success', async () => {
      vi.mocked(api.updateTaskSortOrders).mockResolvedValue(undefined as any)

      const result = await taskApi.updateSortOrders({ 1: 10, 2: 20 }, 1)
      expect(result).toBe(true)
    })

    it('should return false on error', async () => {
      vi.mocked(api.updateTaskSortOrders).mockRejectedValue(new Error('API Error'))

      const result = await taskApi.updateSortOrders({ 1: 10, 2: 20 }, 1)
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

  describe('list', () => {
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

  describe('create', () => {
    it('should create schedule successfully', async () => {
      const mockSchedule = { id: 1, enabled: true, spec: '0 * * * *' }
      vi.mocked(api.createSchedule).mockResolvedValue(mockSchedule as any)

      const result = await scheduleApi.create({ taskId: 1, spec: '0 * * * *' })
      expect(result).toEqual(mockSchedule)
    })

    it('should return null on error', async () => {
      vi.mocked(api.createSchedule).mockRejectedValue(new Error('API Error'))

      const result = await scheduleApi.create({ taskId: 1, spec: '0 * * * *' })
      expect(result).toBeNull()
    })
  })

  describe('update', () => {
    it('should update schedule successfully', async () => {
      vi.mocked(api.updateSchedule).mockResolvedValue(true as any)

      const result = await scheduleApi.update(1, true, '0 * * * *')
      expect(result).toBe(true)
    })

    it('should return false on error', async () => {
      vi.mocked(api.updateSchedule).mockRejectedValue(new Error('API Error'))

      const result = await scheduleApi.update(1, true, '0 * * * *')
      expect(result).toBe(false)
    })
  })

  describe('delete', () => {
    it('should delete schedule successfully', async () => {
      vi.mocked(api.deleteSchedule).mockResolvedValue(true as any)

      const result = await scheduleApi.delete(1)
      expect(result).toBe(true)
    })

    it('should return false on error', async () => {
      vi.mocked(api.deleteSchedule).mockRejectedValue(new Error('API Error'))

      const result = await scheduleApi.delete(1)
      expect(result).toBe(false)
    })
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

    it('should return default pagination on error', async () => {
      vi.mocked(api.listRuns).mockRejectedValue(new Error('API Error'))

      const result = await runApi.list(1, 50)
      expect(result).toEqual({ runs: [], total: 0, page: 1, pageSize: 50 })
    })
  })

  describe('get', () => {
    it('should get run successfully', async () => {
      const mockRun = { id: 1, status: 'finished' }
      vi.mocked(api.getRun).mockResolvedValue(mockRun as any)

      const result = await runApi.get(1)
      expect(result).toEqual(mockRun)
    })

    it('should return null on error', async () => {
      vi.mocked(api.getRun).mockRejectedValue(new Error('API Error'))

      const result = await runApi.get(1)
      expect(result).toBeNull()
    })
  })

  describe('getFiles', () => {
    it('should return files with default pagination', async () => {
      const mockFiles = { items: [{ path: '/test' }], total: 1 }
      vi.mocked(api.getRunFiles).mockResolvedValue(mockFiles as any)

      const result = await runApi.getFiles(1, 0, 100)
      expect(result).toEqual(mockFiles)
    })

    it('should return empty files on error', async () => {
      vi.mocked(api.getRunFiles).mockRejectedValue(new Error('API Error'))

      const result = await runApi.getFiles(1, 0, 100)
      expect(result).toEqual({ items: [], total: 0 })
    })
  })

  describe('delete', () => {
    it('should return true on success', async () => {
      vi.mocked(api.clearRun).mockResolvedValue(undefined as any)

      const result = await runApi.delete(1)
      expect(result).toBe(true)
    })

    it('should return false on error', async () => {
      vi.mocked(api.clearRun).mockRejectedValue(new Error('API Error'))

      const result = await runApi.delete(1)
      expect(result).toBe(false)
    })
  })

  describe('deleteAll', () => {
    it('should return true on success', async () => {
      vi.mocked(api.clearAllRuns).mockResolvedValue(undefined as any)

      const result = await runApi.deleteAll()
      expect(result).toBe(true)
    })

    it('should return false on error', async () => {
      vi.mocked(api.clearAllRuns).mockRejectedValue(new Error('API Error'))

      const result = await runApi.deleteAll()
      expect(result).toBe(false)
    })
  })

  describe('deleteByTask', () => {
    it('should return true on success', async () => {
      vi.mocked(api.clearRunsByTask).mockResolvedValue(undefined as any)

      const result = await runApi.deleteByTask(1)
      expect(result).toBe(true)
    })

    it('should return false on error', async () => {
      vi.mocked(api.clearRunsByTask).mockRejectedValue(new Error('API Error'))

      const result = await runApi.deleteByTask(1)
      expect(result).toBe(false)
    })
  })

  describe('getRunsByTask', () => {
    it('should get runs by task successfully', async () => {
      const mockRuns = [{ id: 1, taskId: 1 }]
      vi.mocked(api.getRunsByTask).mockResolvedValue(mockRuns as any)

      const result = await runApi.getRunsByTask(1)
      expect(result).toEqual(mockRuns)
    })

    it('should return empty array on error', async () => {
      vi.mocked(api.getRunsByTask).mockRejectedValue(new Error('API Error'))

      const result = await runApi.getRunsByTask(1)
      expect(result).toEqual([])
    })
  })
})

describe('useApi - jobApi', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('should return active runs on success', async () => {
    const mockJobs = [{ id: 1, taskId: 1, status: 'running' }]
    vi.mocked(api.getActiveRuns).mockResolvedValue(mockJobs as any)

    const result = await jobApi.list()
    expect(result).toEqual(mockJobs)
  })

  it('should return empty array on error', async () => {
    vi.mocked(api.getActiveRuns).mockRejectedValue(new Error('API Error'))

    const result = await jobApi.list()
    expect(result).toEqual([])
  })
})

describe('useApi - activeTransferApi', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('get', () => {
    it('should return active transfer on success', async () => {
      const mockTransfer = { taskId: 1, status: 'running' }
      vi.mocked(api.getActiveTransfer).mockResolvedValue(mockTransfer as any)

      const result = await activeTransferApi.get(1)
      expect(result).toEqual(mockTransfer)
    })

    it('should return null on error', async () => {
      vi.mocked(api.getActiveTransfer).mockRejectedValue(new Error('API Error'))

      const result = await activeTransferApi.get(1)
      expect(result).toBeNull()
    })
  })

  describe('getCompleted', () => {
    it('should return completed files on success', async () => {
      const mockCompleted = { total: 10, items: [{ path: '/test' }] }
      vi.mocked(api.getActiveTransferCompleted).mockResolvedValue(mockCompleted as any)

      const result = await activeTransferApi.getCompleted(1)
      expect(result).toEqual(mockCompleted)
    })

    it('should return empty on error', async () => {
      vi.mocked(api.getActiveTransferCompleted).mockRejectedValue(new Error('API Error'))

      const result = await activeTransferApi.getCompleted(1)
      expect(result).toEqual({ total: 0, items: [] })
    })
  })

  describe('getPending', () => {
    it('should return pending files on success', async () => {
      const mockPending = { total: 5, items: [{ path: '/pending' }] }
      vi.mocked(api.getActiveTransferPending).mockResolvedValue(mockPending as any)

      const result = await activeTransferApi.getPending(1)
      expect(result).toEqual(mockPending)
    })

    it('should return empty on error', async () => {
      vi.mocked(api.getActiveTransferPending).mockRejectedValue(new Error('API Error'))

      const result = await activeTransferApi.getPending(1)
      expect(result).toEqual({ total: 0, items: [] })
    })
  })
})

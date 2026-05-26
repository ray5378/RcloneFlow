import { describe, it, expect, vi, beforeEach } from 'vitest'
import {
  getTasks,
  createTask,
  updateTask,
  runTask,
  deleteTask,
  killTask,
  getTaskBootstrap,
  updateTaskOptions,
  updateTaskSortOrders,
  exportTasks,
  importTasks,
  clearAllTasks,
  getBisyncLstFiles,
  deleteBisyncLstFile,
  rollbackBisyncLstFile,
  getBisyncLstContent,
  resyncBisync,
  resolveBisyncConflict,
} from './task'

// Mock the client module
vi.mock('./client', () => ({
  get: vi.fn(),
  post: vi.fn(),
  put: vi.fn(),
  del: vi.fn(),
  patch: vi.fn(),
}))

import { get, post, put, del, patch } from './client'

describe('task API', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('getTasks', () => {
    it('should call get with correct path', async () => {
      const mockTasks = [
        { id: 1, name: 'Task 1', mode: 'copy' },
        { id: 2, name: 'Task 2', mode: 'sync' },
      ]
      ;(get as ReturnType<typeof vi.fn>).mockResolvedValueOnce(mockTasks)

      const result = await getTasks()

      expect(get).toHaveBeenCalledWith('/api/tasks')
      expect(result).toEqual(mockTasks)
    })
  })

  describe('createTask', () => {
    it('should call post with task data', async () => {
      const taskData = {
        name: 'New Task',
        mode: 'copy' as const,
        sourceRemote: 'local',
        sourcePath: '/src',
        targetRemote: 'gdrive',
        targetPath: '/dst',
      }
      const mockCreated = { id: 1, ...taskData, createdAt: '2024-01-01' }
      ;(post as ReturnType<typeof vi.fn>).mockResolvedValueOnce(mockCreated)

      const result = await createTask(taskData)

      expect(post).toHaveBeenCalledWith('/api/tasks', taskData)
      expect(result).toEqual(mockCreated)
    })
  })

  describe('updateTask', () => {
    it('should call put with task id and data', async () => {
      const taskData = {
        name: 'Updated Task',
        mode: 'sync' as const,
        sourceRemote: 'local',
        sourcePath: '/src',
        targetRemote: 'gdrive',
        targetPath: '/dst',
      }
      ;(put as ReturnType<typeof vi.fn>).mockResolvedValueOnce(undefined)

      await updateTask(1, taskData)

      expect(put).toHaveBeenCalledWith('/api/tasks', { id: 1, task: taskData })
    })
  })

  describe('runTask', () => {
    it('should call post with task id', async () => {
      const mockResponse = { started: true, taskId: 1 }
      ;(post as ReturnType<typeof vi.fn>).mockResolvedValueOnce(mockResponse)

      const result = await runTask(1)

      expect(post).toHaveBeenCalledWith('/api/tasks/1/run', {})
      expect(result).toEqual(mockResponse)
    })

    it('should preserve singleton blocked response', async () => {
      const mockResponse = { started: false, reason: 'singleton_blocked', message: '单例模式：有其他任务正在运行，跳过本次执行', taskId: 1 }
      ;(post as ReturnType<typeof vi.fn>).mockResolvedValueOnce(mockResponse)

      const result = await runTask(1)

      expect(result).toEqual(mockResponse)
    })
  })

  describe('deleteTask', () => {
    it('should call del with task id', async () => {
      ;(del as ReturnType<typeof vi.fn>).mockResolvedValueOnce(undefined)

      await deleteTask(1)

      expect(del).toHaveBeenCalledWith('/api/tasks/1')
    })
  })

  describe('killTask', () => {
    it('should call post with task id', async () => {
      ;(post as ReturnType<typeof vi.fn>).mockResolvedValueOnce(undefined)

      await killTask(1)

      expect(post).toHaveBeenCalledWith('/api/tasks/1/kill', {})
    })
  })

  describe('getTaskBootstrap', () => {
    it('should call get with correct params', async () => {
      const mockBootstrap = { tasks: [], activeRuns: [] }
      ;(get as ReturnType<typeof vi.fn>).mockResolvedValueOnce(mockBootstrap)

      const result = await getTaskBootstrap(2, 100)

      expect(get).toHaveBeenCalledWith('/api/tasks/bootstrap?page=2&pageSize=100')
      expect(result).toEqual(mockBootstrap)
    })

    it('should use default params when none provided', async () => {
      const mockBootstrap = { tasks: [], activeRuns: [] }
      ;(get as ReturnType<typeof vi.fn>).mockResolvedValueOnce(mockBootstrap)

      await getTaskBootstrap()

      expect(get).toHaveBeenCalledWith('/api/tasks/bootstrap?page=1&pageSize=50')
    })
  })

  describe('updateTaskOptions', () => {
    it('should call patch with options', async () => {
      const options = { verify: true, bwLimit: '1M' }
      ;(patch as ReturnType<typeof vi.fn>).mockResolvedValueOnce(undefined)

      await updateTaskOptions(1, options)

      expect(patch).toHaveBeenCalledWith('/api/tasks', { id: 1, options })
    })
  })

  describe('updateTaskSortOrders', () => {
    it('should call patch with orders', async () => {
      const orders = { 1: 10, 2: 20 }
      ;(patch as ReturnType<typeof vi.fn>).mockResolvedValueOnce(undefined)

      await updateTaskSortOrders(orders, 1)

      expect(patch).toHaveBeenCalledWith('/api/tasks', { orders, priorityTaskId: 1 })
    })
  })

  describe('exportTasks', () => {
    it('should call get with correct path', async () => {
      const mockExport = { tasks: [], schedules: [], version: 1, exportedAt: '2024-01-01' }
      ;(get as ReturnType<typeof vi.fn>).mockResolvedValueOnce(mockExport)

      const result = await exportTasks()

      expect(get).toHaveBeenCalledWith('/api/tasks/export')
      expect(result).toEqual(mockExport)
    })
  })

  describe('importTasks', () => {
    it('should call post with payload', async () => {
      const payload = {
        tasks: [],
        schedules: [],
        conflictStrategy: 'skip' as const,
      }
      const mockResponse = { imported: 0, skipped: 0, overwritten: 0 }
      ;(post as ReturnType<typeof vi.fn>).mockResolvedValueOnce(mockResponse)

      const result = await importTasks(payload)

      expect(post).toHaveBeenCalledWith('/api/tasks/import', payload)
      expect(result).toEqual(mockResponse)
    })

    it('should handle optional remote conflict strategy', async () => {
      const payload = {
        tasks: [],
        schedules: [],
        conflictStrategy: 'overwrite' as const,
        remoteConflictStrategy: 'skip' as const,
      }
      const mockResponse = { imported: 0, skipped: 0, overwritten: 0, remotesAdded: 0, remotesSkipped: 0, remotesOverwritten: 0, remoteErrors: [] }
      ;(post as ReturnType<typeof vi.fn>).mockResolvedValueOnce(mockResponse)

      const result = await importTasks(payload)

      expect(post).toHaveBeenCalledWith('/api/tasks/import', payload)
      expect(result).toEqual(mockResponse)
    })

    it('should handle rclone config in payload', async () => {
      const payload = {
        tasks: [],
        schedules: [],
        conflictStrategy: 'skip' as const,
        rcloneConfig: { myremote: { type: 'local' } },
      }
      const mockResponse = { imported: 0, skipped: 0, overwritten: 0 }
      ;(post as ReturnType<typeof vi.fn>).mockResolvedValueOnce(mockResponse)

      const result = await importTasks(payload)

      expect(post).toHaveBeenCalledWith('/api/tasks/import', payload)
      expect(result).toEqual(mockResponse)
    })
  })

  describe('clearAllTasks', () => {
    it('should call del with correct path', async () => {
      ;(del as ReturnType<typeof vi.fn>).mockResolvedValueOnce(undefined)

      await clearAllTasks()

      expect(del).toHaveBeenCalledWith('/api/tasks/clear')
    })
  })

  describe('Bisync APIs', () => {
    describe('getBisyncLstFiles', () => {
      it('should call get with task id', async () => {
        const mockResponse = { versions: [] }
        ;(get as ReturnType<typeof vi.fn>).mockResolvedValueOnce(mockResponse)

        const result = await getBisyncLstFiles(1)

        expect(get).toHaveBeenCalledWith('/api/tasks/1/bisync/lst-files')
        expect(result).toEqual(mockResponse)
      })
    })

    describe('deleteBisyncLstFile', () => {
      it('should call post with task id and version id', async () => {
        ;(post as ReturnType<typeof vi.fn>).mockResolvedValueOnce(undefined)

        await deleteBisyncLstFile(1, 'version123')

        expect(post).toHaveBeenCalledWith('/api/tasks/1/bisync/delete-lst', { versionId: 'version123' })
      })
    })

    describe('rollbackBisyncLstFile', () => {
      it('should call post with task id and version id', async () => {
        ;(post as ReturnType<typeof vi.fn>).mockResolvedValueOnce(undefined)

        await rollbackBisyncLstFile(1, 'version123')

        expect(post).toHaveBeenCalledWith('/api/tasks/1/bisync/rollback-lst', { versionId: 'version123' })
      })
    })

    describe('getBisyncLstContent', () => {
      it('should call get with task id and encoded file name', async () => {
        const mockResponse = { content: 'file content' }
        ;(get as ReturnType<typeof vi.fn>).mockResolvedValueOnce(mockResponse)

        const result = await getBisyncLstContent(1, 'test file.lst')

        expect(get).toHaveBeenCalledWith('/api/tasks/1/bisync/lst-content?file=test%20file.lst')
        expect(result).toEqual(mockResponse)
      })
    })

    describe('resyncBisync', () => {
      it('should call post with task id', async () => {
        ;(post as ReturnType<typeof vi.fn>).mockResolvedValueOnce(undefined)

        await resyncBisync(1)

        expect(post).toHaveBeenCalledWith('/api/tasks/1/bisync/resync', {})
      })
    })

    describe('resolveBisyncConflict', () => {
      it('should call post with task id, version id, and keep file option', async () => {
        ;(post as ReturnType<typeof vi.fn>).mockResolvedValueOnce(undefined)

        await resolveBisyncConflict(1, 'version123', 'conflict1')

        expect(post).toHaveBeenCalledWith('/api/tasks/1/bisync/resolve-conflict', { versionId: 'version123', keepFile: 'conflict1' })
      })

      it('should handle keepFile option conflict2', async () => {
        ;(post as ReturnType<typeof vi.fn>).mockResolvedValueOnce(undefined)

        await resolveBisyncConflict(1, 'version456', 'conflict2')

        expect(post).toHaveBeenCalledWith('/api/tasks/1/bisync/resolve-conflict', { versionId: 'version456', keepFile: 'conflict2' })
      })
    })
  })
})

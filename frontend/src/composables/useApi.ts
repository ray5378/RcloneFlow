// API wrapper with unified error handling
import * as api from '../api'
import { handleError } from './useError'

/**
 * Task APIs with error handling
 */
export const taskApi = {
  async list() {
    try {
      return await api.listTasks()
    } catch (err) {
      handleError(err, { module: 'Task', operation: '获取任务列表' })
      return []
    }
  },

  async create(task: any) {
    try {
      return await api.createTask(task)
    } catch (err) {
      handleError(err, { module: 'Task', operation: '创建任务' })
      return null
    }
  },

  async update(id: number, task: any) {
    try {
      return await api.updateTask(id, task)
    } catch (err) {
      handleError(err, { module: 'Task', operation: '更新任务' })
      return null
    }
  },

  async delete(id: number) {
    try {
      return await api.deleteTask(id)
    } catch (err) {
      handleError(err, { module: 'Task', operation: '删除任务' })
      return false
    }
  },

  async run(id: number) {
    try {
      return await api.runTask(id)
    } catch (err) {
      handleError(err, { module: 'Task', operation: '运行任务' })
      return null
    }
  },

  async stop(id: number) {
    try {
      return await api.stopTask(id)
    } catch (err) {
      handleError(err, { module: 'Task', operation: '停止任务' })
      return false
    }
  }
}

/**
 * Remote APIs with error handling
 */
export const remoteApi = {
  async list() {
    try {
      return await api.listRemotes()
    } catch (err) {
      handleError(err, { module: 'Remote', operation: '获取远程存储列表' })
      return { remotes: [] }
    }
  }
}

/**
 * Schedule APIs with error handling
 */
export const scheduleApi = {
  async list() {
    try {
      return await api.listSchedules()
    } catch (err) {
      handleError(err, { module: 'Schedule', operation: '获取定时任务列表' })
      return []
    }
  }
}

/**
 * Run APIs with error handling
 */
export const runApi = {
  async list(page = 1, pageSize = 50) {
    try {
      return await api.listRuns(page, pageSize)
    } catch (err) {
      handleError(err, { module: 'Run', operation: '获取历史记录' })
      return []
    }
  },

  async get(id: number) {
    try {
      return await api.getRun(id)
    } catch (err) {
      handleError(err, { module: 'Run', operation: '获取运行详情' })
      return null
    }
  },

  async getFiles(id: number, offset: number, limit: number) {
    try {
      return await api.getRunFiles(id, offset, limit)
    } catch (err) {
      handleError(err, { module: 'Run', operation: '获取运行文件列表' })
      return { files: [], total: 0 }
    }
  },

  async delete(id: number) {
    try {
      return await api.clearRun(id)
    } catch (err) {
      handleError(err, { module: 'Run', operation: '删除历史记录' })
      return false
    }
  },

  async deleteAll() {
    try {
      return await api.clearAllRuns()
    } catch (err) {
      handleError(err, { module: 'Run', operation: '清空历史记录' })
      return false
    }
  }
}

/**
 * Queue APIs with error handling
 */
export const queueApi = {
  async getStatus() {
    try {
      return await api.getQueueStatus()
    } catch (err) {
      handleError(err, { module: 'Queue', operation: '获取队列状态' })
      return null
    }
  },

  async add(taskId: number, options?: any) {
    try {
      return await api.addToQueue(taskId, options)
    } catch (err) {
      handleError(err, { module: 'Queue', operation: '添加到队列' })
      return false
    }
  },

  async remove(jobId: string) {
    try {
      return await api.removeFromQueue(jobId)
    } catch (err) {
      handleError(err, { module: 'Queue', operation: '从队列移除' })
      return false
    }
  },

  async clear() {
    try {
      return await api.clearQueue()
    } catch (err) {
      handleError(err, { module: 'Queue', operation: '清空队列' })
      return false
    }
  }
}

/**
 * Job APIs with error handling
 */
export const jobApi = {
  async list() {
    try {
      return await api.listJobs()
    } catch (err) {
      handleError(err, { module: 'Job', operation: '获取任务列表' })
      return []
    }
  },

  async stop(id: number) {
    try {
      return await api.stopJob(id)
    } catch (err) {
      handleError(err, { module: 'Job', operation: '停止任务' })
      return false
    }
  }
}

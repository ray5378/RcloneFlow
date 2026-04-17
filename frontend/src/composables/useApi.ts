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
      await api.stopTaskTransfer(id)
      return true
    } catch (err) {
      handleError(err, { module: 'Task', operation: '停止任务' })
      return false
    }
  },

  async kill(id: number) {
    try {
      await api.killTask(id)
      return true
    } catch (err) {
      handleError(err, { module: 'Task', operation: '强制停止任务' })
      return false
    }
  },

  async updateOptions(id: number, options: Record<string, any>) {
    try {
      await api.updateTaskOptions(id, options)
      return true
    } catch (err) {
      handleError(err, { module: 'Task', operation: '更新任务配置' })
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
  },

  async create(schedule: any) {
    try {
      return await api.createSchedule(schedule)
    } catch (err) {
      handleError(err, { module: 'Schedule', operation: '创建定时任务' })
      return null
    }
  },

  async update(id: number, enabled: boolean, spec?: string) {
    try {
      return await api.updateSchedule(id, enabled, spec)
    } catch (err) {
      handleError(err, { module: 'Schedule', operation: '更新定时任务' })
      return false
    }
  },

  async delete(id: number) {
    try {
      return await api.deleteSchedule(id)
    } catch (err) {
      handleError(err, { module: 'Schedule', operation: '删除定时任务' })
      return false
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
      return { runs: [], total: 0, page, pageSize }
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
  },

  async deleteByTask(taskId: number) {
    try {
      await api.clearRunsByTask(taskId)
      return true
    } catch (err) {
      handleError(err, { module: 'Run', operation: '删除任务历史记录' })
      return false
    }
  },

  async getRunsByTask(taskId: number) {
    try {
      return await api.getRunsByTask(taskId)
    } catch (err) {
      handleError(err, { module: 'Run', operation: '获取任务历史记录' })
      return []
    }
  }
}

/**
 * Job APIs with error handling
 */
export const jobApi = {
  async list() {
    try {
      return await api.getActiveRuns()
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

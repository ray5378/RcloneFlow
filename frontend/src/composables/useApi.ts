import * as api from '../api'
import { handleError } from './useError'
import { t } from '../i18n'

export const taskApi = {
  async list() {
    try { return await api.listTasks() } catch (err) { handleError(err, { module: 'Task', operation: t('runtime.taskListFetch') }); return [] }
  },
  async bootstrap(page = 1, pageSize = 50) {
    try { return await api.getTaskBootstrap(page, pageSize) } catch (err) { handleError(err, { module: 'Task', operation: t('runtime.taskListFetch') }); return null }
  },
  async create(task: any) {
    try { return await api.createTask(task) } catch (err) { handleError(err, { module: 'Task', operation: t('runtime.taskCreate') }); return null }
  },
  async update(id: number, task: any) {
    try { return await api.updateTask(id, task) } catch (err) { handleError(err, { module: 'Task', operation: t('runtime.taskUpdate') }); return null }
  },
  async delete(id: number) {
    try { return await api.deleteTask(id) } catch (err) { handleError(err, { module: 'Task', operation: t('runtime.taskDelete') }); return false }
  },
  async run(id: number) {
    try { return await api.runTask(id) } catch (err) { handleError(err, { module: 'Task', operation: t('runtime.taskRun') }); return null }
  },
  async kill(id: number) {
    try { await api.killTask(id); return true } catch (err) { handleError(err, { module: 'Task', operation: t('runtime.taskKill') }); return false }
  },
  async updateOptions(id: number, options: Record<string, any>) {
    try { await api.updateTaskOptions(id, options); return true } catch (err) { handleError(err, { module: 'Task', operation: t('runtime.taskUpdateOptions') }); return false }
  }
}

export const remoteApi = {
  async list() {
    try { return await api.listRemotes() } catch (err) { handleError(err, { module: 'Remote', operation: t('runtime.remoteList') }); return { remotes: [] } }
  }
}

export const scheduleApi = {
  async list() {
    try { return await api.listSchedules() } catch (err) { handleError(err, { module: 'Schedule', operation: t('runtime.scheduleList') }); return [] }
  },
  async create(schedule: any) {
    try { return await api.createSchedule(schedule) } catch (err) { handleError(err, { module: 'Schedule', operation: t('runtime.scheduleCreate') }); return null }
  },
  async update(id: number, enabled: boolean, spec?: string) {
    try { return await api.updateSchedule(id, enabled, spec) } catch (err) { handleError(err, { module: 'Schedule', operation: t('runtime.scheduleUpdate') }); return false }
  },
  async delete(id: number) {
    try { return await api.deleteSchedule(id) } catch (err) { handleError(err, { module: 'Schedule', operation: t('runtime.scheduleDelete') }); return false }
  }
}

export const runApi = {
  async list(page = 1, pageSize = 50) {
    try { return await api.listRuns(page, pageSize) } catch (err) { handleError(err, { module: 'Run', operation: t('runtime.runList') }); return { runs: [], total: 0, page, pageSize } }
  },
  async get(id: number) {
    try { return await api.getRun(id) } catch (err) { handleError(err, { module: 'Run', operation: t('runtime.runGet') }); return null }
  },
  async getFiles(id: number, offset: number, limit: number) {
    try { return await api.getRunFiles(id, offset, limit) } catch (err) { handleError(err, { module: 'Run', operation: t('runtime.runFiles') }); return { files: [], total: 0 } }
  },
  async delete(id: number) {
    try { await api.clearRun(id); return true } catch (err) { handleError(err, { module: 'Run', operation: t('runtime.runDelete') }); return false }
  },
  async deleteAll() {
    try { await api.clearAllRuns(); return true } catch (err) { handleError(err, { module: 'Run', operation: t('runtime.runDeleteAll') }); return false }
  },
  async deleteByTask(taskId: number) {
    try { await api.clearRunsByTask(taskId); return true } catch (err) { handleError(err, { module: 'Run', operation: t('runtime.runDeleteByTask') }); return false }
  },
  async getRunsByTask(taskId: number) {
    try { return await api.getRunsByTask(taskId) } catch (err) { handleError(err, { module: 'Run', operation: t('runtime.runListByTask') }); return [] }
  }
}

export const jobApi = {
  async list() {
    try { return await api.getActiveRuns() } catch (err) { handleError(err, { module: 'ActiveRun', operation: t('runtime.activeRunList') }); return [] }
  }
}

export const activeTransferApi = {
  async get(taskId: number) {
    try { return await api.getActiveTransfer(taskId) } catch (err) { handleError(err, { module: 'ActiveTransfer', operation: 'overview' }); return null }
  },
  async getCompleted(taskId: number, offset = 0, limit = 100) {
    try { return await api.getActiveTransferCompleted(taskId, offset, limit) } catch (err) { handleError(err, { module: 'ActiveTransfer', operation: 'completed' }); return { total: 0, items: [] } }
  },
  async getPending(taskId: number, offset = 0, limit = 100) {
    try { return await api.getActiveTransferPending(taskId, offset, limit) } catch (err) { handleError(err, { module: 'ActiveTransfer', operation: 'pending' }); return { total: 0, items: [] } }
  }
}

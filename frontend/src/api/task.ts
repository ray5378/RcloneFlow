/**
 * Task API 模块
 * 对应后端 TaskService
 */
import { get, post, put, del } from './client'
import { getActiveRuns, getRuns, getJobStatus, stopJob } from './run'
import type { Task } from '../types'

/** 获取所有任务 */
export async function getTasks(): Promise<Task[]> {
  return get<Task[]>('/api/tasks')
}

/** 创建任务 */
export async function createTask(task: Omit<Task, 'id' | 'createdAt'>): Promise<Task> {
  return post<Task>('/api/tasks', task)
}

/** 更新任务 */
export async function updateTask(taskId: number, task: Omit<Task, 'id' | 'createdAt'>): Promise<void> {
  return put('/api/tasks', { id: taskId, task })
}

/** 运行任务 */
export async function runTask(taskId: number): Promise<{ jobId: number }> {
  return post<{ jobId: number }>(`/api/tasks/${taskId}/run`, {})
}

/** 删除任务 */
export async function deleteTask(taskId: number): Promise<void> {
  return del(`/api/tasks/${taskId}`)
}

/** 停止指定任务当前传输 */
export async function stopTaskTransfer(taskId: number): Promise<void> {
  const activeRuns = await getActiveRuns()
  const active = activeRuns.find(item => item.runRecord?.taskId === taskId && item.runRecord?.rcJobId)
  if (!active?.runRecord?.rcJobId) {
    throw new Error('当前任务没有正在运行的传输')
  }
  await stopJob(active.runRecord.rcJobId)
}

/** 获取指定任务实时进度 */
export async function getTaskProgress(taskId: number): Promise<Record<string, any>> {
  const activeRuns = await getActiveRuns()
  const active = activeRuns.find(item => item.runRecord?.taskId === taskId)
  if (active?.runRecord) {
    const rt = active.realtimeStatus || {}
    const derived = (active.derivedProgress && typeof active.derivedProgress === 'object') ? active.derivedProgress as Record<string, any> : {}
    const groupStats = (active.groupStats && typeof active.groupStats === 'object') ? active.groupStats as Record<string, any> : {}
    const globalStats = (active.globalStats && typeof active.globalStats === 'object') ? active.globalStats as Record<string, any> : {}
    const progress = (rt.progress && typeof rt.progress === 'object') ? rt.progress as Record<string, any> : {}
    const group = (progress.group && typeof progress.group === 'object') ? progress.group as Record<string, any> : {}
    const bytes = Number(derived.bytes ?? groupStats.bytes ?? progress.bytes ?? group.bytes ?? globalStats.bytes ?? 0)
    const totalBytes = Number(derived.totalBytes ?? derived.total_bytes ?? groupStats.totalBytes ?? groupStats.total_bytes ?? progress.totalBytes ?? progress.total_bytes ?? group.totalBytes ?? group.total_bytes ?? globalStats.totalBytes ?? globalStats.total_bytes ?? 0)
    const speed = Number(derived.speed ?? groupStats.speed ?? progress.speed ?? group.speed ?? globalStats.speed ?? 0)
    const eta = derived.eta ?? groupStats.eta ?? progress.eta ?? group.eta ?? globalStats.eta ?? null
    const percentage = Number(derived.percentage ?? groupStats.percentage ?? progress.percentage ?? group.percentage ?? (totalBytes > 0 ? (bytes / totalBytes) * 100 : 0))

    return {
      taskId,
      jobId: active.runRecord.rcJobId,
      status: active.runRecord.status || (rt.finished ? 'finished' : 'running'),
      bytes,
      totalBytes,
      speed: speed ? `${(speed / 1024 / 1024).toFixed(2)} MB/s` : '',
      eta,
      percentage,
      anomaly: derived.anomaly || '',
      anomalyMessage: derived.anomalyMessage || '',
      currentFileSize: Number(derived.currentFileSize || 0),
      error: (rt.error as string) || active.runRecord.error || '',
      raw: rt,
    }
  }

  const runs = await getRuns()
  const latest = (runs || [])
    .filter(run => run.taskId === taskId)
    .sort((a, b) => new Date(b.startedAt || 0).getTime() - new Date(a.startedAt || 0).getTime())[0]

  if (!latest) {
    return { taskId, status: 'no_runs', percentage: 0 }
  }

  if (latest.rcJobId) {
    try {
      const status = await getJobStatus(latest.rcJobId)
      return {
        taskId,
        jobId: latest.rcJobId,
        status: status.finished ? (status.success ? 'finished' : 'failed') : 'running',
        percentage: Number(status.progress || 0),
        error: status.error || latest.error || '',
        raw: status,
      }
    } catch {
      // ignore and fallback to latest run record
    }
  }

  return {
    taskId,
    jobId: latest.rcJobId,
    status: latest.status || 'unknown',
    percentage: 0,
    error: latest.error || '',
    raw: latest,
  }
}

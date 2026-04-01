/**
 * Schedule API 模块
 * 对应后端 ScheduleService
 */
import type { Schedule } from '../types'

const BASE = '/api'

async function api<T>(path: string, options?: RequestInit): Promise<T> {
  const res = await fetch(BASE + path, {
    headers: { 'Content-Type': 'application/json' },
    ...options,
  })
  const data = await res.json().catch(() => ({}))
  if (!res.ok) throw new Error((data as { error?: string }).error || res.statusText)
  return data as T
}

/** 获取所有定时任务 */
export async function getSchedules(): Promise<Schedule[]> {
  return api<Schedule[]>('/schedules')
}

/** 创建定时任务 */
export async function createSchedule(schedule: Omit<Schedule, 'id' | 'createdAt'>): Promise<Schedule> {
  return api<Schedule>('/schedules', {
    method: 'POST',
    body: JSON.stringify(schedule),
  })
}

/** 删除定时任务 */
export async function deleteSchedule(scheduleId: number): Promise<void> {
  return api(`/schedules/${scheduleId}`, { method: 'DELETE' })
}

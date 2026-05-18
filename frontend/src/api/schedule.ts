/**
 * Schedule API 模块
 * 对应后端 ScheduleService
 */
import { get, post, del, put } from './client'
import type { Schedule } from '../types'

/** 获取所有定时任务 */
export async function getSchedules(): Promise<Schedule[]> {
  return get<Schedule[]>('/api/schedules')
}

/** 创建定时任务 */
export async function createSchedule(schedule: Omit<Schedule, 'id' | 'createdAt'>): Promise<Schedule> {
  return post<Schedule>('/api/schedules', schedule)
}

/** 更新定时任务(启用/禁用/规则) */
export async function updateSchedule(scheduleId: number, enabled: boolean, spec?: string): Promise<void> {
  const body: any = { enabled }
  if (spec && spec.trim()) body.spec = spec
  return put<void>(`/api/schedules/${scheduleId}`, body)
}

/** 删除定时任务 */
export async function deleteSchedule(scheduleId: number): Promise<void> {
  return del(`/api/schedules/${scheduleId}`)
}

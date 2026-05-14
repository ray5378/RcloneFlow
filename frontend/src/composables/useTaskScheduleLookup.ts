import { computed, type Ref } from 'vue'
import type { Schedule } from '../types'

export function useTaskScheduleLookup(schedules: Ref<Schedule[]>) {
  const scheduleByTaskId = computed(() => {
    const index = new Map<number, Schedule>()
    for (const schedule of schedules.value || []) {
      const taskId = Number(schedule.taskId)
      if (taskId > 0) index.set(taskId, schedule)
    }
    return index
  })

  function getScheduleByTaskId(taskId: number): Schedule | undefined {
    return scheduleByTaskId.value.get(Number(taskId))
  }

  return {
    getScheduleByTaskId,
  }
}

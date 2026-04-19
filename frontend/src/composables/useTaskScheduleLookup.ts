import type { Ref } from 'vue'

export function useTaskScheduleLookup(schedules: Ref<any[]>) {
  function getScheduleByTaskId(taskId: number) {
    return schedules.value.find((s: any) => s.taskId === taskId)
  }

  return {
    getScheduleByTaskId,
  }
}

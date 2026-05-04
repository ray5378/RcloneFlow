import { computed, type Ref } from 'vue'

export function useTaskScheduleLookup(schedules: Ref<any[]>) {
  const scheduleByTaskId = computed(() => {
    const index = new Map<number, any>()
    for (const s of schedules.value || []) {
      const taskId = Number((s as any)?.taskId)
      if (taskId > 0) index.set(taskId, s)
    }
    return index
  })

  function getScheduleByTaskId(taskId: number) {
    return scheduleByTaskId.value.get(Number(taskId))
  }

  return {
    getScheduleByTaskId,
  }
}

import { describe, it, expect } from 'vitest'
import { ref } from 'vue'
import { useTaskScheduleLookup } from './useTaskScheduleLookup'

describe('useTaskScheduleLookup', () => {
  it('should get schedule by task id', () => {
    const schedules = ref([
      { id: 1, taskId: 42, enabled: true, spec: '0 12 * * *' } as any,
      { id: 2, taskId: 99, enabled: false, spec: '0 6 * * *' } as any,
    ])
    const { getScheduleByTaskId } = useTaskScheduleLookup(schedules)
    expect(getScheduleByTaskId(42)?.id).toBe(1)
    expect(getScheduleByTaskId(99)?.enabled).toBe(false)
    expect(getScheduleByTaskId(123)).toBeUndefined()
  })

  it('should handle empty schedules', () => {
    const schedules = ref([])
    const { getScheduleByTaskId } = useTaskScheduleLookup(schedules)
    expect(getScheduleByTaskId(1)).toBeUndefined()
  })

  it('should handle null schedules', () => {
    const schedules = ref(null as any)
    const { getScheduleByTaskId } = useTaskScheduleLookup(schedules)
    expect(getScheduleByTaskId(1)).toBeUndefined()
  })
})

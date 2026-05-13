import { describe, it, expect } from 'vitest'
import { ref } from 'vue'
import { useTaskScheduleLookup } from './useTaskScheduleLookup'

describe('useTaskScheduleLookup', () => {
  it('builds lookup by positive taskId and coerces numeric input', () => {
    const schedules = ref<any[]>([
      { id: 1, taskId: 11, spec: 'a' },
      { id: 2, taskId: '12', spec: 'b' },
      { id: 3, taskId: 0, spec: 'ignored-zero' },
      { id: 4, taskId: -1, spec: 'ignored-negative' },
      { id: 5, taskId: 'abc', spec: 'ignored-nan' },
    ])

    const api = useTaskScheduleLookup(schedules)

    expect(api.getScheduleByTaskId(11)).toEqual({ id: 1, taskId: 11, spec: 'a' })
    expect(api.getScheduleByTaskId(12)).toEqual({ id: 2, taskId: '12', spec: 'b' })
    expect(api.getScheduleByTaskId(0)).toBeUndefined()
    expect(api.getScheduleByTaskId(-1)).toBeUndefined()
    expect(api.getScheduleByTaskId(999)).toBeUndefined()
  })

  it('reacts to schedule list changes and latest task mapping wins', () => {
    const schedules = ref<any[]>([{ id: 1, taskId: 21, spec: 'old' }])
    const api = useTaskScheduleLookup(schedules)

    expect(api.getScheduleByTaskId(21)?.spec).toBe('old')

    schedules.value = [
      { id: 2, taskId: 21, spec: 'new' },
      { id: 3, taskId: 22, spec: 'other' },
    ]

    expect(api.getScheduleByTaskId(21)?.spec).toBe('new')
    expect(api.getScheduleByTaskId(22)?.spec).toBe('other')
  })
})

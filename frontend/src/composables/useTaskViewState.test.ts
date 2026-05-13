import { describe, it, expect } from 'vitest'
import { useTaskViewState } from './useTaskViewState'

describe('useTaskViewState', () => {
  it('initializes reactive view state', () => {
    const api = useTaskViewState()
    expect(api.tasks.value).toEqual([])
    expect(api.schedules.value).toEqual([])
    expect(api.runs.value).toEqual([])
    expect(api.runsTotal.value).toBe(0)
    expect(api.taskRuns.value).toEqual([])
    expect(api.runsPage.value).toBe(1)
    expect(api.runsPageSize).toBe(50)
    expect(api.jumpPage.value).toBe(1)
    expect(api.remotes.value).toEqual([])
    expect(api.currentModule.value).toBe('tasks')
    expect(api.historyFilterTaskId.value).toBeNull()
    expect(api.historyStatusFilter.value).toBe('all')
  })
})

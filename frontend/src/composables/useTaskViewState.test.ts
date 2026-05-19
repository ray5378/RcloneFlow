import { describe, it, expect } from 'vitest'
import { useTaskViewState } from './useTaskViewState'

describe('useTaskViewState', () => {
  it('should initialize with default values', () => {
    const state = useTaskViewState()
    expect(state.tasks.value).toEqual([])
    expect(state.schedules.value).toEqual([])
    expect(state.runs.value).toEqual([])
    expect(state.runsTotal.value).toBe(0)
    expect(state.taskRuns.value).toEqual([])
    expect(state.runsPage.value).toBe(1)
    expect(state.runsPageSize).toBe(50)
    expect(state.jumpPage.value).toBe(1)
    expect(state.remotes.value).toEqual([])
    expect(state.currentModule.value).toBe('tasks')
    expect(state.historyFilterTaskId.value).toBeNull()
    expect(state.historyStatusFilter.value).toBe('all')
  })
})

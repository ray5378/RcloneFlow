import { describe, it, expect } from 'vitest'
import { useTaskViewRuntimeState } from './useTaskViewRuntimeState'

describe('useTaskViewRuntimeState', () => {
  it('should initialize with default values', () => {
    const state = useTaskViewRuntimeState()

    expect(state.activeRuns.value).toEqual([])
    expect(state.globalStats.value).toEqual({})
    expect(state.showGlobalStatsModal.value).toBe(false)
    expect(state.lastNonDecreasingTotalsByTask.value).toEqual({})
    expect(state.STUCK_MS).toBe(25000)

    expect(state.tasks.value).toEqual([])
    expect(state.schedules.value).toEqual([])
    expect(state.runs.value).toEqual([])
    expect(state.runsTotal.value).toBe(0)
    expect(state.runsPage.value).toBe(1)
    expect(state.runsPageSize).toBe(50)
    expect(state.currentModule.value).toBe('tasks')
    expect(state.historyFilterTaskId.value).toBeNull()
    expect(state.historyStatusFilter.value).toBe('all')
  })

  it('should have activeRunLookup methods', () => {
    const state = useTaskViewRuntimeState()

    expect(state.activeRunLookup).toHaveProperty('getActiveRunByTaskId')
    expect(state.activeRunLookup).toHaveProperty('getActiveProgressByTaskId')
    expect(state.activeRunLookup).toHaveProperty('getActiveProgressTextByTaskId')
  })
})

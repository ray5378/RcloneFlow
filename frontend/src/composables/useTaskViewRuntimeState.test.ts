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
  })

  it('should have activeRunLookup methods', () => {
    const state = useTaskViewRuntimeState()

    expect(state.activeRunLookup).toHaveProperty('getActiveRunByTaskId')
    expect(state.activeRunLookup).toHaveProperty('getActiveProgressByTaskId')
    expect(state.activeRunLookup).toHaveProperty('getActiveProgressTextByTaskId')
  })
})

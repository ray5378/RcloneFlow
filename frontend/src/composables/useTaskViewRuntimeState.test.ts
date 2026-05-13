import { describe, it, expect } from 'vitest'
import { useTaskViewRuntimeState } from './useTaskViewRuntimeState'

describe('useTaskViewRuntimeState', () => {
  it('initializes runtime state and lookup wrapper', () => {
    const api = useTaskViewRuntimeState()
    expect(api.activeRuns.value).toEqual([])
    expect(api.globalStats.value).toEqual({})
    expect(api.showGlobalStatsModal.value).toBe(false)
    expect(api.lastNonDecreasingTotalsByTask.value).toEqual({})
    expect(api.STUCK_MS).toBe(25000)
    expect(typeof api.activeRunLookup.getActiveRunByTaskId).toBe('function')
  })
})

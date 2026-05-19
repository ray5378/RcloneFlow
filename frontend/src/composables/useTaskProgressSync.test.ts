import { describe, it, expect, vi } from 'vitest'
import { ref } from 'vue'
import { useTaskProgressSync } from './useTaskProgressSync'

describe('useTaskProgressSync', () => {
  const createOptions = () => ({
    runs: ref([]),
    activeRuns: ref([]),
    activeRunLookup: { getActiveRunByTaskId: vi.fn() },
    loadData: vi.fn().mockResolvedValue(undefined),
    loadActiveRuns: vi.fn().mockResolvedValue(undefined),
  })

  it('should return expected methods', () => {
    const options = createOptions()
    const sync = useTaskProgressSync(options)

    expect(sync).toHaveProperty('getRunProgressFromSummary')
    expect(sync).toHaveProperty('getRealtimeProgressByRun')
    expect(sync).toHaveProperty('getRunningProgressByRun')
    expect(sync).toHaveProperty('getTaskCardProgressByTask')
    expect(sync).toHaveProperty('getRunningProgressByTask')
    expect(sync).toHaveProperty('formatBps')
    expect(sync).toHaveProperty('calcEtaFromAvg')
    expect(sync).toHaveProperty('triggerAutoRefresh')
  })

  it('getRunningProgressByTask should return null for non-existent task', () => {
    const options = createOptions()
    const { getRunningProgressByTask } = useTaskProgressSync(options)

    const result = getRunningProgressByTask(999)
    expect(result).toBeNull()
  })

  it('getTaskCardProgressByTask should return null for non-existent task', () => {
    const options = createOptions()
    const { getTaskCardProgressByTask } = useTaskProgressSync(options)

    const result = getTaskCardProgressByTask(999)
    expect(result).toBeNull()
  })

  it('getRunProgressFromSummary should handle empty summary', () => {
    const options = createOptions()
    const { getRunProgressFromSummary } = useTaskProgressSync(options)

    const result = getRunProgressFromSummary({})
    expect(result).toBeNull()
  })

  it('getRealtimeProgressByRun should return null for non-existent run', () => {
    const options = createOptions()
    const { getRealtimeProgressByRun } = useTaskProgressSync(options)

    const result = getRealtimeProgressByRun({ id: 999 })
    expect(result).toBeNull()
  })

  it('triggerAutoRefresh should be a function', () => {
    const options = createOptions()
    const { triggerAutoRefresh } = useTaskProgressSync(options)

    expect(typeof triggerAutoRefresh).toBe('function')
  })
})

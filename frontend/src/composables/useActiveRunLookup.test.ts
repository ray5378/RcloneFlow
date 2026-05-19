import { describe, it, expect, vi } from 'vitest'
import { ref } from 'vue'
import { useActiveRunLookup } from './useActiveRunLookup'

vi.mock('../components/task/runningHint', () => ({
  getActiveProgress: (active: any) => active?.progress ?? 0,
  getActiveProgressText: (active: any) => active?.progressText ?? '',
}))

describe('useActiveRunLookup.ts', () => {
  it('should find active run by task ID', () => {
    const activeRuns = ref([
      { runRecord: { taskId: 1 }, progress: 50 },
      { runRecord: { taskId: 2 }, progress: 75 },
    ])

    const { getActiveRunByTaskId } = useActiveRunLookup(activeRuns)

    expect(getActiveRunByTaskId(1)?.progress).toBe(50)
    expect(getActiveRunByTaskId(2)?.progress).toBe(75)
    expect(getActiveRunByTaskId(99)).toBeUndefined()
  })

  it('should handle alternative taskId fields', () => {
    const activeRuns = ref([
      { taskId: 10 },
      { taskID: 20 },
      { task_id: 30 },
    ])

    const { getActiveRunByTaskId } = useActiveRunLookup(activeRuns)

    expect(getActiveRunByTaskId(10)).toBeDefined()
    expect(getActiveRunByTaskId(20)).toBeDefined()
    expect(getActiveRunByTaskId(30)).toBeDefined()
  })

  it('should handle empty activeRuns', () => {
    const activeRuns = ref([])
    const { getActiveRunByTaskId } = useActiveRunLookup(activeRuns)
    expect(getActiveRunByTaskId(1)).toBeUndefined()
  })

  it('should handle null activeRuns', () => {
    const activeRuns = ref(null as any)
    const { getActiveRunByTaskId } = useActiveRunLookup(activeRuns)
    expect(getActiveRunByTaskId(1)).toBeUndefined()
  })

  it('should get active progress by task ID', () => {
    const activeRuns = ref([
      { runRecord: { taskId: 1 }, progress: 42 },
    ])

    const { getActiveProgressByTaskId } = useActiveRunLookup(activeRuns)
    expect(getActiveProgressByTaskId(1)).toBe(42)
  })

  it('should get active progress text by task ID', () => {
    const activeRuns = ref([
      { runRecord: { taskId: 1 }, progressText: '50%' },
    ])

    const { getActiveProgressTextByTaskId } = useActiveRunLookup(activeRuns)
    expect(getActiveProgressTextByTaskId(1)).toBe('50%')
  })
})

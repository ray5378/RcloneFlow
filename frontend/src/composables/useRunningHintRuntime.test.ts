import { describe, it, expect, vi } from 'vitest'
import { ref } from 'vue'
import { useRunningHintRuntime } from './useRunningHintRuntime'

describe('useRunningHintRuntime', () => {
  it('should return expected properties', () => {
    const activeRuns = ref([])
    const openRunLog = vi.fn()

    const runtime = useRunningHintRuntime(activeRuns, openRunLog)

    expect(runtime).toHaveProperty('runningHintVisible')
    expect(runtime).toHaveProperty('runningHintRun')
    expect(runtime).toHaveProperty('runningHintPhaseText')
    expect(runtime).toHaveProperty('runningHintProgressText')
    expect(runtime).toHaveProperty('openRunningHint')
    expect(runtime).toHaveProperty('closeRunningHint')
    expect(runtime).toHaveProperty('openRunningHintLog')
  })

  it('should open and close hint', () => {
    const activeRuns = ref([{ id: 1, status: 'running' }])
    const openRunLog = vi.fn()

    const { runningHintVisible, runningHintRun, openRunningHint, closeRunningHint } = useRunningHintRuntime(activeRuns, openRunLog)

    expect(runningHintVisible.value).toBe(false)

    openRunningHint()
    expect(runningHintVisible.value).toBe(true)
    expect(runningHintRun.value).not.toBeNull()

    closeRunningHint()
    expect(runningHintVisible.value).toBe(false)
  })
})

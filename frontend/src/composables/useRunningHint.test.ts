import { describe, it, expect, vi } from 'vitest'
import { ref } from 'vue'
import { useRunningHint } from './useRunningHint'

vi.mock('../components/task/runningHint', () => ({
  getActiveProgress: (active: any) => active ? { phase: 'transferring' } : null,
  getActiveProgressText: (active: any) => active ? '50%' : '-',
}))

describe('useRunningHint', () => {
  const openRunLog = vi.fn()

  it('should initialize with default values', () => {
    const activeRuns = ref([])
    const { visible, run, phaseText, progressText } = useRunningHint(activeRuns, openRunLog)
    expect(visible.value).toBe(false)
    expect(run.value).toBeNull()
    expect(phaseText.value).toBe('-')
    expect(progressText.value).toBe('-')
  })

  it('should open and close', () => {
    const activeRuns = ref([])
    const { visible, run, open, close } = useRunningHint(activeRuns, openRunLog)
    open({ taskId: 1 })
    expect(visible.value).toBe(true)
    expect(run.value.taskId).toBe(1)
    close()
    expect(visible.value).toBe(false)
  })

  it('should open log and close', () => {
    const activeRuns = ref([])
    const { open, openLog, visible } = useRunningHint(activeRuns, openRunLog)
    open({ taskId: 1 })
    openLog()
    expect(openRunLog).toHaveBeenCalledWith({ taskId: 1 })
    expect(visible.value).toBe(false)
  })

  it('should find active run by task id', () => {
    const activeRuns = ref([{ taskId: 5, progress: {} }])
    const { open, phaseText } = useRunningHint(activeRuns, openRunLog)
    open({ taskId: 5 })
    expect(phaseText.value).toBe('transferring')
  })
})

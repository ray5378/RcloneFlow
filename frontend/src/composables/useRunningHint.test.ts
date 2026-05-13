import { describe, it, expect, vi } from 'vitest'
import { ref } from 'vue'
import { useRunningHint } from './useRunningHint'

vi.mock('../components/task/runningHint', () => ({
  getActiveProgress: vi.fn((active: any) => active?.progress || null),
  getActiveProgressText: vi.fn((active: any) => active?.progress ? `progress:${active.progress.phase || 'none'}` : '-'),
  getRunningHintDebug: vi.fn((active: any) => active?.debug || { checkText: '-', progressLine: '-', progressJson: '-' }),
}))

describe('useRunningHint', () => {
  it('indexes active runs by task id and exposes computed state', () => {
    const activeRuns = ref<any[]>([
      { taskId: 1, progress: { phase: 'copying' }, debug: { checkText: 'ok', progressLine: 'L1', progressJson: '{}' } },
      { runRecord: { taskId: 2 }, progress: { phase: 'queued' }, debug: { checkText: 'ok2', progressLine: 'L2', progressJson: '{2}' } },
    ])
    const openRunLog = vi.fn()
    const api = useRunningHint(activeRuns, openRunLog, true)

    api.open({ taskId: 2, name: 'task-2' })

    expect(api.visible.value).toBe(true)
    expect(api.run.value).toEqual({ taskId: 2, name: 'task-2' })
    expect(api.phaseText.value).toBe('queued')
    expect(api.progressText.value).toBe('progress:queued')
    expect(api.debugInfo.value.progressLine).toBe('L2')
  })

  it('returns safe defaults when task has no active run', () => {
    const api = useRunningHint(ref([]), vi.fn(), false)
    api.open({ taskId: 999 })

    expect(api.phaseText.value).toBe('-')
    expect(api.progressText.value).toBe('-')
    expect(api.debugInfo.value).toEqual({ checkText: '-', progressLine: '-', progressJson: '-' })
  })

  it('toggles debug only when enabled', () => {
    const disabled = useRunningHint(ref([]), vi.fn(), false)
    disabled.toggleDebug()
    expect(disabled.debugOpen.value).toBe(false)

    const enabled = useRunningHint(ref([]), vi.fn(), true)
    enabled.toggleDebug()
    expect(enabled.debugOpen.value).toBe(true)
    enabled.toggleDebug()
    expect(enabled.debugOpen.value).toBe(false)
  })

  it('openLog delegates then closes modal and debug state', () => {
    const openRunLog = vi.fn()
    const api = useRunningHint(ref([]), openRunLog, true)
    api.open({ taskId: 1 })
    api.toggleDebug()

    api.openLog()

    expect(openRunLog).toHaveBeenCalledWith({ taskId: 1 })
    expect(api.visible.value).toBe(false)
    expect(api.debugOpen.value).toBe(false)
  })
})

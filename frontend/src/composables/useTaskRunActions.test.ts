import { describe, it, expect, vi } from 'vitest'
import { useTaskRunActions } from './useTaskRunActions'

vi.mock('../i18n', () => ({
  t: (key: string) => key,
}))

describe('useTaskRunActions', () => {
  function makeOptions() {
    return {
      loadData: vi.fn().mockResolvedValue(undefined),
      loadActiveRuns: vi.fn().mockResolvedValue(undefined),
      showToast: vi.fn(),
      taskApi: {
        run: vi.fn().mockResolvedValue({ started: true }),
        kill: vi.fn().mockResolvedValue(undefined),
      },
    }
  }

  it('should run task successfully', async () => {
    vi.useFakeTimers()
    const opts = makeOptions()
    const { runTask, runningTaskId } = useTaskRunActions(opts)
    const result = await runTask(1)
    expect(result?.started).toBe(true)
    expect(opts.loadData).toHaveBeenCalled()
    expect(runningTaskId.value).toBe(1)
    vi.runAllTimers()
    expect(runningTaskId.value).toBeNull()
    vi.useRealTimers()
  })

  it('should block when another task is running', async () => {
    const opts = makeOptions()
    const { runTask } = useTaskRunActions(opts)
    await runTask(1)
    await runTask(2)
    expect(opts.showToast).toHaveBeenCalledWith('runtime.singletonBlocked', 'error')
  })

  it('should handle run failure', async () => {
    const opts = makeOptions()
    opts.taskApi.run.mockRejectedValueOnce(new Error('run failed'))
    const { runTask, runningTaskId } = useTaskRunActions(opts)
    await runTask(1)
    expect(runningTaskId.value).toBeNull()
    expect(opts.showToast).toHaveBeenCalledWith('run failed', 'error')
  })

  it('should handle started=false response', async () => {
    const opts = makeOptions()
    opts.taskApi.run.mockResolvedValueOnce({ started: false, message: 'blocked' })
    const { runTask, runningTaskId } = useTaskRunActions(opts)
    const result = await runTask(1)
    expect(result?.started).toBe(false)
    expect(runningTaskId.value).toBeNull()
    expect(opts.showToast).toHaveBeenCalledWith('blocked', 'error')
  })

  it('should stop task', async () => {
    vi.useFakeTimers()
    const opts = makeOptions()
    const { stopTaskAny, stoppedTaskId } = useTaskRunActions(opts)
    await stopTaskAny(1)
    expect(stoppedTaskId.value).toBe(1)
    expect(opts.taskApi.kill).toHaveBeenCalledWith(1)
    vi.runAllTimers()
    expect(stoppedTaskId.value).toBeNull()
    vi.useRealTimers()
  })

  it('should handle stop failure', async () => {
    const opts = makeOptions()
    opts.taskApi.kill.mockRejectedValueOnce(new Error('kill failed'))
    const { stopTaskAny, stoppedTaskId } = useTaskRunActions(opts)
    await stopTaskAny(1)
    expect(stoppedTaskId.value).toBeNull()
    expect(opts.showToast).toHaveBeenCalledWith('kill failed', 'error')
  })
})

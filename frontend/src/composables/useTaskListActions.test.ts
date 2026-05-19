import { describe, it, expect, vi } from 'vitest'
import { ref } from 'vue'
import { useTaskListActions } from './useTaskListActions'

vi.mock('../i18n', () => ({
  t: (key: string) => key,
}))

describe('useTaskListActions', () => {
  function makeOptions() {
    return {
      openMenuId: ref<number | null>(null),
      historyFilterTaskId: ref<number | null>(null),
      schedules: ref([]),
      loadData: vi.fn().mockResolvedValue(undefined),
      showConfirm: vi.fn((_title, _msg, cb) => cb()),
      showToast: vi.fn(),
      clearAllRuns: vi.fn().mockResolvedValue(undefined),
      taskApi: { delete: vi.fn().mockResolvedValue(true) },
      scheduleApi: { delete: vi.fn().mockResolvedValue(undefined), update: vi.fn().mockResolvedValue(undefined) },
    }
  }

  it('should get schedule by task id', () => {
    const opts = makeOptions()
    opts.schedules.value = [{ id: 1, taskId: 42, enabled: true, spec: '0 12 * * *' } as any]
    const { getScheduleByTaskId } = useTaskListActions(opts)
    expect(getScheduleByTaskId(42)).toBeDefined()
    expect(getScheduleByTaskId(99)).toBeUndefined()
  })

  it('should delete task', async () => {
    const opts = makeOptions()
    opts.openMenuId.value = 1
    const { deleteTask } = useTaskListActions(opts)
    await deleteTask(42)
    expect(opts.taskApi.delete).toHaveBeenCalledWith(42)
    expect(opts.openMenuId.value).toBeNull()
    expect(opts.loadData).toHaveBeenCalled()
  })

  it('should clear all runs with confirm when task selected', () => {
    const opts = makeOptions()
    opts.historyFilterTaskId.value = 5
    const { clearAllRunsWithConfirm } = useTaskListActions(opts)
    clearAllRunsWithConfirm()
    expect(opts.showConfirm).toHaveBeenCalled()
  })

  it('should show toast when no task selected for clear', () => {
    const opts = makeOptions()
    opts.historyFilterTaskId.value = null
    const { clearAllRunsWithConfirm } = useTaskListActions(opts)
    clearAllRunsWithConfirm()
    expect(opts.showToast).toHaveBeenCalledWith('runtime.chooseTaskFirst', 'error')
  })
})

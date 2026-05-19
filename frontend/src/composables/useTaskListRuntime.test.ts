import { describe, it, expect, vi } from 'vitest'
import { ref } from 'vue'
import { useTaskListRuntime } from './useTaskListRuntime'

vi.mock('../i18n', () => ({
  t: (key: string) => key,
}))

describe('useTaskListRuntime', () => {
  function makeOptions() {
    return {
      openMenuId: ref<number | null>(null),
      historyFilterTaskId: ref<number | null>(null),
      schedules: ref([]),
      loadData: vi.fn().mockResolvedValue(undefined),
      loadActiveRuns: vi.fn().mockResolvedValue(undefined),
      showConfirm: vi.fn((_t, _m, cb) => cb()),
      showToast: vi.fn(),
      clearAllRuns: vi.fn().mockResolvedValue(undefined),
      currentModule: ref<'history' | 'add' | 'tasks'>('tasks'),
      remotes: ref([]),
      remoteApi: { list: vi.fn().mockResolvedValue({ remotes: [] }) },
      resetTaskFormForCreate: vi.fn(),
      resetTaskPathBrowse: vi.fn(),
      getScheduleByTaskId: vi.fn(),
      fillTaskFormForEdit: vi.fn(),
      restoreTaskPathBrowse: vi.fn().mockResolvedValue(undefined),
      taskApi: {
        delete: vi.fn().mockResolvedValue(true),
        run: vi.fn().mockResolvedValue({ started: true }),
        kill: vi.fn().mockResolvedValue(undefined),
        updateSortOrders: vi.fn().mockResolvedValue(true),
      },
      scheduleApi: {
        delete: vi.fn().mockResolvedValue(undefined),
        update: vi.fn().mockResolvedValue(undefined),
      },
    }
  }

  it('should expose all methods', () => {
    const opts = makeOptions()
    const runtime = useTaskListRuntime(opts)
    expect(typeof runtime.deleteTask).toBe('function')
    expect(typeof runtime.deleteSchedule).toBe('function')
    expect(typeof runtime.clearAllRunsWithConfirm).toBe('function')
    expect(typeof runtime.runTask).toBe('function')
    expect(typeof runtime.stopTaskAny).toBe('function')
    expect(typeof runtime.goToAddTask).toBe('function')
    expect(typeof runtime.editTask).toBe('function')
    expect(typeof runtime.saveTaskSortOrders).toBe('function')
    expect(runtime.runningTaskId).toBeDefined()
    expect(runtime.stoppedTaskId).toBeDefined()
  })

  it('should save task sort orders', async () => {
    const opts = makeOptions()
    const { saveTaskSortOrders } = useTaskListRuntime(opts)
    const result = await saveTaskSortOrders({ 1: 10, 2: 20 })
    expect(result).toBe(true)
    expect(opts.taskApi.updateSortOrders).toHaveBeenCalledWith({ 1: 10, 2: 20 }, undefined)
    expect(opts.loadData).toHaveBeenCalled()
    expect(opts.showToast).toHaveBeenCalledWith('runtime.taskSortSave', 'success')
  })

  it('should return false on sort order save failure', async () => {
    const opts = makeOptions()
    opts.taskApi.updateSortOrders.mockResolvedValueOnce(false)
    const { saveTaskSortOrders } = useTaskListRuntime(opts)
    const result = await saveTaskSortOrders({ 1: 10 })
    expect(result).toBe(false)
  })
})

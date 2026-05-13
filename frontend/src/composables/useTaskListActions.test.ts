import { describe, it, expect, vi } from 'vitest'
import { ref, nextTick } from 'vue'

vi.mock('../i18n', () => ({
  t: (key: string) => ({
    'common.delete': '删除',
    'runtime.deleteTaskConfirm': '确认删除任务',
    'schedule.deleteConfirm': '确认删除定时任务',
    'runtime.chooseTaskFirst': '请先选择任务',
    'runtime.deleteAllHistory': '删除全部历史',
    'runtime.deleteAllHistoryConfirm': '确认删除全部历史',
  }[key] || key),
}))

import { useTaskListActions } from './useTaskListActions'

describe('useTaskListActions', () => {
  it('deletes task via confirm callback and reloads data', async () => {
    const openMenuId = ref<number | null>(7)
    const loadData = vi.fn(async () => {})
    let confirmHandler: (() => void) | null = null
    const showConfirm = vi.fn((_title, _msg, onConfirm) => { confirmHandler = onConfirm })
    const api = useTaskListActions({
      openMenuId,
      historyFilterTaskId: ref(null),
      schedules: ref([]),
      loadData,
      showConfirm,
      showToast: vi.fn(),
      clearAllRuns: vi.fn(async () => true),
      taskApi: { delete: vi.fn(async () => true) },
      scheduleApi: { delete: vi.fn(async () => {}), update: vi.fn(async () => {}) },
    })

    await api.deleteTask(3)
    expect(showConfirm).toHaveBeenCalled()
    await confirmHandler?.()

    expect(openMenuId.value).toBeNull()
    expect(loadData).toHaveBeenCalled()
  })

  it('toggles schedule and clears highlight after timeout', async () => {
    const originalSetTimeout = globalThis.setTimeout
    const timeoutSpy = vi.fn((fn: any) => { fn(); return 1 as any })
    ;(globalThis as any).setTimeout = timeoutSpy

    try {
      const loadData = vi.fn(async () => {})
      const api = useTaskListActions({
        openMenuId: ref(null),
        historyFilterTaskId: ref(null),
        schedules: ref([{ id: 10, taskId: 4, enabled: true } as any]),
        loadData,
        showConfirm: vi.fn(),
        showToast: vi.fn(),
        clearAllRuns: vi.fn(async () => true),
        taskApi: { delete: vi.fn(async () => true) },
        scheduleApi: { delete: vi.fn(async () => {}), update: vi.fn(async () => {}) },
      })

      await api.toggleSchedule(4)
      expect(loadData).toHaveBeenCalled()
      expect(api.scheduleToggledTaskId.value).toBeNull()
    } finally {
      ;(globalThis as any).setTimeout = originalSetTimeout
    }
  })

  it('deletes schedule only when confirm returns true', async () => {
    const originalConfirm = globalThis.confirm
    ;(globalThis as any).confirm = vi.fn(() => true)

    try {
      const loadData = vi.fn(async () => {})
      const scheduleDelete = vi.fn(async () => {})
      const api = useTaskListActions({
        openMenuId: ref(null),
        historyFilterTaskId: ref(null),
        schedules: ref([]),
        loadData,
        showConfirm: vi.fn(),
        showToast: vi.fn(),
        clearAllRuns: vi.fn(async () => true),
        taskApi: { delete: vi.fn(async () => true) },
        scheduleApi: { delete: scheduleDelete, update: vi.fn(async () => {}) },
      })

      await api.deleteSchedule(5)
      expect(scheduleDelete).toHaveBeenCalledWith(5)
      expect(loadData).toHaveBeenCalled()
    } finally {
      ;(globalThis as any).confirm = originalConfirm
    }
  })

  it('shows toast when clearing all runs without selected task and confirms otherwise', async () => {
    const showToast = vi.fn()
    let confirmHandler: (() => void) | null = null
    const showConfirm = vi.fn((_title, _msg, onConfirm) => { confirmHandler = onConfirm })
    const clearAllRuns = vi.fn(async () => true)

    const noTask = useTaskListActions({
      openMenuId: ref(null),
      historyFilterTaskId: ref(null),
      schedules: ref([]),
      loadData: vi.fn(async () => {}),
      showConfirm,
      showToast,
      clearAllRuns,
      taskApi: { delete: vi.fn(async () => true) },
      scheduleApi: { delete: vi.fn(async () => {}), update: vi.fn(async () => {}) },
    })
    noTask.clearAllRunsWithConfirm()
    expect(showToast).toHaveBeenCalledWith('请先选择任务', 'error')

    const yesTask = useTaskListActions({
      openMenuId: ref(null),
      historyFilterTaskId: ref(8),
      schedules: ref([]),
      loadData: vi.fn(async () => {}),
      showConfirm,
      showToast,
      clearAllRuns,
      taskApi: { delete: vi.fn(async () => true) },
      scheduleApi: { delete: vi.fn(async () => {}), update: vi.fn(async () => {}) },
    })
    yesTask.clearAllRunsWithConfirm()
    expect(showConfirm).toHaveBeenCalled()
    await confirmHandler?.()
    expect(clearAllRuns).toHaveBeenCalled()
  })
})

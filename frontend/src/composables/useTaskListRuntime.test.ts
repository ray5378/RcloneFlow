import { describe, it, expect, vi } from 'vitest'
import { ref } from 'vue'

const actionsApi = {
  deleteTask: vi.fn(async () => {}),
  toggleSchedule: vi.fn(async () => {}),
  deleteSchedule: vi.fn(async () => {}),
  clearAllRunsWithConfirm: vi.fn(),
  scheduleToggledTaskId: ref(9),
}
const runApi = {
  runningTaskId: ref(1),
  stoppedTaskId: ref(2),
  stopTaskAny: vi.fn(async () => {}),
  runTask: vi.fn(async () => ({ started: true })),
}
const formApi = {
  goToAddTask: vi.fn(async () => {}),
  editTask: vi.fn(),
}

vi.mock('./useTaskListActions', () => ({ useTaskListActions: () => actionsApi }))
vi.mock('./useTaskRunActions', () => ({ useTaskRunActions: () => runApi }))
vi.mock('./useTaskFormEntry', () => ({ useTaskFormEntry: () => formApi }))

import { useTaskListRuntime } from './useTaskListRuntime'

describe('useTaskListRuntime', () => {
  it('wires list/run/form child composables', async () => {
    const api = useTaskListRuntime({
      openMenuId: ref(null),
      historyFilterTaskId: ref(null),
      schedules: ref([]),
      loadData: vi.fn(async () => {}),
      loadActiveRuns: vi.fn(async () => {}),
      showConfirm: vi.fn(),
      showToast: vi.fn(),
      clearAllRuns: vi.fn(async () => true),
      currentModule: ref('tasks'),
      remotes: ref([]),
      remoteApi: { list: vi.fn(async () => ({ remotes: [] })) },
      resetTaskFormForCreate: vi.fn(),
      resetTaskPathBrowse: vi.fn(),
      getScheduleByTaskId: vi.fn(),
      fillTaskFormForEdit: vi.fn(),
      restoreTaskPathBrowse: vi.fn(async () => {}),
      taskApi: { delete: vi.fn(), run: vi.fn(), kill: vi.fn() },
      scheduleApi: { delete: vi.fn(), update: vi.fn() },
    })

    expect(api.runningTaskId.value).toBe(1)
    expect(api.stoppedTaskId.value).toBe(2)
    expect(api.scheduleToggledTaskId.value).toBe(9)

    await api.deleteTask(1)
    await api.toggleSchedule(2)
    await api.deleteSchedule(3)
    api.clearAllRunsWithConfirm()
    await api.stopTaskAny(4)
    await api.runTask(5)
    await api.goToAddTask()
    api.editTask({ id: 6 } as any)

    expect(actionsApi.deleteTask).toHaveBeenCalledWith(1)
    expect(actionsApi.toggleSchedule).toHaveBeenCalledWith(2)
    expect(actionsApi.deleteSchedule).toHaveBeenCalledWith(3)
    expect(actionsApi.clearAllRunsWithConfirm).toHaveBeenCalled()
    expect(runApi.stopTaskAny).toHaveBeenCalledWith(4)
    expect(runApi.runTask).toHaveBeenCalledWith(5)
    expect(formApi.goToAddTask).toHaveBeenCalled()
    expect(formApi.editTask).toHaveBeenCalledWith({ id: 6 })
  })
})

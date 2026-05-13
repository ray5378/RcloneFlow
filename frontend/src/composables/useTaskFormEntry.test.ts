import { describe, it, expect, vi } from 'vitest'
import { ref } from 'vue'
import { useTaskFormEntry } from './useTaskFormEntry'

describe('useTaskFormEntry', () => {
  it('loads remotes and enters add-task module', async () => {
    const currentModule = ref<'history' | 'add' | 'tasks'>('tasks')
    const openMenuId = ref<number | null>(5)
    const remotes = ref<string[]>([])
    const remoteApi = { list: vi.fn(async () => ({ remotes: ['a', 'b'] })) }
    const resetTaskFormForCreate = vi.fn()
    const resetTaskPathBrowse = vi.fn()
    const api = useTaskFormEntry({
      currentModule,
      openMenuId,
      remotes,
      remoteApi,
      resetTaskFormForCreate,
      resetTaskPathBrowse,
      getScheduleByTaskId: vi.fn(),
      fillTaskFormForEdit: vi.fn(),
      restoreTaskPathBrowse: vi.fn(async () => {}),
    })

    await api.goToAddTask()

    expect(remoteApi.list).toHaveBeenCalled()
    expect(remotes.value).toEqual(['a', 'b'])
    expect(resetTaskFormForCreate).toHaveBeenCalled()
    expect(resetTaskPathBrowse).toHaveBeenCalled()
    expect(currentModule.value).toBe('add')
    expect(openMenuId.value).toBeNull()
  })

  it('fills form from task and schedule when editing', () => {
    const currentModule = ref<'history' | 'add' | 'tasks'>('tasks')
    const openMenuId = ref<number | null>(9)
    const fillTaskFormForEdit = vi.fn()
    const restoreTaskPathBrowse = vi.fn(async () => {})
    const getScheduleByTaskId = vi.fn(() => ({ spec: '*/5 * * * *' }))
    const api = useTaskFormEntry({
      currentModule,
      openMenuId,
      remotes: ref([]),
      remoteApi: { list: vi.fn(async () => ({ remotes: [] })) },
      resetTaskFormForCreate: vi.fn(),
      resetTaskPathBrowse: vi.fn(),
      getScheduleByTaskId,
      fillTaskFormForEdit,
      restoreTaskPathBrowse,
    })
    const task = { id: 7, name: 'demo' } as any

    api.editTask(task)

    expect(getScheduleByTaskId).toHaveBeenCalledWith(7)
    expect(fillTaskFormForEdit).toHaveBeenCalledWith(task, '*/5 * * * *')
    expect(restoreTaskPathBrowse).toHaveBeenCalledWith(task)
    expect(currentModule.value).toBe('add')
    expect(openMenuId.value).toBeNull()
  })
})

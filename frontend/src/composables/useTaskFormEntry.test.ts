import { describe, it, expect, vi } from 'vitest'
import { ref } from 'vue'
import { useTaskFormEntry } from './useTaskFormEntry'

describe('useTaskFormEntry', () => {
  function makeOptions() {
    return {
      currentModule: ref<'history' | 'add' | 'tasks'>('tasks'),
      openMenuId: ref<number | null>(null),
      remotes: ref<string[]>([]),
      remoteApi: { list: vi.fn().mockResolvedValue({ remotes: ['remote1', 'remote2'] }) },
      resetTaskFormForCreate: vi.fn(),
      resetTaskPathBrowse: vi.fn(),
      fillTaskFormForEdit: vi.fn(),
      restoreTaskPathBrowse: vi.fn().mockResolvedValue(undefined),
    }
  }

  it('should go to add task', async () => {
    const opts = makeOptions()
    opts.openMenuId.value = 1
    const { goToAddTask } = useTaskFormEntry(opts)
    await goToAddTask()
    expect(opts.currentModule.value).toBe('add')
    expect(opts.openMenuId.value).toBeNull()
    expect(opts.remotes.value).toEqual(['remote1', 'remote2'])
    expect(opts.resetTaskFormForCreate).toHaveBeenCalled()
    expect(opts.resetTaskPathBrowse).toHaveBeenCalled()
  })

  it('should edit task', async () => {
    const opts = makeOptions()
    const { editTask } = useTaskFormEntry(opts)
    const task = { id: 1, name: 'test' } as any
    await editTask(task)
    expect(opts.fillTaskFormForEdit).toHaveBeenCalledWith(task)
    expect(opts.restoreTaskPathBrowse).toHaveBeenCalledWith(task)
    expect(opts.currentModule.value).toBe('add')
  })
})

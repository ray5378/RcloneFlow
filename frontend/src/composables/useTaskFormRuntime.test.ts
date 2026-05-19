import { describe, it, expect, vi } from 'vitest'
import { ref } from 'vue'
import { useTaskFormRuntime } from './useTaskFormRuntime'

vi.mock('../api', () => ({
  listPath: vi.fn().mockResolvedValue({ items: [] }),
}))

vi.mock('../i18n', () => ({
  t: (key: string) => key,
}))

describe('useTaskFormRuntime', () => {
  function makeOptions() {
    return {
      schedules: ref([]),
      currentModule: ref<'history' | 'add' | 'tasks'>('tasks'),
      normalizeTaskOptions: vi.fn((raw) => raw || {}),
      loadData: vi.fn().mockResolvedValue(undefined),
      taskApi: {
        create: vi.fn().mockResolvedValue({ id: 1 }),
        update: vi.fn().mockResolvedValue({}),
      },
      showToast: vi.fn(),
      parseRcloneCommand: vi.fn(),
    }
  }

  it('should expose all expected properties', () => {
    const opts = makeOptions()
    const runtime = useTaskFormRuntime(opts)
    expect(runtime.createForm).toBeDefined()
    expect(runtime.commandMode).toBeDefined()
    expect(runtime.commandText).toBeDefined()
    expect(runtime.editingTask).toBeDefined()
    expect(runtime.showAdvancedOptions).toBeDefined()
    expect(runtime.creatingState).toBeDefined()
    expect(typeof runtime.createTask).toBe('function')
    expect(typeof runtime.resetTaskFormForCreate).toBe('function')
    expect(typeof runtime.fillTaskFormForEdit).toBe('function')
  })
})

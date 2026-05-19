import { describe, it, expect, vi } from 'vitest'
import { ref } from 'vue'
import { useTaskFormOrchestrator } from './useTaskFormOrchestrator'

vi.mock('../i18n', () => ({
  t: (key: string) => key,
}))

describe('useTaskFormOrchestrator', () => {
  function makeOptions() {
    return {
      createForm: ref({
        name: 'test',
        mode: 'copy' as const,
        sourceRemote: 'src',
        sourcePath: '/src',
        targetRemote: 'dst',
        targetPath: '/dst',
        options: {},
      }),
      editingTask: ref(null),
      currentModule: ref<'history' | 'add' | 'tasks'>('tasks'),
      normalizeTaskOptions: vi.fn((raw) => raw || {}),
      loadData: vi.fn().mockResolvedValue(undefined),
      taskApi: {
        create: vi.fn().mockResolvedValue({ id: 1 }),
        update: vi.fn().mockResolvedValue({}),
      },
      commandMode: ref(false),
      commandText: ref(''),
      parseRcloneCommand: vi.fn(),
      showToast: vi.fn(),
    }
  }

  it('should expose creatingState and methods', () => {
    const opts = makeOptions()
    const { creatingState, handleTaskFormDoneClick, validateTaskForm, executeTaskFormSubmit, validateTaskFormBeforeSubmit, runTaskFormFlow, createTask } = useTaskFormOrchestrator(opts)
    expect(creatingState.value).toBe('idle')
    expect(typeof handleTaskFormDoneClick).toBe('function')
    expect(typeof validateTaskForm).toBe('function')
    expect(typeof executeTaskFormSubmit).toBe('function')
    expect(typeof validateTaskFormBeforeSubmit).toBe('function')
    expect(typeof runTaskFormFlow).toBe('function')
    expect(typeof createTask).toBe('function')
  })

  it('should handle done click', () => {
    const opts = makeOptions()
    const { handleTaskFormDoneClick, creatingState } = useTaskFormOrchestrator(opts)
    creatingState.value = 'done'
    expect(handleTaskFormDoneClick()).toBe(true)
  })

  it('should validate form', () => {
    const opts = makeOptions()
    const { validateTaskForm } = useTaskFormOrchestrator(opts)
    expect(validateTaskForm()).toBe('')
  })

  it('should execute submit', async () => {
    const opts = makeOptions()
    const { executeTaskFormSubmit } = useTaskFormOrchestrator(opts)
    const result = await executeTaskFormSubmit()
    expect(result).toBe('')
  })

  it('should create task', async () => {
    const opts = makeOptions()
    const { createTask } = useTaskFormOrchestrator(opts)
    await createTask()
    expect(opts.taskApi.create).toHaveBeenCalled()
  })
})

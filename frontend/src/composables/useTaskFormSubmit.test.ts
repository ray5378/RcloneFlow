import { describe, it, expect, vi } from 'vitest'
import { ref } from 'vue'
import { useTaskFormSubmit } from './useTaskFormSubmit'

vi.mock('../i18n', () => ({
  t: (key: string) => key,
}))

describe('useTaskFormSubmit', () => {
  function makeOptions() {
    return {
      createForm: ref({
        name: 'test-task',
        mode: 'copy' as const,
        sourceRemote: 'src',
        sourcePath: '/src',
        targetRemote: 'dst',
        targetPath: '/dst',
        options: {},
      }),
      editingTask: ref(null),
      creatingState: ref<'idle' | 'loading' | 'done'>('idle'),
      currentModule: ref<'history' | 'add' | 'tasks'>('tasks'),
      normalizeTaskOptions: vi.fn((raw) => raw || {}),
      loadData: vi.fn().mockResolvedValue(undefined),
      showToast: vi.fn(),
      taskApi: {
        create: vi.fn().mockResolvedValue({ id: 1 }),
        update: vi.fn().mockResolvedValue({}),
      },
    }
  }

  it('should handle done click when state is done', () => {
    const opts = makeOptions()
    opts.creatingState.value = 'done'
    const { handleTaskFormDoneClick } = useTaskFormSubmit(opts)
    const result = handleTaskFormDoneClick()
    expect(result).toBe(true)
    expect(opts.creatingState.value).toBe('idle')
    expect(opts.currentModule.value).toBe('tasks')
  })

  it('should return false when state is not done', () => {
    const opts = makeOptions()
    opts.creatingState.value = 'idle'
    const { handleTaskFormDoneClick } = useTaskFormSubmit(opts)
    expect(handleTaskFormDoneClick()).toBe(false)
  })

  it('should validate form', () => {
    const opts = makeOptions()
    const { validateTaskForm } = useTaskFormSubmit(opts)
    expect(validateTaskForm()).toBe('')

    opts.createForm.value.name = ''
    expect(validateTaskForm()).toBe('runtime.enterTaskName')

    opts.createForm.value.name = 'test'
    opts.createForm.value.sourceRemote = ''
    expect(validateTaskForm()).toBe('runtime.chooseSourceTarget')
  })

  it('should build task payload', () => {
    const opts = makeOptions()
    const { buildTaskPayload } = useTaskFormSubmit(opts)
    const payload = buildTaskPayload()
    expect(payload.name).toBe('test-task')
    expect(payload.mode).toBe('copy')
    expect(payload.sourceRemote).toBe('src')
  })

  it('should submit create task', async () => {
    const opts = makeOptions()
    const { submitTaskForm } = useTaskFormSubmit(opts)
    const result = await submitTaskForm()
    expect(result.kind).toBe('create')
    expect(opts.taskApi.create).toHaveBeenCalled()
  })

  it('should submit update task', async () => {
    const opts = makeOptions()
    opts.editingTask.value = { id: 1 } as any
    const { submitTaskForm } = useTaskFormSubmit(opts)
    const result = await submitTaskForm()
    expect(result.kind).toBe('update')
    expect(opts.taskApi.update).toHaveBeenCalledWith(1, expect.any(Object))
  })

  it('should complete task form submit', async () => {
    const opts = makeOptions()
    const { completeTaskFormSubmit } = useTaskFormSubmit(opts)
    await completeTaskFormSubmit('create')
    expect(opts.editingTask.value).toBeNull()
    expect(opts.loadData).toHaveBeenCalled()
    expect(opts.showToast).toHaveBeenCalledWith('runtime.taskCreateSuccess', 'success')
    expect(opts.currentModule.value).toBe('tasks')
    expect(opts.creatingState.value).toBe('done')
  })

  it('should execute task form submit', async () => {
    const opts = makeOptions()
    const { executeTaskFormSubmit } = useTaskFormSubmit(opts)
    const result = await executeTaskFormSubmit()
    expect(result).toBe('')
    expect(opts.creatingState.value).toBe('done')
  })

  it('should handle submit error', async () => {
    const opts = makeOptions()
    opts.taskApi.create.mockRejectedValueOnce(new Error('create failed'))
    const { executeTaskFormSubmit } = useTaskFormSubmit(opts)
    const result = await executeTaskFormSubmit()
    expect(result).toBe('create failed')
    expect(opts.creatingState.value).toBe('idle')
  })
})

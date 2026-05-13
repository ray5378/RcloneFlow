import { describe, it, expect, vi } from 'vitest'
import { ref } from 'vue'
import { useTaskFormSubmit } from './useTaskFormSubmit'

function makeOptions() {
  const createForm = ref({
    name: 'task-1',
    mode: 'copy',
    sourceRemote: 'src',
    sourcePath: '/from',
    targetRemote: 'dst',
    targetPath: '/to',
    options: { transfers: 2 },
    enableSchedule: false,
    scheduleMinute: '05',
    scheduleHour: '*',
    scheduleDay: '1',
    scheduleMonth: '2',
    scheduleWeek: '3',
  })

  return {
    createForm,
    editingTask: ref<any>(null),
    creatingState: ref<'idle' | 'loading' | 'done'>('idle'),
    currentModule: ref<'history' | 'add' | 'tasks'>('add'),
    normalizeTaskOptions: vi.fn((raw) => ({ ...(raw || {}), normalized: true })),
    getScheduleByTaskId: vi.fn(),
    loadData: vi.fn(async () => {}),
    taskApi: {
      create: vi.fn(async (task) => ({ id: 77, ...task })),
      update: vi.fn(async () => ({})),
    },
    scheduleApi: {
      create: vi.fn(async () => ({})),
      update: vi.fn(async () => ({})),
    },
  }
}

describe('useTaskFormSubmit', () => {
  it('handles done click only when state is done', () => {
    const options = makeOptions()
    const api = useTaskFormSubmit(options)

    expect(api.handleTaskFormDoneClick()).toBe(false)
    expect(options.creatingState.value).toBe('idle')

    options.creatingState.value = 'done'
    options.currentModule.value = 'add'
    expect(api.handleTaskFormDoneClick()).toBe(true)
    expect(options.creatingState.value).toBe('idle')
    expect(options.currentModule.value).toBe('tasks')
  })

  it('validates required task fields and builds normalized payload/spec', () => {
    const options = makeOptions()
    const api = useTaskFormSubmit(options)

    options.createForm.value.name = ''
    expect(api.validateTaskForm()).toBeTruthy()

    options.createForm.value.name = 'task-1'
    options.createForm.value.sourceRemote = ''
    expect(api.validateTaskForm()).toBeTruthy()

    options.createForm.value.sourceRemote = 'src'
    expect(api.validateTaskForm()).toBe('')
    expect(api.buildTaskPayload()).toEqual({
      name: 'task-1',
      mode: 'copy',
      sourceRemote: 'src',
      sourcePath: '/from',
      targetRemote: 'dst',
      targetPath: '/to',
      options: { transfers: 2, normalized: true },
    })
    expect(api.buildScheduleSpec()).toBe('05|*|1|2|3')

    options.createForm.value.scheduleMinute = ''
    options.createForm.value.scheduleHour = ''
    options.createForm.value.scheduleDay = ''
    options.createForm.value.scheduleMonth = ''
    options.createForm.value.scheduleWeek = ''
    expect(api.buildScheduleSpec()).toBe('00|*|*|*|*')
  })

  it('creates a task without schedule when scheduling is disabled', async () => {
    const options = makeOptions()
    const api = useTaskFormSubmit(options)

    await expect(api.submitTaskForm()).resolves.toEqual({
      kind: 'create',
      task: expect.objectContaining({ id: 77, name: 'task-1' }),
    })
    expect(options.taskApi.create).toHaveBeenCalledWith({
      name: 'task-1',
      mode: 'copy',
      sourceRemote: 'src',
      sourcePath: '/from',
      targetRemote: 'dst',
      targetPath: '/to',
      options: { transfers: 2, normalized: true },
    })
    expect(options.scheduleApi.create).not.toHaveBeenCalled()
  })

  it('creates a task and schedule when scheduling is enabled', async () => {
    const options = makeOptions()
    options.createForm.value.enableSchedule = true
    const api = useTaskFormSubmit(options)

    const result = await api.submitTaskForm()
    expect(result.kind).toBe('create')
    expect(options.scheduleApi.create).toHaveBeenCalledWith({
      taskId: 77,
      spec: '05|*|1|2|3',
      enabled: true,
    })
  })

  it('updates task and existing schedule when editing with scheduling enabled', async () => {
    const options = makeOptions()
    options.editingTask.value = { id: 9, name: 'old' }
    options.createForm.value.enableSchedule = true
    options.getScheduleByTaskId.mockReturnValue({ id: 12 })
    const api = useTaskFormSubmit(options)

    await expect(api.submitTaskForm()).resolves.toEqual({ kind: 'update' })
    expect(options.taskApi.update).toHaveBeenCalledWith(9, expect.objectContaining({ name: 'task-1' }))
    expect(options.scheduleApi.update).toHaveBeenCalledWith(12, true, '05|*|1|2|3')
    expect(options.scheduleApi.create).not.toHaveBeenCalled()
  })

  it('updates task and creates schedule when editing without prior schedule', async () => {
    const options = makeOptions()
    options.editingTask.value = { id: 10, name: 'old' }
    options.createForm.value.enableSchedule = true
    options.getScheduleByTaskId.mockReturnValue(undefined)
    const api = useTaskFormSubmit(options)

    await api.submitTaskForm()
    expect(options.scheduleApi.create).toHaveBeenCalledWith({
      taskId: 10,
      spec: '05|*|1|2|3',
      enabled: true,
    })
  })

  it('disables prior schedule when editing and scheduling is turned off', async () => {
    const options = makeOptions()
    options.editingTask.value = { id: 11, name: 'old' }
    options.createForm.value.enableSchedule = false
    options.getScheduleByTaskId.mockReturnValue({ id: 13 })
    const api = useTaskFormSubmit(options)

    await api.submitTaskForm()
    expect(options.scheduleApi.update).toHaveBeenCalledWith(13, false)
  })

  it('completes submit flow and resets state on success', async () => {
    const options = makeOptions()
    options.creatingState.value = 'loading'
    options.editingTask.value = { id: 15, name: 'edit' }
    const api = useTaskFormSubmit(options)

    await api.completeTaskFormSubmit()
    expect(options.editingTask.value).toBe(null)
    expect(options.loadData).toHaveBeenCalledTimes(1)
    expect(options.currentModule.value).toBe('add')
    expect(options.creatingState.value).toBe('done')

    options.creatingState.value = 'loading'
    api.resetTaskFormSubmitState()
    expect(options.creatingState.value).toBe('idle')
  })

  it('executes submit flow and surfaces errors while restoring idle state', async () => {
    const options = makeOptions()
    const api = useTaskFormSubmit(options)

    await expect(api.executeTaskFormSubmit()).resolves.toBe('')
    expect(options.creatingState.value).toBe('done')
    expect(options.currentModule.value).toBe('add')

    const failed = makeOptions()
    failed.taskApi.create = vi.fn(async () => {
      throw new Error('boom')
    })
    const failedApi = useTaskFormSubmit(failed)

    await expect(failedApi.executeTaskFormSubmit()).resolves.toBe('boom')
    expect(failed.creatingState.value).toBe('idle')
  })
})

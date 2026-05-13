import { describe, it, expect, vi } from 'vitest'
import { ref } from 'vue'

const mocks = vi.hoisted(() => {
  const submitApi = {
    handleTaskFormDoneClick: vi.fn(() => true),
    validateTaskForm: vi.fn(() => ''),
    executeTaskFormSubmit: vi.fn(async () => ''),
  }
  const prepareApi = {
    validateTaskFormBeforeSubmit: vi.fn(() => ''),
  }
  const flowApi = {
    runTaskFormFlow: vi.fn(async () => ''),
  }
  const entrySubmitApi = {
    createTask: vi.fn(async () => true),
  }

  return {
    submitApi,
    prepareApi,
    flowApi,
    entrySubmitApi,
    submitSpy: vi.fn(() => submitApi),
    prepareSpy: vi.fn(() => prepareApi),
    flowSpy: vi.fn(() => flowApi),
    entrySubmitSpy: vi.fn(() => entrySubmitApi),
  }
})

vi.mock('./useTaskFormSubmit', () => ({ useTaskFormSubmit: mocks.submitSpy }))
vi.mock('./useTaskFormPrepare', () => ({ useTaskFormPrepare: mocks.prepareSpy }))
vi.mock('./useTaskFormFlow', () => ({ useTaskFormFlow: mocks.flowSpy }))
vi.mock('./useTaskFormEntrySubmit', () => ({ useTaskFormEntrySubmit: mocks.entrySubmitSpy }))

import { useTaskFormOrchestrator } from './useTaskFormOrchestrator'

describe('useTaskFormOrchestrator', () => {
  it('wires create state and all child composables together', async () => {
    const options = {
      createForm: ref({ name: 'demo', options: {} }),
      editingTask: ref(null),
      currentModule: ref<'history' | 'add' | 'tasks'>('add'),
      normalizeTaskOptions: vi.fn((raw) => raw || {}),
      getScheduleByTaskId: vi.fn(),
      loadData: vi.fn(async () => {}),
      taskApi: { create: vi.fn(), update: vi.fn() },
      scheduleApi: { create: vi.fn(), update: vi.fn() },
      commandMode: ref(false),
      commandText: ref(''),
      parseRcloneCommand: vi.fn(),
      showToast: vi.fn(),
    }

    const api = useTaskFormOrchestrator(options)

    expect(api.creatingState.value).toBe('idle')
    expect(mocks.submitSpy).toHaveBeenCalledTimes(1)
    expect(mocks.prepareSpy).toHaveBeenCalledTimes(1)
    expect(mocks.flowSpy).toHaveBeenCalledTimes(1)
    expect(mocks.entrySubmitSpy).toHaveBeenCalledTimes(1)

    const submitArgs = mocks.submitSpy.mock.calls[0][0]
    expect(submitArgs.createForm).toBe(options.createForm)
    expect(submitArgs.editingTask).toBe(options.editingTask)
    expect(submitArgs.currentModule).toBe(options.currentModule)
    expect(submitArgs.normalizeTaskOptions).toBe(options.normalizeTaskOptions)
    expect(submitArgs.getScheduleByTaskId).toBe(options.getScheduleByTaskId)
    expect(submitArgs.loadData).toBe(options.loadData)
    expect(submitArgs.taskApi).toBe(options.taskApi)
    expect(submitArgs.scheduleApi).toBe(options.scheduleApi)
    expect(submitArgs.creatingState).toBe(api.creatingState)

    const prepareArgs = mocks.prepareSpy.mock.calls[0][0]
    expect(prepareArgs.createForm).toBe(options.createForm)
    expect(prepareArgs.commandMode).toBe(options.commandMode)
    expect(prepareArgs.commandText).toBe(options.commandText)
    expect(prepareArgs.normalizeTaskOptions).toBe(options.normalizeTaskOptions)
    expect(prepareArgs.parseRcloneCommand).toBe(options.parseRcloneCommand)
    expect(prepareArgs.validateTaskForm).toBe(mocks.submitApi.validateTaskForm)

    const flowArgs = mocks.flowSpy.mock.calls[0][0]
    expect(flowArgs.handleTaskFormDoneClick).toBe(mocks.submitApi.handleTaskFormDoneClick)
    expect(flowArgs.validateTaskFormBeforeSubmit).toBe(mocks.prepareApi.validateTaskFormBeforeSubmit)
    expect(flowArgs.executeTaskFormSubmit).toBe(mocks.submitApi.executeTaskFormSubmit)

    const entryArgs = mocks.entrySubmitSpy.mock.calls[0][0]
    expect(entryArgs.runTaskFormFlow).toBe(mocks.flowApi.runTaskFormFlow)
    expect(entryArgs.showToast).toBe(options.showToast)

    expect(api.handleTaskFormDoneClick).toBe(mocks.submitApi.handleTaskFormDoneClick)
    expect(api.validateTaskForm).toBe(mocks.submitApi.validateTaskForm)
    expect(api.executeTaskFormSubmit).toBe(mocks.submitApi.executeTaskFormSubmit)
    expect(api.validateTaskFormBeforeSubmit).toBe(mocks.prepareApi.validateTaskFormBeforeSubmit)
    expect(api.runTaskFormFlow).toBe(mocks.flowApi.runTaskFormFlow)
    expect(api.createTask).toBe(mocks.entrySubmitApi.createTask)

    expect(api.handleTaskFormDoneClick()).toBe(true)
    expect(api.validateTaskForm()).toBe('')
    await expect(api.executeTaskFormSubmit()).resolves.toBe('')
    await expect(api.runTaskFormFlow()).resolves.toBe('')
    await expect(api.createTask()).resolves.toBe(true)
  })
})

import { describe, it, expect, vi } from 'vitest'

const mocks = vi.hoisted(() => {
  const stateApi = {
    createForm: { value: { name: 'task', sourceRemote: 'src', targetRemote: 'dst' } },
    commandMode: { value: false },
    commandText: { value: 'rclone copy a: b:' },
    editingTask: { value: null },
    showAdvancedOptions: { value: true },
    resetTaskFormForCreate: vi.fn(),
    fillTaskFormForEdit: vi.fn(),
  }

  const scheduleLookupApi = {
    getScheduleByTaskId: vi.fn((taskId: number) => ({ id: taskId + 100 })),
  }

  const orchestratorApi = {
    creatingState: { value: 'idle' as 'idle' | 'loading' | 'done' },
    handleTaskFormDoneClick: vi.fn(() => true),
    validateTaskForm: vi.fn(() => ''),
    executeTaskFormSubmit: vi.fn(async () => ''),
    validateTaskFormBeforeSubmit: vi.fn(() => ''),
    runTaskFormFlow: vi.fn(async () => ''),
    createTask: vi.fn(async () => true),
  }

  const pathBrowseApi = {
    sourcePathOptions: { value: [{ Path: '/src' }] },
    targetPathOptions: { value: [{ Path: '/dst' }] },
    showSourcePathInput: { value: false },
    showTargetPathInput: { value: true },
    sourceCurrentPath: { value: '/from' },
    targetCurrentPath: { value: '/to' },
    sourceBreadcrumbs: { value: [{ name: 'src:', path: '' }] },
    targetBreadcrumbs: { value: [{ name: 'dst:', path: '' }] },
    setShowSourcePathInput: vi.fn(),
    setShowTargetPathInput: vi.fn(),
    resetTaskPathBrowse: vi.fn(),
    restoreTaskPathBrowse: vi.fn(async () => {}),
    onSourceRemoteChange: vi.fn(),
    onTargetRemoteChange: vi.fn(),
    onSourceBreadcrumbClick: vi.fn(),
    onTargetBreadcrumbClick: vi.fn(),
    loadSourcePath: vi.fn(async () => {}),
    loadTargetPath: vi.fn(async () => {}),
    onSourceClick: vi.fn(),
    onSourceArrow: vi.fn(),
    onTargetClick: vi.fn(),
    onTargetArrow: vi.fn(),
  }

  return {
    stateApi,
    scheduleLookupApi,
    orchestratorApi,
    pathBrowseApi,
    useTaskFormStateSpy: vi.fn(() => stateApi),
    useTaskScheduleLookupSpy: vi.fn(() => scheduleLookupApi),
    useTaskFormOrchestratorSpy: vi.fn(() => orchestratorApi),
    useTaskPathBrowseSpy: vi.fn(() => pathBrowseApi),
    listPathSpy: vi.fn(),
  }
})

vi.mock('../api', () => ({
  listPath: mocks.listPathSpy,
}))
vi.mock('./useTaskFormState', () => ({
  useTaskFormState: mocks.useTaskFormStateSpy,
}))
vi.mock('./useTaskScheduleLookup', () => ({
  useTaskScheduleLookup: mocks.useTaskScheduleLookupSpy,
}))
vi.mock('./useTaskFormOrchestrator', () => ({
  useTaskFormOrchestrator: mocks.useTaskFormOrchestratorSpy,
}))
vi.mock('./useTaskPathBrowse', () => ({
  useTaskPathBrowse: mocks.useTaskPathBrowseSpy,
}))

import { useTaskFormRuntime } from './useTaskFormRuntime'

describe('useTaskFormRuntime', () => {
  it('wires state, schedule lookup, orchestrator, and path browse into one runtime api', async () => {
    const options = {
      schedules: { value: [{ id: 1, taskId: 8 }] },
      currentModule: { value: 'add' as 'history' | 'add' | 'tasks' },
      normalizeTaskOptions: vi.fn((raw) => raw || {}),
      loadData: vi.fn(async () => {}),
      taskApi: { create: vi.fn(), update: vi.fn() },
      scheduleApi: { create: vi.fn(), update: vi.fn() },
      showToast: vi.fn(),
      parseRcloneCommand: vi.fn(),
    }

    const api = useTaskFormRuntime(options)

    expect(mocks.useTaskFormStateSpy).toHaveBeenCalledTimes(1)
    expect(mocks.useTaskScheduleLookupSpy).toHaveBeenCalledWith(options.schedules)
    expect(mocks.useTaskFormOrchestratorSpy).toHaveBeenCalledTimes(1)
    expect(mocks.useTaskPathBrowseSpy).toHaveBeenCalledTimes(1)

    const orchestratorArgs = mocks.useTaskFormOrchestratorSpy.mock.calls[0][0]
    expect(orchestratorArgs.createForm).toBe(mocks.stateApi.createForm)
    expect(orchestratorArgs.editingTask).toBe(mocks.stateApi.editingTask)
    expect(orchestratorArgs.currentModule).toBe(options.currentModule)
    expect(orchestratorArgs.normalizeTaskOptions).toBe(options.normalizeTaskOptions)
    expect(orchestratorArgs.getScheduleByTaskId).toBe(mocks.scheduleLookupApi.getScheduleByTaskId)
    expect(orchestratorArgs.loadData).toBe(options.loadData)
    expect(orchestratorArgs.taskApi).toBe(options.taskApi)
    expect(orchestratorArgs.scheduleApi).toBe(options.scheduleApi)
    expect(orchestratorArgs.commandMode).toBe(mocks.stateApi.commandMode)
    expect(orchestratorArgs.commandText).toBe(mocks.stateApi.commandText)
    expect(orchestratorArgs.parseRcloneCommand).toBe(options.parseRcloneCommand)
    expect(orchestratorArgs.showToast).toBe(options.showToast)

    const pathBrowseArgs = mocks.useTaskPathBrowseSpy.mock.calls[0][0]
    expect(pathBrowseArgs.createForm).toBe(mocks.stateApi.createForm)
    expect(pathBrowseArgs.listPath).toBe(mocks.listPathSpy)

    expect(api.createForm).toBe(mocks.stateApi.createForm)
    expect(api.commandMode).toBe(mocks.stateApi.commandMode)
    expect(api.commandText).toBe(mocks.stateApi.commandText)
    expect(api.editingTask).toBe(mocks.stateApi.editingTask)
    expect(api.showAdvancedOptions).toBe(mocks.stateApi.showAdvancedOptions)
    expect(api.resetTaskFormForCreate).toBe(mocks.stateApi.resetTaskFormForCreate)
    expect(api.fillTaskFormForEdit).toBe(mocks.stateApi.fillTaskFormForEdit)

    expect(api.getScheduleByTaskId).toBe(mocks.scheduleLookupApi.getScheduleByTaskId)
    expect(api.creatingState).toBe(mocks.orchestratorApi.creatingState)
    expect(api.handleTaskFormDoneClick).toBe(mocks.orchestratorApi.handleTaskFormDoneClick)
    expect(api.validateTaskForm).toBe(mocks.orchestratorApi.validateTaskForm)
    expect(api.executeTaskFormSubmit).toBe(mocks.orchestratorApi.executeTaskFormSubmit)
    expect(api.validateTaskFormBeforeSubmit).toBe(mocks.orchestratorApi.validateTaskFormBeforeSubmit)
    expect(api.runTaskFormFlow).toBe(mocks.orchestratorApi.runTaskFormFlow)
    expect(api.createTask).toBe(mocks.orchestratorApi.createTask)

    expect(api.sourcePathOptions).toBe(mocks.pathBrowseApi.sourcePathOptions)
    expect(api.targetPathOptions).toBe(mocks.pathBrowseApi.targetPathOptions)
    expect(api.showSourcePathInput).toBe(mocks.pathBrowseApi.showSourcePathInput)
    expect(api.showTargetPathInput).toBe(mocks.pathBrowseApi.showTargetPathInput)
    expect(api.sourceCurrentPath).toBe(mocks.pathBrowseApi.sourceCurrentPath)
    expect(api.targetCurrentPath).toBe(mocks.pathBrowseApi.targetCurrentPath)
    expect(api.sourceBreadcrumbs).toBe(mocks.pathBrowseApi.sourceBreadcrumbs)
    expect(api.targetBreadcrumbs).toBe(mocks.pathBrowseApi.targetBreadcrumbs)
    expect(api.setShowSourcePathInput).toBe(mocks.pathBrowseApi.setShowSourcePathInput)
    expect(api.setShowTargetPathInput).toBe(mocks.pathBrowseApi.setShowTargetPathInput)
    expect(api.resetTaskPathBrowse).toBe(mocks.pathBrowseApi.resetTaskPathBrowse)
    expect(api.restoreTaskPathBrowse).toBe(mocks.pathBrowseApi.restoreTaskPathBrowse)
    expect(api.onSourceRemoteChange).toBe(mocks.pathBrowseApi.onSourceRemoteChange)
    expect(api.onTargetRemoteChange).toBe(mocks.pathBrowseApi.onTargetRemoteChange)
    expect(api.onSourceBreadcrumbClick).toBe(mocks.pathBrowseApi.onSourceBreadcrumbClick)
    expect(api.onTargetBreadcrumbClick).toBe(mocks.pathBrowseApi.onTargetBreadcrumbClick)
    expect(api.loadSourcePath).toBe(mocks.pathBrowseApi.loadSourcePath)
    expect(api.loadTargetPath).toBe(mocks.pathBrowseApi.loadTargetPath)
    expect(api.onSourceClick).toBe(mocks.pathBrowseApi.onSourceClick)
    expect(api.onSourceArrow).toBe(mocks.pathBrowseApi.onSourceArrow)
    expect(api.onTargetClick).toBe(mocks.pathBrowseApi.onTargetClick)
    expect(api.onTargetArrow).toBe(mocks.pathBrowseApi.onTargetArrow)

    expect(api.getScheduleByTaskId(8)).toEqual({ id: 108 })
    expect(api.handleTaskFormDoneClick()).toBe(true)
    expect(api.validateTaskForm()).toBe('')
    await expect(api.executeTaskFormSubmit()).resolves.toBe('')
    await expect(api.runTaskFormFlow()).resolves.toBe('')
    await expect(api.createTask()).resolves.toBe(true)
  })
})

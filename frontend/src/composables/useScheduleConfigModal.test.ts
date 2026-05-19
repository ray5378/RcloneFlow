import { describe, it, expect, vi, beforeEach } from 'vitest'
import { useScheduleConfigModal } from './useScheduleConfigModal'

describe('useScheduleConfigModal', () => {
  const mockOptions = () => ({
    getScheduleByTaskId: vi.fn(),
    scheduleApi: {
      create: vi.fn().mockResolvedValue({}),
      update: vi.fn().mockResolvedValue({}),
    },
    loadData: vi.fn().mockResolvedValue(undefined),
    showToast: vi.fn(),
  })

  it('should initialize with closed modal', () => {
    const options = mockOptions()
    const { scheduleConfigVisible, scheduleConfigSaving } = useScheduleConfigModal(options)

    expect(scheduleConfigVisible.value).toBe(false)
    expect(scheduleConfigSaving.value).toBe(false)
  })

  it('should open modal for task', () => {
    const options = mockOptions()
    options.getScheduleByTaskId.mockReturnValue(null)
    const { scheduleConfigVisible, openScheduleConfigForTask } = useScheduleConfigModal(options)

    const task = { id: 1, name: 'test-task' }
    openScheduleConfigForTask(task as any)

    expect(scheduleConfigVisible.value).toBe(true)
  })

  it('should close modal when not saving', () => {
    const options = mockOptions()
    const { scheduleConfigVisible, openScheduleConfigForTask, closeScheduleConfig } = useScheduleConfigModal(options)

    openScheduleConfigForTask({ id: 1, name: 'test' } as any)
    expect(scheduleConfigVisible.value).toBe(true)

    closeScheduleConfig()
    expect(scheduleConfigVisible.value).toBe(false)
  })

  it('should not close modal when saving', () => {
    const options = mockOptions()
    const { scheduleConfigVisible, scheduleConfigSaving, openScheduleConfigForTask, closeScheduleConfig } = useScheduleConfigModal(options)

    openScheduleConfigForTask({ id: 1, name: 'test' } as any)
    scheduleConfigSaving.value = true

    closeScheduleConfig()
    expect(scheduleConfigVisible.value).toBe(true)
  })

  it('should create new schedule when saving with enable', async () => {
    const options = mockOptions()
    options.getScheduleByTaskId.mockReturnValue(null)
    const { scheduleConfigTask, openScheduleConfigForTask, saveScheduleConfig } = useScheduleConfigModal(options)

    openScheduleConfigForTask({ id: 1, name: 'test' } as any)
    await saveScheduleConfig({ enableSchedule: true, weekdays: [1], time: '12:00' } as any)

    expect(options.scheduleApi.create).toHaveBeenCalledWith({
      taskId: 1,
      spec: expect.any(String),
      enabled: true,
    })
    expect(options.showToast).toHaveBeenCalled()
  })

  it('should update existing schedule when saving with enable', async () => {
    const options = mockOptions()
    options.getScheduleByTaskId.mockReturnValue({ id: 10, spec: 'old', enabled: true })
    const { openScheduleConfigForTask, saveScheduleConfig } = useScheduleConfigModal(options)

    openScheduleConfigForTask({ id: 1, name: 'test' } as any)
    await saveScheduleConfig({ enableSchedule: true, weekdays: [1], time: '12:00' } as any)

    expect(options.scheduleApi.update).toHaveBeenCalledWith(10, true, expect.any(String))
  })

  it('should disable schedule when saving without enable', async () => {
    const options = mockOptions()
    options.getScheduleByTaskId.mockReturnValue({ id: 10, spec: 'old', enabled: true })
    const { openScheduleConfigForTask, saveScheduleConfig } = useScheduleConfigModal(options)

    openScheduleConfigForTask({ id: 1, name: 'test' } as any)
    await saveScheduleConfig({ enableSchedule: false } as any)

    expect(options.scheduleApi.update).toHaveBeenCalledWith(10, false)
  })

  it('should do nothing when task has no id', async () => {
    const options = mockOptions()
    const { openScheduleConfigForTask, saveScheduleConfig } = useScheduleConfigModal(options)

    openScheduleConfigForTask({ id: null, name: 'test' } as any)
    await saveScheduleConfig({ enableSchedule: true } as any)

    expect(options.scheduleApi.create).not.toHaveBeenCalled()
    expect(options.scheduleApi.update).not.toHaveBeenCalled()
  })
})

import { describe, it, expect, vi } from 'vitest'
import { useTaskViewAuxRuntime } from './useTaskViewAuxRuntime'

describe('useTaskViewAuxRuntime', () => {
  const mockOptions = () => ({
    loadData: vi.fn().mockResolvedValue(undefined),
    showToast: vi.fn(),
    taskApi: {
      updateOptions: vi.fn().mockResolvedValue(true),
    },
    getFinalSummary: vi.fn().mockReturnValue({}),
  })

  it('should return all expected properties', () => {
    const options = mockOptions()
    const aux = useTaskViewAuxRuntime(options)

    expect(aux).toHaveProperty('showWebhookModal')
    expect(aux).toHaveProperty('webhookForm')
    expect(aux).toHaveProperty('setWebhook')
    expect(aux).toHaveProperty('saveWebhook')
    expect(aux).toHaveProperty('testWebhook')
    expect(aux).toHaveProperty('showSingletonModal')
    expect(aux).toHaveProperty('singletonForm')
    expect(aux).toHaveProperty('setSingletonMode')
    expect(aux).toHaveProperty('saveSingleton')
    expect(aux).toHaveProperty('showLogModal')
    expect(aux).toHaveProperty('logModalTitle')
    expect(aux).toHaveProperty('logContent')
    expect(aux).toHaveProperty('openRunLog')
    expect(aux).toHaveProperty('openMenuId')
    expect(aux).toHaveProperty('confirmModal')
    expect(aux).toHaveProperty('toggleMenu')
    expect(aux).toHaveProperty('closeMenus')
    expect(aux).toHaveProperty('showConfirm')
    expect(aux).toHaveProperty('closeConfirm')
    expect(aux).toHaveProperty('confirmAndClose')
    expect(aux).toHaveProperty('formatTime')
    expect(aux).toHaveProperty('formatDuration')
    expect(aux).toHaveProperty('getRunDurationText')
    expect(aux).toHaveProperty('getStatusClass')
    expect(aux).toHaveProperty('getStatusText')
    expect(aux).toHaveProperty('formatScheduleSpec')
  })
})

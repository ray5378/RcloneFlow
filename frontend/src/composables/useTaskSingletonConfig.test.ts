import { describe, it, expect, vi } from 'vitest'
import { useTaskSingletonConfig } from './useTaskSingletonConfig'

describe('useTaskSingletonConfig', () => {
  const mockOptions = () => ({
    taskApi: {
      updateOptions: vi.fn().mockResolvedValue(true),
    },
    loadData: vi.fn().mockResolvedValue(undefined),
    showToast: vi.fn(),
  })

  it('should initialize with closed modal', () => {
    const options = mockOptions()
    const { showSingletonModal, singletonForm } = useTaskSingletonConfig(options)

    expect(showSingletonModal.value).toBe(false)
    expect(singletonForm.value.taskId).toBeNull()
    expect(singletonForm.value.singletonEnabled).toBe(false)
  })

  it('should open modal and parse options', () => {
    const options = mockOptions()
    const { showSingletonModal, singletonForm, setSingletonMode } = useTaskSingletonConfig(options)

    setSingletonMode({ id: 1, options: { singletonMode: true } })

    expect(showSingletonModal.value).toBe(true)
    expect(singletonForm.value.taskId).toBe(1)
    expect(singletonForm.value.singletonEnabled).toBe(true)
  })

  it('should handle string options', () => {
    const options = mockOptions()
    const { singletonForm, setSingletonMode } = useTaskSingletonConfig(options)

    setSingletonMode({ id: 2, options: JSON.stringify({ singletonMode: true }) })

    expect(singletonForm.value.taskId).toBe(2)
    expect(singletonForm.value.singletonEnabled).toBe(true)
  })

  it('should handle invalid options', () => {
    const options = mockOptions()
    const { singletonForm, setSingletonMode } = useTaskSingletonConfig(options)

    setSingletonMode({ id: 3, options: 'invalid-json' })

    expect(singletonForm.value.singletonEnabled).toBe(false)
  })

  it('should save singleton settings', async () => {
    const options = mockOptions()
    const { setSingletonMode, saveSingleton } = useTaskSingletonConfig(options)

    setSingletonMode({ id: 1, options: { singletonMode: true } })
    await saveSingleton()

    expect(options.taskApi.updateOptions).toHaveBeenCalledWith(1, { singletonMode: true })
    expect(options.loadData).toHaveBeenCalled()
  })

  it('should show error on save failure', async () => {
    const options = mockOptions()
    options.taskApi.updateOptions.mockRejectedValue(new Error('API error'))
    const { setSingletonMode, saveSingleton } = useTaskSingletonConfig(options)

    setSingletonMode({ id: 1, options: {} })
    await saveSingleton()

    expect(options.showToast).toHaveBeenCalledWith('API error', 'error')
  })

  it('should do nothing when taskId is null', async () => {
    const options = mockOptions()
    const { saveSingleton, showSingletonModal } = useTaskSingletonConfig(options)

    showSingletonModal.value = true
    await saveSingleton()

    expect(options.taskApi.updateOptions).not.toHaveBeenCalled()
    expect(showSingletonModal.value).toBe(false)
  })
})

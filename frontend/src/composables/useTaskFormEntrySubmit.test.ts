import { describe, it, expect, vi } from 'vitest'
import { useTaskFormEntrySubmit } from './useTaskFormEntrySubmit'

describe('useTaskFormEntrySubmit', () => {
  it('should create task successfully', async () => {
    const showToast = vi.fn()
    const runTaskFormFlow = vi.fn().mockResolvedValue('')
    const { createTask } = useTaskFormEntrySubmit({ runTaskFormFlow, showToast })
    await createTask()
    expect(showToast).not.toHaveBeenCalled()
  })

  it('should show error on failure', async () => {
    const showToast = vi.fn()
    const runTaskFormFlow = vi.fn().mockResolvedValue('validation error')
    const { createTask } = useTaskFormEntrySubmit({ runTaskFormFlow, showToast })
    await createTask()
    expect(showToast).toHaveBeenCalledWith('validation error', 'error')
  })
})

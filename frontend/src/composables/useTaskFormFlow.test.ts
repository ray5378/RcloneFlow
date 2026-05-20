import { describe, it, expect, vi } from 'vitest'
import { useTaskFormFlow } from './useTaskFormFlow'

describe('useTaskFormFlow', () => {
  it('should return validation error', async () => {
    const options = {
      validateTaskFormBeforeSubmit: vi.fn().mockReturnValue('Name required'),
      executeTaskFormSubmit: vi.fn(),
    }
    const { runTaskFormFlow } = useTaskFormFlow(options)
    const result = await runTaskFormFlow()
    expect(result).toBe('Name required')
    expect(options.executeTaskFormSubmit).not.toHaveBeenCalled()
  })

  it('should execute submit when valid', async () => {
    const options = {
      validateTaskFormBeforeSubmit: vi.fn().mockReturnValue(''),
      executeTaskFormSubmit: vi.fn().mockResolvedValue('success'),
    }
    const { runTaskFormFlow } = useTaskFormFlow(options)
    const result = await runTaskFormFlow()
    expect(result).toBe('success')
    expect(options.executeTaskFormSubmit).toHaveBeenCalled()
  })
})
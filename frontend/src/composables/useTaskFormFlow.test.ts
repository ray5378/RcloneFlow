import { describe, it, expect, vi } from 'vitest'
import { useTaskFormFlow } from './useTaskFormFlow'

describe('useTaskFormFlow', () => {
  it('short-circuits when done click handler consumes flow', async () => {
    const api = useTaskFormFlow({
      handleTaskFormDoneClick: () => true,
      validateTaskFormBeforeSubmit: vi.fn(() => 'bad'),
      executeTaskFormSubmit: vi.fn(async () => 'ok'),
    })

    await expect(api.runTaskFormFlow()).resolves.toBe('')
  })

  it('returns validation error before submit', async () => {
    const executeTaskFormSubmit = vi.fn(async () => 'ok')
    const api = useTaskFormFlow({
      handleTaskFormDoneClick: () => false,
      validateTaskFormBeforeSubmit: () => '字段不完整',
      executeTaskFormSubmit,
    })

    await expect(api.runTaskFormFlow()).resolves.toBe('字段不完整')
    expect(executeTaskFormSubmit).not.toHaveBeenCalled()
  })

  it('executes submit when validation passes', async () => {
    const executeTaskFormSubmit = vi.fn(async () => 'created')
    const api = useTaskFormFlow({
      handleTaskFormDoneClick: () => false,
      validateTaskFormBeforeSubmit: () => '',
      executeTaskFormSubmit,
    })

    await expect(api.runTaskFormFlow()).resolves.toBe('created')
    expect(executeTaskFormSubmit).toHaveBeenCalledTimes(1)
  })
})

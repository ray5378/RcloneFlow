import { describe, it, expect, vi } from 'vitest'
import { useToastCenter } from './useToastCenter'

describe('useToastCenter', () => {
  it('should initialize with empty toasts', () => {
    const { toasts } = useToastCenter()
    expect(toasts.value).toEqual([])
  })

  it('should add a toast with default type', () => {
    const { toasts, showToast } = useToastCenter()
    vi.useFakeTimers()
    showToast('Hello')
    expect(toasts.value).toHaveLength(1)
    expect(toasts.value[0]).toMatchObject({ message: 'Hello', type: 'info' })
    vi.runAllTimers()
    expect(toasts.value).toHaveLength(0)
    vi.useRealTimers()
  })

  it('should add a toast with explicit type', () => {
    const { toasts, showToast } = useToastCenter()
    vi.useFakeTimers()
    showToast('Success!', 'success')
    expect(toasts.value[0].type).toBe('success')
    showToast('Error!', 'error')
    expect(toasts.value[1].type).toBe('error')
    vi.runAllTimers()
    vi.useRealTimers()
  })

  it('should auto-remove toast after timeout', () => {
    const { toasts, showToast } = useToastCenter()
    vi.useFakeTimers()
    showToast('Temporary')
    expect(toasts.value).toHaveLength(1)
    vi.advanceTimersByTime(2999)
    expect(toasts.value).toHaveLength(1)
    vi.advanceTimersByTime(1)
    expect(toasts.value).toHaveLength(0)
    vi.useRealTimers()
  })

  it('should handle multiple toasts with unique IDs', () => {
    const { toasts, showToast } = useToastCenter()
    vi.useFakeTimers()
    showToast('A')
    showToast('B')
    showToast('C')
    const ids = toasts.value.map(t => t.id)
    expect(new Set(ids).size).toBe(3)
    vi.runAllTimers()
    vi.useRealTimers()
  })
})

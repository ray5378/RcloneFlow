import { describe, it, expect, vi } from 'vitest'
import { useRunDetailEntry } from './useRunDetailEntry'

describe('useRunDetailEntry', () => {
  it('should open running hint for running status', () => {
    const openRunningHint = vi.fn()
    const openRunDetailModal = vi.fn()
    const openRunDetailFiles = vi.fn()
    const closeRunDetailModal = vi.fn()

    const { showRunDetail } = useRunDetailEntry({
      openRunningHint,
      openRunDetailModal,
      openRunDetailFiles,
      closeRunDetailModal,
    })

    showRunDetail({ id: 1, status: 'running' })

    expect(openRunningHint).toHaveBeenCalledWith({ id: 1, status: 'running' })
    expect(openRunDetailModal).not.toHaveBeenCalled()
    expect(openRunDetailFiles).not.toHaveBeenCalled()
  })

  it('should open detail modal and files for non-running status', () => {
    const openRunningHint = vi.fn()
    const openRunDetailModal = vi.fn()
    const openRunDetailFiles = vi.fn()
    const closeRunDetailModal = vi.fn()

    const { showRunDetail } = useRunDetailEntry({
      openRunningHint,
      openRunDetailModal,
      openRunDetailFiles,
      closeRunDetailModal,
    })

    showRunDetail({ id: 2, status: 'finished' })

    expect(openRunningHint).not.toHaveBeenCalled()
    expect(openRunDetailModal).toHaveBeenCalledWith({ id: 2, status: 'finished' })
    expect(openRunDetailFiles).toHaveBeenCalledWith({ id: 2, status: 'finished' })
  })

  it('should close detail modal', () => {
    const openRunningHint = vi.fn()
    const openRunDetailModal = vi.fn()
    const openRunDetailFiles = vi.fn()
    const closeRunDetailModal = vi.fn()

    const { closeRunDetail } = useRunDetailEntry({
      openRunningHint,
      openRunDetailModal,
      openRunDetailFiles,
      closeRunDetailModal,
    })

    closeRunDetail()

    expect(closeRunDetailModal).toHaveBeenCalled()
  })
})

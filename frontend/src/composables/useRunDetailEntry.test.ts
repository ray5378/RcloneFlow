import { describe, it, expect, vi } from 'vitest'
import { useRunDetailEntry } from './useRunDetailEntry'

describe('useRunDetailEntry', () => {
  it('routes running runs to running hint only', () => {
    const openRunningHint = vi.fn()
    const openRunDetailModal = vi.fn()
    const openRunDetailFiles = vi.fn()
    const closeRunDetailModal = vi.fn()
    const api = useRunDetailEntry({ openRunningHint, openRunDetailModal, openRunDetailFiles, closeRunDetailModal })

    api.showRunDetail({ id: 1, status: 'running' })

    expect(openRunningHint).toHaveBeenCalledWith({ id: 1, status: 'running' })
    expect(openRunDetailModal).not.toHaveBeenCalled()
    expect(openRunDetailFiles).not.toHaveBeenCalled()
  })

  it('routes historical runs to modal and files', () => {
    const openRunningHint = vi.fn()
    const openRunDetailModal = vi.fn()
    const openRunDetailFiles = vi.fn()
    const closeRunDetailModal = vi.fn()
    const api = useRunDetailEntry({ openRunningHint, openRunDetailModal, openRunDetailFiles, closeRunDetailModal })

    api.showRunDetail({ id: 2, status: 'finished' })

    expect(openRunDetailModal).toHaveBeenCalledWith({ id: 2, status: 'finished' })
    expect(openRunDetailFiles).toHaveBeenCalledWith({ id: 2, status: 'finished' })
    expect(openRunningHint).not.toHaveBeenCalled()
  })

  it('closes detail modal', () => {
    const closeRunDetailModal = vi.fn()
    const api = useRunDetailEntry({ openRunningHint: vi.fn(), openRunDetailModal: vi.fn(), openRunDetailFiles: vi.fn(), closeRunDetailModal })
    api.closeRunDetail()
    expect(closeRunDetailModal).toHaveBeenCalledTimes(1)
  })
})

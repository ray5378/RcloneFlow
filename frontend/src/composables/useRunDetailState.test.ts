import { describe, it, expect } from 'vitest'
import { useRunDetailState } from './useRunDetailState'

describe('useRunDetailState', () => {
  it('should initialize with closed modal', () => {
    const { showDetailModal, runDetail } = useRunDetailState()
    expect(showDetailModal.value).toBe(false)
    expect(runDetail.value).toEqual({})
  })

  it('should open modal with run data', () => {
    const { showDetailModal, runDetail, openRunDetailModal } = useRunDetailState()
    const run = { id: 1, status: 'finished' }

    openRunDetailModal(run)

    expect(showDetailModal.value).toBe(true)
    expect(runDetail.value).toEqual(run)
  })

  it('should close modal', () => {
    const { showDetailModal, openRunDetailModal, closeRunDetailModal } = useRunDetailState()
    openRunDetailModal({ id: 1 })
    expect(showDetailModal.value).toBe(true)

    closeRunDetailModal()

    expect(showDetailModal.value).toBe(false)
  })
})

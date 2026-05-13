import { describe, it, expect } from 'vitest'
import { useRunDetailState } from './useRunDetailState'

describe('useRunDetailState', () => {
  it('opens modal with selected run', () => {
    const api = useRunDetailState()
    const run = { id: 123, taskName: 'demo' }

    api.openRunDetailModal(run)

    expect(api.showDetailModal.value).toBe(true)
    expect(api.runDetail.value).toEqual(run)
  })

  it('closes modal without clearing run detail', () => {
    const api = useRunDetailState()
    const run = { id: 456 }
    api.openRunDetailModal(run)

    api.closeRunDetailModal()

    expect(api.showDetailModal.value).toBe(false)
    expect(api.runDetail.value).toEqual(run)
  })
})

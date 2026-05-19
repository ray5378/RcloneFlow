import { describe, it, expect } from 'vitest'
import { useRunDetailRuntime } from './useRunDetailRuntime'

describe('useRunDetailRuntime', () => {
  it('should initialize and return expected properties', () => {
    const runApi = {
      getRunStatus: async () => ({ data: { log: '' } }),
    }

    const runtime = useRunDetailRuntime({ runApi })

    expect(runtime).toHaveProperty('showDetailModal')
    expect(runtime).toHaveProperty('runDetail')
    expect(runtime).toHaveProperty('openRunDetailModal')
    expect(runtime).toHaveProperty('closeRunDetailModal')
    expect(runtime).toHaveProperty('runFilesTotal')
    expect(runtime).toHaveProperty('openRunDetailFiles')
    expect(runtime).toHaveProperty('getFinalSummary')
    expect(runtime).toHaveProperty('finalFiles')
    expect(runtime).toHaveProperty('setFinalFilter')
  })

  it('should open and close modal', () => {
    const runApi = {
      getRunStatus: async () => ({ data: { log: '' } }),
    }

    const { showDetailModal, openRunDetailModal, closeRunDetailModal, runDetail } = useRunDetailRuntime({ runApi })

    expect(showDetailModal.value).toBe(false)

    openRunDetailModal({ id: 1, status: 'finished' })
    expect(showDetailModal.value).toBe(true)
    expect(runDetail.value.id).toBe(1)

    closeRunDetailModal()
    expect(showDetailModal.value).toBe(false)
  })
})

import { describe, it, expect, vi } from 'vitest'
import { ref } from 'vue'

const stateApi = {
  showDetailModal: ref(true),
  runDetail: ref({ id: 1 }),
  openRunDetailModal: vi.fn(),
  closeRunDetailModal: vi.fn(),
}
const filesApi = {
  runFiles: ref([{ name: 'a' }, { name: 'b' }]),
  runFilesPage: ref(2),
  openRunDetailFiles: vi.fn(),
  visibleRunFiles: ref([{ name: 'a' }, { name: 'b' }]),
  pagedRunFiles: ref([{ name: 'a' }]),
  totalRunFilesPages: ref(3),
  goPrevFilesPage: vi.fn(),
  goNextFilesPage: vi.fn(),
}
const computedApi = {
  getFinalSummary: vi.fn(() => ({ counts: { total: 2 } })),
  hasFinalSummaryFiles: ref(true),
  finalFiles: ref([{ name: 'fa' }]),
  finalCountAll: ref(2),
  finalCountSuccess: ref(1),
  finalCountFailed: ref(1),
  finalCountOther: ref(0),
  setFinalFilter: vi.fn(),
  finalFilesPage: ref(1),
  finalFilesTotal: ref(2),
  totalFinalFilesPages: ref(1),
  pagedFinalFiles: ref([{ name: 'fa' }]),
  finalFilesJump: ref(1),
  goPrevFinalFilesPage: vi.fn(),
  goNextFinalFilesPage: vi.fn(),
  jumpFinalFilesPage: vi.fn(),
}

vi.mock('./useRunDetailState', () => ({ useRunDetailState: () => stateApi }))
vi.mock('./useRunDetailFiles', () => ({ useRunDetailFiles: () => filesApi }))
vi.mock('./useRunDetailComputed', () => ({ useRunDetailComputed: () => computedApi }))

import { useRunDetailRuntime } from './useRunDetailRuntime'

describe('useRunDetailRuntime', () => {
  it('wires child composables and exposes computed totals', () => {
    const api = useRunDetailRuntime({ runApi: { getFiles: vi.fn() } })

    expect(api.showDetailModal.value).toBe(true)
    expect(api.runDetail.value).toEqual({ id: 1 })
    expect(api.runFilesTotal.value).toBe(2)
    expect(api.runFilesPage.value).toBe(2)
    expect(api.pagedRunFiles.value).toEqual([{ name: 'a' }])
    expect(api.totalRunFilesPages.value).toBe(3)
    expect(api.hasFinalSummaryFiles.value).toBe(true)
    expect(api.finalFiles.value).toEqual([{ name: 'fa' }])
    expect(api.finalCountAll.value).toBe(2)
    expect(api.finalFilesTotal.value).toBe(2)
  })

  it('passes through action handlers', () => {
    const api = useRunDetailRuntime({ runApi: { getFiles: vi.fn() } })
    api.openRunDetailModal({ id: 9 })
    api.closeRunDetailModal()
    api.openRunDetailFiles({ id: 9 })
    api.setFinalFilter('failed')
    api.goPrevFilesPage()
    api.goNextFilesPage()
    api.goPrevFinalFilesPage()
    api.goNextFinalFilesPage()
    api.jumpFinalFilesPage()

    expect(stateApi.openRunDetailModal).toHaveBeenCalledWith({ id: 9 })
    expect(stateApi.closeRunDetailModal).toHaveBeenCalled()
    expect(filesApi.openRunDetailFiles).toHaveBeenCalledWith({ id: 9 })
    expect(computedApi.setFinalFilter).toHaveBeenCalledWith('failed')
    expect(filesApi.goPrevFilesPage).toHaveBeenCalled()
    expect(filesApi.goNextFilesPage).toHaveBeenCalled()
    expect(computedApi.goPrevFinalFilesPage).toHaveBeenCalled()
    expect(computedApi.goNextFinalFilesPage).toHaveBeenCalled()
    expect(computedApi.jumpFinalFilesPage).toHaveBeenCalled()
  })
})

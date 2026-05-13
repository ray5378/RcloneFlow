import { describe, it, expect, vi } from 'vitest'
import { ref } from 'vue'
import { useTaskHistoryPagingEntry } from './useTaskHistoryPagingEntry'

describe('useTaskHistoryPagingEntry', () => {
  it('clamps jump target into valid range and triggers load', () => {
    const loadData = vi.fn()
    const api = useTaskHistoryPagingEntry({
      jumpPage: ref(99),
      runsPage: ref(1),
      currentTotalPages: ref(7),
      loadData,
    })

    api.jumpToPage()

    expect(loadData).toHaveBeenCalledTimes(1)
  })

  it('updates both jumpPage and runsPage with normalized page', () => {
    const jumpPage = ref(0)
    const runsPage = ref(5)
    const api = useTaskHistoryPagingEntry({
      jumpPage,
      runsPage,
      currentTotalPages: ref(3),
      loadData: vi.fn(),
    })

    api.jumpToPage()
    expect(runsPage.value).toBe(1)
    expect(jumpPage.value).toBe(1)

    jumpPage.value = 3
    api.jumpToPage()
    expect(runsPage.value).toBe(3)
    expect(jumpPage.value).toBe(3)
  })
})

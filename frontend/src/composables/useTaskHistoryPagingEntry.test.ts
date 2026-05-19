import { describe, it, expect, vi } from 'vitest'
import { ref } from 'vue'
import { useTaskHistoryPagingEntry } from './useTaskHistoryPagingEntry'

describe('useTaskHistoryPagingEntry', () => {
  it('should jump to valid page', () => {
    const jumpPage = ref(3)
    const runsPage = ref(1)
    const currentTotalPages = ref(5)
    const loadData = vi.fn()

    const { jumpToPage } = useTaskHistoryPagingEntry({
      jumpPage,
      runsPage,
      currentTotalPages,
      loadData,
    })

    jumpToPage()

    expect(runsPage.value).toBe(3)
    expect(jumpPage.value).toBe(3)
    expect(loadData).toHaveBeenCalled()
  })

  it('should clamp page to max', () => {
    const jumpPage = ref(10)
    const runsPage = ref(1)
    const currentTotalPages = ref(3)
    const loadData = vi.fn()

    const { jumpToPage } = useTaskHistoryPagingEntry({
      jumpPage,
      runsPage,
      currentTotalPages,
      loadData,
    })

    jumpToPage()

    expect(runsPage.value).toBe(3)
    expect(jumpPage.value).toBe(3)
  })

  it('should clamp page to min', () => {
    const jumpPage = ref(0)
    const runsPage = ref(1)
    const currentTotalPages = ref(5)
    const loadData = vi.fn()

    const { jumpToPage } = useTaskHistoryPagingEntry({
      jumpPage,
      runsPage,
      currentTotalPages,
      loadData,
    })

    jumpToPage()

    expect(runsPage.value).toBe(1)
    expect(jumpPage.value).toBe(1)
  })
})

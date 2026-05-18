import type { Ref } from 'vue'

export function useTaskHistoryPagingEntry(options: {
  jumpPage: Ref<number>
  runsPage: Ref<number>
  currentTotalPages: Ref<number>
  loadData: () => Promise<void> | void
}) {
  function jumpToPage() {
    const page = Math.min(Math.max(1, options.jumpPage.value || 1), options.currentTotalPages.value)
    options.runsPage.value = page
    options.jumpPage.value = page
    options.loadData()
  }

  return {
    jumpToPage,
  }
}

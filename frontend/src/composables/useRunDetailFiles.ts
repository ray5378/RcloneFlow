import { computed, ref, type Ref } from 'vue'
import type { RunFileRow } from '../api/run'

type FinalFilterType = 'all' | 'success' | 'failed' | 'other'

interface UseRunDetailFilesOptions {
  runDetail: Ref<any>
  currentFinalFilter?: Ref<FinalFilterType>
  runApi: {
    getFiles: (runId: number, offset: number, limit: number) => Promise<{ items?: any[]; total?: number }>
  }
}

export function useRunDetailFiles(options: UseRunDetailFilesOptions) {
  const runFiles = ref<any[]>([])
  const runFilesTotal = ref(0)
  const runFilesPage = ref(1)
  const runFilesPageSize = ref(Math.max(10, Math.floor((window.innerHeight - 380) / 32)))

  function resetRunFiles() {
    runFilesPage.value = 1
    runFiles.value = []
    runFilesTotal.value = 0
  }

  async function reloadRunFiles() {
    try {
      if (!options.runDetail.value?.id) return
      const pageSize = Math.min(1000, Math.max(100, runFilesPageSize.value * 4))
      const allItems: any[] = []
      let offset = 0
      let total = 0
      while (true) {
        const res = await options.runApi.getFiles(options.runDetail.value.id, offset, pageSize)
        const items = res.items || []
        total = res.total || total || 0
        allItems.push(...items)
        if (items.length < pageSize) break
        if (total > 0 && allItems.length >= total) break
        offset += pageSize
      }
      runFiles.value = allItems
      runFilesTotal.value = total || allItems.length
    } catch (e) {
      console.error(e)
    }
  }

  function openRunDetailFiles(run: any) {
    options.runDetail.value = run
    resetRunFiles()
    void reloadRunFiles()
  }

  const moveFilteredRunFiles = computed<RunFileRow[]>(() => {
    const items = Array.isArray(runFiles.value) ? (runFiles.value as RunFileRow[]) : []
    if (options.runDetail.value?.taskMode === 'move') {
      return items.filter(it => (it.status || '') !== 'deleted')
    }
    return items
  })

  const visibleRunFiles = computed<RunFileRow[]>(() => {
    const items = moveFilteredRunFiles.value
    const filter = options.currentFinalFilter?.value || 'all'
    if (filter === 'success') return items.filter(it => (it.status || '') === 'success')
    if (filter === 'failed') return items.filter(it => (it.status || '') === 'failed')
    if (filter === 'other') return items.filter(it => (it.status || '') === 'skipped')
    return items
  })

  const pagedRunFiles = computed(() => {
    const page = runFilesPage.value || 1
    const pageSize = runFilesPageSize.value || 1
    const start = (page - 1) * pageSize
    return visibleRunFiles.value.slice(start, start + pageSize)
  })
  const totalRunFilesPages = computed(() => Math.max(1, Math.ceil((visibleRunFiles.value.length || 0) / runFilesPageSize.value)))

  function goPrevFilesPage() {
    if (runFilesPage.value > 1) {
      runFilesPage.value--
    }
  }

  function goNextFilesPage() {
    if (runFilesPage.value < totalRunFilesPages.value) {
      runFilesPage.value++
    }
  }

  return {
    runFiles,
    runFilesTotal,
    runFilesPage,
    runFilesPageSize,
    visibleRunFiles,
    resetRunFiles,
    reloadRunFiles,
    openRunDetailFiles,
    pagedRunFiles,
    totalRunFilesPages,
    goPrevFilesPage,
    goNextFilesPage,
  }
}

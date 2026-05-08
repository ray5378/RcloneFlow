import { computed, ref, type Ref } from 'vue'

interface UseRunDetailFilesOptions {
  runDetail: Ref<any>
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

  const pagedRunFiles = computed(() => {
    const page = runFilesPage.value || 1
    const pageSize = runFilesPageSize.value || 1
    const start = (page - 1) * pageSize
    return runFiles.value.slice(start, start + pageSize)
  })
  const totalRunFilesPages = computed(() => Math.max(1, Math.ceil((runFilesTotal.value || 0) / runFilesPageSize.value)))

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
    resetRunFiles,
    reloadRunFiles,
    openRunDetailFiles,
    pagedRunFiles,
    totalRunFilesPages,
    goPrevFilesPage,
    goNextFilesPage,
  }
}

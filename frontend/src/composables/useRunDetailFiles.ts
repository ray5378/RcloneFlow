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
      const pageOffset = (runFilesPage.value - 1) * runFilesPageSize.value
      const res = await options.runApi.getFiles(options.runDetail.value.id, pageOffset, runFilesPageSize.value)
      runFiles.value = res.items || []
      runFilesTotal.value = res.total || 0
    } catch (e) {
      console.error(e)
    }
  }

  function openRunDetailFiles(run: any) {
    options.runDetail.value = run
    resetRunFiles()
    void reloadRunFiles()
  }

  const pagedRunFiles = computed(() => runFiles.value)
  const totalRunFilesPages = computed(() => Math.max(1, Math.ceil((runFilesTotal.value || 0) / runFilesPageSize.value)))

  function goPrevFilesPage() {
    if (runFilesPage.value > 1) {
      runFilesPage.value--
      reloadRunFiles()
    }
  }

  function goNextFilesPage() {
    if (runFilesPage.value < totalRunFilesPages.value) {
      runFilesPage.value++
      reloadRunFiles()
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

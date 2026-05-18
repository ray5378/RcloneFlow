import { computed, ref, watch, type Ref } from 'vue'
import type { RunFileRow } from '../api/run'

type FinalFilterType = 'all' | 'success' | 'failed' | 'other'
type RunFileKind = 'success' | 'failed' | 'skipped' | 'deleted' | 'unknown'

interface UseRunDetailFilesOptions {
  runDetail: Ref<any>
  currentFinalFilter?: Ref<FinalFilterType>
  runApi: {
    getFiles: (runId: number, offset: number, limit: number, filter?: FinalFilterType) => Promise<{ items?: any[]; total?: number }>
  }
}

function getRunFileKind(row: Partial<RunFileRow> & { action?: string }): RunFileKind {
  const status = String(row.status || '').trim().toLowerCase()
  const action = String(row.action || '').trim().toLowerCase()
  const joined = `${status} ${action}`.trim()

  if (joined.includes('deleted')) return 'deleted'
  if (joined.includes('failed')) return 'failed'
  if (joined.includes('skipped')) return 'skipped'
  if (joined.includes('success') || joined.includes('copied') || joined.includes('new') || joined.includes('moved') || joined.includes('synced')) return 'success'
  return 'unknown'
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
      const page = runFilesPage.value || 1
      const pageSize = runFilesPageSize.value || 1
      const filter = options.currentFinalFilter?.value || 'all'
      const offset = (page - 1) * pageSize
      const res = await options.runApi.getFiles(options.runDetail.value.id, offset, pageSize, filter)
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

  const visibleRunFiles = computed<RunFileRow[]>(() => {
    return Array.isArray(runFiles.value) ? (runFiles.value as RunFileRow[]) : []
  })

  const pagedRunFiles = computed(() => visibleRunFiles.value)
  const totalRunFilesPages = computed(() => Math.max(1, Math.ceil((runFilesTotal.value || 0) / runFilesPageSize.value)))

  watch(() => options.currentFinalFilter?.value, () => {
    runFilesPage.value = 1
    void reloadRunFiles()
  })

  watch(() => options.runDetail.value?.id, () => {
    resetRunFiles()
    void reloadRunFiles()
  })

  watch([runFilesPage, runFilesPageSize], () => {
    const totalPages = totalRunFilesPages.value
    if (runFilesPage.value > totalPages) {
      runFilesPage.value = totalPages
      return
    }
    if (runFilesPage.value < 1) {
      runFilesPage.value = 1
      return
    }
    void reloadRunFiles()
  })

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

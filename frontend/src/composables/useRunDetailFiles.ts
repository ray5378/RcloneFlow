import { computed, ref, watch, type Ref } from 'vue'
import type { RunFileRow } from '../api/run'

type RunFileKind = 'success' | 'failed' | 'skipped' | 'deleted' | 'unknown'

interface UseRunDetailFilesOptions {
  runDetail: Ref<any>
  runApi: {
    getFiles: (runId: number, offset: number, limit: number) => Promise<{ items?: any[]; total?: number }>
  }
}

const DEFAULT_PAGE_SIZE = 10
const MIN_PAGE_SIZE = 1

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
  const runFilesPageSize = ref(Math.max(DEFAULT_PAGE_SIZE, Math.floor((window.innerHeight - 380) / 32)))

  function resetRunFiles() {
    runFilesPage.value = 1
    runFiles.value = []
    runFilesTotal.value = 0
  }

  async function reloadRunFiles() {
    if (!options.runDetail.value?.id) return
    const page = runFilesPage.value || 1
    const pageSize = runFilesPageSize.value || MIN_PAGE_SIZE
    const offset = (page - 1) * pageSize
    const res = await options.runApi.getFiles(options.runDetail.value.id, offset, pageSize)
    runFiles.value = res.items || []
    runFilesTotal.value = res.total || 0
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

  const validatedRunFilesPage = computed({
    get: () => {
      const total = totalRunFilesPages.value
      const current = runFilesPage.value
      if (current < 1) return 1
      if (current > total) return total
      return current
    },
    set: (val: number) => {
      runFilesPage.value = val
    }
  })

  watch(() => options.runDetail.value?.id, () => {
    resetRunFiles()
    void reloadRunFiles()
  })

  watch([validatedRunFilesPage, runFilesPageSize], () => {
    void reloadRunFiles()
  })

  function goPrevFilesPage() {
    if (validatedRunFilesPage.value > 1) {
      validatedRunFilesPage.value--
    }
  }

  function goNextFilesPage() {
    if (validatedRunFilesPage.value < totalRunFilesPages.value) {
      validatedRunFilesPage.value++
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

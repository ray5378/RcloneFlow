import { computed, ref, type Ref } from 'vue'

type FinalFilterType = 'all' | 'success' | 'failed' | 'other'

interface UseRunDetailComputedOptions {
  runDetail?: Ref<any>
  finalFilesPage?: Ref<number>
  finalFilesPageSize?: Ref<number>
}

export function useRunDetailComputed(options?: UseRunDetailComputedOptions) {
  function getFinalSummary(run: any) {
    try {
      const sum = typeof run?.summary === 'string' ? JSON.parse(run.summary) : run?.summary
      if (sum && typeof sum === 'object' && sum.finalSummary) return sum.finalSummary
    } catch {}
    return null
  }

  function getPreflight(run: any) {
    try {
      const sum = typeof run?.summary === 'string' ? JSON.parse(run.summary) : run?.summary
      return sum?.preflight || null
    } catch {
      return null
    }
  }

  const finalFiles = computed(() => {
    if (!options?.runDetail) return [] as any[]
    return (getFinalSummary(options.runDetail.value)?.files || []) as any[]
  })

  const finalCountAll = computed(() => finalFiles.value.length)
  const finalCountSuccess = computed(() => finalFiles.value.filter(it => (it.status || '') === 'success').length)
  const finalCountFailed = computed(() => finalFiles.value.filter(it => (it.status || '') === 'failed').length)
  const finalCountOther = computed(() => finalFiles.value.filter(it => (it.status || '') === 'skipped').length)

  const currentFinalFilter = ref<FinalFilterType>('all')
  function setFinalFilter(filter: FinalFilterType) {
    currentFinalFilter.value = filter
    if (options?.finalFilesPage) options.finalFilesPage.value = 1
  }

  const finalFilteredFiles = computed(() => {
    if (currentFinalFilter.value === 'success') return finalFiles.value.filter(it => (it.status || '') === 'success')
    if (currentFinalFilter.value === 'failed') return finalFiles.value.filter(it => (it.status || '') === 'failed')
    if (currentFinalFilter.value === 'other') return finalFiles.value.filter(it => (it.status || '') === 'skipped')
    return finalFiles.value
  })

  const finalFilesTotal = computed(() => finalFilteredFiles.value.length)
  const totalFinalFilesPages = computed(() => {
    const pageSize = options?.finalFilesPageSize?.value || 1
    return Math.max(1, Math.ceil((finalFilesTotal.value || 0) / pageSize))
  })
  const pagedFinalFiles = computed(() => {
    const page = options?.finalFilesPage?.value || 1
    const pageSize = options?.finalFilesPageSize?.value || 1
    const start = (page - 1) * pageSize
    return finalFilteredFiles.value.slice(start, start + pageSize)
  })

  const finalFilesJump = ref<number | null>(null)

  function goPrevFinalFilesPage() {
    if (!options?.finalFilesPage) return
    if (options.finalFilesPage.value > 1) options.finalFilesPage.value--
  }

  function goNextFinalFilesPage() {
    if (!options?.finalFilesPage) return
    if (options.finalFilesPage.value < totalFinalFilesPages.value) options.finalFilesPage.value++
  }

  function jumpFinalFilesPage() {
    if (!options?.finalFilesPage || !finalFilesJump.value) return
    const p = Math.min(Math.max(1, finalFilesJump.value), totalFinalFilesPages.value)
    options.finalFilesPage.value = p
  }

  return {
    getFinalSummary,
    getPreflight,
    finalFiles,
    finalCountAll,
    finalCountSuccess,
    finalCountFailed,
    finalCountOther,
    currentFinalFilter,
    setFinalFilter,
    finalFilteredFiles,
    finalFilesTotal,
    totalFinalFilesPages,
    pagedFinalFiles,
    finalFilesJump,
    goPrevFinalFilesPage,
    goNextFinalFilesPage,
    jumpFinalFilesPage,
  }
}

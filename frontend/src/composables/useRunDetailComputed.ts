import { computed, ref, type Ref } from 'vue'

type FinalFilterType = 'all' | 'success' | 'failed' | 'other'

interface UseRunDetailComputedOptions {
  runDetail?: Ref<any>
  finalFilesPage?: Ref<number>
  finalFilesPageSize?: Ref<number>
}

export function useRunDetailComputed(options?: UseRunDetailComputedOptions) {
  const finalFilesPage = options?.finalFilesPage ?? ref(1)
  const finalFilesPageSize = options?.finalFilesPageSize ?? ref(Math.max(10, Math.floor((window.innerHeight - 420) / 34)))

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
    finalFilesPage.value = 1
  }

  const finalFilteredFiles = computed(() => {
    if (currentFinalFilter.value === 'success') return finalFiles.value.filter(it => (it.status || '') === 'success')
    if (currentFinalFilter.value === 'failed') return finalFiles.value.filter(it => (it.status || '') === 'failed')
    if (currentFinalFilter.value === 'other') return finalFiles.value.filter(it => (it.status || '') === 'skipped')
    return finalFiles.value
  })

  const finalFilesTotal = computed(() => finalFilteredFiles.value.length)
  const totalFinalFilesPages = computed(() => {
    const pageSize = finalFilesPageSize.value || 1
    return Math.max(1, Math.ceil((finalFilesTotal.value || 0) / pageSize))
  })
  const pagedFinalFiles = computed(() => {
    const page = finalFilesPage.value || 1
    const pageSize = finalFilesPageSize.value || 1
    const start = (page - 1) * pageSize
    return finalFilteredFiles.value.slice(start, start + pageSize)
  })

  const finalFilesJump = ref<number | null>(null)

  function goPrevFinalFilesPage() {
    if (finalFilesPage.value > 1) finalFilesPage.value--
  }

  function goNextFinalFilesPage() {
    if (finalFilesPage.value < totalFinalFilesPages.value) finalFilesPage.value++
  }

  function jumpFinalFilesPage() {
    if (!finalFilesJump.value) return
    const p = Math.min(Math.max(1, finalFilesJump.value), totalFinalFilesPages.value)
    finalFilesPage.value = p
  }

  return {
    getFinalSummary,
    getPreflight,
    finalFilesPage,
    finalFilesPageSize,
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

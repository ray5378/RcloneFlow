import { computed, ref, type Ref } from 'vue'
import type { FinalSummary } from '../api/run'
import type { Run, RunSummaryPayload } from '../types'

type FinalFilterType = 'all' | 'success' | 'failed' | 'other'

interface UseRunDetailComputedOptions {
  runDetail?: Ref<any>
  finalFilesPage?: Ref<number>
  finalFilesPageSize?: Ref<number>
}

export function useRunDetailComputed(options?: UseRunDetailComputedOptions) {
  const finalFilesPage = options?.finalFilesPage ?? ref(1)
  const finalFilesPageSize = options?.finalFilesPageSize ?? ref(Math.max(10, Math.floor((window.innerHeight - 420) / 34)))
  const finalSummaryByRunId = new Map<number, FinalSummary | null>()
  const finalSummaryBySummaryText = new Map<string, FinalSummary | null>()

  function parseFinalSummary(run: Run | null | undefined): FinalSummary | null {
    try {
      const summary = run?.summary
      if (typeof summary === 'string') {
        if (finalSummaryBySummaryText.has(summary)) return finalSummaryBySummaryText.get(summary) || null
        const sum = JSON.parse(summary) as RunSummaryPayload | undefined
        const finalSummary = (sum && typeof sum === 'object' && sum.finalSummary) ? (sum.finalSummary as FinalSummary) : null
        finalSummaryBySummaryText.set(summary, finalSummary)
        if (run?.id) finalSummaryByRunId.set(run.id, finalSummary)
        return finalSummary
      }
      const runId = run?.id
      if (runId && finalSummaryByRunId.has(runId)) return finalSummaryByRunId.get(runId) || null
      const sum = summary as RunSummaryPayload | undefined
      // 历史详情只读 finalSummary；
      // 不从 progress / completedFreezeByTask 倒推最终总结。
      const finalSummary = (sum && typeof sum === 'object' && sum.finalSummary) ? (sum.finalSummary as FinalSummary) : null
      if (runId) finalSummaryByRunId.set(runId, finalSummary)
      return finalSummary
    } catch {
      return null
    }
  }

  function getFinalSummary(run: Run | null | undefined): FinalSummary | null {
    return parseFinalSummary(run)
  }

  const finalFiles = computed(() => {
    if (!options?.runDetail) return [] as any[]
    return (getFinalSummary(options.runDetail.value)?.files || []) as any[]
  })

  function getSummaryCounts(run: Run | null | undefined) {
    const fs = getFinalSummary(run)
    const counts = (fs?.counts && typeof fs.counts === 'object') ? fs.counts : null
    return {
      all: Number(counts?.total || finalFiles.value.length || 0),
      copied: Number(counts?.copied || 0),
      deleted: Number(counts?.deleted || 0),
      failed: Number(counts?.failed || 0),
      skipped: Number(counts?.skipped || 0),
    }
  }

  const finalCountAll = computed(() => getSummaryCounts(options?.runDetail?.value).all)
  const finalCountSuccess = computed(() => {
    const counts = getSummaryCounts(options?.runDetail?.value)
    return options?.runDetail?.value?.taskMode === 'move'
      ? counts.copied
      : counts.copied + counts.deleted
  })
  const finalCountFailed = computed(() => getSummaryCounts(options?.runDetail?.value).failed)
  const finalCountOther = computed(() => getSummaryCounts(options?.runDetail?.value).skipped)

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

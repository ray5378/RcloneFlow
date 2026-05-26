import { computed, ref, watch, type Ref } from 'vue'
import type { FinalSummary, FinalSummaryFile } from '../api/run'
import type { Run, RunSummaryPayload } from '../types'

const DEFAULT_PAGE_SIZE = 10
const MIN_PAGE_SIZE = 1

interface UseRunDetailComputedOptions {
  runDetail?: Ref<any>
  detailFiles?: Ref<any[]>
  finalFilesPage?: Ref<number>
  finalFilesPageSize?: Ref<number>
}

export function useRunDetailComputed(options?: UseRunDetailComputedOptions) {
  const finalFilesPage = options?.finalFilesPage ?? ref(1)
  const finalFilesPageSize = options?.finalFilesPageSize ?? ref(Math.max(DEFAULT_PAGE_SIZE, Math.floor((window.innerHeight - 420) / 34)))
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
      const finalSummary = (sum && typeof sum === 'object' && sum.finalSummary) ? (sum.finalSummary as FinalSummary) : null
      if (runId) finalSummaryByRunId.set(runId, finalSummary)
      return finalSummary
    } catch (e) {
      console.warn('[useRunDetailComputed] Failed to parse final summary:', e)
      return null
    }
  }

  const summaryFiles = computed(() => {
    if (!options?.runDetail) return [] as FinalSummaryFile[]
    const detail = options.runDetail.value as Record<string, unknown>
    return (parseFinalSummary(detail)?.files || []) as FinalSummaryFile[]
  })

  const detailFiles = computed(() => {
    if (!options?.detailFiles?.value || !Array.isArray(options.detailFiles.value)) return [] as FinalSummaryFile[]
    return options.detailFiles.value as FinalSummaryFile[]
  })

  const hasFinalSummaryFiles = computed(() => summaryFiles.value.length > 0)

  const finalFiles = computed(() => {
    if (hasFinalSummaryFiles.value) return summaryFiles.value
    return detailFiles.value
  })

  function getSummaryCounts(run: Run | null | undefined) {
    const fs = parseFinalSummary(run)
    const counts = (fs?.counts && typeof fs.counts === 'object') ? fs.counts : null
    return {
      all: Number(counts?.total || finalFiles.value.length || 0),
      copied: Number(counts?.copied || 0),
      deleted: Number(counts?.deleted || 0),
      failed: Number(counts?.failed || 0),
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

  const finalFilesTotal = computed(() => finalFiles.value.length)
  const totalFinalFilesPages = computed(() => {
    const pageSize = finalFilesPageSize.value || MIN_PAGE_SIZE
    return Math.max(1, Math.ceil((finalFilesTotal.value || 0) / pageSize))
  })
  const pagedFinalFiles = computed(() => {
    const page = finalFilesPage.value || 1
    const pageSize = finalFilesPageSize.value || MIN_PAGE_SIZE
    const start = (page - 1) * pageSize
    return finalFiles.value.slice(start, start + pageSize)
  })

  const finalFilesJump = ref<number | null>(null)

  function clampFinalFilesPage() {
    const totalPages = totalFinalFilesPages.value
    if (finalFilesPage.value > totalPages) finalFilesPage.value = totalPages
    if (finalFilesPage.value < 1) finalFilesPage.value = 1
  }

  watch(() => options?.runDetail?.value?.id, () => {
    finalFilesPage.value = 1
    finalFilesJump.value = null
  }, { immediate: true })

  watch([finalFilesTotal, finalFilesPageSize], clampFinalFilesPage, { immediate: true })

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
    getFinalSummary: parseFinalSummary,
    finalFilesPage,
    finalFilesPageSize,
    hasFinalSummaryFiles,
    finalFiles,
    finalCountAll,
    finalCountSuccess,
    finalCountFailed,
    finalFilesTotal,
    totalFinalFilesPages,
    pagedFinalFiles,
    finalFilesJump,
    goPrevFinalFilesPage,
    goNextFinalFilesPage,
    jumpFinalFilesPage,
  }
}

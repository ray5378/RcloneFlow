import { computed, type Ref } from 'vue'
import type { Run } from '../types'

interface UseTaskHistoryComputedOptions {
  runs: Ref<Run[]>
  runsTotal: Ref<number>
  taskRuns: Ref<Run[]>
  historyFilterTaskId: Ref<number | null>
  historyStatusFilter: Ref<string>
  runsPage: Ref<number>
  runsPageSize: number
  getFinalSummary: (run: Run) => any
}

export function useTaskHistoryComputed(options: UseTaskHistoryComputedOptions) {
  const finalSummaryCache = new Map<number, { summaryKey: string; value: any }>()

  function getCachedFinalSummary(run: Run) {
    if (!run?.id) return options.getFinalSummary(run)
    const summaryKey = typeof run.summary === 'string' ? run.summary : JSON.stringify(run.summary || null)
    const cached = finalSummaryCache.get(run.id)
    if (cached && cached.summaryKey === summaryKey) return cached.value
    const value = options.getFinalSummary(run)
    finalSummaryCache.set(run.id, { summaryKey, value })
    return value
  }

  const filteredRunsBase = computed(() => {
    const source = options.historyFilterTaskId.value === null ? options.runs.value : options.taskRuns.value
    if (options.historyStatusFilter.value === 'hasTransfer') {
      return source.filter(r => {
        const sum = getCachedFinalSummary(r)
        return sum && (sum.totalCount > 1 || sum.transferredBytes > 0)
      })
    }
    if (options.historyStatusFilter.value !== 'all') {
      return source.filter(r => r.status === options.historyStatusFilter.value)
    }
    return source
  })

  const filteredRuns = computed(() => {
    const start = (options.runsPage.value - 1) * options.runsPageSize
    const end = start + options.runsPageSize
    return filteredRunsBase.value.slice(start, end)
  })

  const filteredRunsTotal = computed(() => filteredRunsBase.value.length)

  const currentTotal = computed(() => {
    return options.historyFilterTaskId.value === null ? (options.runsTotal.value || 0) : filteredRunsTotal.value
  })

  const currentTotalPages = computed(() => Math.max(1, Math.ceil(currentTotal.value / options.runsPageSize)))

  return {
    filteredRuns,
    filteredRunsTotal,
    currentTotal,
    currentTotalPages,
  }
}

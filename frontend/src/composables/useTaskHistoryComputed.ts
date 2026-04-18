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
  const filteredRuns = computed(() => {
    const source = options.historyFilterTaskId.value === null ? options.runs.value : options.taskRuns.value
    let result = [...source]
    if (options.historyStatusFilter.value === 'hasTransfer') {
      result = result.filter(r => {
        const sum = options.getFinalSummary(r)
        return sum && (sum.totalCount > 1 || sum.transferredBytes > 0)
      })
    } else if (options.historyStatusFilter.value !== 'all') {
      result = result.filter(r => r.status === options.historyStatusFilter.value)
    }
    const start = (options.runsPage.value - 1) * options.runsPageSize
    const end = start + options.runsPageSize
    return result.slice(start, end)
  })

  const filteredRunsTotal = computed(() => {
    const source = options.historyFilterTaskId.value === null ? options.runs.value : options.taskRuns.value
    let result = [...source]
    if (options.historyStatusFilter.value === 'hasTransfer') {
      result = result.filter(r => {
        const sum = options.getFinalSummary(r)
        return sum && (sum.totalCount > 1 || sum.transferredBytes > 0)
      })
    } else if (options.historyStatusFilter.value !== 'all') {
      result = result.filter(r => r.status === options.historyStatusFilter.value)
    }
    return result.length
  })

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

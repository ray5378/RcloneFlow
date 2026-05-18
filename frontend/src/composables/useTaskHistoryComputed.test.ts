import { describe, it, expect } from 'vitest'
import { ref } from 'vue'
import { useTaskHistoryComputed } from './useTaskHistoryComputed'

describe('useTaskHistoryComputed', () => {
  it('keeps single-file transfer runs in hasTransfer filter', () => {
    const runs = ref<any[]>([{ id: 11, status: 'finished', summary: '{}' }])
    const taskRuns = ref<any[]>([])
    const runsTotal = ref(1)
    const historyFilterTaskId = ref<number | null>(null)
    const historyStatusFilter = ref('hasTransfer')
    const runsPage = ref(1)

    const { filteredRuns, filteredRunsTotal } = useTaskHistoryComputed({
      runs,
      runsTotal,
      taskRuns,
      historyFilterTaskId,
      historyStatusFilter,
      runsPage,
      runsPageSize: 50,
      getFinalSummary: () => ({
        totalCount: 1,
        transferredBytes: 0,
      }),
    })

    expect(filteredRunsTotal.value).toBe(1)
    expect(filteredRuns.value).toHaveLength(1)
    expect(filteredRuns.value[0]?.id).toBe(11)
  })
})

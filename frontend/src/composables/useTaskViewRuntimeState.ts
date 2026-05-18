import { ref } from 'vue'
import { useActiveRunLookup } from './useActiveRunLookup'

export function useTaskViewRuntimeState() {
  const activeRuns = ref<any[]>([])
  const globalStats = ref<any>({})
  const showGlobalStatsModal = ref(false)
  const activeRunLookup = useActiveRunLookup(activeRuns)
  const lastNonDecreasingTotalsByTask = ref<Record<number, { runId?: number; totalBytes: number; totalCount: number }>>({})

  const STUCK_MS = 25000

  return {
    activeRuns,
    globalStats,
    showGlobalStatsModal,
    activeRunLookup,
    lastNonDecreasingTotalsByTask,
    STUCK_MS,
  }
}

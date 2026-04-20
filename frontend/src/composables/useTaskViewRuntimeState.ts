import { ref } from 'vue'
import { useActiveRunLookup } from './useActiveRunLookup'

export function useTaskViewRuntimeState() {
  const activeRuns = ref<any[]>([])
  const globalStats = ref<any>({})
  const showGlobalStatsModal = ref(false)
  const activeRunLookup = useActiveRunLookup(activeRuns)
  const lastRunningProgressByTask = ref<Record<number, { sp: any; at: number }>>({})
  const lastNonDecreasingTotalsByTask = ref<Record<number, { totalBytes: number; totalCount: number }>>({})

  const LINGER_MS = 20000
  const STUCK_MS = 25000

  return {
    activeRuns,
    globalStats,
    showGlobalStatsModal,
    activeRunLookup,
    lastRunningProgressByTask,
    lastNonDecreasingTotalsByTask,
    LINGER_MS,
    STUCK_MS,
  }
}

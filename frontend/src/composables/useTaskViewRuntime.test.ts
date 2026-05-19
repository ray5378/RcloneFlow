import { describe, it, expect, vi } from 'vitest'
import { ref } from 'vue'
import { useTaskViewRuntime } from './useTaskViewRuntime'

describe('useTaskViewRuntime', () => {
  const createOptions = () => ({
    tasks: ref([]),
    remotes: ref([]),
    schedules: ref([]),
    runs: ref([]),
    runsTotal: ref(0),
    runsPage: ref(1),
    runsPageSize: 10,
    activeRuns: ref([]),
    globalStats: ref({}),
    showGlobalStatsModal: ref(false),
    activeRunLookup: { getActiveRunByTaskId: vi.fn() },
    lastNonDecreasingTotalsByTask: ref({}),
    currentModule: ref<'history' | 'add' | 'tasks'>('tasks'),
    stuckMs: 25000,
    taskApi: { list: vi.fn().mockResolvedValue([]) },
    remoteApi: { list: vi.fn().mockResolvedValue({ remotes: [] }) },
    scheduleApi: { list: vi.fn().mockResolvedValue([]) },
    runApi: { list: vi.fn().mockResolvedValue({ runs: [], total: 0 }) },
    jobApi: { list: vi.fn().mockResolvedValue([]) },
  })

  it('should return expected methods', () => {
    const options = createOptions()
    const runtime = useTaskViewRuntime(options)

    expect(runtime).toHaveProperty('loadData')
    expect(runtime).toHaveProperty('loadActiveRuns')
    expect(runtime).toHaveProperty('loadGlobalStats')
    expect(runtime).toHaveProperty('openGlobalStats')
    expect(runtime).toHaveProperty('setupRealtimeSync')
    expect(runtime).toHaveProperty('getRunProgressFromSummary')
    expect(runtime).toHaveProperty('getRealtimeProgressByRun')
    expect(runtime).toHaveProperty('getRunningProgressByRun')
    expect(runtime).toHaveProperty('getTaskCardProgressByTask')
    expect(runtime).toHaveProperty('getRunningProgressByTask')
    expect(runtime).toHaveProperty('formatBps')
    expect(runtime).toHaveProperty('calcEtaFromAvg')
    expect(runtime).toHaveProperty('triggerAutoRefresh')
  })
})

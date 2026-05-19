import { describe, it, expect, vi } from 'vitest'
import { ref } from 'vue'
import { useTaskViewDataSync } from './useTaskViewDataSync'

describe('useTaskViewDataSync', () => {
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
    currentModule: ref<'history' | 'add' | 'tasks'>('tasks'),
    lastNonDecreasingTotalsByTask: ref({}),
    taskApi: { list: vi.fn().mockResolvedValue([]) },
    remoteApi: { list: vi.fn().mockResolvedValue({ remotes: [] }) },
    scheduleApi: { list: vi.fn().mockResolvedValue([]) },
    runApi: { list: vi.fn().mockResolvedValue({ runs: [], total: 0 }) },
    jobApi: { list: vi.fn().mockResolvedValue([]) },
  })

  it('should return expected methods', () => {
    const options = createOptions()
    const sync = useTaskViewDataSync(options)

    expect(sync).toHaveProperty('loadData')
    expect(sync).toHaveProperty('loadActiveRuns')
    expect(sync).toHaveProperty('loadGlobalStats')
    expect(sync).toHaveProperty('openGlobalStats')
    expect(sync).toHaveProperty('setupRealtimeSync')
  })

  it('openGlobalStats should show modal', () => {
    const options = createOptions()
    const { openGlobalStats } = useTaskViewDataSync(options)

    expect(options.showGlobalStatsModal.value).toBe(false)
    openGlobalStats()
    expect(options.showGlobalStatsModal.value).toBe(true)
  })

  it('loadActiveRuns should be a function', () => {
    const options = createOptions()
    const { loadActiveRuns } = useTaskViewDataSync(options)

    expect(typeof loadActiveRuns).toBe('function')
  })

  it('loadGlobalStats should be a function', () => {
    const options = createOptions()
    const { loadGlobalStats } = useTaskViewDataSync(options)

    expect(typeof loadGlobalStats).toBe('function')
  })

  it('setupRealtimeSync should be a function', () => {
    const options = createOptions()
    const { setupRealtimeSync } = useTaskViewDataSync(options)

    expect(typeof setupRealtimeSync).toBe('function')
  })
})

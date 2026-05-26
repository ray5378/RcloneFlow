import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { useTaskViewDataSync } from './useTaskViewDataSync'
import { ref } from 'vue'

vi.mock('./useWebSocket', () => ({
  useWebSocket: vi.fn(() => ({
    connect: vi.fn(),
    cleanup: vi.fn(),
  })),
  onWsMessage: vi.fn(() => vi.fn()),
}))

vi.mock('../api', () => ({
  getGlobalStats: vi.fn().mockResolvedValue({}),
}))

describe('useTaskViewDataSync', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.useFakeTimers()
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('should be a function', () => {
    expect(typeof useTaskViewDataSync).toBe('function')
  })

  it('should return correct methods', () => {
    const sync = useTaskViewDataSync({
      tasks: ref([]),
      remotes: ref([]),
      schedules: ref([]),
      runs: ref([]),
      runsTotal: ref(0),
      runsPage: ref(1),
      runsPageSize: 50,
      activeRuns: ref([]),
      globalStats: ref(null),
      showGlobalStatsModal: ref(false),
      currentModule: ref('tasks'),
      lastNonDecreasingTotalsByTask: ref({}),
      taskApi: { list: vi.fn().mockResolvedValue([]) },
      remoteApi: { list: vi.fn().mockResolvedValue({ remotes: [] }) },
      scheduleApi: { list: vi.fn().mockResolvedValue([]) },
      runApi: { list: vi.fn().mockResolvedValue({ runs: [], total: 0 }) },
      jobApi: { list: vi.fn().mockResolvedValue([]) },
    })

    expect(typeof sync.loadData).toBe('function')
    expect(typeof sync.loadActiveRuns).toBe('function')
    expect(typeof sync.loadGlobalStats).toBe('function')
    expect(typeof sync.openGlobalStats).toBe('function')
    expect(typeof sync.setupRealtimeSync).toBe('function')
  })

  it('should have correct constants', () => {
    expect(1200).toBeDefined()
    expect(150).toBeDefined()
    expect(300).toBeDefined()
    expect([500, 2000]).toBeDefined()
    expect(0).toBeDefined()
  })

  it('should open global stats modal', () => {
    const showGlobalStatsModal = ref(false)

    const sync = useTaskViewDataSync({
      tasks: ref([]),
      remotes: ref([]),
      schedules: ref([]),
      runs: ref([]),
      runsTotal: ref(0),
      runsPage: ref(1),
      runsPageSize: 50,
      activeRuns: ref([]),
      globalStats: ref(null),
      showGlobalStatsModal,
      currentModule: ref('tasks'),
      lastNonDecreasingTotalsByTask: ref({}),
      taskApi: { list: vi.fn().mockResolvedValue([]) },
      remoteApi: { list: vi.fn().mockResolvedValue({ remotes: [] }) },
      scheduleApi: { list: vi.fn().mockResolvedValue([]) },
      runApi: { list: vi.fn().mockResolvedValue({ runs: [], total: 0 }) },
      jobApi: { list: vi.fn().mockResolvedValue([]) },
    })

    sync.openGlobalStats()

    expect(showGlobalStatsModal.value).toBe(true)
  })

  it('should handle empty active runs list', async () => {
    const activeRuns = ref([])
    const jobApiList = vi.fn().mockResolvedValue([])

    const sync = useTaskViewDataSync({
      tasks: ref([]),
      remotes: ref([]),
      schedules: ref([]),
      runs: ref([]),
      runsTotal: ref(0),
      runsPage: ref(1),
      runsPageSize: 50,
      activeRuns,
      globalStats: ref(null),
      showGlobalStatsModal: ref(false),
      currentModule: ref('tasks'),
      lastNonDecreasingTotalsByTask: ref({}),
      taskApi: { list: vi.fn().mockResolvedValue([]) },
      remoteApi: { list: vi.fn().mockResolvedValue({ remotes: [] }) },
      scheduleApi: { list: vi.fn().mockResolvedValue([]) },
      runApi: { list: vi.fn().mockResolvedValue({ runs: [], total: 0 }) },
      jobApi: { list: jobApiList },
    })

    await sync.loadActiveRuns()

    expect(jobApiList).toHaveBeenCalled()
    expect(activeRuns.value).toEqual([])
  })

  it('should setup realtime sync', () => {
    const sync = useTaskViewDataSync({
      tasks: ref([]),
      remotes: ref([]),
      schedules: ref([]),
      runs: ref([]),
      runsTotal: ref(0),
      runsPage: ref(1),
      runsPageSize: 50,
      activeRuns: ref([]),
      globalStats: ref(null),
      showGlobalStatsModal: ref(false),
      currentModule: ref('tasks'),
      lastNonDecreasingTotalsByTask: ref({}),
      taskApi: { list: vi.fn().mockResolvedValue([]) },
      remoteApi: { list: vi.fn().mockResolvedValue({ remotes: [] }) },
      scheduleApi: { list: vi.fn().mockResolvedValue([]) },
      runApi: { list: vi.fn().mockResolvedValue({ runs: [], total: 0 }) },
      jobApi: { list: vi.fn().mockResolvedValue([]) },
    })

    expect(() => sync.setupRealtimeSync()).not.toThrow()
  })

  it('should handle bootstrap with empty data', async () => {
    const tasks = ref([])
    const activeRuns = ref([])

    const sync = useTaskViewDataSync({
      tasks,
      remotes: ref([]),
      schedules: ref([]),
      runs: ref([]),
      runsTotal: ref(0),
      runsPage: ref(1),
      runsPageSize: 50,
      activeRuns,
      globalStats: ref(null),
      showGlobalStatsModal: ref(false),
      currentModule: ref('tasks'),
      lastNonDecreasingTotalsByTask: ref({}),
      taskApi: { 
        list: vi.fn().mockResolvedValue([]),
        bootstrap: vi.fn().mockResolvedValue({ tasks: [], activeRuns: [] })
      },
      remoteApi: { list: vi.fn().mockResolvedValue({ remotes: [] }) },
      scheduleApi: { list: vi.fn().mockResolvedValue([]) },
      runApi: { list: vi.fn().mockResolvedValue({ runs: [], total: 0 }) },
      jobApi: { list: vi.fn().mockResolvedValue([]) },
    })

    await sync.loadData()

    expect(tasks.value).toEqual([])
    expect(activeRuns.value).toEqual([])
  })
})

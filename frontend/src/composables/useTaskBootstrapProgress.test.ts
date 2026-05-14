import { describe, it, expect } from 'vitest'
import { createApp, defineComponent, h, nextTick, ref } from 'vue'
import { useTaskViewDataSync } from './useTaskViewDataSync'
import { useTaskProgressSync } from './useTaskProgressSync'
import { useActiveRunLookup } from './useActiveRunLookup'

describe('task bootstrap active run progress', () => {
  it('lets task cards read total count from bootstrap activeRuns on first load', async () => {
    let captured: any = null

    const Host = defineComponent({
      setup() {
        const tasks = ref<any[]>([])
        const remotes = ref<string[]>([])
        const schedules = ref<any[]>([])
        const runs = ref<any[]>([])
        const runsTotal = ref(0)
        const runsPage = ref(1)
        const activeRuns = ref<any[]>([])
        const globalStats = ref<any>({})
        const showGlobalStatsModal = ref(false)
        const currentModule = ref<'history' | 'add' | 'tasks'>('tasks')
        const lastNonDecreasingTotalsByTask = ref<Record<number, { runId?: number; totalBytes: number; totalCount: number }>>({})

        const activeRunLookup = useActiveRunLookup(activeRuns)
        const dataSync = useTaskViewDataSync({
          tasks,
          remotes,
          schedules,
          runs,
          runsTotal,
          runsPage,
          runsPageSize: 50,
          activeRuns,
          globalStats,
          showGlobalStatsModal,
          currentModule,
          lastNonDecreasingTotalsByTask,
          taskApi: {
            bootstrap: async () => ({
              tasks: [{ id: 7, name: 'demo task' }],
              activeRuns: [{
                runRecord: { id: 31, taskId: 7, status: 'running' },
                progress: {
                  bytes: 10,
                  totalBytes: 100,
                  speed: 1,
                  percentage: 10,
                  eta: 90,
                  completedFiles: 1,
                  plannedFiles: 2050,
                  logicalTotalCount: 2050,
                  totalCount: 2050,
                },
              }],
            }),
            list: async () => [],
          },
          remoteApi: { list: async () => ({ remotes: [] }) },
          scheduleApi: { list: async () => [] },
          runApi: { list: async () => ({ runs: [], total: 0 }) },
          jobApi: { list: async () => [] },
        })

        const progressSync = useTaskProgressSync({
          runs,
          activeRuns,
          activeRunLookup,
          loadData: dataSync.loadData,
          loadActiveRuns: async () => {},
        })

        captured = {
          loadData: dataSync.loadData,
          getTaskCardProgressByTask: progressSync.getTaskCardProgressByTask,
        }

        return () => h('div')
      },
    })

    const container = document.createElement('div')
    const app = createApp(Host)
    app.mount(container)

    try {
      await captured.loadData()
      await nextTick()
      const progress = captured.getTaskCardProgressByTask(7)
      expect(progress).toBeTruthy()
      expect(progress?.totalCount).toBe(2050)
      expect((progress as any)?.logicalTotalCount ?? progress?.totalCount).toBe(2050)
      expect((progress as any)?.plannedFiles ?? progress?.totalCount).toBe(2050)
    } finally {
      app.unmount()
      container.remove()
    }
  })
})

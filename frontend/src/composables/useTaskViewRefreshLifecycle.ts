import { onMounted, onUnmounted, watch, type Ref } from 'vue'

const INITIAL_POLL_DELAY_MS = 1500
const MODULE_CHANGE_POLL_DELAY_MS = 300
const STUCK_CHECK_INTERVAL_MS = 1000

export function useTaskViewRefreshLifecycle(options: {
  tasks: Ref<any[]>
  activeRuns: Ref<any[]>
  currentModule?: Ref<'history' | 'add' | 'tasks'>
  getRunningProgressByTask: (taskId: number) => any
  loadData: () => Promise<void> | void
  loadActiveRuns: () => Promise<void>
  setupRealtimeSync?: () => void
  stuckMs: number
}) {
  const ACTIVE_POLL_FAST_MS = 3000
  const ACTIVE_POLL_IDLE_MS = 12000

  let lastRenderedSignature = ''
  let lastStuckRefreshTime = 0
  let stuckTimer: number | null = null
  let activePollTimer: number | null = null

  function stopActivePollLoop() {
    if (activePollTimer) {
      clearTimeout(activePollTimer)
      activePollTimer = null
    }
  }

  function getActivePollDelay() {
    const activeCount = Array.isArray(options.activeRuns.value) ? options.activeRuns.value.length : 0
    return activeCount > 0 ? ACTIVE_POLL_FAST_MS : ACTIVE_POLL_IDLE_MS
  }

  function scheduleNextActivePoll(delay?: number) {
    stopActivePollLoop()
    if (options.currentModule && options.currentModule.value !== 'tasks') return
    const nextDelay = typeof delay === 'number' ? delay : getActivePollDelay()
    activePollTimer = window.setTimeout(async () => {
      activePollTimer = null
      try {
        if (document.visibilityState === 'visible') {
          await options.loadActiveRuns()
        }
      } catch (e) {
        console.error('[useTaskViewRefreshLifecycle] Failed to load active runs:', e)
      } finally {
        scheduleNextActivePoll()
      }
    }, nextDelay)
  }

  function restartActivePollLoop(delay?: number) {
    scheduleNextActivePoll(delay)
  }

  if (options.currentModule) {
    watch(options.currentModule, (next) => {
      if (next === 'tasks') {
        Promise.all([
          Promise.resolve(options.loadData()).catch((e) => {
            console.error('[useTaskViewRefreshLifecycle] Failed to load data on module change:', e)
          }),
          options.loadActiveRuns().catch((e) => {
            console.error('[useTaskViewRefreshLifecycle] Failed to load active runs on module change:', e)
          }),
        ]).catch((e) => {
          console.error('[useTaskViewRefreshLifecycle] Failed to handle module change:', e)
        })
        setTimeout(() => {
          options.loadActiveRuns().catch((e) => {
            console.error('[useTaskViewRefreshLifecycle] Failed to load active runs after module change:', e)
          })
        }, MODULE_CHANGE_POLL_DELAY_MS)
        restartActivePollLoop(INITIAL_POLL_DELAY_MS)
      } else {
        stopActivePollLoop()
      }
    })
  }

  watch(() => (options.activeRuns.value || []).length, () => {
    if (options.currentModule && options.currentModule.value !== 'tasks') return
    restartActivePollLoop()
  })

  onMounted(() => {
    Promise.all([
      Promise.resolve(options.loadData()).catch((e) => {
        console.error('[useTaskViewRefreshLifecycle] Failed to load data on mount:', e)
      }),
      options.loadActiveRuns().catch((e) => {
        console.error('[useTaskViewRefreshLifecycle] Failed to load active runs on mount:', e)
      }),
    ]).catch((e) => {
      console.error('[useTaskViewRefreshLifecycle] Failed to initialize on mount:', e)
    })
    options.setupRealtimeSync?.()

    restartActivePollLoop(INITIAL_POLL_DELAY_MS)

    stuckTimer = window.setInterval(() => {
      try {
        if (options.currentModule && options.currentModule.value !== 'tasks') return
        const activeTasks = new Set<number>()
        for (const item of options.activeRuns.value || []) {
          const taskId = Number(item?.runRecord?.taskId ?? item?.taskId ?? item?.taskID ?? item?.task_id)
          if (taskId > 0) activeTasks.add(taskId)
        }
        if (activeTasks.size === 0) {
          lastRenderedSignature = ''
          return
        }
        const sigParts: string[] = []
        for (const taskId of activeTasks) {
          const progress = options.getRunningProgressByTask(taskId) as any
          const pct = progress ? Number(progress.percentage || 0).toFixed(3) : 'na'
          const c = progress ? Number(progress.completedFiles || 0) : -1
          sigParts.push(`${taskId}:${pct}:${c}`)
        }
        const sig = `${activeTasks.size}|${sigParts.join(',')}`
        if (sig === lastRenderedSignature) {
          const now = Date.now()
          const last = lastStuckRefreshTime
          if (now - last > options.stuckMs) {
            lastStuckRefreshTime = now
            options.loadData()
          }
        } else {
          lastRenderedSignature = sig
        }
      } catch (e) {
        console.error('[useTaskViewRefreshLifecycle] Error in stuck detection timer:', e)
      }
    }, STUCK_CHECK_INTERVAL_MS)

  })

  onUnmounted(() => {
    if (stuckTimer) {
      clearInterval(stuckTimer)
      stuckTimer = null
    }
    stopActivePollLoop()
  })

  return {}
}

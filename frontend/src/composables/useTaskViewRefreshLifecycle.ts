import { onMounted, onUnmounted, watch, type Ref } from 'vue'

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
  let lastRenderedSignature = ''
  let stuckTimer: number | null = null
  let activePollTimer: number | null = null

  function stopActivePollLoop() {
    if (activePollTimer) {
      clearInterval(activePollTimer)
      activePollTimer = null
    }
  }

  function startActivePollLoop() {
    stopActivePollLoop()
    if (options.currentModule && options.currentModule.value !== 'tasks') return
    activePollTimer = window.setInterval(() => {
      if (document.visibilityState === 'visible') {
        options.loadActiveRuns().catch(console.error)
      }
    }, 3000)
  }

  if (options.currentModule) {
    watch(options.currentModule, (next) => {
      if (next === 'tasks') {
        Promise.all([
          Promise.resolve(options.loadData()).catch(console.error),
          options.loadActiveRuns().catch(console.error),
        ]).catch(console.error)
        setTimeout(() => {
          options.loadActiveRuns().catch(console.error)
        }, 300)
        startActivePollLoop()
      } else {
        stopActivePollLoop()
      }
    })
  }

  onMounted(() => {
    Promise.all([
      Promise.resolve(options.loadData()).catch(console.error),
      options.loadActiveRuns().catch(console.error),
    ]).catch(console.error)
    options.setupRealtimeSync?.()

    startActivePollLoop()

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
          const last = (window as any).__last_stuck_refresh || 0
          if (now - last > options.stuckMs) {
            ;(window as any).__last_stuck_refresh = now
            options.loadData()
          }
        } else {
          lastRenderedSignature = sig
        }
      } catch {}
    }, 1000)

  })

  onUnmounted(() => {
    if (stuckTimer) {
      clearInterval(stuckTimer)
      stuckTimer = null
    }
    if (activePollTimer) {
      clearInterval(activePollTimer)
      activePollTimer = null
    }
  })

  return {}
}

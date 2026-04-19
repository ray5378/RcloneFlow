import { onMounted, onUnmounted, type Ref } from 'vue'

export function useTaskViewRefreshLifecycle(options: {
  tasks: Ref<any[]>
  getDeNoisedStableByTask: (taskId: number) => any
  loadData: () => Promise<void> | void
  loadActiveRuns: () => Promise<void>
  stuckMs: number
}) {
  let lastRenderedSignature = ''
  let stuckTimer: number | null = null
  let activePollTimer: number | null = null

  onMounted(() => {
    stuckTimer = window.setInterval(() => {
      try {
        const sigParts: string[] = []
        for (const t of options.tasks.value || []) {
          const sp = options.getDeNoisedStableByTask((t as any).id) as any
          const pct = sp ? Number(sp.percentage || 0).toFixed(3) : 'na'
          const c = sp ? Number(sp.completedFiles || 0) : -1
          sigParts.push(`${(t as any).id}:${pct}:${c}`)
        }
        const sig = `${options.tasks.value?.length || 0}|${sigParts.join(',')}`
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

    activePollTimer = window.setInterval(() => {
      if (document.visibilityState === 'visible') {
        options.loadActiveRuns().catch(console.error)
      }
    }, 3000)
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

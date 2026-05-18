import { computed, ref, type Ref } from 'vue'
import type { ActiveRun } from '../api/run'
import { getActiveProgress, getActiveProgressText } from '../components/task/runningHint'

export function useRunningHint(activeRuns: Ref<ActiveRun[]>, openRunLog: (run: any) => void) {
  const visible = ref(false)
  const run = ref<any>(null)

  const activeRunByTaskId = computed(() => {
    const index = new Map<number, ActiveRun>()
    for (const item of activeRuns.value || []) {
      const activeTaskId = Number(item?.runRecord?.taskId ?? item?.taskId)
      if (activeTaskId > 0) index.set(activeTaskId, item)
    }
    return index
  })

  const active = computed(() => {
    const taskId = Number(run.value?.taskId)
    if (!taskId) return null
    return activeRunByTaskId.value.get(taskId) || null
  })

  const phaseText = computed(() => getActiveProgress(active.value)?.phase || '-')
  const progressText = computed(() => getActiveProgressText(active.value))

  function open(nextRun: any) {
    run.value = nextRun
    visible.value = true
  }

  function close() {
    visible.value = false
  }

  function openLog() {
    openRunLog(run.value)
    close()
  }

  return {
    visible,
    run,
    phaseText,
    progressText,
    open,
    close,
    openLog,
  }
}

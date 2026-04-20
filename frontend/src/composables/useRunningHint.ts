import { computed, ref, type Ref } from 'vue'
import { getActiveProgress, getActiveProgressText, getRunningHintDebug } from '../components/task/runningHint'

const EMPTY_DEBUG_INFO = {
  checkText: '-',
  progressLine: '-',
  progressJson: '-',
}

const RUNNING_HINT_DEBUG_ENABLED = import.meta.env.DEV

export function useRunningHint(activeRuns: Ref<any[]>, openRunLog: (run: any) => void) {
  const visible = ref(false)
  const run = ref<any>(null)
  const debugOpen = ref(false)

  const active = computed(() => {
    const taskId = run.value?.taskId
    if (!taskId) return null
    return (activeRuns.value || []).find((item: any) => {
      const activeTaskId = item?.runRecord?.taskId ?? item?.taskId
      return activeTaskId === taskId
    }) || null
  })

  const phaseText = computed(() => getActiveProgress(active.value)?.phase || '-')
  const progressText = computed(() => getActiveProgressText(active.value))
  const debugInfo = computed(() => {
    const info = getRunningHintDebug(active.value)
    return info || EMPTY_DEBUG_INFO
  })

  function open(nextRun: any) {
    run.value = nextRun
    visible.value = true
  }

  function close() {
    visible.value = false
    debugOpen.value = false
  }

  function toggleDebug() {
    if (!RUNNING_HINT_DEBUG_ENABLED) return
    debugOpen.value = !debugOpen.value
  }

  function openLog() {
    openRunLog(run.value)
    close()
  }

  return {
    visible,
    run,
    debugOpen,
    debugEnabled: RUNNING_HINT_DEBUG_ENABLED,
    phaseText,
    progressText,
    debugInfo,
    open,
    close,
    toggleDebug,
    openLog,
  }
}

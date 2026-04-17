import { computed, ref, type Ref } from 'vue'
import { getActiveProgress, getActiveProgressText, getRunningHintDebug } from '../components/task/runningHint'

export function useRunningHint(activeRuns: Ref<any[]>, openRunLog: (run: any) => void) {
  const visible = ref(false)
  const run = ref<any>(null)
  const debugOpen = ref(false)

  const active = computed(() => {
    if (!run.value?.taskId) return null
    return (activeRuns.value || []).find((item: any) => item?.taskId === run.value.taskId) || null
  })

  const phaseText = computed(() => getActiveProgress(active.value)?.phase || '-')
  const progressText = computed(() => getActiveProgressText(active.value))
  const debugInfo = computed(() => getRunningHintDebug(active.value))

  function open(nextRun: any) {
    run.value = nextRun
    visible.value = true
  }

  function close() {
    visible.value = false
    debugOpen.value = false
  }

  function toggleDebug() {
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
    phaseText,
    progressText,
    debugInfo,
    open,
    close,
    toggleDebug,
    openLog,
  }
}

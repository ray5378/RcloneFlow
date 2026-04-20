import { computed, ref, type Ref } from 'vue'
import type { ActiveRun } from '../api/run'
import { getActiveProgress, getActiveProgressText, getRunningHintDebug } from '../components/task/runningHint'

const EMPTY_DEBUG_INFO = {
  checkText: '-',
  progressLine: '-',
  progressJson: '-',
}

// 运行中提示小窗的调试详情不属于主展示链，只是辅助排障能力。
// 现在它由默认配置 RUNNING_HINT_DEBUG_ENABLED 控制：
// - 默认关闭
// - 开启后才允许展开自检 / 日志原文 / 接口进度 JSON
export function useRunningHint(activeRuns: Ref<ActiveRun[]>, openRunLog: (run: any) => void, debugEnabled = false) {
  const visible = ref(false)
  const run = ref<any>(null)
  const debugOpen = ref(false)

  const active = computed(() => {
    const taskId = run.value?.taskId
    if (!taskId) return null
    return (activeRuns.value || []).find((item: ActiveRun & { taskId?: number }) => {
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
    if (!debugEnabled) return
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
    debugEnabled,
    phaseText,
    progressText,
    debugInfo,
    open,
    close,
    toggleDebug,
    openLog,
  }
}

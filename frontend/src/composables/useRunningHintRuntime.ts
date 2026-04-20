import type { Ref } from 'vue'
import { useRunningHint } from './useRunningHint'

export function useRunningHintRuntime(activeRuns: Ref<any[]>, openRunLog: (run: any) => void, debugEnabled = false) {
  const hint = useRunningHint(activeRuns, openRunLog, debugEnabled)

  return {
    runningHintVisible: hint.visible,
    runningHintRun: hint.run,
    runningHintDebugEnabled: hint.debugEnabled,
    runningHintDebugOpen: hint.debugOpen,
    runningHintPhaseText: hint.phaseText,
    runningHintProgressText: hint.progressText,
    runningHintDebugInfo: hint.debugInfo,
    openRunningHint: hint.open,
    closeRunningHint: hint.close,
    toggleRunningHintDebug: hint.toggleDebug,
    openRunningHintLog: hint.openLog,
  }
}

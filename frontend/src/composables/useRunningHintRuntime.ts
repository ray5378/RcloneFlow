import type { Ref } from 'vue'
import { useRunningHint } from './useRunningHint'

export function useRunningHintRuntime(activeRuns: Ref<any[]>, openRunLog: (run: any) => void) {
  const hint = useRunningHint(activeRuns, openRunLog)

  return {
    runningHintVisible: hint.visible,
    runningHintRun: hint.run,
    runningHintPhaseText: hint.phaseText,
    runningHintProgressText: hint.progressText,
    openRunningHint: hint.open,
    closeRunningHint: hint.close,
    openRunningHintLog: hint.openLog,
  }
}

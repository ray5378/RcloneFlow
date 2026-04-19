import { computed, type Ref } from 'vue'
import { getActiveProgress, getActiveProgressText } from '../components/task/runningHint'

export function useActiveRunLookup(activeRuns: Ref<any[]>) {
  function getActiveRunByTaskId(taskId: number) {
    const targetId = Number(taskId)
    const cur = (activeRuns.value || []).find((item: any) => {
      const candidateId = Number(item?.runRecord?.taskId ?? item?.taskId ?? item?.taskID ?? item?.task_id)
      return candidateId > 0 && candidateId === targetId
    })
    if (cur) return cur
    return undefined as any
  }

  function getActiveProgressByTaskId(taskId: number) {
    const active = getActiveRunByTaskId(taskId)
    return getActiveProgress(active)
  }

  function getActiveProgressTextByTaskId(taskId: number) {
    const active = getActiveRunByTaskId(taskId)
    return getActiveProgressText(active)
  }

  return {
    getActiveRunByTaskId,
    getActiveProgressByTaskId,
    getActiveProgressTextByTaskId,
  }
}

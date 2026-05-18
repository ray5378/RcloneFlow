import { computed, type Ref } from 'vue'
import { getActiveProgress, getActiveProgressText } from '../components/task/runningHint'

export function useActiveRunLookup(activeRuns: Ref<any[]>) {
  const activeRunByTaskId = computed(() => {
    const index = new Map<number, any>()
    for (const item of activeRuns.value || []) {
      const candidateId = Number(item?.runRecord?.taskId ?? item?.taskId ?? item?.taskID ?? item?.task_id)
      if (candidateId > 0) index.set(candidateId, item)
    }
    return index
  })

  function getActiveRunByTaskId(taskId: number) {
    return activeRunByTaskId.value.get(Number(taskId))
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

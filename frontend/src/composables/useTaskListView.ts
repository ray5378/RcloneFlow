import { computed, ref, watch, type Ref } from 'vue'
import type { Task } from '../types'

export function useTaskListView(tasks: Ref<Task[]>, runningTaskIds?: Ref<Set<number>>) {
  const tasksPage = ref(1)
  const tasksPageSize = 10
  const tasksJumpPage = ref(1)
  const taskSearch = ref('')

  const filteredTasksRaw = computed(() => {
    let result = tasks.value
    const q = taskSearch.value
    if (q) {
      const lq = q.toLowerCase()
      result = result.filter(t =>
        t.name.toLowerCase().includes(lq) ||
        t.sourceRemote.toLowerCase().includes(lq) ||
        t.targetRemote.toLowerCase().includes(lq) ||
        t.mode.toLowerCase().includes(lq)
      )
    }
    if (runningTaskIds?.value?.size) {
      const ids = runningTaskIds.value
      const running: Task[] = []
      const others: Task[] = []
      for (const t of result) {
        if (ids.has(t.id)) { running.push(t) } else { others.push(t) }
      }
      running.sort((a, b) => a.id - b.id)
      result = [...running, ...others]
    }
    return result
  })

  const tasksTotal = computed(() => filteredTasksRaw.value.length)
  const currentTasksPages = computed(() => Math.max(1, Math.ceil(tasksTotal.value / tasksPageSize)))

  const filteredTasks = computed(() => {
    const start = (tasksPage.value - 1) * tasksPageSize
    const end = start + tasksPageSize
    return filteredTasksRaw.value.slice(start, end)
  })

  watch(taskSearch, () => {
    tasksPage.value = 1
    tasksJumpPage.value = 1
  })

  watch(currentTasksPages, (pages) => {
    if (tasksPage.value > pages) {
      tasksPage.value = pages
    }
    if ((tasksJumpPage.value || 1) > pages) {
      tasksJumpPage.value = pages
    }
    if ((tasksJumpPage.value || 1) < 1) {
      tasksJumpPage.value = 1
    }
  })

  function jumpToTasksPage() {
    const page = Math.min(Math.max(1, tasksJumpPage.value || 1), currentTasksPages.value)
    tasksPage.value = page
    tasksJumpPage.value = page
  }

  return {
    tasksPage,
    tasksPageSize,
    tasksJumpPage,
    taskSearch,
    tasksTotal,
    currentTasksPages,
    filteredTasksRaw,
    filteredTasks,
    jumpToTasksPage,
  }
}

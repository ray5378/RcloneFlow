import { computed, ref, type Ref } from 'vue'
import type { Task } from '../types'

export function useTaskListView(tasks: Ref<Task[]>) {
  const tasksPage = ref(1)
  const tasksPageSize = 20
  const tasksJumpPage = ref(1)
  const taskSearch = ref('')

  const filteredTasksRaw = computed(() => {
    if (!taskSearch.value) return tasks.value
    const q = taskSearch.value.toLowerCase()
    return tasks.value.filter(t =>
      t.name.toLowerCase().includes(q) ||
      t.sourceRemote.toLowerCase().includes(q) ||
      t.targetRemote.toLowerCase().includes(q) ||
      t.mode.toLowerCase().includes(q)
    )
  })

  const tasksTotal = computed(() => filteredTasksRaw.value.length)
  const currentTasksPages = computed(() => Math.max(1, Math.ceil(tasksTotal.value / tasksPageSize)))

  const filteredTasks = computed(() => {
    const start = (tasksPage.value - 1) * tasksPageSize
    const end = start + tasksPageSize
    return filteredTasksRaw.value.slice(start, end)
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

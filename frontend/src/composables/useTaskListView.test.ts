import { describe, it, expect } from 'vitest'
import { ref } from 'vue'
import { useTaskListView } from './useTaskListView'

describe('useTaskListView', () => {
  const createTasks = (count: number) =>
    Array.from({ length: count }, (_, i) => ({
      id: i + 1,
      name: `task-${i + 1}`,
      mode: 'copy',
      sourceRemote: 'src',
      targetRemote: 'dst',
      sourcePath: '/src',
      targetPath: '/dst',
    }))

  it('should initialize with default values', () => {
    const tasks = ref(createTasks(5))
    const { tasksPage, tasksPageSize, tasksJumpPage, taskSearch } = useTaskListView(tasks)

    expect(tasksPage.value).toBe(1)
    expect(tasksPageSize).toBe(10)
    expect(tasksJumpPage.value).toBe(1)
    expect(taskSearch.value).toBe('')
  })

  it('should filter tasks by search', async () => {
    const tasks = ref(createTasks(3))
    tasks.value[0].name = 'backup-task'
    const { taskSearch, filteredTasksRaw } = useTaskListView(tasks)

    taskSearch.value = 'backup'
    await new Promise(r => setTimeout(r, 0))

    expect(filteredTasksRaw.value).toHaveLength(1)
    expect(filteredTasksRaw.value[0].name).toBe('backup-task')
  })

  it('should paginate tasks', async () => {
    const tasks = ref(createTasks(15))
    const { tasksPage, filteredTasks, tasksTotal } = useTaskListView(tasks)

    expect(tasksTotal.value).toBe(15)
    expect(filteredTasks.value).toHaveLength(10)

    tasksPage.value = 2
    await new Promise(r => setTimeout(r, 0))

    expect(filteredTasks.value).toHaveLength(5)
  })

  it('should reset page on search change', async () => {
    const tasks = ref(createTasks(25))
    const { tasksPage, taskSearch, tasksJumpPage } = useTaskListView(tasks)

    tasksPage.value = 3
    tasksJumpPage.value = 3

    taskSearch.value = 'test'
    await new Promise(r => setTimeout(r, 0))

    expect(tasksPage.value).toBe(1)
    expect(tasksJumpPage.value).toBe(1)
  })

  it('should jump to page', () => {
    const tasks = ref(createTasks(25))
    const { tasksJumpPage, tasksPage, currentTasksPages, jumpToTasksPage } = useTaskListView(tasks)

    expect(currentTasksPages.value).toBe(3)

    tasksJumpPage.value = 2
    jumpToTasksPage()

    expect(tasksPage.value).toBe(2)
    expect(tasksJumpPage.value).toBe(2)
  })

  it('should clamp jump page to valid range', () => {
    const tasks = ref(createTasks(15))
    const { tasksJumpPage, tasksPage, jumpToTasksPage } = useTaskListView(tasks)

    tasksJumpPage.value = 99
    jumpToTasksPage()
    expect(tasksPage.value).toBe(2)

    tasksJumpPage.value = 0
    jumpToTasksPage()
    expect(tasksPage.value).toBe(1)
  })
})

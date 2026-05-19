import { describe, it, expect, vi } from 'vitest'
import { ref } from 'vue'
import { useTaskViewPagingBridge } from './useTaskViewPagingBridge'

describe('useTaskViewPagingBridge', () => {
  const createOptions = () => ({
    taskSearch: ref(''),
    tasksJumpPage: ref<number | null>(1),
    historyStatusFilter: ref(''),
    jumpPage: ref(1),
    tasksPage: ref(1),
    runsPage: ref(1),
    currentModule: ref('tasks'),
    loadData: vi.fn().mockResolvedValue(undefined),
  })

  it('should set search and paging values', () => {
    const opts = createOptions()
    const { setTaskSearch, setTasksJumpPageValue, setHistoryStatusFilter, setJumpPageValue } = useTaskViewPagingBridge(opts)

    setTaskSearch('test')
    setTasksJumpPageValue(5)
    setHistoryStatusFilter('success')
    setJumpPageValue(3)

    expect(opts.taskSearch.value).toBe('test')
    expect(opts.tasksJumpPage.value).toBe(5)
    expect(opts.historyStatusFilter.value).toBe('success')
    expect(opts.jumpPage.value).toBe(3)
  })

  it('should navigate tasks pages', () => {
    const opts = createOptions()
    const { prevTasksPage, nextTasksPage } = useTaskViewPagingBridge(opts)

    nextTasksPage()
    expect(opts.tasksPage.value).toBe(2)

    prevTasksPage()
    expect(opts.tasksPage.value).toBe(1)
  })

  it('should navigate runs pages and load data', async () => {
    const opts = createOptions()
    const { prevRunsPage, nextRunsPage } = useTaskViewPagingBridge(opts)

    await nextRunsPage()
    expect(opts.runsPage.value).toBe(2)
    expect(opts.loadData).toHaveBeenCalledTimes(1)

    await prevRunsPage()
    expect(opts.runsPage.value).toBe(1)
    expect(opts.loadData).toHaveBeenCalledTimes(2)
  })

  it('should switch back to tasks module', () => {
    const opts = createOptions()
    opts.currentModule.value = 'history'
    const { backToTasks } = useTaskViewPagingBridge(opts)

    backToTasks()
    expect(opts.currentModule.value).toBe('tasks')
  })
})

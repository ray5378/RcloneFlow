import { describe, it, expect, vi } from 'vitest'
import { nextTick, ref } from 'vue'
import { useRunDetailFiles } from './useRunDetailFiles'

describe('useRunDetailFiles', () => {
  it('requests paged detail rows from backend with filter changes', async () => {
    const getFiles = vi.fn(async (_runId: number, offset: number, limit: number, filter = 'all') => {
      const allItems = [
        { name: 'copied-a', status: 'success' },
        { name: 'failed-a', status: 'failed' },
        { name: 'copied-b', status: 'success' },
        { name: 'skipped-a', status: 'skipped' },
      ]
      const byFilter: Record<string, any[]> = {
        all: allItems,
        success: allItems.filter(it => it.status === 'success'),
        failed: allItems.filter(it => it.status === 'failed'),
        other: allItems.filter(it => it.status === 'skipped'),
      }
      const items = byFilter[filter] || allItems
      return {
        items: items.slice(offset, offset + limit),
        total: items.length,
      }
    })
    const runDetail = ref<any>({ id: 9, taskMode: 'move' })
    const currentFinalFilter = ref<'all' | 'success' | 'failed' | 'other'>('all')

    const api = useRunDetailFiles({
      runDetail,
      currentFinalFilter,
      runApi: { getFiles },
    })

    api.runFilesPageSize.value = 10
    await api.reloadRunFiles()
    await nextTick()

    expect(getFiles).toHaveBeenCalledWith(9, 0, 10, 'all')
    expect(api.visibleRunFiles.value.map(it => it.name)).toEqual(['copied-a', 'failed-a', 'copied-b', 'skipped-a'])
    expect(api.pagedRunFiles.value.map(it => it.name)).toEqual(['copied-a', 'failed-a', 'copied-b', 'skipped-a'])
    expect(api.totalRunFilesPages.value).toBe(1)
    expect(api.runFilesTotal.value).toBe(4)

    currentFinalFilter.value = 'success'
    await nextTick()
    await nextTick()
    expect(getFiles).toHaveBeenLastCalledWith(9, 0, 10, 'success')
    expect(api.visibleRunFiles.value.map(it => it.name)).toEqual(['copied-a', 'copied-b'])

    currentFinalFilter.value = 'failed'
    await nextTick()
    await nextTick()
    expect(getFiles).toHaveBeenLastCalledWith(9, 0, 10, 'failed')
    expect(api.visibleRunFiles.value.map(it => it.name)).toEqual(['failed-a'])

    currentFinalFilter.value = 'other'
    await nextTick()
    await nextTick()
    expect(getFiles).toHaveBeenLastCalledWith(9, 0, 10, 'other')
    expect(api.visibleRunFiles.value.map(it => it.name)).toEqual(['skipped-a'])
  })

  it('resets file pagination and requests the next page from backend', async () => {
    const getFiles = vi.fn(async (_runId: number, offset: number, limit: number) => ({
      items: [{ name: `file-${offset + 1}` }],
      total: 4,
    }))
    const runDetail = ref<any>({ id: 1 })

    const api = useRunDetailFiles({
      runDetail,
      runApi: { getFiles },
    })

    api.runFilesPageSize.value = 1
    await api.reloadRunFiles()
    await nextTick()

    expect(api.runFilesPage.value).toBe(1)
    expect(api.runFilesTotal.value).toBe(4)
    expect(api.visibleRunFiles.value.length).toBe(1)
    expect(api.pagedRunFiles.value.map(it => it.name)).toEqual(['file-1'])

    api.goNextFilesPage()
    await nextTick()
    await nextTick()
    expect(api.runFilesPage.value).toBe(2)
    expect(getFiles).toHaveBeenLastCalledWith(1, 1, 1, 'all')
    expect(api.pagedRunFiles.value.map(it => it.name)).toEqual(['file-2'])

    api.resetRunFiles()
    expect(api.runFilesPage.value).toBe(1)
    expect(api.runFilesTotal.value).toBe(0)
    expect(api.pagedRunFiles.value).toEqual([])
  })
})

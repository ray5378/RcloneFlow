import { describe, it, expect, vi } from 'vitest'
import { nextTick, ref } from 'vue'
import { useRunDetailFiles } from './useRunDetailFiles'

describe('useRunDetailFiles', () => {
  it('requests paged detail rows from backend', async () => {
    const getFiles = vi.fn(async (_runId: number, offset: number, limit: number) => {
      const allItems = [
        { name: 'copied-a', status: 'success' },
        { name: 'failed-a', status: 'failed' },
        { name: 'copied-b', status: 'success' },
        { name: 'skipped-a', status: 'skipped' },
      ]
      return {
        items: allItems.slice(offset, offset + limit),
        total: allItems.length,
      }
    })
    const runDetail = ref<any>({ id: 9, taskMode: 'move' })

    const api = useRunDetailFiles({
      runDetail,
      runApi: { getFiles },
    })

    api.runFilesPageSize.value = 10
    await api.reloadRunFiles()
    await nextTick()

    expect(getFiles).toHaveBeenCalledWith(9, 0, 10)
    expect(api.visibleRunFiles.value.map(it => it.name)).toEqual(['copied-a', 'failed-a', 'copied-b', 'skipped-a'])
    expect(api.pagedRunFiles.value.map(it => it.name)).toEqual(['copied-a', 'failed-a', 'copied-b', 'skipped-a'])
    expect(api.totalRunFilesPages.value).toBe(1)
    expect(api.runFilesTotal.value).toBe(4)
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
    expect(getFiles).toHaveBeenLastCalledWith(1, 1, 1)
    expect(api.pagedRunFiles.value.map(it => it.name)).toEqual(['file-2'])

    api.resetRunFiles()
    expect(api.runFilesPage.value).toBe(1)
    expect(api.runFilesTotal.value).toBe(0)
    expect(api.pagedRunFiles.value).toEqual([])
  })
})

import { describe, it, expect, vi } from 'vitest'
import { nextTick, ref } from 'vue'
import { useRunDetailFiles } from './useRunDetailFiles'

describe('useRunDetailFiles', () => {
  it('filters deleted rows from move run details', async () => {
    const getFiles = vi.fn()
      .mockResolvedValueOnce({
        items: [
          { name: 'copied-a', status: 'success' },
          { name: 'source-a', status: 'deleted' },
          { name: 'copied-b', status: 'success' },
        ],
        total: 3,
      })
    const runDetail = ref<any>({ id: 9, taskMode: 'move' })

    const api = useRunDetailFiles({
      runDetail,
      runApi: { getFiles },
    })

    api.runFilesPageSize.value = 10
    await api.reloadRunFiles()
    await nextTick()

    expect(api.visibleRunFiles.value.map(it => it.name)).toEqual(['copied-a', 'copied-b'])
    expect(api.pagedRunFiles.value.map(it => it.name)).toEqual(['copied-a', 'copied-b'])
    expect(api.totalRunFilesPages.value).toBe(1)
  })

  it('resets file pagination when a new run detail is opened', async () => {
    const getFiles = vi.fn()
      .mockResolvedValueOnce({ items: [{ name: 'file-1' }, { name: 'file-2' }], total: 4 })
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
    expect(api.pagedRunFiles.value.map(it => it.name)).toEqual(['file-1'])

    api.goNextFilesPage()
    expect(api.runFilesPage.value).toBe(2)

    api.resetRunFiles()
    expect(api.runFilesPage.value).toBe(1)
    expect(api.runFilesTotal.value).toBe(0)
    expect(api.pagedRunFiles.value).toEqual([])
  })
})

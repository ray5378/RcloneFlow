import { describe, it, expect, vi } from 'vitest'
import { nextTick, ref } from 'vue'
import { useRunDetailFiles } from './useRunDetailFiles'

describe('useRunDetailFiles', () => {
  it('filters deleted rows from move run details and applies summary filters to visible detail rows', async () => {
    const getFiles = vi.fn()
      .mockResolvedValueOnce({
        items: [
          { name: 'copied-a', status: 'success' },
          { name: 'source-a', status: 'Deleted', action: 'Deleted' },
          { name: 'failed-a', status: 'Failed' },
          { name: 'skipped-a', status: 'Skipped' },
          { name: 'copied-b', status: 'Copied', action: 'Copied' },
        ],
        total: 5,
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

    expect(api.visibleRunFiles.value.map(it => it.name)).toEqual(['copied-a', 'failed-a', 'skipped-a', 'copied-b'])
    expect(api.pagedRunFiles.value.map(it => it.name)).toEqual(['copied-a', 'failed-a', 'skipped-a', 'copied-b'])
    expect(api.totalRunFilesPages.value).toBe(1)

    currentFinalFilter.value = 'success'
    await nextTick()
    expect(api.visibleRunFiles.value.map(it => it.name)).toEqual(['copied-a', 'copied-b'])

    currentFinalFilter.value = 'failed'
    await nextTick()
    expect(api.visibleRunFiles.value.map(it => it.name)).toEqual(['failed-a'])

    currentFinalFilter.value = 'other'
    await nextTick()
    expect(api.visibleRunFiles.value.map(it => it.name)).toEqual(['skipped-a'])
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
    expect(api.visibleRunFiles.value.length).toBe(2)
    expect(api.pagedRunFiles.value.map(it => it.name)).toEqual(['file-1'])

    api.goNextFilesPage()
    expect(api.runFilesPage.value).toBe(2)

    api.resetRunFiles()
    expect(api.runFilesPage.value).toBe(1)
    expect(api.runFilesTotal.value).toBe(0)
    expect(api.pagedRunFiles.value).toEqual([])
  })

  it('loads multiple pages, updates run detail via open, and clamps page on filter change', async () => {
    const getFiles = vi.fn()
      .mockResolvedValueOnce({
        items: Array.from({ length: 100 }, (_, i) => ({ name: `success-${i + 1}`, status: 'success' })),
        total: 150,
      })
      .mockResolvedValueOnce({
        items: Array.from({ length: 50 }, (_, i) => ({ name: `failed-${i + 1}`, status: 'failed' })),
        total: 150,
      })
    const runDetail = ref<any>({ id: 0, taskMode: 'copy' })
    const currentFinalFilter = ref<'all' | 'success' | 'failed' | 'other'>('all')

    const api = useRunDetailFiles({
      runDetail,
      currentFinalFilter,
      runApi: { getFiles },
    })

    api.runFilesPageSize.value = 25
    api.openRunDetailFiles({ id: 77, taskMode: 'copy' })
    await Promise.resolve()
    await Promise.resolve()
    await nextTick()

    expect(runDetail.value.id).toBe(77)
    expect(getFiles).toHaveBeenNthCalledWith(1, 77, 0, 100)
    expect(getFiles).toHaveBeenNthCalledWith(2, 77, 100, 100)
    expect(api.runFiles.value).toHaveLength(150)
    expect(api.runFilesTotal.value).toBe(150)
    expect(api.totalRunFilesPages.value).toBe(6)

    api.runFilesPage.value = 6
    currentFinalFilter.value = 'failed'
    await nextTick()
    expect(api.runFilesPage.value).toBe(1)
    expect(api.totalRunFilesPages.value).toBe(2)

    api.goPrevFilesPage()
    expect(api.runFilesPage.value).toBe(1)
    api.goPrevFilesPage()
    expect(api.runFilesPage.value).toBe(1)
    api.goNextFilesPage()
    expect(api.runFilesPage.value).toBe(2)
    api.goNextFilesPage()
    expect(api.runFilesPage.value).toBe(2)
  })

  it('returns early without id and swallows reload errors', async () => {
    const errorSpy = vi.spyOn(console, 'error').mockImplementation(() => {})
    const getFiles = vi.fn().mockRejectedValueOnce(new Error('boom'))
    const runDetail = ref<any>({})
    const api = useRunDetailFiles({ runDetail, runApi: { getFiles } })

    await api.reloadRunFiles()
    expect(getFiles).not.toHaveBeenCalled()

    runDetail.value = { id: 1 }
    await api.reloadRunFiles()
    expect(errorSpy).toHaveBeenCalled()
  })
})

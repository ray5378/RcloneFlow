import { describe, it, expect } from 'vitest'
import { nextTick, ref } from 'vue'
import { useRunDetailComputed } from './useRunDetailComputed'

describe('useRunDetailComputed', () => {
  it('keeps final file pagination state in sync with the active data source', async () => {
    const runDetail = ref({
      id: 101,
      taskMode: 'copy',
      summary: {
        finalSummary: {
          counts: { total: 5, copied: 3, deleted: 0, failed: 1, skipped: 1 },
          files: [
            { name: 'a', status: 'success' },
            { name: 'b', status: 'success' },
            { name: 'c', status: 'failed' },
            { name: 'd', status: 'skipped' },
            { name: 'e', status: 'success' },
          ],
        },
      },
    })
    const detailFiles = ref([
      { name: 'x', status: 'success' },
      { name: 'y', status: 'failed' },
      { name: 'z', status: 'success' },
    ])
    const finalFilesPage = ref(1)
    const finalFilesPageSize = ref(2)

    const api = useRunDetailComputed({ runDetail, detailFiles, finalFilesPage, finalFilesPageSize })
    await nextTick()

    expect(api.totalFinalFilesPages.value).toBe(3)
    expect(api.pagedFinalFiles.value.map(it => it.name)).toEqual(['a', 'b'])

    api.goNextFinalFilesPage()
    expect(finalFilesPage.value).toBe(2)
    expect(api.pagedFinalFiles.value.map(it => it.name)).toEqual(['c', 'd'])

    api.goNextFinalFilesPage()
    expect(finalFilesPage.value).toBe(3)
    expect(api.pagedFinalFiles.value.map(it => it.name)).toEqual(['e'])

    api.finalFilesJump.value = 1
    api.jumpFinalFilesPage()
    expect(finalFilesPage.value).toBe(1)
    expect(api.pagedFinalFiles.value.map(it => it.name)).toEqual(['a', 'b'])

    runDetail.value = {
      id: 102,
      taskMode: 'copy',
      summary: {
        finalSummary: {
          counts: { total: 2, copied: 2, deleted: 0, failed: 0, skipped: 0 },
          files: [
            { name: 'p', status: 'success' },
            { name: 'q', status: 'success' },
          ],
        },
      },
    }
    await nextTick()

    expect(finalFilesPage.value).toBe(1)
    expect(api.totalFinalFilesPages.value).toBe(1)
    expect(api.pagedFinalFiles.value.map(it => it.name)).toEqual(['p', 'q'])
  })
})

import { describe, it, expect, vi } from 'vitest'
import { ref } from 'vue'
import { useTaskTags } from './useTaskTags'
import * as tagsApi from '../api/tags'

vi.mock('../api/tags', () => ({
  fetchTags: vi.fn()
}))

describe('useTaskTags', () => {
  it('should separate action and keyword tags', async () => {
    vi.mocked(tagsApi.fetchTags).mockResolvedValue([
      { id: 1, tag: 'sync', type: 'action' },
      { id: 2, tag: 'copy', type: 'action' },
      { id: 3, tag: '备份', type: 'keyword' },
    ])

    const tasks = ref([])
    const { tags, actionTags, keywordTags, reload } = useTaskTags(tasks)

    await reload()

    expect(actionTags.value).toHaveLength(2)
    expect(keywordTags.value).toHaveLength(1)
    expect(actionTags.value[0].tag).toBe('sync')
    expect(keywordTags.value[0].tag).toBe('备份')
  })

  it('should return empty on fetch error', async () => {
    vi.mocked(tagsApi.fetchTags).mockRejectedValue(new Error('fail'))

    const tasks = ref([])
    const { tags, reload } = useTaskTags(tasks)

    await reload()

    expect(tags.value).toEqual([])
  })
})
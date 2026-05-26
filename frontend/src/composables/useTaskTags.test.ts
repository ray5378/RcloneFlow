import { describe, it, expect, vi, beforeEach } from 'vitest'
import { ref } from 'vue'
import { useTaskTags } from './useTaskTags'
import * as tagsApi from '../api/tags'

vi.mock('../api/tags', () => ({
  fetchTags: vi.fn(),
  selectTag: vi.fn(),
  createTag: vi.fn(),
  deleteTag: vi.fn(),
}))

describe('useTaskTags', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('should return all expected methods and properties', () => {
    const tasks = ref([])
    const result = useTaskTags(tasks)

    expect(result).toHaveProperty('tags')
    expect(result).toHaveProperty('actionTags')
    expect(result).toHaveProperty('keywordTags')
    expect(result).toHaveProperty('selectedKeywordTags')
    expect(result).toHaveProperty('suggestedTags')
    expect(result).toHaveProperty('loading')
    expect(result).toHaveProperty('reload')
    expect(result).toHaveProperty('toggleTag')
    expect(result).toHaveProperty('createManualTag')
    expect(result).toHaveProperty('deleteManualTag')
  })

  it('should separate action and keyword tags', async () => {
    vi.mocked(tagsApi.fetchTags).mockResolvedValue([
      { id: 1, tag: 'sync', type: 'action', selected: false },
      { id: 2, tag: 'copy', type: 'action', selected: false },
      { id: 3, tag: '备份', type: 'keyword', selected: true },
    ])

    const tasks = ref([])
    const { tags, actionTags, keywordTags, selectedKeywordTags, suggestedTags, reload } = useTaskTags(tasks)

    await reload()

    expect(actionTags.value).toHaveLength(2)
    expect(keywordTags.value).toHaveLength(1)
    expect(selectedKeywordTags.value).toHaveLength(1)
    expect(suggestedTags.value).toHaveLength(0)
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

  it('should toggleTag without error', async () => {
    const tasks = ref([])
    const { toggleTag } = useTaskTags(tasks)
    await expect(toggleTag('test-tag', true)).resolves.not.toThrow()
    expect(tagsApi.selectTag).toHaveBeenCalled()
  })

  it('should createManualTag without error', async () => {
    const tasks = ref([])
    const { createManualTag } = useTaskTags(tasks)
    await expect(createManualTag('new-tag')).resolves.not.toThrow()
    expect(tagsApi.createTag).toHaveBeenCalled()
  })

  it('should deleteManualTag without error', async () => {
    const tasks = ref([])
    const { deleteManualTag } = useTaskTags(tasks)
    await expect(deleteManualTag('old-tag')).resolves.not.toThrow()
    expect(tagsApi.deleteTag).toHaveBeenCalled()
  })
})

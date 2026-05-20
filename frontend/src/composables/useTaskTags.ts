import { ref, computed, type Ref } from 'vue'
import { fetchTags, selectTag, createTag, deleteTag, type Tag } from '../api/tags'
import type { Task } from '../types'

export function useTaskTags(_tasks: Ref<Task[]>) {
  const tags = ref<Tag[]>([])
  const loading = ref(false)

  const actionTags = computed(() => tags.value.filter(t => t.type === 'action'))
  const keywordTags = computed(() => tags.value.filter(t => t.type === 'keyword'))

  const selectedKeywordTags = computed(() => tags.value.filter(t => t.type === 'keyword' && t.selected))
  const suggestedTags = computed(() => tags.value.filter(t => t.type === 'keyword' && !t.selected))

  async function load() {
    loading.value = true
    try {
      tags.value = await fetchTags()
    } catch {
      tags.value = []
    } finally {
      loading.value = false
    }
  }

  async function toggleTag(tag: string, selected: boolean) {
    await selectTag(tag, selected)
    await load()
  }

  async function createManualTag(tag: string) {
    await createTag(tag)
    await load()
  }

  async function deleteManualTag(tag: string) {
    await deleteTag(tag)
    await load()
  }

  return {
    tags,
    actionTags,
    keywordTags,
    selectedKeywordTags,
    suggestedTags,
    loading,
    reload: load,
    toggleTag,
    createManualTag,
    deleteManualTag,
  }
}
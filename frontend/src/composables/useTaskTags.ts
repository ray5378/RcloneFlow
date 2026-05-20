import { ref, computed, type Ref } from 'vue'
import { fetchTags, type Tag } from '../api/tags'
import type { Task } from '../types'

export function useTaskTags(_tasks: Ref<Task[]>) {
  const tags = ref<Tag[]>([])
  const loading = ref(false)

  const actionTags = computed(() => tags.value.filter(t => t.type === 'action'))
  const keywordTags = computed(() => tags.value.filter(t => t.type === 'keyword'))

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

  return { tags, actionTags, keywordTags, loading, reload: load }
}
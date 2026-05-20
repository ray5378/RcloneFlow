<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { t } from '../../i18n'
import type { Tag } from '../../api/tags'

defineProps<{
  search: string
  sorting?: boolean
  savingSort?: boolean
  actionTags: Tag[]
  selectedKeywordTags: Tag[]
}>()

const emit = defineEmits<{
  (e: 'update:search', value: string): void
  (e: 'add'): void
  (e: 'toggle-sort'): void
  (e: 'save-sort'): void
  (e: 'cancel-sort'): void
  (e: 'open-tag-manager'): void
}>()

const showDropdown = ref(false)
const searchWrapperRef = ref<HTMLElement | null>(null)

function onSearchInput(event: Event) {
  const target = event.target as HTMLInputElement
  emit('update:search', target.value)
}

function selectTag(tag: string) {
  emit('update:search', tag)
  showDropdown.value = false
}

function clearSearch() {
  emit('update:search', '')
  showDropdown.value = false
}

function onFocus() {
  showDropdown.value = true
}

function onOutsideClick(event: MouseEvent) {
  const target = event.target as HTMLElement
  if (searchWrapperRef.value && !searchWrapperRef.value.contains(target)) {
    showDropdown.value = false
  }
}

onMounted(() => document.addEventListener('click', onOutsideClick))
onUnmounted(() => document.removeEventListener('click', onOutsideClick))
</script>

<template>
  <div class="card-header">
    <div class="title">{{ t('taskUI.taskList') }}</div>
    <div class="header-actions">
      <div class="search-wrapper" ref="searchWrapperRef">
        <input
          :value="search"
          type="text"
          :placeholder="t('taskUI.searchTask')"
          class="search-input"
          @input="onSearchInput"
          @focus="onFocus"
        />
        <button v-if="search" class="search-clear-btn" @click="clearSearch">&times;</button>
        <div v-if="showDropdown" class="search-dropdown">
          <div v-if="actionTags.length" class="dropdown-group">
            <div class="dropdown-label">{{ t('taskUI.tagsAction') }}</div>
            <div class="dropdown-tags">
              <button
                v-for="tag in actionTags"
                :key="tag.id"
                class="tag-pill tag-action"
                @click="selectTag(tag.tag)"
              >{{ t(`taskUI.${tag.tag}`) }}</button>
            </div>
          </div>
          <div v-if="selectedKeywordTags.length" class="dropdown-group">
            <div class="dropdown-label">{{ t('taskUI.tagsKeyword') }}</div>
            <div class="dropdown-tags">
              <button
                v-for="tag in selectedKeywordTags"
                :key="tag.id"
                class="tag-pill tag-keyword"
                @click="selectTag(tag.tag)"
              >{{ tag.tag }}</button>
            </div>
          </div>
          <div v-if="search" class="dropdown-group">
            <button class="dropdown-clear" @click="clearSearch">{{ t('taskUI.tagsClear') }}</button>
          </div>
        </div>
      </div>
      <button class="ghost small task-header-action-btn" @click="emit('open-tag-manager')">{{ t('taskUI.tagsManage') }}</button>
      <button v-if="!sorting" class="ghost small task-header-action-btn" @click="emit('toggle-sort')">{{ t('taskUI.taskSort') }}</button>
      <template v-else>
        <button class="primary small task-header-action-btn" :disabled="savingSort" @click="emit('save-sort')">{{ t('taskUI.saveSort') }}</button>
        <button class="ghost small task-header-action-btn" :disabled="savingSort" @click="emit('cancel-sort')">{{ t('taskUI.cancelSort') }}</button>
      </template>
      <button class="primary small task-header-action-btn add-task-btn" @click="emit('add')">{{ t('taskUI.addTask') }}</button>
    </div>
  </div>
</template>

<style scoped>
.header-actions > button.task-header-action-btn,
.header-actions > button.primary.task-header-action-btn,
.header-actions > button.ghost.task-header-action-btn,
.header-actions > button.primary.small.task-header-action-btn,
.header-actions > button.ghost.small.task-header-action-btn {
  inline-size: 88px !important;
  min-inline-size: 88px !important;
  max-inline-size: 88px !important;
  block-size: 32px !important;
  min-block-size: 32px !important;
  max-block-size: 32px !important;
  padding: 0 10px !important;
  line-height: 1 !important;
  white-space: nowrap;
  display: inline-flex !important;
  align-items: center;
  justify-content: center;
  flex: 0 0 88px !important;
  box-sizing: border-box;
}

.add-task-btn {
  white-space: nowrap;
}

.search-wrapper { position: relative; display: flex; align-items: center; }
.search-input { min-width: 200px; }
.search-clear-btn {
  position: absolute; right: 8px; background: none; border: none; color: var(--muted);
  font-size: 16px; cursor: pointer; padding: 0; line-height: 1;
}

.search-dropdown {
  position: absolute; top: 100%; left: 0; right: 0; margin-top: 4px;
  background: var(--card); border: 1px solid var(--border); border-radius: 8px;
  padding: 8px; z-index: 100; box-shadow: 0 4px 12px rgba(0,0,0,0.15);
  max-height: 50vh; overflow-y: auto;
}
body.light .search-dropdown { background: #fff; box-shadow: 0 4px 12px rgba(0,0,0,0.1); }

.dropdown-group { margin-bottom: 8px; }
.dropdown-group:last-child { margin-bottom: 0; }
.dropdown-label {
  font-size: 11px; color: var(--muted); text-transform: uppercase;
  margin-bottom: 6px; padding: 0 4px;
}
.dropdown-tags { display: flex; flex-wrap: wrap; gap: 4px; }

.tag-pill {
  padding: 4px 10px; border-radius: 12px; border: 1px solid var(--border);
  font-size: 12px; cursor: pointer; background: var(--surface); color: var(--text);
  transition: background 0.15s; white-space: nowrap;
}
.tag-pill:hover { background: var(--accent); color: #fff; border-color: var(--accent); }
.tag-action { border-color: var(--accent); color: var(--accent); }
.tag-keyword { border-color: var(--border); }
.dropdown-clear {
  display: block; width: 100%; text-align: center; background: none; border: none;
  color: var(--muted); font-size: 12px; cursor: pointer; padding: 4px;
}
.dropdown-clear:hover { color: var(--text); }

@media (max-width: 768px) {
  .header-actions {
    width: 100%;
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
  }

  .search-wrapper {
    flex: 1 1 100%;
  }

  .search-input {
    min-width: 0;
    flex: 1 1 100%;
    width: 100%;
  }

  .header-actions > button.task-header-action-btn,
  .header-actions > button.primary.task-header-action-btn,
  .header-actions > button.ghost.task-header-action-btn,
  .header-actions > button.primary.small.task-header-action-btn,
  .header-actions > button.ghost.small.task-header-action-btn {
    inline-size: 88px !important;
    min-inline-size: 88px !important;
    max-inline-size: 88px !important;
    block-size: 32px !important;
    min-block-size: 32px !important;
    max-block-size: 32px !important;
    padding: 0 8px !important;
    font-size: 12px;
    flex: 0 0 88px !important;
  }
}
</style>
<script setup lang="ts">
import { t } from '../../i18n'

const props = defineProps<{
  search: string
  sortingMode?: boolean
  sortDirty?: boolean
}>()
const emit = defineEmits<{
  (e: 'update:search', value: string): void
  (e: 'add'): void
  (e: 'enter-sort-mode'): void
  (e: 'cancel-sort'): void
  (e: 'save-sort'): void
}>()

function onSearchInput(event: Event) {
  const target = event.target as HTMLInputElement
  emit('update:search', target.value)
}
</script>

<template>
  <div class="card-header">
    <div class="title">{{ t('taskUI.taskList') }}</div>
    <div class="header-actions">
      <input :value="search" type="text" :placeholder="t('taskUI.searchTask')" class="search-input" :disabled="sortingMode" @input="onSearchInput" />
      <template v-if="sortingMode">
        <button class="ghost small" @click="emit('cancel-sort')">{{ t('taskUI.cancelSort') }}</button>
        <button class="primary small" :disabled="!sortDirty" @click="emit('save-sort')">{{ t('taskUI.saveSort') }}</button>
      </template>
      <template v-else>
        <button class="ghost small" @click="emit('enter-sort-mode')">↕ {{ t('taskUI.enterSortMode') }}</button>
        <button class="primary small" @click="emit('add')">+ {{ t('taskUI.addTask') }}</button>
      </template>
    </div>
  </div>
</template>

<script setup lang="ts">
import { t } from '../../i18n'

defineProps<{
  search: string
  sorting: boolean
  saveDisabled?: boolean
}>()
const emit = defineEmits<{
  (e: 'update:search', value: string): void
  (e: 'add'): void
  (e: 'toggle-sort'): void
  (e: 'save-sort'): void
  (e: 'cancel-sort'): void
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
      <input :value="search" type="text" :placeholder="t('taskUI.searchTask')" class="search-input" @input="onSearchInput" />
      <template v-if="!sorting">
        <button class="ghost small" @click="emit('toggle-sort')">{{ t('taskUI.taskSort') }}</button>
        <button class="primary small" @click="emit('add')">+ {{ t('taskUI.addTask') }}</button>
      </template>
      <template v-else>
        <button class="ghost small" @click="emit('cancel-sort')">{{ t('taskUI.cancelSort') }}</button>
        <button class="primary small" :disabled="saveDisabled" @click="emit('save-sort')">{{ t('taskUI.saveSort') }}</button>
      </template>
    </div>
  </div>
</template>

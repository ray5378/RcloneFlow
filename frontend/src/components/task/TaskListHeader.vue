<script setup lang="ts">
import { t } from '../../i18n'

defineProps<{ search: string }>()
const emit = defineEmits<{ (e: 'update:search', value: string): void; (e: 'add'): void }>()

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
      <button class="primary small" @click="emit('add')">+ {{ t('taskUI.addTask') }}</button>
    </div>
  </div>
</template>

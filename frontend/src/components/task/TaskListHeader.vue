<script setup lang="ts">
import { t } from '../../i18n'

defineProps<{ search: string; sorting?: boolean; savingSort?: boolean }>()
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

@media (max-width: 768px) {
  .header-actions {
    width: 100%;
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
  }

  .search-input {
    min-width: 0;
    flex: 1 1 100%;
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

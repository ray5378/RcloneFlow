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
      <button v-if="!sorting" class="ghost small action-btn" @click="emit('toggle-sort')">{{ t('taskUI.taskSort') }}</button>
      <template v-else>
        <button class="primary small action-btn" :disabled="savingSort" @click="emit('save-sort')">{{ t('taskUI.saveSort') }}</button>
        <button class="ghost small action-btn" :disabled="savingSort" @click="emit('cancel-sort')">{{ t('taskUI.cancelSort') }}</button>
      </template>
      <button class="primary small action-btn add-task-btn" @click="emit('add')">{{ t('taskUI.addTask') }}</button>
    </div>
  </div>
</template>

<style scoped>
.action-btn {
  min-width: 88px;
  height: 32px;
  padding: 6px 12px;
  white-space: nowrap;
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

  .action-btn {
    font-size: 12px;
    padding: 8px 10px;
    min-width: 88px;
    flex-shrink: 1;
  }
}
</style>

<script setup lang="ts">
import { ref } from 'vue'
import { t } from '../../i18n'
import type { Tag } from '../../api/tags'

const props = defineProps<{
  visible: boolean
  suggestedTags: Tag[]
  selectedKeywordTags: Tag[]
  actionTags: Tag[]
}>()

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'create-tag', tag: string): void
  (e: 'select-tag', tag: string): void
  (e: 'unselect-tag', tag: string): void
  (e: 'delete-tag', tag: string): void
}>()

const inputValue = ref('')

function onInputKeydown(event: KeyboardEvent) {
  if (event.key !== 'Enter') return
  const value = inputValue.value.trim()
  if (!value) return
  emit('create-tag', value)
  inputValue.value = ''
}

function onClickSuggested(tag: Tag) {
  emit('select-tag', tag.tag)
}

function onClickSelected(tag: Tag) {
  emit('unselect-tag', tag.tag)
}

function onClickDelete(tag: Tag) {
  emit('delete-tag', tag.tag)
}
</script>

<template>
  <div v-if="visible" class="modal-overlay" @click.self="emit('close')">
    <div class="modal-content" style="max-width:520px">
      <div class="modal-header">
        <h3>{{ t('taskUI.tagsManageTitle') }}</h3>
        <button class="close-btn" @click="emit('close')">×</button>
      </div>
      <div class="modal-body">
        <div class="tag-input-row">
          <input
            v-model="inputValue"
            type="text"
            :placeholder="t('taskUI.tagsInputPlaceholder')"
            class="tag-input"
            @keydown="onInputKeydown"
          />
        </div>

        <div v-if="actionTags.length + selectedKeywordTags.length > 0" class="tag-section">
          <div class="tag-section-label">{{ t('taskUI.tagsSelected') }}</div>
          <div class="tag-pills">
            <span
              v-for="tag in actionTags"
              :key="tag.id"
              class="tag-pill tag-pill-action"
            >{{ t('taskUI.' + tag.tag) }}</span>
            <span
              v-for="tag in selectedKeywordTags"
              :key="tag.id"
              class="tag-pill tag-pill-selected"
              @click="onClickSelected(tag)"
              :title="t('taskUI.tagsClear')"
            >
              {{ tag.tag }}
              <button class="tag-remove-btn" @click.stop="onClickDelete(tag)">×</button>
            </span>
          </div>
        </div>

        <div v-if="suggestedTags.length > 0" class="tag-section">
          <div class="tag-section-label">{{ t('taskUI.tagsSuggested') }}</div>
          <div class="tag-pills">
            <button
              v-for="tag in suggestedTags"
              :key="tag.id"
              class="tag-pill tag-pill-suggested"
              @click="onClickSuggested(tag)"
            >{{ tag.tag }}</button>
          </div>
        </div>
      </div>
      <div class="modal-footer">
        <button class="primary" @click="emit('close')">{{ t('common.save') }}</button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.7);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  padding: 20px;
}

.modal-content {
  background: var(--surface);
  border-radius: 12px;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.5);
  width: 100%;
  max-height: 90vh;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 20px;
  border-bottom: 1px solid var(--border);
  flex-shrink: 0;
}

.modal-header h3 {
  margin: 0;
  font-size: 18px;
  color: var(--text);
}

.close-btn {
  background: none;
  border: none;
  font-size: 24px;
  color: var(--muted);
  cursor: pointer;
  padding: 0;
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 6px;
  transition: background 0.2s;
}

.close-btn:hover {
  background: var(--hover);
}

.modal-body {
  padding: 20px;
  overflow-y: auto;
  flex: 1;
}

.modal-footer {
  padding: 16px 20px;
  border-top: 1px solid var(--border);
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  flex-shrink: 0;
}

.tag-input-row {
  margin-bottom: 16px;
}

.tag-input {
  width: 100%;
  padding: 10px 14px;
  border-radius: 8px;
  border: 1px solid #333;
  background: #252525;
  color: #e0e0e0;
  font-size: 14px;
  box-sizing: border-box;
}

body.light .tag-input {
  background: #fff;
  border-color: #ddd;
  color: #333;
}

.tag-section {
  margin-bottom: 16px;
}

.tag-section:last-child {
  margin-bottom: 0;
}

.tag-section-label {
  font-size: 12px;
  color: var(--muted);
  margin-bottom: 8px;
  font-weight: 600;
}

.tag-pills {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.tag-pill {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 5px 12px;
  border-radius: 14px;
  font-size: 13px;
  border: 1px solid var(--border);
  background: var(--surface);
  color: var(--text);
  white-space: nowrap;
  transition: background 0.15s, border-color 0.15s;
}

.tag-pill-action {
  border-color: var(--accent);
  color: var(--accent);
  background: transparent;
  cursor: default;
}

.tag-pill-selected {
  border-color: var(--accent);
  background: rgba(79, 70, 241, 0.12);
  cursor: pointer;
}

.tag-pill-selected:hover {
  background: rgba(239, 68, 68, 0.15);
  border-color: #ef4444;
}

.tag-pill-suggested {
  cursor: pointer;
  border-color: var(--border);
  background: var(--surface);
}

.tag-pill-suggested:hover {
  background: var(--accent);
  color: #fff;
  border-color: var(--accent);
}

.tag-remove-btn {
  background: none;
  border: none;
  color: var(--muted);
  font-size: 14px;
  cursor: pointer;
  padding: 0 2px;
  line-height: 1;
}

.tag-remove-btn:hover {
  color: #ef4444;
}

@media (max-width: 640px) {
  .modal-overlay {
    padding: 10px;
  }

  .modal-content {
    max-height: 85vh;
    border-radius: 16px 16px 16px 16px;
  }

  .modal-header {
    padding: 14px 16px;
  }

  .modal-header h3 {
    font-size: 16px;
  }

  .modal-body {
    padding: 16px;
  }

  .modal-footer {
    padding: 14px 16px;
  }
}
</style>
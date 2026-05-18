<script setup lang="ts">
import { t } from '../../i18n'

defineProps<{
  visible: boolean
  singletonEnabled: boolean
}>()

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'save'): void
  (e: 'update:singletonEnabled', value: boolean): void
}>()

function onToggle(event: Event) {
  const target = event.target as HTMLInputElement
  emit('update:singletonEnabled', target.checked)
}
</script>

<template>
  <div v-if="visible" class="modal-overlay" @click.self="emit('close')">
    <div class="modal-content" style="max-width:560px">
      <div class="modal-header">
        <h3>{{ t('singleton.title') }}</h3>
        <button class="close-btn" @click="emit('close')">×</button>
      </div>
      <div class="modal-body">
        <div class="detail-item full-width">
          <label class="trigger-opt">
            <input :checked="singletonEnabled" type="checkbox" @change="onToggle" />
            <span>{{ t('singleton.enable') }}</span>
          </label>
          <p class="hint">{{ t('singleton.hint') }}</p>
        </div>
      </div>
      <div class="modal-footer">
        <button class="primary" @click="emit('save')">{{ t('common.save') }}</button>
        <button class="ghost" @click="emit('close')">{{ t('common.cancel') }}</button>
      </div>
    </div>
  </div>
</template>

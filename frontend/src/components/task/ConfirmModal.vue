<script setup lang="ts">
import { t } from '../../i18n'

defineProps<{
  visible: boolean
  title: string
  message: string
}>()

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'confirm'): void
}>()
</script>

<template>
  <div v-if="visible" class="modal-overlay" @click.self="emit('close')">
    <div class="modal-content confirm-modal">
      <div class="modal-header">
        <h3>{{ title }}</h3>
        <button class="close-btn" @click="emit('close')">×</button>
      </div>
      <div class="modal-body">
        <p>{{ message }}</p>
      </div>
      <div class="modal-footer">
        <button class="ghost" @click="emit('close')">{{ t('modal.cancel') }}</button>
        <button class="primary danger" @click="emit('confirm')">{{ t('modal.confirm') }}</button>
      </div>
    </div>
  </div>
</template>

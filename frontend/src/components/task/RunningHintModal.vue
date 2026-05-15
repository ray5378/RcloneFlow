<script setup lang="ts">
import { t } from '../../i18n'

defineProps<{
  visible: boolean
  run: any | null
  phaseText: string
  progressText: string
}>()

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'open-log', run: any): void
}>()
</script>

<template>
  <div v-if="visible" class="modal-overlay" @click.self="emit('close')">
    <div class="modal-content" style="max-width:520px">
      <div class="modal-header">
        <h3>{{ t('modal.runningTitle') }}</h3>
        <button class="close-btn" @click="emit('close')">×</button>
      </div>
      <div class="modal-body">
        <p>{{ t('modal.runningHint1') }}</p>
        <p>{{ t('modal.runningHint2') }}</p>
        <div class="hint-box">
          <div class="detail-item"><label>{{ t('modal.task') }}</label><span>{{ run?.taskName || `#${run?.taskId}` }}</span></div>
          <div class="detail-item"><label>{{ t('modal.phase') }}</label><span>{{ phaseText || '-' }}</span></div>
          <div class="detail-item"><label>{{ t('modal.live') }}</label><span>{{ progressText || '-' }}</span></div>
        </div>
      </div>
      <div class="modal-footer">
        <button class="primary" @click="emit('open-log', run)">{{ t('modal.openTransferLog') }}</button>
        <button class="ghost" @click="emit('close')">{{ t('modal.gotIt') }}</button>
      </div>
    </div>
  </div>
</template>

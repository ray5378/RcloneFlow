<script setup lang="ts">
import { t } from '../../../i18n'
import type { ActiveTransferCompletedFile, ActiveTransferCurrentFile, ActiveTransferPendingFile, ActiveTransferSummary, TrackingMode } from '../../../api/activeTransfer'
import TransferSummaryBar from './TransferSummaryBar.vue'
import TransferCurrentFileCard from './TransferCurrentFileCard.vue'
import TransferCompletedList from './TransferCompletedList.vue'
import TransferPendingList from './TransferPendingList.vue'

const props = defineProps<{
  visible: boolean
  trackingMode: TrackingMode
  summary: ActiveTransferSummary | null
  currentFile: ActiveTransferCurrentFile | null | undefined
  completedItems: ActiveTransferCompletedFile[]
  pendingItems: ActiveTransferPendingFile[]
  completedTotal?: number
  pendingTotal?: number
  degraded?: boolean
  loading?: boolean
  error?: string
}>()

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'refresh'): void
  (e: 'load-more-completed'): void
  (e: 'load-more-pending'): void
}>()
</script>

<template>
  <div v-if="visible" class="modal-overlay" @click.self="emit('close')">
    <div class="modal-content" style="max-width:1280px">
      <div class="modal-header">
        <h3>{{ t('activeTransfer.title') }}</h3>
        <div class="actions">
          <button class="ghost small" @click="emit('refresh')">{{ t('activeTransfer.refresh') }}</button>
          <button class="close-btn" @click="emit('close')">×</button>
        </div>
      </div>
      <div class="modal-body">
        <div v-if="degraded" class="degraded">{{ t('activeTransfer.degraded') }}</div>
        <div v-if="loading" class="state-box">{{ t('common.loading') }}</div>
        <div v-else-if="error" class="state-box error">{{ error }}</div>
        <div v-else-if="!summary && !currentFile && !completedItems.length && !pendingItems.length" class="state-box">{{ t('activeTransfer.empty') }}</div>
        <template v-else>
          <TransferSummaryBar :summary="summary" />
          <TransferCurrentFileCard :current-file="currentFile" :tracking-mode="trackingMode" />
          <div class="two-col">
            <TransferCompletedList :items="completedItems" :total="completedTotal" :tracking-mode="trackingMode" @load-more="emit('load-more-completed')" />
            <TransferPendingList :items="pendingItems" :total="pendingTotal" :tracking-mode="trackingMode" @load-more="emit('load-more-pending')" />
          </div>
        </template>
      </div>
    </div>
  </div>
</template>

<style scoped>
.modal-overlay { position:fixed; inset:0; background:rgba(0,0,0,.45); display:flex; align-items:center; justify-content:center; z-index:1000; }
.modal-content { width:min(96vw, 1280px); max-height:88vh; overflow:auto; background:var(--bg, #111); border-radius:12px; padding:16px; }
body.light .modal-content { background:#fff; }
.modal-header { display:flex; justify-content:space-between; align-items:center; margin-bottom:12px; }
.actions { display:flex; gap:8px; align-items:center; }
.close-btn { background:none; border:none; font-size:20px; cursor:pointer; }
.two-col { display:grid; grid-template-columns:minmax(0, 1fr) minmax(0, 1fr); gap:18px; width:100%; align-items:start; }
.degraded { margin-bottom:10px; padding:8px 10px; border-radius:8px; background:#f59e0b22; color:#f59e0b; font-size:12px; }
.state-box { padding:16px; border:1px dashed #444; border-radius:8px; color:#999; text-align:center; }
.state-box.error { color:#ef4444; border-color:#ef444466; }
@media (max-width: 768px) { .two-col { grid-template-columns:1fr; } }
</style>

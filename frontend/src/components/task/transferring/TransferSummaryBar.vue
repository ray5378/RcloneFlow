<script setup lang="ts">
import { t } from '../../../i18n'
import { formatBytes, formatBytesPerSec, formatEta } from '../../../utils/format'
import type { ActiveTransferSummary } from '../../../api/activeTransfer'

defineProps<{
  summary: ActiveTransferSummary | null
}>()
</script>

<template>
  <div class="summary-box" v-if="summary">
    <div class="summary-main">
      <strong>{{ (summary.percentage || 0).toFixed(2) }}%</strong>
      <span>{{ summary.completedCount }}/{{ summary.totalCount }}</span>
    </div>
    <div class="summary-meta">
      <span>{{ formatBytes(summary.bytes || 0) }} / {{ formatBytes(summary.totalBytes || 0) }}</span>
      <span>{{ t('run.speed') }}: {{ formatBytesPerSec(summary.speed || 0) }}</span>
      <span v-if="summary.eta != null">ETA: {{ formatEta(summary.eta || 0) }}</span>
    </div>
  </div>
</template>

<style scoped>
.summary-box { padding: 10px 12px; background: #1a1a1a; border-radius: 8px; margin-bottom: 12px; }
body.light .summary-box { background: #f5f5f5; }
.summary-main { display:flex; justify-content:space-between; align-items:center; margin-bottom:6px; }
.summary-meta { display:flex; gap:12px; flex-wrap:wrap; font-size:12px; color:#999; }
</style>

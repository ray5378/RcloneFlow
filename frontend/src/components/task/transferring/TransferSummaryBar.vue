<script setup lang="ts">
import { computed } from 'vue'
import type { ActiveTransferSummary } from '../../../api/activeTransfer'
import { getUnifiedProgressText } from '../progressText'

const props = defineProps<{
  summary: ActiveTransferSummary | null
}>()

const progressText = computed(() => {
  if (!props.summary) return '-'
  return getUnifiedProgressText({
    percentage: props.summary.percentage,
    bytes: props.summary.bytes,
    totalBytes: props.summary.totalBytes,
    speed: props.summary.speed,
    eta: props.summary.eta,
    completedFiles: props.summary.completedCount,
    totalCount: props.summary.totalCount,
  })
})
</script>

<template>
  <div class="summary-box" v-if="summary">
    <div class="summary-main">
      <strong>{{ (summary.percentage || 0).toFixed(2) }}%</strong>
      <span>{{ summary.completedCount }}/{{ summary.totalCount }}</span>
    </div>
    <div class="summary-meta">{{ progressText }}</div>
  </div>
</template>

<style scoped>
.summary-box { padding: 10px 12px; background: #1a1a1a; border-radius: 8px; margin-bottom: 12px; }
body.light .summary-box { background: #f5f5f5; }
.summary-main { display:flex; justify-content:space-between; align-items:center; margin-bottom:6px; }
.summary-meta { font-size:12px; color:#999; line-height:1.5; word-break:break-word; }
</style>

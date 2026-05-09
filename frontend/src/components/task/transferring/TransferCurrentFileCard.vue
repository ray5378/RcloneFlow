<script setup lang="ts">
import { t } from '../../../i18n'
import { formatBytes, formatBytesPerSec } from '../../../utils/format'
import type { ActiveTransferCurrentFile, TrackingMode } from '../../../api/activeTransfer'
import { getTrackingLabels } from './transferringLabels'

const props = defineProps<{
  currentFile: ActiveTransferCurrentFile | null | undefined
  trackingMode: TrackingMode
}>()
</script>

<template>
  <div class="current-box">
    <div class="title">{{ getTrackingLabels(props.trackingMode).current }}</div>
    <template v-if="currentFile">
      <div class="name">{{ currentFile.path || currentFile.name }}</div>
      <div class="meta">
        <span>{{ currentFile.percentage != null ? `${currentFile.percentage.toFixed(2)}%` : '--' }}</span>
        <span>{{ formatBytes(currentFile.bytes || 0) }} / {{ formatBytes(currentFile.totalBytes || 0) }}</span>
        <span>{{ formatBytesPerSec(currentFile.speed || 0) }}</span>
      </div>
    </template>
    <div v-else class="empty">{{ t('activeTransfer.noCurrentFile') }}</div>
  </div>
</template>

<style scoped>
.current-box { padding: 10px 12px; border: 1px solid #333; border-radius: 8px; margin-bottom: 12px; }
body.light .current-box { border-color: #ddd; }
.title { font-size: 12px; color:#999; margin-bottom: 6px; }
.name { word-break: break-all; margin-bottom: 6px; }
.meta { display:flex; gap:12px; flex-wrap:wrap; font-size:12px; color:#999; }
.empty { font-size:12px; color:#999; }
</style>

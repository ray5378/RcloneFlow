<script setup lang="ts">
import { computed } from 'vue'
import { t } from '../../../i18n'
import { formatBytes, formatBytesPerSec } from '../../../utils/format'
import type { ActiveTransferCurrentFile, TrackingMode } from '../../../api/activeTransfer'
import { getTrackingLabels } from './transferringLabels'

const props = defineProps<{
  currentFile: ActiveTransferCurrentFile | null | undefined
  currentFiles?: ActiveTransferCurrentFile[]
  transferSlots?: number
  trackingMode: TrackingMode
}>()

const files = computed(() => {
  const list = (props.currentFiles || []).filter(Boolean)
  if (list.length) return list
  return props.currentFile ? [props.currentFile] : []
})

const slotCount = computed(() => Math.max(1, Number(props.transferSlots || files.value.length || 1)))
const placeholderCount = computed(() => Math.max(0, slotCount.value - files.value.length))
const placeholders = computed(() => Array.from({ length: placeholderCount.value }, (_, index) => index))
</script>

<template>
  <div class="current-box">
    <div class="title">{{ getTrackingLabels(props.trackingMode).current }} <span>({{ files.length }}/{{ slotCount }})</span></div>
    <div class="file-grid">
      <div v-for="file in files" :key="file.path || file.name" class="file-card">
        <div class="name">{{ file.path || file.name }}</div>
        <div class="meta">
          <span>{{ file.percentage != null ? `${file.percentage.toFixed(2)}%` : '--' }}</span>
          <span>{{ formatBytes(file.bytes || 0) }} / {{ formatBytes(file.totalBytes || 0) }}</span>
          <span>{{ formatBytesPerSec(file.speed || 0) }}</span>
        </div>
      </div>
      <div v-for="slot in placeholders" :key="`current-placeholder-${slot}`" class="file-card placeholder-card">
        <div class="name placeholder-name">{{ t('activeTransfer.waitingTransfer') }}</div>
        <div class="meta placeholder-meta">
          <span>--</span>
          <span>— / —</span>
          <span>—</span>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.current-box { padding: 10px 12px; border: 1px solid #333; border-radius: 8px; margin-bottom: 12px; }
body.light .current-box { border-color: #ddd; }
.title { font-size: 12px; color:#999; margin-bottom: 8px; }
.file-grid { display:grid; grid-template-columns:repeat(auto-fit, minmax(300px, 1fr)); gap:10px; align-items:start; }
.file-card { padding:10px 12px; border:1px dashed #2f2f2f; border-radius:8px; min-width:0; min-height: 75px; box-sizing: border-box; }
body.light .file-card { border-color:#e5e5e5; }
.name { word-break: break-all; margin-bottom: 6px; }
.meta { display:flex; gap:12px; flex-wrap:wrap; font-size:12px; color:#999; }
.placeholder-card { color:#777; opacity:.55; }
.placeholder-name { font-style:italic; }
.placeholder-meta { color:#777; }
@media (max-width: 768px) {
  .file-grid { grid-template-columns:1fr; }
}
</style>

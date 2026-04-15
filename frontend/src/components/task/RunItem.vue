<script setup lang="ts">
import type { RunRecord } from './types'
import { formatBytes, formatDuration } from '../../utils/format'

const props = defineProps<{
  run: RunRecord
}>()

const emit = defineEmits<{
  click: [run: RunRecord]
}>()

function getStatusClass(status: string) {
  switch (status) {
    case 'finished': return 'success'
    case 'failed': return 'danger'
    case 'skipped': return 'warning'
    case 'running': return 'info'
    default: return ''
  }
}

function formatTime(dateStr: string) {
  if (!dateStr) return ''
  const d = new Date(dateStr)
  return d.toLocaleString('zh-CN', { month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit' })
}
</script>

<template>
  <div class="item run-item" @click="emit('click', run)">
    <div class="run-left">
      <span class="chip" :class="getStatusClass(run.status)">{{ run.status }}</span>
      <span class="run-name">{{ run.taskName || `任务 #${run.taskId}` }}</span>
    </div>
    <div class="run-time">{{ formatTime(run.createdAt) }}</div>
    <div class="run-size">{{ formatBytes(run.bytesTransferred || 0) }}</div>
    <div class="run-speed">{{ run.speed || '-' }}</div>
    <div class="run-trigger chip mini">{{ run.trigger }}</div>
  </div>
</template>

<style scoped>
.run-item {
  display: grid;
  grid-template-columns: minmax(150px, 1fr) 100px 80px 80px auto;
  gap: 12px;
  align-items: center;
  padding: 12px 16px;
  cursor: pointer;
  border-bottom: 1px solid #333;
}
.run-item:hover {
  background: rgba(255,255,255,0.03);
}
.run-left {
  display: flex;
  align-items: center;
  gap: 8px;
  min-width: 0;
}
.run-name {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.run-time {
  color: #888;
  font-size: 12px;
}
.run-size, .run-speed {
  text-align: right;
  font-size: 12px;
  color: #aaa;
}
.chip.mini {
  padding: 2px 6px;
  font-size: 10px;
}
body.light .run-item {
  border-bottom-color: #eee;
}
body.light .run-item:hover {
  background: rgba(0,0,0,0.03);
}
</style>

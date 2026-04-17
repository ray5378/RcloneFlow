<script setup lang="ts">
import { formatBytes } from '../../utils/format'

interface FileItem {
  path?: string
  name?: string
  status?: string
  at?: string
  sizeBytes?: number
}

defineProps<{
  item: FileItem
}>()
</script>

<template>
  <div class="files-row">
    <span class="name" :title="item.path || item.name">
      {{ ((item.path || item.name || '').replace(/\\/g,'/').split('/').pop()) }}
    </span>
    <span class="status" :class="item.status">{{ item.status }}</span>
    <span class="time">{{ item.at || '-' }}</span>
    <span class="size">{{ item.sizeBytes ? formatBytes(item.sizeBytes) : '-' }}</span>
  </div>
</template>

<style scoped>
.files-row {
  display: grid;
  grid-template-columns: 1fr 80px 100px 80px;
  gap: 8px;
  padding: 8px 0;
  border-bottom: 1px solid #333;
  font-size: 12px;
}
.name {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: #ccc;
}
.status {
  text-align: center;
  padding: 2px 6px;
  border-radius: 4px;
  font-size: 10px;
}
.status.New { background: #22c55e33; color: #22c55e; }
.status.Copied { background: #3b82f633; color: #3b82f6; }
.status.Deleted { background: #ef444433; color: #ef4444; }
.status.Failed { background: #dc262633; color: #dc2626; }
.status.Skipped { background: #f59e0b33; color: #f59e0b; }
.time {
  color: #888;
  text-align: center;
}
.size {
  color: #888;
  text-align: right;
}
</style>

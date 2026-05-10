<script setup lang="ts">
import { computed } from 'vue'
import { locale } from '../../i18n'
import { formatBytes } from '../../utils/format'

interface FileItem {
  path?: string
  name?: string
  status?: string
  at?: string
  sizeBytes?: number
}

const props = defineProps<{
  item: FileItem
}>()

const formattedAt = computed(() => {
  const value = props.item.at
  if (!value) return '-'
  try {
    return new Date(value).toLocaleString(locale.value === 'zh' ? 'zh-CN' : 'en-US', {
      year: 'numeric', month: '2-digit', day: '2-digit',
      hour: '2-digit', minute: '2-digit', second: '2-digit'
    })
  } catch {
    return value || '-'
  }
})

const formattedSize = computed(() => {
  const value = props.item.sizeBytes
  if (value === undefined || value === null || Number.isNaN(Number(value))) return '-'
  return formatBytes(Number(value))
})
</script>

<template>
  <div class="files-row">
    <span class="name" :title="item.path || item.name">
      {{ ((item.path || item.name || '').replace(/\\/g,'/').split('/').pop()) }}
    </span>
    <span class="status" :class="item.status">{{ item.status }}</span>
    <span class="time">{{ formattedAt }}</span>
    <span class="size">{{ formattedSize }}</span>
  </div>
</template>

<style scoped>
.files-row {
  display: grid;
  grid-template-columns: 1fr 140px 200px 140px;
  gap: 18px;
  padding: 10px 16px;
  border-bottom: 1px solid #333;
  font-size: 12px;
  align-items: center;
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

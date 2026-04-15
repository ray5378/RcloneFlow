<script setup lang="ts">
interface PathItem {
  Name?: string
  Path?: string
  IsDir?: boolean
}

defineProps<{
  item: PathItem
}>()

const emit = defineEmits<{
  click: [item: PathItem]
  arrowClick: [item: PathItem]
}>()
</script>

<template>
  <div class="path-item" :class="{ 'is-dir': item.IsDir }" @click="emit('click', item)">
    <span v-if="item.IsDir" class="folder-icon" @click.stop="emit('arrowClick', item)">📁</span>
    <span v-else class="file-icon">📄</span>
    <span class="item-name">{{ item.Name }}</span>
  </div>
</template>

<style scoped>
.path-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  cursor: pointer;
  border-radius: 6px;
  transition: background 0.2s;
}
.path-item:hover {
  background: rgba(255,255,255,0.05);
}
.path-item.is-dir {
  color: #fbbf24;
}
.path-item:not(.is-dir) {
  color: #9ca3af;
}
.folder-icon, .file-icon {
  font-size: 16px;
  cursor: pointer;
}
.item-name {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
</style>

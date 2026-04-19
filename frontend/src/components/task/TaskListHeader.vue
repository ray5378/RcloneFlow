<script setup lang="ts">
defineProps<{
  search: string
}>()

const emit = defineEmits<{
  (e: 'update:search', value: string): void
  (e: 'add'): void
}>()

function onSearchInput(event: Event) {
  const target = event.target as HTMLInputElement
  emit('update:search', target.value)
}
</script>

<template>
  <div class="card-header">
    <div class="title">任务列表</div>
    <div class="header-actions">
      <input :value="search" type="text" placeholder="搜索任务..." class="search-input" @input="onSearchInput" />
      <button class="primary small" @click="emit('add')">+ 添加任务</button>
    </div>
  </div>
</template>

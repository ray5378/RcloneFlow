<script setup lang="ts">
defineProps<{
  page: number
  totalPages: number
  jumpPage: number | null
}>()

const emit = defineEmits<{
  (e: 'prev'): void
  (e: 'next'): void
  (e: 'jump'): void
  (e: 'update:jumpPage', value: number | null): void
}>()

function onJumpInput(event: Event) {
  const target = event.target as HTMLInputElement
  emit('update:jumpPage', target.value === '' ? null : Number(target.value))
}
</script>

<template>
  <div class="pagination">
    <span class="page-current">第 {{ page }} / {{ totalPages }} 页</span>
    <button class="page-btn" :disabled="page <= 1" @click="emit('prev')">上一页</button>
    <button class="page-btn" :disabled="page >= totalPages" @click="emit('next')">下一页</button>
    <input type="number" class="page-input" :value="jumpPage ?? ''" min="1" :max="totalPages" @input="onJumpInput" @keyup.enter="emit('jump')" />
    <button class="page-btn" @click="emit('jump')">跳转</button>
  </div>
</template>

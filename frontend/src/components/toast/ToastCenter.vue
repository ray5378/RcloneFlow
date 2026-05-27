<script setup lang="ts">
import ToastItem from './ToastItem.vue'

defineProps<{
  toasts: Array<{ id: string | number; [key: string]: any }>
}>()
</script>

<template>
  <div class="toast-container">
    <ToastItem v-for="toast in toasts" :key="toast.id" :toast="toast" />
  </div>
</template>

<style scoped>
.toast-container {
  position: fixed;
  top: 12px;
  right: 12px;
  z-index: 9999;
  display: flex;
  flex-direction: column;
  gap: 8px;
  pointer-events: none;
}
.toast-container > * {
  pointer-events: auto;
}
</style>

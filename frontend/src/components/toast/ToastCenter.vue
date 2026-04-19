<script setup lang="ts">
import { ToastItem } from './'

defineProps<{
  toasts: Array<{ id: string | number; [key: string]: any }>
}>()
</script>

<template>
  <div class="toast-container">
    <ToastItem v-for="toast in toasts" :key="toast.id" :toast="toast" />
  </div>
</template>

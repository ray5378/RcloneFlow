<script setup lang="ts">
interface Toast {
  id: number
  message: string
  type: 'info' | 'success' | 'error'
}

defineProps<{
  toast: Toast
}>()
</script>

<template>
  <div :class="['toast', toast.type]">
    {{ toast.message }}
  </div>
</template>

<style scoped>
.toast {
  padding: 10px 16px;
  border-radius: 8px;
  font-size: 14px;
  animation: slideIn 0.3s ease;
  box-shadow: 0 4px 12px rgba(0,0,0,0.3);
}
.toast.info {
  background: #4f46e5;
  color: white;
}
.toast.success {
  background: #22c55e;
  color: white;
}
.toast.error {
  background: #ef4444;
  color: white;
}
@keyframes slideIn {
  from {
    transform: translateX(100%);
    opacity: 0;
  }
  to {
    transform: translateX(0);
    opacity: 1;
  }
}
</style>

<script setup lang="ts">
import { ref, onMounted } from 'vue'

const toasts = ref<Toast[]>([])
let toastId = 0

interface Toast {
  id: number
  message: string
  type: 'success' | 'error' | 'warning' | 'info'
}

function showToast(message: string, type: Toast['type'] = 'info', duration = 3000) {
  const id = ++toastId
  toasts.value.push({ id, message, type })
  
  setTimeout(() => {
    toasts.value = toasts.value.filter(t => t.id !== id)
  }, duration)
}

function removeToast(id: number) {
  toasts.value = toasts.value.filter(t => t.id !== id)
}

// 暴露给外部使用
defineExpose({ showToast })

// 全局事件处理
onMounted(() => {
  window.addEventListener('show-toast', ((e: CustomEvent) => {
    showToast(e.detail.message, e.detail.type, e.detail.duration)
  }) as EventListener)
})
</script>

<template>
  <div class="toast-container">
    <TransitionGroup name="toast">
      <div
        v-for="toast in toasts"
        :key="toast.id"
        :class="['toast', `toast-${toast.type}`]"
        @click="removeToast(toast.id)"
      >
        <span class="toast-icon">
          <template v-if="toast.type === 'success'">✓</template>
          <template v-else-if="toast.type === 'error'">✕</template>
          <template v-else-if="toast.type === 'warning'">⚠</template>
          <template v-else>ℹ</template>
        </span>
        <span class="toast-message">{{ toast.message }}</span>
      </div>
    </TransitionGroup>
  </div>
</template>

<style scoped>
.toast-container {
  position: fixed;
  top: 20px;
  right: 20px;
  z-index: 9999;
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.toast {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 12px 20px;
  border-radius: 8px;
  color: white;
  font-size: 14px;
  cursor: pointer;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  min-width: 200px;
  max-width: 400px;
}

.toast-success {
  background: #4caf50;
}

.toast-error {
  background: #f44336;
}

.toast-warning {
  background: #ff9800;
}

.toast-info {
  background: #2196f3;
}

.toast-icon {
  font-size: 16px;
  font-weight: bold;
}

.toast-message {
  flex: 1;
}

/* 动画 */
.toast-enter-active,
.toast-leave-active {
  transition: all 0.3s ease;
}

.toast-enter-from {
  opacity: 0;
  transform: translateX(100%);
}

.toast-leave-to {
  opacity: 0;
  transform: translateX(100%);
}
</style>

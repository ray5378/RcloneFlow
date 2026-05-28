<script setup lang="ts">
defineProps<{
  title: string
  show: boolean
}>()

const emit = defineEmits<{
  close: []
}>()
</script>

<template>
  <Teleport to="body">
    <div v-if="show" class="modal-overlay" @click.self="emit('close')">
      <div class="modal">
        <div class="modal-header">
          <h2>{{ title }}</h2>
          <button class="modal-close" @click="emit('close')">&times;</button>
        </div>
        <slot />
      </div>
    </div>
  </Teleport>
</template>

<style scoped>
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.7);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  padding: 20px;
}

.modal {
  background: #111827;
  color: #e5e7eb;
  border: 1px solid #374151;
  border-radius: 12px;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.5);
  width: 100%;
  max-height: 90vh;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
}

body.light .modal {
  background: #fff;
  color: #111827;
  border: 1px solid #e5e7eb;
}

.modal-header h2 {
  color: #e5e7eb;
  margin: 0;
  font-size: 18px;
}

body.light .modal-header h2 {
  color: #111827;
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 20px;
  border-bottom: 1px solid #374151;
  flex-shrink: 0;
}

body.light .modal-header {
  border-bottom: 1px solid #e5e7eb;
}

.modal-close {
  background: none;
  border: none;
  font-size: 24px;
  color: #9ca3af;
  cursor: pointer;
  padding: 0;
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 6px;
  transition: background 0.2s;
}

.modal-close:hover {
  background: rgba(255, 255, 255, 0.1);
}

body.light .modal-close:hover {
  background: rgba(0, 0, 0, 0.05);
}

@media (max-width: 640px) {
  .modal-overlay {
    padding: 10px;
  }

  .modal {
    max-height: 85vh;
    border-radius: 16px;
  }

  .modal-header {
    padding: 14px 16px;
  }

  .modal-header h2 {
    font-size: 16px;
  }
}
</style>

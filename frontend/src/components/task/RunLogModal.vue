<script setup lang="ts">
defineProps<{
  visible: boolean
  title: string
  content: string
}>()

const emit = defineEmits<{
  (e: 'close'): void
}>()
</script>

<template>
  <div v-if="visible" class="modal-overlay" @click.self="emit('close')">
    <div class="modal-content log-modal">
      <div class="modal-header">
        <h3>{{ title }}</h3>
        <button class="close-btn" @click="emit('close')">×</button>
      </div>
      <div class="modal-body">
        <div class="log-box">
          <pre class="log-pre">{{ content }}</pre>
        </div>
      </div>
      <div class="modal-footer">
        <button class="ghost" @click="emit('close')">关闭</button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.modal-content.log-modal {
  width: 92vw !important;
  max-width: 1200px !important;
  max-height: 80vh;
  display: flex;
  flex-direction: column;
}

.log-modal .modal-body {
  padding: 12px 16px;
  width: 100%;
  flex: 1;
  overflow: hidden;
  display: flex;
}

.log-box {
  width: 100%;
  display: flex;
  justify-content: center;
}

.log-pre {
  background: #0b1220;
  color: #e5e7eb;
  padding: 12px;
  border-radius: 8px;
  height: 100%;
  overflow: auto;
  white-space: pre-wrap;
  width: calc(100% - 64px);
  max-width: 1100px;
  box-sizing: border-box;
  margin: 0;
  border: 1px solid #334155;
}
</style>

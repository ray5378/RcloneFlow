<script setup lang="ts">
defineProps<{
  visible: boolean
  stats: any
  formatBytes: (bytes: number) => string
  formatBytesPerSec: (bytes: number) => string
  formatEta: (seconds: number) => string
}>()

const emit = defineEmits<{
  (e: 'close'): void
}>()
</script>

<template>
  <div v-if="visible" class="modal-overlay" @click.self="emit('close')">
    <div class="modal-content">
      <div class="modal-header">
        <h3>全局实时数据</h3>
        <button class="close-btn" @click="emit('close')">×</button>
      </div>
      <div class="modal-body">
        <div class="detail-item">
          <label>已传输：</label>
          <span>{{ formatBytes(stats.bytes) || '-' }}</span>
        </div>
        <div class="detail-item">
          <label>总大小：</label>
          <span>{{ formatBytes(stats.totalBytes) || '-' }}</span>
        </div>
        <div class="detail-item">
          <label>当前速度：</label>
          <span>{{ formatBytesPerSec(stats.speed) || '-' }}</span>
        </div>
        <div class="detail-item">
          <label>平均速度：</label>
          <span>{{ formatBytesPerSec(stats.speedAvg) || '-' }}</span>
        </div>
        <div class="detail-item">
          <label>预计剩余时间：</label>
          <span>{{ formatEta(stats.eta) || '-' }}</span>
        </div>
        <div class="detail-item">
          <label>进度：</label>
          <span>{{ stats.percentage !== undefined ? stats.percentage.toFixed(2) + '%' : '-' }}</span>
        </div>
        <div class="progress-bar-container">
          <div class="progress-bar" :style="{ width: (stats.percentage || 0) + '%' }"></div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.progress-bar-container {
  width: 100%;
  height: 8px;
  background: rgba(255, 255, 255, 0.08);
  border-radius: 999px;
  overflow: hidden;
  margin-top: 8px;
}

.progress-bar {
  height: 100%;
  background: linear-gradient(90deg, var(--accent, #4f46e5), #22c55e);
  border-radius: 999px;
  transition: width 0.2s ease;
}
</style>

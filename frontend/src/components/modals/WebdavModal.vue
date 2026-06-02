<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { getWebdavStatus, startWebdav, stopWebdav } from '../../api/webdav'
import { t } from '../../i18n'

const emit = defineEmits<{
  (e: 'close'): void
}>()

const loading = ref(true)
const running = ref(false)
const enabled = ref(false)
const statusError = ref('')
const startError = ref('')
const stopError = ref('')
const starting = ref(false)
const stopping = ref(false)
const davAddress = ref('')

async function loadStatus() {
  loading.value = true
  statusError.value = ''
  try {
    const status = await getWebdavStatus()
    running.value = status.running
    enabled.value = status.enabled
    const origin = window.location.origin
    davAddress.value = `${origin}${status.path}`
  } catch (e: any) {
    statusError.value = e.message || '获取状态失败'
  } finally {
    loading.value = false
  }
}

async function onStart() {
  starting.value = true
  startError.value = ''
  try {
    await startWebdav()
    running.value = true
    enabled.value = true
    const origin = window.location.origin
    davAddress.value = `${origin}/dav/`
  } catch (e: any) {
    startError.value = e.message || '开启失败'
  } finally {
    starting.value = false
  }
}

async function onStop() {
  stopping.value = true
  stopError.value = ''
  try {
    await stopWebdav()
    running.value = false
    enabled.value = false
  } catch (e: any) {
    stopError.value = e.message || '关闭失败'
  } finally {
    stopping.value = false
  }
}

onMounted(loadStatus)
</script>

<template>
  <div class="modal-overlay" @click.self="$emit('close')">
    <div class="modal-content webdav-modal">
      <div class="modal-header">
        <h3>{{ t('webdav.title') }}</h3>
        <button class="close-btn" @click="$emit('close')">×</button>
      </div>
      <div class="modal-body" v-if="!loading">
        <p class="webdav-desc">{{ t('webdav.description') }}</p>

        <div class="webdav-status-bar">
          <span class="status-dot" :class="{ active: running }"></span>
          <span class="status-text">{{ running ? t('webdav.running') : t('webdav.stopped') }}</span>
        </div>

        <div class="webdav-note">
          <div class="note-label">{{ t('webdav.address') }}</div>
          <code class="note-address">{{ davAddress }}</code>
          <div class="note-hint">{{ t('webdav.addressNote') }}</div>
        </div>

        <div v-if="startError" class="error">{{ startError }}</div>
        <div v-if="stopError" class="error">{{ stopError }}</div>

        <div class="webdav-actions">
          <button
            v-if="!running"
            class="primary start-btn"
            @click="onStart"
            :disabled="starting"
          >
            {{ starting ? t('webdav.starting') : t('webdav.start') }}
          </button>
          <button
            v-if="running"
            class="danger stop-btn"
            @click="onStop"
            :disabled="stopping"
          >
            {{ stopping ? t('webdav.stopping') : t('webdav.stop') }}
          </button>
        </div>

        <div class="webdav-tips">
          <div class="tips-title">{{ t('webdav.tips') }}</div>
          <ul>
            <li>{{ t('webdav.tip1') }}</li>
            <li>{{ t('webdav.tip2') }}</li>
            <li>{{ t('webdav.tip3') }}</li>
            <li>{{ t('webdav.tip4') }}</li>
            <li>{{ t('webdav.tip5') }}</li>
          </ul>
        </div>
      </div>
      <div v-if="loading" class="modal-body">
        <p>{{ t('common.loading') }}</p>
      </div>
      <div v-if="statusError" class="modal-body">
        <p class="error">{{ statusError }}</p>
      </div>
    </div>
  </div>
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

.modal-content {
  background: var(--surface);
  border-radius: 12px;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.5);
  width: 100%;
  max-height: 90vh;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
}

.webdav-modal {
  max-width: 520px;
  width: 92vw;
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 20px;
  border-bottom: 1px solid var(--border);
  flex-shrink: 0;
}

.modal-header h3 {
  margin: 0;
  font-size: 18px;
  color: var(--text);
}

.close-btn {
  background: none;
  border: none;
  font-size: 24px;
  color: var(--muted);
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

.close-btn:hover {
  background: var(--hover);
}

.modal-body {
  padding: 20px;
  overflow-y: auto;
  flex: 1;
}

.webdav-desc {
  color: var(--muted);
  font-size: 14px;
  margin: 0 0 16px 0;
  line-height: 1.5;
}

.webdav-status-bar {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 16px;
  padding: 12px;
  background: var(--hover);
  border-radius: 8px;
}

.status-dot {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  background: #666;
}

.status-dot.active {
  background: #10b981;
  box-shadow: 0 0 6px rgba(16, 185, 129, 0.5);
}

.status-text {
  font-size: 14px;
  color: var(--text);
}

.webdav-note {
  margin-bottom: 16px;
  padding: 14px;
  background: rgba(100, 181, 246, 0.08);
  border: 1px solid rgba(100, 181, 246, 0.2);
  border-radius: 8px;
}

.note-label {
  font-size: 12px;
  color: var(--muted);
  margin-bottom: 6px;
  font-weight: 600;
}

.note-address {
  display: block;
  padding: 10px 12px;
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: 8px;
  font-size: 14px;
  color: #64b5f6;
  word-break: break-all;
  font-family: monospace;
}

.note-hint {
  margin-top: 8px;
  font-size: 12px;
  color: var(--muted);
  line-height: 1.5;
}

.webdav-actions {
  margin-bottom: 16px;
}

.start-btn {
  width: 100%;
  padding: 12px;
  background: #10b981;
  color: #fff;
  border: none;
  border-radius: 8px;
  font-size: 15px;
  cursor: pointer;
  transition: background 0.2s;
}

.start-btn:hover:not(:disabled) {
  background: #059669;
}

.start-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.stop-btn {
  width: 100%;
  padding: 12px;
  background: #ef4444;
  color: #fff;
  border: none;
  border-radius: 8px;
  font-size: 15px;
  cursor: pointer;
  transition: background 0.2s;
}

.stop-btn:hover:not(:disabled) {
  background: #dc2626;
}

.stop-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.error {
  color: #ef4444;
  font-size: 13px;
  margin-bottom: 12px;
}

.webdav-tips {
  padding: 16px;
  background: rgba(100, 181, 246, 0.08);
  border: 1px solid rgba(100, 181, 246, 0.2);
  border-radius: 8px;
}

.tips-title {
  font-size: 13px;
  font-weight: 600;
  color: var(--text);
  margin-bottom: 8px;
}

.webdav-tips ul {
  margin: 0;
  padding-left: 18px;
}

.webdav-tips li {
  font-size: 12px;
  color: var(--muted);
  line-height: 1.6;
  margin-bottom: 4px;
}
</style>
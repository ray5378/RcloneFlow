<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { getWebdavStatus, startWebdav, stopWebdav, getCredentials, saveCredentials, saveCacheSettings, cleanupCache, setWebdavMode } from '../../api/webdav'
import { showSuccessToast, showErrorToast } from '../../api/errors'
import { t } from '../../i18n'

const emit = defineEmits<{
  (e: 'close'): void
}>()

const loading = ref(true)
const running = ref(false)
const enabled = ref(false)
const statusError = ref('')
const actionError = ref('')
const starting = ref(false)
const stopping = ref(false)
const davAddress = ref('')
const cacheSaving = ref(false)
const cacheCleaning = ref(false)

const username = ref('')
const password = ref('')
const showPassword = ref(false)

const cacheMaxSize = ref('1G')
const cacheCleanupInterval = ref('24h')
const accessMode = ref('cache')
const modeSwitching = ref(false)

async function loadStatus() {
  loading.value = true
  statusError.value = ''
  try {
    const [status, cred] = await Promise.all([
      getWebdavStatus(),
      getCredentials()
    ])
    running.value = status.running
    enabled.value = status.enabled
    const origin = window.location.origin
    davAddress.value = `${origin}${status.path}`
    if (cred.ok) {
      username.value = cred.username
      password.value = cred.password
    }
    cacheMaxSize.value = status.cache_max_size || '1G'
    cacheCleanupInterval.value = status.cache_cleanup_interval || '24h'
    accessMode.value = status.mode || 'cache'
  } catch (e: any) {
    statusError.value = e.message || '获取状态失败'
  } finally {
    loading.value = false
  }
}

async function saveAndStart() {
  if (!username.value || !password.value) return
  actionError.value = ''

  starting.value = true
  try {
    await saveCredentials(username.value, password.value)
    await startWebdav()
    running.value = true
    enabled.value = true
    const origin = window.location.origin
    davAddress.value = `${origin}/dav/`
    showSuccessToast(t('webdav.startSuccess'))
  } catch (e: any) {
    actionError.value = e.message || '操作失败'
    showErrorToast(e.message || t('webdav.startFailed'))
  } finally {
    starting.value = false
  }
}

async function onStop() {
  stopping.value = true
  actionError.value = ''
  try {
    await stopWebdav()
    running.value = false
    enabled.value = false
    showSuccessToast(t('webdav.stopSuccess'))
  } catch (e: any) {
    actionError.value = e.message || '关闭失败'
    showErrorToast(e.message || t('webdav.stopFailed'))
  } finally {
    stopping.value = false
  }
}

async function onSaveCache() {
  cacheSaving.value = true
  actionError.value = ''
  try {
    await saveCacheSettings(cacheMaxSize.value, cacheCleanupInterval.value)
    showSuccessToast(t('webdav.saveCacheSuccess'))
  } catch (e: any) {
    actionError.value = e.message || '保存缓存设置失败'
    showErrorToast(e.message || '保存缓存设置失败')
  } finally {
    cacheSaving.value = false
  }
}

async function onCleanupCache() {
  cacheCleaning.value = true
  actionError.value = ''
  try {
    await cleanupCache()
    showSuccessToast(t('webdav.cleanupSuccess'))
  } catch (e: any) {
    actionError.value = e.message || '清理缓存失败'
    showErrorToast(e.message || '清理缓存失败')
  } finally {
    cacheCleaning.value = false
  }
}

async function onSwitchMode() {
  const newMode = accessMode.value === 'cache' ? 'stream' : 'cache'
  modeSwitching.value = true
  actionError.value = ''
  try {
    await setWebdavMode(newMode)
    accessMode.value = newMode
    showSuccessToast(t('webdav.modeSwitchSuccess'))
  } catch (e: any) {
    showErrorToast(e.message || '切换模式失败')
  } finally {
    modeSwitching.value = false
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

        <div class="mode-section" v-if="running">
          <div class="mode-title">{{ t('webdav.accessMode') }}</div>
          <div class="mode-toggle-bar">
            <button
              class="mode-btn"
              :class="{ active: accessMode === 'cache' }"
              :disabled="modeSwitching"
              @click="accessMode !== 'cache' && onSwitchMode()"
            >
              {{ t('webdav.modeCache') }}
            </button>
            <button
              class="mode-btn"
              :class="{ active: accessMode === 'stream' }"
              :disabled="modeSwitching"
              @click="accessMode !== 'stream' && onSwitchMode()"
            >
              {{ t('webdav.modeStream') }}
            </button>
          </div>
          <div class="mode-hint">
            {{ accessMode === 'cache' ? t('webdav.modeCacheHint') : t('webdav.modeStreamHint') }}
          </div>
        </div>

        <div class="cred-section">
          <div class="cred-title">{{ t('webdav.credentials') }}</div>
          <div class="field-item">
            <label>{{ t('webdav.username') }}</label>
            <input
              v-model="username"
              type="text"
              :placeholder="t('webdav.usernamePlaceholder')"
              :disabled="running"
            />
          </div>
          <div class="field-item">
            <label>{{ t('webdav.password') }}</label>
            <div class="pw-input-wrap">
              <input
                v-model="password"
                :type="showPassword ? 'text' : 'password'"
                :placeholder="t('webdav.passwordPlaceholder')"
                :disabled="running"
              />
              <button
                class="toggle-pw-btn"
                @click="showPassword = !showPassword"
                :title="showPassword ? t('webdav.hide') : t('webdav.show')"
              >
                <span v-if="showPassword">🙈</span>
                <span v-else>👁️</span>
              </button>
            </div>
          </div>
        </div>

        <div class="cache-section">
          <div class="cache-title">{{ t('webdav.cacheSettings') }}</div>

          <div class="field-item">
            <label>{{ t('webdav.cacheMaxSize') }}</label>
            <input
              v-model="cacheMaxSize"
              type="text"
              :placeholder="t('webdav.cacheMaxSizePlaceholder')"
            />
            <div class="field-hint">{{ t('webdav.cacheMaxSizeHint') }}</div>
          </div>
          <div class="field-item">
            <label>{{ t('webdav.cacheCleanupInterval') }}</label>
            <input
              v-model="cacheCleanupInterval"
              type="text"
              :placeholder="t('webdav.cacheCleanupIntervalPlaceholder')"
            />
          </div>
          <div class="cache-actions">
            <button
              class="secondary cache-save-btn"
              @click="onSaveCache"
              :disabled="cacheSaving"
            >
              {{ cacheSaving ? t('common.saving') : t('webdav.saveCache') }}
            </button>
            <button
              class="secondary cache-cleanup-btn"
              @click="onCleanupCache"
              :disabled="cacheCleaning"
            >
              {{ cacheCleaning ? t('webdav.cleaning') : t('webdav.cleanupNow') }}
            </button>
          </div>
        </div>

        <div v-if="actionError" class="error">{{ actionError }}</div>

        <div class="webdav-actions">
          <button
            v-if="!running"
            class="primary start-btn"
            @click="saveAndStart"
            :disabled="starting || !username || !password"
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

/* 模式切换 */
.mode-section {
  margin-bottom: 16px;
  padding: 12px;
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: 8px;
}

.mode-title {
  font-size: 13px;
  font-weight: 600;
  color: var(--text);
  margin-bottom: 8px;
}

.mode-toggle-bar {
  display: flex;
  gap: 0;
  border: 1px solid var(--border);
  border-radius: 8px;
  overflow: hidden;
}

.mode-btn {
  flex: 1;
  padding: 8px 12px;
  border: none;
  background: var(--surface);
  color: var(--muted);
  font-size: 13px;
  cursor: pointer;
  transition: all 0.2s;
}

.mode-btn:not(:last-child) {
  border-right: 1px solid var(--border);
}

.mode-btn.active {
  background: #64b5f6;
  color: #fff;
  font-weight: 600;
}

.mode-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.mode-hint {
  margin-top: 8px;
  font-size: 12px;
  color: var(--muted);
  line-height: 1.5;
}

.cred-section {
  margin-bottom: 16px;
  padding: 14px;
  background: var(--hover);
  border-radius: 8px;
}

.cred-title {
  font-size: 13px;
  font-weight: 600;
  color: var(--text);
  margin-bottom: 12px;
}

.cache-section {
  margin-bottom: 16px;
  padding: 14px;
  background: var(--hover);
  border-radius: 8px;
}

.cache-title {
  font-size: 13px;
  font-weight: 600;
  color: var(--text);
  margin-bottom: 12px;
}

.cache-actions {
  display: flex;
  gap: 8px;
  margin-top: 12px;
}

.cache-actions button {
  flex: 1;
  padding: 8px 12px;
  border-radius: 8px;
  font-size: 13px;
  cursor: pointer;
  transition: background 0.2s;
}

.cache-save-btn {
  background: var(--surface);
  border: 1px solid var(--border);
  color: var(--text);
}

.cache-save-btn:hover:not(:disabled) {
  background: var(--hover);
}

.cache-save-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.cache-cleanup-btn {
  background: var(--surface);
  border: 1px solid var(--border);
  color: #f59e0b;
}

.cache-cleanup-btn:hover:not(:disabled) {
  background: rgba(245, 158, 11, 0.1);
  border-color: #f59e0b;
}

.cache-cleanup-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.field-item {
  margin-bottom: 12px;
}

.field-item:last-child {
  margin-bottom: 0;
}

.field-item label {
  display: block;
  font-size: 13px;
  color: var(--muted);
  margin-bottom: 4px;
}

.field-item input {
  width: 100%;
  padding: 10px 12px;
  border: 1px solid var(--border);
  border-radius: 8px;
  background: var(--surface);
  color: var(--text);
  font-size: 14px;
  box-sizing: border-box;
}

.field-item input:focus {
  outline: none;
  border-color: #64b5f6;
}

.field-item input:disabled {
  opacity: 0.6;
}

.field-hint {
  margin-top: 4px;
  font-size: 12px;
  color: var(--muted);
  line-height: 1.5;
}

.pw-input-wrap {
  display: flex;
  align-items: center;
  gap: 0;
}

.pw-input-wrap input {
  flex: 1;
  border-top-right-radius: 0;
  border-bottom-right-radius: 0;
}

.toggle-pw-btn {
  padding: 10px 12px;
  border: 1px solid var(--border);
  border-left: none;
  border-radius: 0 8px 8px 0;
  background: var(--surface);
  color: var(--muted);
  cursor: pointer;
  font-size: 16px;
  line-height: 1;
  transition: background 0.2s;
}

.toggle-pw-btn:hover {
  background: var(--hover);
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
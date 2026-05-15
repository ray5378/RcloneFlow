<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { getSettings, saveSettings, resetSettings } from '../api/settings'
import { t } from '../i18n'

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'settings-saved', values: Record<string, string>): void
}>()

const loading = ref(true)
const saving = ref(false)
const saved = ref(false)
const showResetConfirm = ref(false)
const form = ref<Record<string, string>>({})
const errors = ref<Record<string, string>>({})
const saveFailed = ref(false)
const durationRe = /^\s*\d+\s*(ms|s|m|h|d)\s*$/i

function validate() {
  errors.value = {}
  const intFields = ['FINAL_SUMMARY_RETENTION_DAYS', 'CLEANUP_INTERVAL_HOURS', 'WEBHOOK_MAX_FILES']
  for (const k of intFields) {
    const v = (form.value as any)[k]
    if (v !== '' && (isNaN(Number(v)) || !Number.isFinite(Number(v)) || Number(v) < 0)) {
      errors.value[k] = t('defaults.errNonNegative')
    }
  }
  const mb = (form.value as any).PROGRESS_FLUSH_MIN_DELTA_BYTES
  if (mb !== '' && (isNaN(Number(mb)) || Number(mb) < 0)) {
    errors.value.PROGRESS_FLUSH_MIN_DELTA_BYTES = t('defaults.errNonNegativeMb')
  }
  const pct = (form.value as any).PROGRESS_FLUSH_MIN_DELTA_PCT
  if (pct !== '' && (isNaN(Number(pct)) || Number(pct) < 0 || Number(pct) > 100)) {
    errors.value.PROGRESS_FLUSH_MIN_DELTA_PCT = t('defaults.errPct')
  }
  const durFields = ['ACCESS_TOKEN_TTL', 'REFRESH_TOKEN_TTL', 'PROGRESS_FLUSH_INTERVAL', 'FINISH_WAIT_INTERVAL', 'FINISH_WAIT_TIMEOUT']
  for (const k of durFields) {
    const v = (form.value as any)[k]
    if (v && !durationRe.test(String(v))) {
      errors.value[k] = t('defaults.errDuration')
    }
  }
  return Object.keys(errors.value).length === 0
}

function flat(resp: any) {
  const out: Record<string, string> = {}
  const patch = (grp: any) => { Object.keys(grp || {}).forEach(k => out[k] = grp[k]?.effective ?? '') }
  patch(resp.auth)
  patch(resp.log)
  patch(resp.history)
  patch(resp.precheck)
  patch(resp.progress)
  patch(resp.webdav)
  patch(resp.webhook)
  return out
}

async function load() {
  loading.value = true
  try {
    const resp = await getSettings()
    form.value = flat(resp)
    const b = Number((form.value as any).PROGRESS_FLUSH_MIN_DELTA_BYTES || 0)
    if (!isNaN(b) && isFinite(b) && b > 0) {
      ;(form.value as any).PROGRESS_FLUSH_MIN_DELTA_BYTES = String(Math.round((b / 1048576) * 100) / 100)
    }
  } finally {
    loading.value = false
  }
}

async function onSave() {
  if (!validate()) return
  const payload: Record<string, string> = {}
  Object.keys(form.value).forEach(k => {
    const v = (form.value as any)[k]
    payload[k] = v === undefined || v === null ? '' : String(v)
  })
  const mb = Number(payload.PROGRESS_FLUSH_MIN_DELTA_BYTES)
  if (!isNaN(mb) && isFinite(mb) && mb >= 0) {
    payload.PROGRESS_FLUSH_MIN_DELTA_BYTES = String(Math.round(mb * 1048576))
  }
  saving.value = true
  saveFailed.value = false
  try {
    await saveSettings(payload)
    await load()
    emit('settings-saved', { ...form.value })
    saved.value = true
    setTimeout(() => { saved.value = false }, 10000)
  } catch (e: any) {
    console.error(e)
    saveFailed.value = true
    setTimeout(() => { saveFailed.value = false }, 3000)
  } finally {
    saving.value = false
  }
}

function onSavedClick() {
  if (saved.value) emit('close')
}

async function onReset() {
  showResetConfirm.value = true
}

async function doConfirmReset() {
  saving.value = true
  try {
    await resetSettings()
    await load()
    emit('settings-saved', { ...form.value })
    saved.value = true
    setTimeout(() => { saved.value = false }, 10000)
  } catch (e: any) {
    alert(e?.message || e)
  } finally {
    saving.value = false
    showResetConfirm.value = false
  }
}

onMounted(load)
</script>

<template>
  <div class="modal-overlay" @click.self="$emit('close')">
    <div class="modal-content large">
      <div class="modal-header">
        <h3>{{ t('settings.defaults') }}</h3>
        <button class="close-btn" @click="$emit('close')">×</button>
      </div>
      <div class="modal-body" v-if="!loading">
        <div class="section">
          <div class="section-title">{{ t('defaults.auth') }}</div>
          <div class="grid">
            <label :title="t('defaults.accessTokenTtlTitle')">{{ t('defaults.accessTokenTtl') }}</label>
            <input v-model="form.ACCESS_TOKEN_TTL" :placeholder="t('defaults.durationPlaceholder24h')" />
            <div class="error" v-if="errors.ACCESS_TOKEN_TTL">{{ errors.ACCESS_TOKEN_TTL }}</div>
            <label :title="t('defaults.refreshTokenTtlTitle')">{{ t('defaults.refreshTokenTtl') }}</label>
            <input v-model="form.REFRESH_TOKEN_TTL" :placeholder="t('defaults.durationPlaceholder90d')" />
            <div class="error" v-if="errors.REFRESH_TOKEN_TTL">{{ errors.REFRESH_TOKEN_TTL }}</div>
          </div>
        </div>

        <div class="section">
          <div class="section-title">{{ t('defaults.logRealtime') }}</div>
          <div class="grid">
            <label :title="t('defaults.logLevelTitle')">{{ t('defaults.logLevel') }}</label>
            <select v-model="form.LOG_LEVEL">
              <option value="debug">{{ t('defaults.logDebug') }}</option>
              <option value="info">{{ t('defaults.logInfo') }}</option>
              <option value="warn">{{ t('defaults.logWarn') }}</option>
              <option value="error">{{ t('defaults.logError') }}</option>
            </select>
          </div>
        </div>

        <div class="section">
          <div class="section-title">{{ t('defaults.historyRetention') }}</div>
          <div class="grid">
            <label :title="t('defaults.retentionDaysTitle')">{{ t('defaults.retentionDays') }}</label>
            <input v-model="form.FINAL_SUMMARY_RETENTION_DAYS" type="number" min="0" />
            <div class="error" v-if="errors.FINAL_SUMMARY_RETENTION_DAYS">{{ errors.FINAL_SUMMARY_RETENTION_DAYS }}</div>
            <label :title="t('defaults.cleanupIntervalHoursTitle')">{{ t('defaults.cleanupIntervalHours') }}</label>
            <input v-model="form.CLEANUP_INTERVAL_HOURS" type="number" min="0" />
            <div class="error" v-if="errors.CLEANUP_INTERVAL_HOURS">{{ errors.CLEANUP_INTERVAL_HOURS }}</div>
          </div>
        </div>

        <div class="section">
          <div class="section-title">{{ t('defaults.progressFlush') }}</div>
          <div class="grid">
            <label :title="t('defaults.flushIntervalTitle')">{{ t('defaults.flushInterval') }}</label>
            <input v-model="form.PROGRESS_FLUSH_INTERVAL" :placeholder="t('defaults.durationPlaceholder5s')" />
            <div class="error" v-if="errors.PROGRESS_FLUSH_INTERVAL">{{ errors.PROGRESS_FLUSH_INTERVAL }}</div>
            <label :title="t('defaults.flushPctTitle')">{{ t('defaults.flushPct') }}</label>
            <input v-model="form.PROGRESS_FLUSH_MIN_DELTA_PCT" type="number" min="0" step="0.1" />
            <div class="error" v-if="errors.PROGRESS_FLUSH_MIN_DELTA_PCT">{{ errors.PROGRESS_FLUSH_MIN_DELTA_PCT }}</div>
            <label :title="t('defaults.flushMbTitle')">{{ t('defaults.flushMb') }}</label>
            <input v-model="form.PROGRESS_FLUSH_MIN_DELTA_BYTES" type="number" min="0" step="0.01" />
            <div class="error" v-if="errors.PROGRESS_FLUSH_MIN_DELTA_BYTES">{{ errors.PROGRESS_FLUSH_MIN_DELTA_BYTES }}</div>
          </div>
        </div>

        <div class="section">
          <div class="section-title">{{ t('modal.webdavFinalize') }}</div>
          <div class="grid">
            <label :title="t('defaults.webdavIntervalTitle')">{{ t('defaults.webdavInterval') }}</label>
            <input v-model="form.FINISH_WAIT_INTERVAL" :placeholder="t('defaults.durationPlaceholder5s')" />
            <div class="error" v-if="errors.FINISH_WAIT_INTERVAL">{{ errors.FINISH_WAIT_INTERVAL }}</div>
            <label :title="t('defaults.webdavTimeoutTitle')">{{ t('defaults.webdavTimeout') }}</label>
            <input v-model="form.FINISH_WAIT_TIMEOUT" :placeholder="t('defaults.durationPlaceholder5h')" />
            <div class="error" v-if="errors.FINISH_WAIT_TIMEOUT">{{ errors.FINISH_WAIT_TIMEOUT }}</div>
          </div>
        </div>

        <div class="section">
          <div class="section-title">{{ t('modal.webhookNotify') }}</div>
          <div class="grid">
            <label :title="t('defaults.webhookFileLimitTitle')">{{ t('defaults.webhookFileLimit') }}</label>
            <input v-model="form.WEBHOOK_MAX_FILES" type="number" min="0" :placeholder="t('defaults.zeroUnlimited')" />
          </div>
        </div>

        <div class="section">
          <div class="section-title">{{ t('defaults.runningDebug') }}</div>
          <div class="grid">
            <label :title="t('defaults.runningDebugTitle')">{{ t('defaults.runningDebugLabel') }}</label>
            <select v-model="form.RUNNING_HINT_DEBUG_ENABLED">
              <option value="false">{{ t('modal.off') }}</option>
              <option value="true">{{ t('modal.on') }}</option>
            </select>
          </div>
        </div>
      </div>
      <div class="modal-footer">
        <button class="ghost" @click="onReset" :disabled="saving">{{ t('modal.resetDefaults') }}</button>
        <button :class="['primary', { saved: saved, failed: saveFailed }]" @click="saved ? onSavedClick() : onSave()" :disabled="saving">
          {{ saved ? t('modal.savedActive') : (saveFailed ? t('modal.saveFailed') : t('common.save')) }}
        </button>
      </div>
    </div>
  </div>

  <div v-if="showResetConfirm" class="modal-overlay" @click.self="showResetConfirm=false">
    <div class="modal-content confirm-modal">
      <div class="modal-header">
        <h3>{{ t('modal.resetDefaults') }}</h3>
        <button class="close-btn" @click="showResetConfirm=false">×</button>
      </div>
      <div class="modal-body">
        <p>{{ t('modal.resetDefaultsConfirm') }}</p>
      </div>
      <div class="modal-footer">
        <button class="ghost" @click="showResetConfirm=false">{{ t('modal.cancel') }}</button>
        <button class="primary danger" @click="doConfirmReset">{{ t('modal.confirm') }}</button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.modal-content.large { max-width: 720px; width: 92vw; }
.section { margin-bottom: 16px; }
.section-title { font-weight: 700; margin-bottom: 8px; color: var(--text); }
.grid { display: grid; grid-template-columns: 220px 1fr; gap: 8px 12px; }
input, select { padding: 8px 10px; border-radius: 8px; border: 1px solid var(--border); background: var(--surface); color: var(--text); }
.modal-footer .primary.saved { background: var(--success); color: #fff; border: none; }
.modal-footer .primary.failed { background: var(--danger); color: #fff; border: none; }
.error { color: var(--danger); font-size: 12px; margin-top: 4px; }

@media (max-width: 768px) {
  .modal-content.large {
    width: min(100vw - 16px, 100%);
    max-width: none;
  }

  .grid {
    grid-template-columns: 1fr;
    gap: 6px;
  }

  .grid > label {
    margin-top: 4px;
  }

  .grid > input,
  .grid > select {
    width: 100%;
    min-width: 0;
  }
}
</style>

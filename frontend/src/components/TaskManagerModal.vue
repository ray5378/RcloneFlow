<script setup lang="ts">
import { ref } from 'vue'
import { exportTasks, importTasks, clearAllTasks } from '../api/task'
import { clearAllRuns } from '../api/run'
import { getTasks } from '../api/task'
import { t } from '../i18n'

const emit = defineEmits<{
  (e: 'close'): void
}>()

const showConflictModal = ref(false)
const showClearTasksConfirm = ref(false)
const showClearHistoryConfirm = ref(false)
const conflictCount = ref(0)
const importResult = ref<{ imported: number; skipped: number; overwritten: number } | null>(null)
const importing = ref(false)
const processing = ref(false)

let pendingImportData: any = null

async function handleExport() {
  try {
    const data = await exportTasks()
    const json = JSON.stringify(data, null, 2)
    const blob = new Blob([json], { type: 'application/json' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `rcloneflow-tasks-${new Date().toISOString().slice(0, 10)}.json`
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
    URL.revokeObjectURL(url)
  } catch (e: any) {
    alert(e?.message || t('taskManager.exportFailed'))
  }
}

function handleImportFile(event: Event) {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) return

  const reader = new FileReader()
  reader.onload = async (e) => {
    try {
      const raw = e.target?.result
      if (raw == null) {
        alert(t('taskManager.invalidFile') + ': 文件读取结果为空')
        return
      }
      const text = typeof raw === 'string' ? raw.trim() : new TextDecoder().decode(raw as ArrayBuffer).trim()
      if (!text) {
        alert(t('taskManager.invalidFile') + ': 文件内容为空')
        return
      }
      const data = JSON.parse(text)
      if (!data.tasks || !Array.isArray(data.tasks)) {
        alert(t('taskManager.invalidFile') + ': 缺少 tasks 字段，文件包含的键: ' + JSON.stringify(Object.keys(data)))
        return
      }
      if (data.tasks.length === 0) {
        alert(t('taskManager.noTasksToImport'))
        return
      }

      const existingTasks = await getTasks()
      const existingNames = new Set(existingTasks.map(t => t.name.toLowerCase()))
      const incomingNames = data.tasks.map((t: any) => (t.name || '').toLowerCase()).filter(Boolean)
      const conflicts = incomingNames.filter((n: string) => existingNames.has(n))

      if (conflicts.length > 0) {
        conflictCount.value = conflicts.length
        pendingImportData = data
        showConflictModal.value = true
      } else {
        await doImport('skip')
      }
    } catch (err: any) {
      alert(t('taskManager.invalidFile') + ': ' + (err?.message || String(err)))
    }
  }
  reader.onerror = () => {
    alert(t('taskManager.invalidFile') + ': 文件读取失败')
  }
  reader.readAsText(file)
  input.value = ''
}

async function doImport(strategy: 'skip' | 'overwrite') {
  if (!pendingImportData) return
  importing.value = true
  try {
    const result = await importTasks({
      tasks: pendingImportData.tasks,
      schedules: pendingImportData.schedules || [],
      conflictStrategy: strategy,
    })
    importResult.value = result
    showConflictModal.value = false
    pendingImportData = null
  } catch (e: any) {
    alert(e?.message || t('taskManager.importFailed'))
  } finally {
    importing.value = false
  }
}

async function handleClearTasks() {
  processing.value = true
  try {
    await clearAllTasks()
    showClearTasksConfirm.value = false
    alert(t('taskManager.clearTasksSuccess'))
    emit('close')
  } catch (e: any) {
    alert(e?.message || t('taskManager.operationFailed'))
  } finally {
    processing.value = false
  }
}

async function handleClearHistory() {
  processing.value = true
  try {
    await clearAllRuns()
    showClearHistoryConfirm.value = false
    alert(t('taskManager.clearHistorySuccess'))
    emit('close')
  } catch (e: any) {
    alert(e?.message || t('taskManager.operationFailed'))
  } finally {
    processing.value = false
  }
}
</script>

<template>
  <div class="modal-overlay" @click.self="$emit('close')">
    <div class="modal-content">
      <div class="modal-header">
        <h3>{{ t('taskManager.title') }}</h3>
        <button class="close-btn" @click="$emit('close')">×</button>
      </div>
      <div class="modal-body">
        <div class="action-list">
          <div class="action-item" @click="handleExport">
            <span class="action-icon">📤</span>
            <div class="action-info">
              <div class="action-name">{{ t('taskManager.exportTasks') }}</div>
              <div class="action-hint">{{ t('taskManager.exportHint') }}</div>
            </div>
            <span class="action-arrow">›</span>
          </div>

          <label class="action-item" for="import-file-input">
            <span class="action-icon">📥</span>
            <div class="action-info">
              <div class="action-name">{{ t('taskManager.importTasks') }}</div>
              <div class="action-hint">{{ t('taskManager.importHint') }}</div>
            </div>
            <span class="action-arrow">›</span>
            <input
              id="import-file-input"
              type="file"
              accept=".json"
              @change="handleImportFile"
              style="display: none"
            />
          </label>

          <div class="action-item danger" @click="showClearTasksConfirm = true">
            <span class="action-icon">🗑️</span>
            <div class="action-info">
              <div class="action-name">{{ t('taskManager.clearAllTasks') }}</div>
              <div class="action-hint">{{ t('taskManager.clearAllTasksHint') }}</div>
            </div>
            <span class="action-arrow">›</span>
          </div>

          <div class="action-item danger" @click="showClearHistoryConfirm = true">
            <span class="action-icon">🧹</span>
            <div class="action-info">
              <div class="action-name">{{ t('taskManager.clearHistory') }}</div>
              <div class="action-hint">{{ t('taskManager.clearHistoryHint') }}</div>
            </div>
            <span class="action-arrow">›</span>
          </div>
        </div>

        <div v-if="importResult" class="result-banner">
          {{ t('taskManager.importSuccess', {
            imported: importResult.imported,
            skipped: importResult.skipped,
            overwritten: importResult.overwritten,
          }) }}
        </div>
      </div>
      <div class="modal-footer">
        <button class="ghost" @click="$emit('close')">{{ t('modal.close') }}</button>
      </div>
    </div>
  </div>

  <div v-if="showConflictModal" class="modal-overlay" @click.self="showConflictModal = false">
    <div class="modal-content confirm-modal">
      <div class="modal-header">
        <h3>{{ t('taskManager.conflictTitle') }}</h3>
        <button class="close-btn" @click="showConflictModal = false">×</button>
      </div>
      <div class="modal-body">
        <p>{{ t('taskManager.conflictMessage', { count: conflictCount }) }}</p>
      </div>
      <div class="modal-footer">
        <button class="ghost" @click="showConflictModal = false; pendingImportData = null">{{ t('common.cancel') }}</button>
        <button class="ghost" @click="doImport('skip')" :disabled="importing">{{ t('taskManager.skipAll') }}</button>
        <button class="primary" @click="doImport('overwrite')" :disabled="importing">{{ t('taskManager.overwriteAll') }}</button>
      </div>
    </div>
  </div>

  <div v-if="showClearTasksConfirm" class="modal-overlay" @click.self="showClearTasksConfirm = false">
    <div class="modal-content confirm-modal">
      <div class="modal-header">
        <h3>{{ t('taskManager.clearAllTasks') }}</h3>
        <button class="close-btn" @click="showClearTasksConfirm = false">×</button>
      </div>
      <div class="modal-body">
        <p>{{ t('taskManager.clearTasksConfirm') }}</p>
      </div>
      <div class="modal-footer">
        <button class="ghost" @click="showClearTasksConfirm = false">{{ t('common.cancel') }}</button>
        <button class="primary danger" @click="handleClearTasks" :disabled="processing">{{ t('modal.confirm') }}</button>
      </div>
    </div>
  </div>

  <div v-if="showClearHistoryConfirm" class="modal-overlay" @click.self="showClearHistoryConfirm = false">
    <div class="modal-content confirm-modal">
      <div class="modal-header">
        <h3>{{ t('taskManager.clearHistory') }}</h3>
        <button class="close-btn" @click="showClearHistoryConfirm = false">×</button>
      </div>
      <div class="modal-body">
        <p>{{ t('taskManager.clearHistoryConfirm') }}</p>
      </div>
      <div class="modal-footer">
        <button class="ghost" @click="showClearHistoryConfirm = false">{{ t('common.cancel') }}</button>
        <button class="primary danger" @click="handleClearHistory" :disabled="processing">{{ t('modal.confirm') }}</button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.action-list { display: flex; flex-direction: column; gap: 10px; }
.action-item { display: flex; align-items: center; gap: 10px; padding: 14px; border-radius: 12px; background: var(--surface, #222); border: 1px solid var(--border, #2f2f2f); cursor: pointer; }
.action-item:hover { border-color: #64b5f6; }
.action-item.danger .action-icon, .action-item.danger .action-name { color: #ff8a80; }
.action-icon { font-size: 18px; }
.action-info { flex: 1; }
.action-name { font-weight: 600; font-size: 14px; }
.action-hint { font-size: 12px; color: #888; margin-top: 2px; }
.action-arrow { color: #666; font-size: 18px; }
.result-banner { margin-top: 12px; padding: 10px 14px; border-radius: 8px; background: rgba(100, 181, 246, 0.1); border: 1px solid rgba(100, 181, 246, 0.3); color: #64b5f6; font-size: 13px; text-align: center; }
.confirm-modal { width: min(420px, 92vw); }
.modal-body p { margin: 0; line-height: 1.6; color: #ccc; }
.modal-footer .primary.danger { background: #ff5252; }
.modal-footer .primary.danger:hover { background: #ff1744; }
.modal-footer .primary:disabled, .modal-footer .ghost:disabled { opacity: 0.5; cursor: not-allowed; }

body.light .action-item { background: #f8f9fb; border-color: #e8ebef; color: #1f2937; }
body.light .action-item:hover { border-color: #64b5f6; }
body.light .action-hint { color: #666; }
body.light .action-arrow { color: #999; }
body.light .result-banner { background: rgba(100, 181, 246, 0.08); border-color: rgba(100, 181, 246, 0.25); }
body.light .modal-body p { color: #444; }
</style>

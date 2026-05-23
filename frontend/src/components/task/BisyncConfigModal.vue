<script setup lang="ts">
import { ref, watch } from 'vue'
import { t } from '../../i18n'
import type { BisyncOptions } from './types'
import { getBisyncFiles, deleteBisyncFile } from '../../api/task'

const props = defineProps<{
  visible: boolean
  title?: string
  modelValue: BisyncOptions
  taskId?: number
  taskName?: string
}>()

const emit = defineEmits<{
  'update:modelValue': [value: BisyncOptions]
  save: [value: BisyncOptions]
  close: []
}>()

const draft = ref<BisyncOptions>({})
const loadingFiles = ref(false)
const bisyncFiles = ref<string[]>([])
const deletingFile = ref<string | null>(null)

function cloneOptions(value?: Partial<BisyncOptions> | null): BisyncOptions {
  return {
    compare: value?.compare || '',
    maxDelete: value?.maxDelete || '',
    checkAccess: !!value?.checkAccess,
    checkFilename: value?.checkFilename || '',
    conflictResolve: value?.conflictResolve || '',
    conflictLoser: value?.conflictLoser || '',
    conflictSuffix: value?.conflictSuffix || '',
    backupDir1: value?.backupDir1 || '',
    backupDir2: value?.backupDir2 || '',
    createEmptySrcDirs: !!value?.createEmptySrcDirs,
    removeEmptyDirs: !!value?.removeEmptyDirs,
    recover: !!value?.recover,
    resync: !!value?.resync,
  }
}

watch(() => [props.visible, props.modelValue] as const, () => {
  draft.value = cloneOptions(props.modelValue)
  if (props.visible && props.taskId) {
    loadBisyncFiles()
  }
}, { immediate: true, deep: true })

async function loadBisyncFiles() {
  if (!props.taskId) return
  loadingFiles.value = true
  try {
    const result = await getBisyncFiles(props.taskId)
    bisyncFiles.value = result.files || []
  } catch (error) {
    console.error('Failed to load bisync files:', error)
  } finally {
    loadingFiles.value = false
  }
}

async function handleDeleteFile(fileName: string) {
  if (!props.taskId || deletingFile.value) return
  deletingFile.value = fileName
  try {
    await deleteBisyncFile(props.taskId, fileName)
    await loadBisyncFiles()
  } catch (error) {
    console.error('Failed to delete bisync file:', error)
  } finally {
    deletingFile.value = null
  }
}

function save() {
  const next = cloneOptions(draft.value)
  emit('update:modelValue', next)
  emit('save', next)
}

function toggleResync() {
  draft.value = {
    ...draft.value,
    resync: !draft.value.resync,
  }
}
</script>

<template>
  <div v-if="visible" class="modal-overlay" @click.self="emit('close')">
    <div class="modal-content bisync-modal">
      <div class="modal-header">
        <h3>{{ title || t('bisync.configTitle') }}</h3>
        <button class="close-btn" @click="emit('close')">×</button>
      </div>
      <div class="modal-body">
        <div class="section">
          <div class="section-title">{{ t('bisync.basicOptions') }}</div>
          <div class="option-row">
            <label>{{ t('bisync.compare') }}</label>
            <input v-model="draft.compare" type="text" :placeholder="t('bisync.comparePlaceholder')" />
          </div>
          <div class="option-row">
            <label>{{ t('bisync.maxDelete') }}</label>
            <input v-model="draft.maxDelete" type="text" :placeholder="t('bisync.maxDeletePlaceholder')" />
          </div>
          <div class="option-row checkbox-row">
            <label class="inline-label">
              <input type="checkbox" v-model="draft.checkAccess" />
              <span>{{ t('bisync.checkAccess') }}</span>
            </label>
          </div>
          <div class="option-row">
            <label>{{ t('bisync.checkFilename') }}</label>
            <input v-model="draft.checkFilename" type="text" :placeholder="t('bisync.checkFilenamePlaceholder')" />
          </div>
        </div>

        <div class="section">
          <div class="section-title">{{ t('bisync.conflictOptions') }}</div>
          <div class="option-row">
            <label>{{ t('bisync.conflictResolve') }}</label>
            <select v-model="draft.conflictResolve">
              <option value="">{{ t('bisync.conflictResolvePlaceholder') }}</option>
              <option value="path1">{{ t('bisync.conflictResolvePath1') }}</option>
              <option value="path2">{{ t('bisync.conflictResolvePath2') }}</option>
              <option value="newer">{{ t('bisync.conflictResolveNewer') }}</option>
              <option value="older">{{ t('bisync.conflictResolveOlder') }}</option>
            </select>
          </div>
          <div class="option-row">
            <label>{{ t('bisync.conflictLoser') }}</label>
            <select v-model="draft.conflictLoser">
              <option value="">{{ t('bisync.conflictLoserPlaceholder') }}</option>
              <option value="backup">{{ t('bisync.conflictLoserBackup') }}</option>
              <option value="delete">{{ t('bisync.conflictLoserDelete') }}</option>
            </select>
          </div>
          <div class="option-row">
            <label>{{ t('bisync.conflictSuffix') }}</label>
            <input v-model="draft.conflictSuffix" type="text" :placeholder="t('bisync.conflictSuffixPlaceholder')" />
          </div>
        </div>

        <div class="section">
          <div class="section-title">{{ t('bisync.backupOptions') }}</div>
          <div class="option-row">
            <label>{{ t('bisync.backupDir1') }}</label>
            <input v-model="draft.backupDir1" type="text" :placeholder="t('bisync.backupDirPlaceholder')" />
          </div>
          <div class="option-row">
            <label>{{ t('bisync.backupDir2') }}</label>
            <input v-model="draft.backupDir2" type="text" :placeholder="t('bisync.backupDirPlaceholder')" />
          </div>
        </div>

        <div class="section">
          <div class="section-title">{{ t('bisync.otherOptions') }}</div>
          <div class="option-row checkbox-row">
            <label class="inline-label">
              <input type="checkbox" v-model="draft.createEmptySrcDirs" />
              <span>{{ t('bisync.createEmptySrcDirs') }}</span>
            </label>
          </div>
          <div class="option-row checkbox-row">
            <label class="inline-label">
              <input type="checkbox" v-model="draft.removeEmptyDirs" />
              <span>{{ t('bisync.removeEmptyDirs') }}</span>
            </label>
          </div>
          <div class="option-row checkbox-row">
            <label class="inline-label">
              <input type="checkbox" v-model="draft.recover" />
              <span>{{ t('bisync.recover') }}</span>
            </label>
          </div>
          <div class="option-row checkbox-row">
            <label class="inline-label">
              <input type="checkbox" :checked="!!draft.resync" @change="toggleResync" />
              <span>{{ t('bisync.resync') }}</span>
            </label>
            <p class="hint">{{ t('bisync.resyncHint') }}</p>
          </div>
        </div>

        <div class="section">
          <div class="section-title">{{ t('bisync.lstFiles') }}</div>
          <div v-if="loadingFiles" class="loading-text">{{ t('common.loading') }}</div>
          <div v-else-if="bisyncFiles.length === 0" class="empty-text">{{ t('bisync.noLstFiles') }}</div>
          <div v-else class="file-list">
            <div v-for="file in bisyncFiles" :key="file" class="file-item">
              <span class="file-name">{{ file }}</span>
              <button
                type="button"
                class="ghost small delete-btn"
                :disabled="deletingFile === file"
                @click="handleDeleteFile(file)"
              >
                {{ deletingFile === file ? t('common.deleting') : t('common.delete') }}
              </button>
            </div>
          </div>
        </div>

        <div class="modal-actions">
          <button class="ghost" @click="emit('close')">{{ t('common.cancel') }}</button>
          <button class="primary" @click="save">{{ t('common.save') }}</button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.bisync-modal { width: min(720px, 92vw); max-width: 720px; }
.section { margin-bottom: 20px; }
.section-title { font-size: 14px; font-weight: 700; color: var(--text); margin-bottom: 12px; }
.option-row { display: flex; flex-direction: column; gap: 6px; margin-bottom: 12px; }
.option-row label { font-size: 12px; color: #888; }
.option-row input,
.option-row select {
  width: 100%;
  padding: 10px 12px;
  border: 1px solid var(--border);
  border-radius: 8px;
  background: var(--surface);
  color: var(--text);
  box-sizing: border-box;
}
.checkbox-row { flex-direction: row; align-items: center; gap: 8px; }
.checkbox-row label.inline-label { display: flex; align-items: center; gap: 8px; margin: 0; }
.checkbox-row input[type="checkbox"] { width: 16px; height: 16px; }
.hint { font-size: 11px; color: #888; margin-top: 4px; }
.file-list { display: flex; flex-direction: column; gap: 8px; }
.file-item { display: flex; justify-content: space-between; align-items: center; padding: 10px 12px; border: 1px solid var(--border); border-radius: 8px; background: var(--surface); }
.file-name { font-size: 13px; color: var(--text); }
.delete-btn { color: #ef4444; }
.loading-text,
.empty-text { font-size: 13px; color: #888; text-align: center; padding: 20px; }
.modal-actions { display: flex; justify-content: flex-end; gap: 10px; margin-top: 20px; }
body.light .option-row input,
body.light .option-row select { background: #fff; border-color: #ddd; color: #333; }
body.light .file-item { background: #f8fafc; border-color: #e5e7eb; }
</style>

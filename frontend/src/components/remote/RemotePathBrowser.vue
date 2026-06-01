<script setup lang="ts">
import { ref, watch, computed } from 'vue'
import * as api from '../../api'
import type { FileItem } from '../../api'

const props = defineProps<{
  modelValue: string
}>()

const emit = defineEmits<{
  'update:modelValue': [value: string]
}>()

const remotes = ref<string[]>([])
const selectedRemote = ref('')
const currentPath = ref('')
const items = ref<FileItem[]>([])
const loading = ref(false)
const errorMsg = ref('')

const breadcrumbs = computed(() => {
  const parts = currentPath.value.split('/').filter(Boolean)
  const crumbs: { name: string; path: string }[] = [{ name: selectedRemote.value, path: '' }]
  let accumulated = ''
  for (const part of parts) {
    accumulated += '/' + part
    crumbs.push({ name: part, path: accumulated })
  }
  return crumbs
})

const fullPath = computed(() => {
  if (!selectedRemote.value) return ''
  return selectedRemote.value + ':' + currentPath.value
})

watch(fullPath, (val) => {
  if (val && val !== props.modelValue) {
    emit('update:modelValue', val)
  }
})

async function loadRemotes() {
  try {
    const data = await api.getRemotes()
    remotes.value = data.remotes || []
  } catch {
    remotes.value = []
  }
}

async function listDir(path: string) {
  if (!selectedRemote.value) return
  loading.value = true
  errorMsg.value = ''
  try {
    const data = await api.listPath(selectedRemote.value, path || '')
    items.value = (data.items || []).filter(item => item.IsDir)
  } catch (e) {
    errorMsg.value = (e as Error).message
    items.value = []
  } finally {
    loading.value = false
  }
}

async function reload() {
  await loadRemotes()
  if (props.modelValue) {
    const colonIndex = props.modelValue.indexOf(':')
    if (colonIndex > 0) {
      selectedRemote.value = props.modelValue.substring(0, colonIndex)
      currentPath.value = props.modelValue.substring(colonIndex + 1)
      await listDir(currentPath.value)
      return
    }
  }
  if (!selectedRemote.value && remotes.value.length > 0) {
    selectedRemote.value = remotes.value[0]
  }
  if (selectedRemote.value) {
    currentPath.value = ''
    await listDir('')
  }
}

watch(selectedRemote, async (newVal) => {
  if (newVal) {
    currentPath.value = ''
    await listDir('')
  } else {
    items.value = []
    currentPath.value = ''
  }
})

function enterDir(item: FileItem) {
  currentPath.value = item.Path
  listDir(item.Path)
}

function navigateCrumb(path: string) {
  currentPath.value = path
  listDir(path)
}

function selectCurrentDir() {
  if (fullPath.value) {
    emit('update:modelValue', fullPath.value)
  }
}

loadRemotes().then(() => {
  if (props.modelValue) {
    const colonIndex = props.modelValue.indexOf(':')
    if (colonIndex > 0) {
      selectedRemote.value = props.modelValue.substring(0, colonIndex)
      currentPath.value = props.modelValue.substring(colonIndex + 1)
      listDir(currentPath.value)
    }
  }
})
</script>

<template>
  <div class="remote-path-browser">
    <div class="field-row">
      <select v-model="selectedRemote">
        <option value="">{{ $t('addTask.selectSourceStorage') }}</option>
        <option v-for="r in remotes" :key="r" :value="r">{{ r }}</option>
      </select>
      <button type="button" class="reload-btn" @click="reload" :title="$t('browserView.refresh')">↻</button>
    </div>

    <div v-if="selectedRemote" class="path-browse">
      <div class="pathbar">
        <template v-for="(crumb, i) in breadcrumbs" :key="crumb.path">
          <span v-if="i > 0" class="sep">/</span>
          <button
            class="crumb"
            :class="{ current: i === breadcrumbs.length - 1 }"
            @click="crumb.path !== currentPath && navigateCrumb(crumb.path)"
          >
            {{ crumb.name }}
          </button>
        </template>
      </div>
      <div class="path-list">
        <div v-if="loading" class="path-loading">{{ $t('common.loading') }}</div>
        <div v-else-if="errorMsg" class="path-error">{{ errorMsg }}</div>
        <div
          v-for="item in items"
          :key="item.Path"
          class="path-item is-dir"
          @click="enterDir(item)"
        >
          <span class="item-icon">📁</span>
          <span class="item-name">{{ item.Name }}</span>
        </div>
        <div v-if="!loading && !errorMsg && !items.length" class="path-empty">{{ $t('addTask.emptyDir') }}</div>
      </div>
    </div>

    <div v-if="selectedRemote && fullPath" class="selected-info">
      <span class="selected-label">{{ $t('remote.typeLabel') }}:</span>
      <code>{{ fullPath }}</code>
      <button type="button" class="select-btn" @click="selectCurrentDir">{{ $t('browserView.confirm') }}</button>
    </div>
  </div>
</template>

<style scoped>
.remote-path-browser {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.field-row {
  display: flex;
  gap: 8px;
  align-items: center;
}

.field-row select {
  flex: 1;
  padding: 10px 12px;
  border-radius: 10px;
  border: 1px solid #d1d5db;
  font: inherit;
  background: #fff;
}

.reload-btn {
  padding: 8px 12px;
  border-radius: 10px;
  border: 1px solid #d1d5db;
  background: #f3f4f6;
  cursor: pointer;
  font-size: 16px;
  line-height: 1;
}

.path-browse {
  border: 1px solid #d1d5db;
  border-radius: 10px;
  background: #fafafa;
  overflow: hidden;
}

.pathbar {
  display: flex;
  flex-wrap: wrap;
  gap: 0;
  padding: 8px 10px;
  border-bottom: 1px solid #e5e7eb;
  background: #f3f4f6;
}

.sep {
  color: #9ca3af;
  margin: 0 4px;
}

.crumb {
  background: transparent;
  border: none;
  color: #2563eb;
  cursor: pointer;
  padding: 0;
  font-size: 13px;
}

.crumb.current {
  color: #111827;
  font-weight: 600;
  cursor: default;
}

.path-list {
  max-height: 180px;
  overflow-y: auto;
  padding: 6px;
}

.path-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 7px 10px;
  cursor: pointer;
  border-radius: 6px;
  transition: background 0.2s;
}

.path-item:hover {
  background: rgba(37, 99, 235, 0.08);
}

.path-item.is-dir {
  color: #d97706;
}

.item-icon {
  font-size: 16px;
  flex-shrink: 0;
}

.item-name {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: 14px;
}

.path-loading,
.path-error,
.path-empty {
  padding: 16px;
  text-align: center;
  color: #9ca3af;
  font-size: 13px;
}

.path-error {
  color: #dc2626;
}

.selected-info {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  background: #f0f9ff;
  border-radius: 8px;
  font-size: 13px;
}

.selected-label {
  color: #6b7280;
  flex-shrink: 0;
}

.selected-info code {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: #1d4ed8;
  font-size: 12px;
}

.select-btn {
  padding: 4px 10px;
  border-radius: 6px;
  border: none;
  background: #2563eb;
  color: #fff;
  cursor: pointer;
  font-size: 12px;
  flex-shrink: 0;
}
</style>
<script setup lang="ts">
import { ref, watch, onMounted } from 'vue'
import { getBisyncLstFiles, deleteBisyncLstFile, rollbackBisyncLstFile, resyncBisync } from '../../api/task'
import type { BisyncLstVersion } from './types'

const props = defineProps<{
  visible: boolean
  taskId: number
}>()

const emit = defineEmits<{
  (e: 'close'): void
}>()

const versions = ref<BisyncLstVersion[]>([])
const loading = ref(false)
const actionLoading = ref<string | null>(null)
const deletingVersion = ref<string | null>(null)
const selectedVersionId = ref<string>('')

async function loadLstFiles() {
  if (!props.taskId) return
  loading.value = true
  try {
    const res = await getBisyncLstFiles(props.taskId)
    versions.value = res.versions || []
  } catch (error) {
    console.error('Failed to load lst files:', error)
  } finally {
    loading.value = false
  }
}

async function handleDelete(versionId: string) {
  if (!props.taskId) return
  deletingVersion.value = versionId
  actionLoading.value = 'delete'
  try {
    await deleteBisyncLstFile(props.taskId, versionId)
    await loadLstFiles()
  } catch (error) {
    console.error('Failed to delete lst version:', error)
  } finally {
    deletingVersion.value = null
    actionLoading.value = null
  }
}

function handleResolveConflict(version: any) {
  const confirmed = window.confirm(
    `检测到冲突文件：\n\n${version.conflict1 || ''}\n${version.conflict2 || ''}\n\n建议：\n1. 先备份冲突文件（如果有需要）\n2. 选择"确定"删除冲突文件\n3. 重新运行同步任务解决冲突\n\n是否删除这些冲突文件？`
  )
  
  if (confirmed) {
    handleDelete(version.id)
  }
}

async function handleRollback() {
  if (!props.taskId || !selectedVersionId.value) return
  actionLoading.value = 'rollback'
  try {
    await rollbackBisyncLstFile(props.taskId, selectedVersionId.value)
    await loadLstFiles()
    selectedVersionId.value = ''
  } catch (error) {
    console.error('Failed to rollback lst version:', error)
  } finally {
    actionLoading.value = null
  }
}

function formatTimestamp(timestamp: string): string {
  try {
    const date = new Date(timestamp)
    return date.toLocaleString('zh-CN', {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit'
    })
  } catch {
    return timestamp
  }
}

async function handleResync() {
  if (!props.taskId) return
  actionLoading.value = 'resync'
  try {
    await resyncBisync(props.taskId)
    await loadLstFiles()
  } catch (error) {
    console.error('Failed to resync:', error)
  } finally {
    actionLoading.value = null
  }
}

watch(
  () => props.visible,
  (newVal) => {
    if (newVal) {
      loadLstFiles()
    }
  }
)
</script>

<template>
  <div v-if="visible" class="modal-overlay" @click.self="emit('close')">
    <div class="modal-content" style="max-width: 700px">
      <div class="modal-header">
        <h3>Bisync 状态文件管理</h3>
        <button class="close-btn" @click="emit('close')">×</button>
      </div>
      <div class="modal-body">
        <div class="detail-item full-width">
          <p class="hint">管理双向同步的历史状态文件，可以回滚到之前的状态或清理旧文件</p>
        </div>

        <div class="detail-item full-width" style="margin-top: 8px">
          <button class="primary small" :disabled="actionLoading === 'resync'" @click="handleResync">
            {{ actionLoading === 'resync' ? '处理中...' : '重新同步 (Resync)' }}
          </button>
        </div>

        <div class="detail-item full-width" style="margin-top: 16px">
          <div class="rollback-section">
            <div class="rollback-header">
              <label>选择版本进行恢复：</label>
              <select v-model="selectedVersionId" class="version-select">
                <option value="">请选择要恢复的版本</option>
                <option v-for="version in versions" :key="version.id" :value="version.id">
                  {{ formatTimestamp(version.timestamp) }}
                  {{ version.path1Lst || version.path2Lst ? '(' + (version.path1Lst ? 'path1' : '') + (version.path1Lst && version.path2Lst ? ' + ' : '') + (version.path2Lst ? 'path2' : '') + ')' : '' }}
                </option>
              </select>
              <button 
                class="primary small" 
                :disabled="!selectedVersionId || actionLoading === 'rollback'"
                @click="handleRollback"
              >
                {{ actionLoading === 'rollback' ? '恢复中...' : '恢复到原配置' }}
              </button>
            </div>
          </div>
        </div>

        <div class="detail-item full-width" style="margin-top: 16px">
          <label>历史状态版本</label>
          <div v-if="loading" class="path-empty">加载中...</div>
          <div v-else-if="!versions.length" class="path-empty">暂无历史状态版本</div>
          <div v-else class="files-list">
            <div v-for="version in versions" :key="version.id" class="file-item" :class="{ selected: selectedVersionId === version.id, 'current': version.type === 'current', 'conflict': version.type === 'conflict' }">
              <div class="version-info">
                <div class="version-header">
                  <span class="version-badge" :class="version.type">
                    {{ version.type === 'current' ? '当前版本' : version.type === 'conflict' ? '冲突文件' : '历史备份' }}
                  </span>
                  <span class="version-time">
                    {{ version.id === 'current' ? '当前' : formatTimestamp(version.timestamp) }}
                  </span>
                </div>
                <div class="version-details">
                  <span v-if="version.path1Lst">📄 {{ version.path1Lst }}</span>
                  <span v-if="version.path1Lst && version.path2Lst"> + </span>
                  <span v-if="version.path2Lst">📄 {{ version.path2Lst }}</span>
                  <span v-if="version.type === 'conflict'" class="conflict-files">
                    <span v-if="version.conflict1">⚠️ {{ version.conflict1 }}</span>
                    <span v-if="version.conflict1 && version.conflict2"> | </span>
                    <span v-if="version.conflict2">⚠️ {{ version.conflict2 }}</span>
                  </span>
                </div>
              </div>
              <div class="file-actions">
                <button
                  v-if="version.type === 'conflict'"
                  class="ghost small"
                  :disabled="actionLoading !== null"
                  @click="handleResolveConflict(version)"
                >
                  处理冲突
                </button>
                <button
                  v-if="version.type !== 'current'"
                  class="ghost small"
                  style="color: #ef4444"
                  :disabled="actionLoading !== null"
                  @click="handleDelete(version.id)"
                >
                  {{ deletingVersion === version.id ? '删除中...' : '删除' }}
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>
      <div class="modal-footer">
        <button class="ghost" @click="emit('close')">关闭</button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.hint {
  margin-top: 8px;
  color: var(--muted, #94a3b8);
  font-size: 13px;
  line-height: 1.5;
}

.rollback-section {
  padding: 12px;
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: 8px;
}

.rollback-header {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}

.rollback-header label {
  font-size: 14px;
  font-weight: 500;
  color: var(--text);
  white-space: nowrap;
}

.version-select {
  flex: 1;
  min-width: 200px;
  padding: 8px 12px;
  border: 1px solid var(--border);
  border-radius: 6px;
  background: var(--bg);
  color: var(--text);
  font-size: 13px;
  cursor: pointer;
}

.version-select:focus {
  outline: none;
  border-color: var(--primary, #3b82f6);
}

.files-list {
  border: 1px solid var(--border);
  border-radius: 8px;
  background: var(--surface);
  max-height: 400px;
  overflow-y: auto;
}

.file-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  border-bottom: 1px solid var(--border);
}

.file-item:last-child {
  border-bottom: none;
}

.file-item.selected {
  background: rgba(59, 130, 246, 0.1);
  border-left: 3px solid var(--primary, #3b82f6);
}

.version-info {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.version-header {
  display: flex;
  align-items: center;
  gap: 8px;
}

.version-badge {
  display: inline-block;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 11px;
  font-weight: 600;
  text-transform: uppercase;
}

.version-badge.current {
  background: rgba(34, 197, 94, 0.15);
  color: #22c55e;
}

.version-badge.backup {
  background: rgba(59, 130, 246, 0.15);
  color: #3b82f6;
}

.version-badge.conflict {
  background: rgba(239, 68, 68, 0.15);
  color: #ef4444;
  animation: pulse 2s ease-in-out infinite;
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.7; }
}

.file-item.current {
  background: rgba(34, 197, 94, 0.05);
  border-left: 3px solid #22c55e;
}

.file-item.conflict {
  background: rgba(239, 68, 68, 0.08);
  border-left: 3px solid #ef4444;
}

.version-time {
  font-size: 14px;
  font-weight: 500;
  color: var(--text);
}

.version-details {
  font-size: 12px;
  color: var(--muted);
  font-family: monospace;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.conflict-files {
  margin-top: 4px;
  padding: 6px 8px;
  background: rgba(239, 68, 68, 0.1);
  border-radius: 4px;
  color: #ef4444;
  font-size: 11px;
}

.file-actions {
  display: flex;
  gap: 8px;
}

.path-empty {
  padding: 20px;
  text-align: center;
  color: #666;
  font-size: 13px;
}
</style>

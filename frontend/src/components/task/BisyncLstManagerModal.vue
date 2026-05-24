<script setup lang="ts">
import { ref, watch, onMounted } from 'vue'
import { getBisyncLstFiles, deleteBisyncLstFile, rollbackBisyncLstFile, resyncBisync } from '../../api/task'

const props = defineProps<{
  visible: boolean
  taskId: number
}>()

const emit = defineEmits<{
  (e: 'close'): void
}>()

const lstFiles = ref<string[]>([])
const loading = ref(false)
const actionLoading = ref<string | null>(null)
const deletingFile = ref<string | null>(null)
const rollbackingFile = ref<string | null>(null)

async function loadLstFiles() {
  if (!props.taskId) return
  loading.value = true
  try {
    const res = await getBisyncLstFiles(props.taskId)
    lstFiles.value = res.files || []
  } catch (error) {
    console.error('Failed to load lst files:', error)
  } finally {
    loading.value = false
  }
}

async function handleDelete(filename: string) {
  if (!props.taskId) return
  deletingFile.value = filename
  actionLoading.value = 'delete'
  try {
    await deleteBisyncLstFile(props.taskId, filename)
    await loadLstFiles()
  } catch (error) {
    console.error('Failed to delete lst file:', error)
  } finally {
    deletingFile.value = null
    actionLoading.value = null
  }
}

async function handleRollback(filename: string) {
  if (!props.taskId) return
  rollbackingFile.value = filename
  actionLoading.value = 'rollback'
  try {
    await rollbackBisyncLstFile(props.taskId, filename)
    await loadLstFiles()
  } catch (error) {
    console.error('Failed to rollback lst file:', error)
  } finally {
    rollbackingFile.value = null
    actionLoading.value = null
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
          <label>历史状态文件</label>
          <div v-if="loading" class="path-empty">加载中...</div>
          <div v-else-if="!lstFiles.length" class="path-empty">暂无历史状态文件</div>
          <div v-else class="files-list">
            <div v-for="file in lstFiles" :key="file" class="file-item">
              <span class="file-name">{{ file }}</span>
              <div class="file-actions">
                <button
                  class="ghost small"
                  :disabled="actionLoading !== null"
                  @click="handleRollback(file)"
                >
                  {{ rollbackingFile === file ? '回滚中...' : '回滚到此版本' }}
                </button>
                <button
                  class="ghost small"
                  style="color: #ef4444"
                  :disabled="actionLoading !== null"
                  @click="handleDelete(file)"
                >
                  {{ deletingFile === file ? '删除中...' : '删除' }}
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

.file-name {
  font-family: monospace;
  font-size: 13px;
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

<script setup lang="ts">
import { ref, watch, computed } from 'vue'
import { getBisyncLstFiles, getBisyncLstContent, deleteBisyncLstFile, rollbackBisyncLstFile, resyncBisync, updateTask, resolveBisyncConflict } from '../../api/task'
import type { BisyncLstVersion, BisyncOptions } from './types'
import type { Task } from '../../types'

const props = defineProps<{
  visible: boolean
  taskId: number
  bisyncOptions: BisyncOptions
}>()

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'update:bisyncOptions', value: BisyncOptions): void
}>()

const versions = ref<BisyncLstVersion[]>([])
const loading = ref(false)
const actionLoading = ref<string | null>(null)
const savingConfig = ref(false)

const localBisyncOptions = ref<BisyncOptions>({ ...props.bisyncOptions })

watch(() => props.bisyncOptions, (val) => {
  localBisyncOptions.value = { ...val }
}, { deep: true })

const contentModal = ref<{
  show: boolean
  title: string
  content: string
  loading: boolean
}>({ show: false, title: '', content: '', loading: false })

const confirmState = ref<{
  show: boolean
  title: string
  message: string
  onConfirm: () => void
}>({ show: false, title: '', message: '', onConfirm: () => {} })

const currentVersion = computed(() => versions.value.find(v => v.type === 'current'))
function isOldFormat(v: BisyncLstVersion): boolean {
  return !!(v.path1Lst?.endsWith('-old') || v.path2Lst?.endsWith('-old'))
}
const oldFormatBackups = computed(() => versions.value.filter(v => v.type === 'backup' && isOldFormat(v)))
const bakFormatBackups = computed(() => versions.value.filter(v => v.type === 'backup' && !isOldFormat(v)))
const conflictVersions = computed(() => versions.value.filter(v => v.type === 'conflict'))

function closeConfirm() {
  confirmState.value.show = false
}

function showConfirm(title: string, message: string, onConfirm: () => void) {
  confirmState.value = { show: true, title, message, onConfirm }
}

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

function handleViewContent(version: BisyncLstVersion) {
  const title = version.type === 'conflict'
    ? `冲突文件 - ${version.id}`
    : formatTimestamp(version.timestamp)
  contentModal.value = {
    show: true,
    title,
    content: '',
    loading: true
  }
  loadContent(version)
}

async function loadContent(version: BisyncLstVersion) {
  const parts: string[] = []
  try {
    if (version.type === 'conflict') {
      if (version.conflict1) {
        const r1 = await getBisyncLstContent(props.taskId, version.conflict1)
        parts.push(`=== 冲突版本 1: ${version.conflict1} ===\n${r1.content}`)
      }
      if (version.conflict2) {
        const r2 = await getBisyncLstContent(props.taskId, version.conflict2)
        parts.push(`=== 冲突版本 2: ${version.conflict2} ===\n${r2.content}`)
      }
    } else {
      if (version.path1Lst) {
        const r1 = await getBisyncLstContent(props.taskId, version.path1Lst)
        parts.push(`=== ${version.path1Lst} ===\n${r1.content}`)
      }
      if (version.path2Lst) {
        const r2 = await getBisyncLstContent(props.taskId, version.path2Lst)
        parts.push(`=== ${version.path2Lst} ===\n${r2.content}`)
      }
    }
    contentModal.value.content = parts.join('\n\n')
  } catch (error) {
    contentModal.value.content = '加载失败：' + (error as any)?.message || String(error)
  } finally {
    contentModal.value.loading = false
  }
}

function closeContentModal() {
  contentModal.value.show = false
}

function handleDeleteClick(version: BisyncLstVersion) {
  const label = version.type === 'conflict' ? '冲突文件' : formatTimestamp(version.timestamp)
  showConfirm(
    '删除确认',
    `确定要删除版本 "${label}" 的状态文件吗？此操作不可撤销。`,
    () => doDelete(version.id)
  )
}

async function doDelete(versionId: string) {
  if (!props.taskId) return
  actionLoading.value = 'delete'
  try {
    await deleteBisyncLstFile(props.taskId, versionId)
    await loadLstFiles()
  } catch (error) {
    console.error('Failed to delete lst version:', error)
  } finally {
    actionLoading.value = null
  }
}

function handleRollbackClick(version: BisyncLstVersion) {
  const label = formatTimestamp(version.timestamp)
  showConfirm(
    '恢复确认',
    `确定要将状态文件恢复到版本 "${label}" 吗？\n当前版本将被自动备份。`,
    () => doRollback(version.id)
  )
}

async function doRollback(versionId: string) {
  if (!props.taskId) return
  actionLoading.value = 'rollback'
  try {
    await rollbackBisyncLstFile(props.taskId, versionId)
    await loadLstFiles()
  } catch (error) {
    console.error('Failed to rollback lst version:', error)
  } finally {
    actionLoading.value = null
  }
}

function handleResyncClick() {
  showConfirm(
    '重新同步确认',
    '重新同步将执行一次双向同步的初始扫描（--resync），期间可能产生大量数据传输。\n建议在两端文件状态不一致时使用。\n\n确定要执行重新同步吗？',
    () => doResync()
  )
}

async function doResync() {
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

function handleResolveConflict(version: BisyncLstVersion, keepFile: 'conflict1' | 'conflict2') {
  const label = keepFile === 'conflict1' ? version.conflict1 : version.conflict2
  showConfirm(
    '保留文件确认',
    `确定保留 "${label}" 吗？\n另一个冲突文件将被删除，保留的文件将恢复为原始文件名。`,
    () => doResolveConflict(version.id, keepFile)
  )
}

async function doResolveConflict(versionId: string, keepFile: 'conflict1' | 'conflict2') {
  if (!props.taskId) return
  actionLoading.value = 'resolve'
  try {
    await resolveBisyncConflict(props.taskId, versionId, keepFile)
    await loadLstFiles()
  } catch (error) {
    console.error('Failed to resolve conflict:', error)
  } finally {
    actionLoading.value = null
  }
}

function updateOption<K extends keyof BisyncOptions>(key: K, value: BisyncOptions[K]) {
  localBisyncOptions.value = {
    ...localBisyncOptions.value,
    [key]: value,
  }
}

async function saveConfig() {
  if (!props.taskId) return
  savingConfig.value = true
  try {
    await updateTask(props.taskId, {
      name: '',
      mode: 'bisync',
      sourceRemote: '',
      targetRemote: '',
      bisyncOptions: localBisyncOptions.value,
    } as any)
    emit('update:bisyncOptions', { ...localBisyncOptions.value })
  } catch (error) {
    console.error('Failed to save bisync config:', error)
  } finally {
    savingConfig.value = false
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

function formatFileSize(bytes: number | undefined | null): string {
  if (bytes === undefined || bytes === null) return ''
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  let size = bytes
  let unitIndex = 0
  while (size >= 1024 && unitIndex < units.length - 1) {
    size /= 1024
    unitIndex++
  }
  return `${size.toFixed(unitIndex > 0 ? 2 : 0)} ${units[unitIndex]}`
}

watch(
  () => props.visible,
  (newVal) => {
    if (newVal) {
      localBisyncOptions.value = { ...props.bisyncOptions }
      loadLstFiles()
    }
  }
)
</script>

<template>
  <div v-if="visible" class="modal-overlay" @click.self="emit('close')">
    <div class="modal-content" style="max-width: 800px">
      <div class="modal-header">
        <h3>Bisync 状态文件管理</h3>
        <button class="close-btn" @click="emit('close')">×</button>
      </div>

      <div class="modal-body">
        <div class="hint">管理双向同步的历史状态文件（.lst），可以回滚到之前的版本或清理旧文件</div>

        <div class="section" style="margin-top: 12px">
          <button class="primary small" :disabled="actionLoading !== null" @click="handleResyncClick">
            {{ actionLoading === 'resync' ? '处理中...' : '重新同步 (Resync)' }}
          </button>
          <span class="hint" style="margin-left: 8px">重新扫描两端差异，适用于文件状态不一致时</span>
        </div>

        <div v-if="loading" class="section" style="margin-top: 16px">
          <div class="path-empty">加载中...</div>
        </div>

        <template v-else>
          <div v-if="currentVersion" class="section current-section">
            <div class="section-title">
              <span class="badge current">当前版本 {{ formatTimestamp(currentVersion.timestamp) }}</span>
            </div>
            <div class="version-card current-card">
              <div class="version-header" style="margin-bottom: 6px; display: flex; align-items: center; gap: 8px">
                <span class="version-time">{{ formatTimestamp(currentVersion.timestamp) }}</span>
              </div>
              <div class="current-files">
                <div class="current-file-line" v-if="currentVersion.path1Lst">
                  <span class="file-name">📄 {{ currentVersion.path1Lst }}</span>
                  <span class="file-size" v-if="currentVersion.path1Size !== undefined">({{ formatFileSize(currentVersion.path1Size) }})</span>
                </div>
                <div class="current-file-line" v-if="currentVersion.path2Lst">
                  <span class="file-name">📄 {{ currentVersion.path2Lst }}</span>
                  <span class="file-size" v-if="currentVersion.path2Size !== undefined">({{ formatFileSize(currentVersion.path2Size) }})</span>
                </div>
              </div>
            </div>
          </div>

          <div v-if="oldFormatBackups.length > 0" class="section" style="margin-top: 16px">
            <div class="section-title">
              <span class="badge old">上一版本备份 ({{ oldFormatBackups.length }})</span>
            </div>
            <div class="backup-table-wrap">
              <table class="backup-table">
                <thead>
                  <tr>
                    <th>时间</th>
                    <th>path1 文件</th>
                    <th>path2 文件</th>
                    <th class="col-actions">操作</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="version in oldFormatBackups" :key="version.id">
                    <td class="col-time">{{ formatTimestamp(version.timestamp) }}</td>
                    <td class="col-file">
                      <div class="file-cell">
                        <span class="file-name">📄 {{ version.path1Lst }}</span>
                        <span class="file-size" v-if="version.path1Size !== undefined">({{ formatFileSize(version.path1Size) }})</span>
                      </div>
                    </td>
                    <td class="col-file">
                      <div class="file-cell" v-if="version.path2Lst">
                        <span class="file-name">📄 {{ version.path2Lst }}</span>
                        <span class="file-size" v-if="version.path2Size !== undefined">({{ formatFileSize(version.path2Size) }})</span>
                      </div>
                    </td>
                    <td class="col-actions">
                      <div class="action-btns">
                        <button
                          class="ghost small"
                          :disabled="actionLoading !== null"
                          @click="handleViewContent(version)"
                        >查看</button>
                        <button
                          class="ghost small"
                          :disabled="actionLoading !== null"
                          @click="handleRollbackClick(version)"
                        >恢复</button>
                        <button
                          class="ghost small danger"
                          :disabled="actionLoading !== null"
                          @click="handleDeleteClick(version)"
                        >删除</button>
                      </div>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>

          <div class="section" style="margin-top: 16px">
            <div class="section-title">
              <span class="badge conflict">冲突文件 ({{ conflictVersions.length }})</span>
            </div>
            <div v-if="conflictVersions.length > 0">
              <div
                v-for="version in conflictVersions"
                :key="version.id"
                class="conflict-pair"
              >
                <div class="conflict-files-title">⚠️ {{ version.id }}</div>
                <div class="conflict-comparison">
                  <div class="conflict-card" v-if="version.conflict1">
                    <div class="conflict-card-header">冲突版本 1</div>
                    <div class="conflict-card-body">
                      <div class="conflict-info-row">
                        <span class="info-label">文件</span>
                        <span class="info-value">{{ version.conflict1 }}</span>
                      </div>
                      <div class="conflict-info-row">
                        <span class="info-label">大小</span>
                        <span class="info-value">{{ formatFileSize(version.conflict1Size) }}</span>
                      </div>
                      <div class="conflict-info-row">
                        <span class="info-label">修改时间</span>
                        <span class="info-value">{{ version.conflict1Mtime ? formatTimestamp(version.conflict1Mtime) : '--' }}</span>
                      </div>
                    </div>
                    <div class="conflict-card-actions">
                      <button
                        class="ghost small"
                        :disabled="actionLoading !== null"
                        @click="handleViewContent(version)"
                      >查看</button>
                      <button
                        class="primary small"
                        :disabled="actionLoading !== null"
                        @click="handleResolveConflict(version, 'conflict1')"
                      >保留此版本</button>
                    </div>
                  </div>
                  <div class="conflict-card" v-if="version.conflict2">
                    <div class="conflict-card-header">冲突版本 2</div>
                    <div class="conflict-card-body">
                      <div class="conflict-info-row">
                        <span class="info-label">文件</span>
                        <span class="info-value">{{ version.conflict2 }}</span>
                      </div>
                      <div class="conflict-info-row">
                        <span class="info-label">大小</span>
                        <span class="info-value">{{ formatFileSize(version.conflict2Size) }}</span>
                      </div>
                      <div class="conflict-info-row">
                        <span class="info-label">修改时间</span>
                        <span class="info-value">{{ version.conflict2Mtime ? formatTimestamp(version.conflict2Mtime) : '--' }}</span>
                      </div>
                    </div>
                    <div class="conflict-card-actions">
                      <button
                        class="ghost small"
                        :disabled="actionLoading !== null"
                        @click="handleViewContent(version)"
                      >查看</button>
                      <button
                        class="primary small"
                        :disabled="actionLoading !== null"
                        @click="handleResolveConflict(version, 'conflict2')"
                      >保留此版本</button>
                    </div>
                  </div>
                </div>
              </div>
            </div>
            <div v-else class="path-empty" style="padding: 12px">暂无冲突文件</div>
          </div>

          <div v-if="bakFormatBackups.length > 0" class="section" style="margin-top: 16px">
            <div class="section-title">
              <span class="badge backup">备份列表 ({{ bakFormatBackups.length }})</span>
            </div>
            <div class="backup-table-wrap">
              <table class="backup-table">
                <thead>
                  <tr>
                    <th>时间</th>
                    <th>path1 文件</th>
                    <th>path2 文件</th>
                    <th class="col-actions">操作</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="version in bakFormatBackups" :key="version.id">
                    <td class="col-time">{{ formatTimestamp(version.timestamp) }}</td>
                    <td class="col-file">
                      <div class="file-cell">
                        <span class="file-name">📄 {{ version.path1Lst }}</span>
                        <span class="file-size" v-if="version.path1Size !== undefined">({{ formatFileSize(version.path1Size) }})</span>
                      </div>
                    </td>
                    <td class="col-file">
                      <div class="file-cell" v-if="version.path2Lst">
                        <span class="file-name">📄 {{ version.path2Lst }}</span>
                        <span class="file-size" v-if="version.path2Size !== undefined">({{ formatFileSize(version.path2Size) }})</span>
                      </div>
                    </td>
                    <td class="col-actions">
                      <div class="action-btns">
                        <button
                          class="ghost small"
                          :disabled="actionLoading !== null"
                          @click="handleViewContent(version)"
                        >查看</button>
                        <button
                          class="ghost small"
                          :disabled="actionLoading !== null"
                          @click="handleRollbackClick(version)"
                        >恢复</button>
                        <button
                          class="ghost small danger"
                          :disabled="actionLoading !== null"
                          @click="handleDeleteClick(version)"
                        >删除</button>
                      </div>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>

          <!-- Bisync 配置 -->
          <div class="section config-section" style="margin-top: 16px">
            <div class="section-title">Bisync 配置</div>
            <div class="config-grid">
              <div class="detail-item">
                <label>比较方式 (Compare)</label>
                <select
                  :value="localBisyncOptions.compare"
                  @change="updateOption('compare', ($event.target as HTMLSelectElement).value || undefined)"
                >
                  <option value="">默认</option>
                  <option value="size">仅检查大小</option>
                  <option value="modtime">检查修改时间</option>
                  <option value="checksum">检查校验和</option>
                </select>
              </div>

              <div class="detail-item">
                <label>最大删除比例 (Max Delete)</label>
                <input
                  type="text"
                  :value="localBisyncOptions.maxDelete"
                  @input="updateOption('maxDelete', ($event.target as HTMLInputElement).value || undefined)"
                  placeholder="例如: 25"
                />
              </div>

              <div class="detail-item check-item">
                <label class="inline-label">
                  <input
                    :checked="localBisyncOptions.checkAccess"
                    type="checkbox"
                    @change="updateOption('checkAccess', ($event.target as HTMLInputElement).checked)"
                  />
                  <span>检查访问权限 (Check Access)</span>
                </label>
              </div>

              <div class="detail-item">
              <label>冲突解决策略 (Conflict Resolve)</label>
              <select
                :value="localBisyncOptions.conflictResolve"
                @change="updateOption('conflictResolve', ($event.target as HTMLSelectElement).value || undefined)"
              >
                <option value="">默认（不自动解决，双方均改名保留）</option>
                <option value="path1">优先源端</option>
                <option value="path2">优先目标端</option>
                <option value="newer">优先较新文件</option>
                <option value="older">优先较旧文件</option>
              </select>
              </div>

              <div class="detail-item">
              <label>冲突处理 (Conflict Loser)</label>
              <select
                :value="localBisyncOptions.conflictLoser"
                @change="updateOption('conflictLoser', ($event.target as HTMLSelectElement).value || undefined)"
              >
                <option value="">默认（双方自动改名，格式：filename.conflict.1.ext）</option>
                <option value="backup">备份冲突文件</option>
                <option value="delete">删除冲突文件</option>
              </select>
              </div>

              <div class="detail-item">
                <label>历史备份数量</label>
                <input
                  type="number"
                  :value="localBisyncOptions.lstBackupCount || 5"
                  min="1"
                  max="50"
                  @input="updateOption('lstBackupCount', parseInt(($event.target as HTMLInputElement).value) || 5)"
                  placeholder="默认: 5"
                />
              </div>
            </div>
            <div class="config-save">
              <button class="primary small" :disabled="savingConfig" @click="saveConfig">
                {{ savingConfig ? '保存中...' : '保存配置' }}
              </button>
            </div>
          </div>

          <div v-if="!currentVersion && oldFormatBackups.length === 0 && bakFormatBackups.length === 0 && conflictVersions.length === 0" class="section" style="margin-top: 16px">
            <div class="path-empty">暂无状态文件</div>
          </div>
        </template>
      </div>

      <div class="modal-footer">
        <button class="ghost" @click="emit('close')">关闭</button>
      </div>
    </div>

    <div v-if="confirmState.show" class="modal-overlay" @click.self="closeConfirm">
      <div class="modal-content confirm-dialog">
        <div class="modal-header">
          <h3>{{ confirmState.title }}</h3>
          <button class="close-btn" @click="closeConfirm">×</button>
        </div>
        <div class="modal-body">
          <pre class="confirm-message">{{ confirmState.message }}</pre>
        </div>
        <div class="modal-footer">
          <button class="ghost" @click="closeConfirm">取消</button>
          <button class="primary danger" @click="() => { closeConfirm(); confirmState.onConfirm(); }">确定</button>
        </div>
      </div>
    </div>

    <div v-if="contentModal.show" class="modal-overlay" @click.self="closeContentModal">
      <div class="modal-content content-dialog">
        <div class="modal-header">
          <h3>文件内容 - {{ contentModal.title }}</h3>
          <button class="close-btn" @click="closeContentModal">×</button>
        </div>
        <div class="modal-body">
          <div v-if="contentModal.loading" class="path-empty">加载中...</div>
          <pre v-else class="content-pre">{{ contentModal.content }}</pre>
        </div>
        <div class="modal-footer">
          <button class="ghost" @click="closeContentModal">关闭</button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.modal-body {
  display: flex;
  flex-direction: column;
}

.hint {
  color: var(--muted, #94a3b8);
  font-size: 13px;
  line-height: 1.5;
}

.section {
  margin-top: 8px;
}

.section-title {
  margin-bottom: 8px;
  font-size: 13px;
  font-weight: 600;
  display: flex;
  align-items: center;
  gap: 8px;
}

.badge {
  display: inline-block;
  padding: 2px 10px;
  border-radius: 4px;
  font-size: 11px;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.3px;
}

.badge.current {
  background: rgba(34, 197, 94, 0.18);
  color: #22c55e;
}

.badge.backup {
  background: rgba(59, 130, 246, 0.15);
  color: #60a5fa;
}

.badge.old {
  background: rgba(168, 85, 247, 0.15);
  color: #a855f7;
}

.badge.conflict {
  background: rgba(239, 68, 68, 0.15);
  color: #ef4444;
  animation: pulse 2s ease-in-out infinite;
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.7; }
}

.current-section {
  padding: 12px;
  background: rgba(34, 197, 94, 0.04);
  border: 1px solid rgba(34, 197, 94, 0.2);
  border-radius: 8px;
}

.current-card {
  background: transparent !important;
  border: none !important;
  padding: 4px 0 !important;
}

.current-files {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.current-file-line {
  display: flex;
  align-items: baseline;
  gap: 6px;
  flex-wrap: wrap;
}

.current-file-line .file-name {
  font-size: 12px;
  font-family: monospace;
  word-break: break-all;
  color: var(--text);
}

.current-file-line .file-size {
  font-size: 11px;
  color: #888;
  white-space: nowrap;
}

.version-card {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  padding: 10px 14px;
  border: 1px solid var(--border);
  border-radius: 6px;
  background: var(--surface);
  margin-bottom: 6px;
  gap: 12px;
}

.version-card:last-child {
  margin-bottom: 0;
}

.version-card.is-conflict {
  border-color: rgba(239, 68, 68, 0.3);
  background: rgba(239, 68, 68, 0.04);
}

.version-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 6px;
}

.version-time {
  font-size: 13px;
  font-weight: 500;
  color: var(--text);
}

.version-body {
  flex: 1;
  min-width: 0;
}

.version-details {
  font-size: 12px;
  color: var(--muted, #94a3b8);
  font-family: monospace;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.file-line {
  display: flex;
  align-items: baseline;
  gap: 6px;
  flex-wrap: wrap;
}

.file-name {
  word-break: break-all;
}

.file-size {
  font-size: 11px;
  color: #888;
  white-space: nowrap;
}

.version-actions {
  display: flex;
  gap: 6px;
  flex-shrink: 0;
  align-items: center;
  margin-top: 4px;
}

.conflict-line {
  color: #ef4444;
  font-size: 11px;
}

.version-list {
  max-height: 360px;
  overflow-y: auto;
}

.backup-table-wrap {
  max-height: 400px;
  overflow-y: auto;
  border: 1px solid var(--border);
  border-radius: 6px;
}

.backup-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 12px;
}

.backup-table th {
  position: sticky;
  top: 0;
  background: var(--surface);
  padding: 8px 10px;
  text-align: left;
  font-weight: 600;
  color: var(--text);
  border-bottom: 1px solid var(--border);
  white-space: nowrap;
}

.backup-table td {
  padding: 10px;
  border-bottom: 1px solid var(--border);
  vertical-align: middle;
}

.backup-table tr:last-child td {
  border-bottom: none;
}

.col-time {
  white-space: nowrap;
  min-width: 130px;
  font-weight: 500;
  color: var(--text);
}

.col-file {
  min-width: 140px;
}

.col-actions {
  width: 1%;
  white-space: nowrap;
}

.file-cell {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.file-cell .file-name {
  font-size: 11px;
  font-family: monospace;
  color: var(--text);
}

.file-cell .file-size {
  font-size: 10px;
}

.action-btns {
  display: flex;
  gap: 4px;
  align-items: center;
}

.ghost.small.danger {
  color: #ef4444;
}

.ghost.small.warn {
  color: #f59e0b;
}

.confirm-dialog {
  max-width: 420px;
}

.content-dialog {
  max-width: 700px;
  max-height: 80vh;
}

.content-pre {
  white-space: pre-wrap;
  word-break: break-all;
  font-family: monospace;
  font-size: 11px;
  line-height: 1.5;
  max-height: 50vh;
  overflow-y: auto;
  padding: 12px;
  background: var(--bg);
  border: 1px solid var(--border);
  border-radius: 4px;
  color: var(--text);
  margin: 0;
}

.confirm-message {
  white-space: pre-wrap;
  font-family: inherit;
  font-size: 13px;
  line-height: 1.6;
  margin: 0;
  color: var(--text);
}

.path-empty {
  padding: 32px;
  text-align: center;
  color: var(--muted, #94a3b8);
  font-size: 13px;
}

/* Config section styles */
.config-section {
  border: 1px solid var(--border);
  border-radius: 8px;
  padding: 12px 14px;
}

.config-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px;
}

.config-grid input,
.config-grid select {
  width: 100%;
  padding: 10px 12px;
  border: 1px solid var(--border);
  border-radius: 8px;
  background: var(--surface);
  color: var(--text);
  box-sizing: border-box;
}

body.light .config-grid input,
body.light .config-grid select {
  background: #fff;
  border-color: #ddd;
  color: #333;
}

.config-grid label {
  display: block;
  margin-bottom: 6px;
  font-size: 13px;
  color: var(--muted);
}

.config-grid .detail-item {
  display: flex;
  flex-direction: column;
}

.config-grid .check-item {
  justify-content: flex-end;
}

.config-grid .inline-label {
  display: flex !important;
  align-items: center;
  gap: 8px;
  margin: 0 0 6px 0;
}

.config-grid .inline-label input[type="checkbox"] {
  width: 16px;
  height: 16px;
}

.config-save {
  margin-top: 12px;
  display: flex;
  justify-content: flex-end;
}

@media (max-width: 600px) {
  .config-grid {
    grid-template-columns: 1fr;
  }
}

/* Conflict comparison styles */
.conflict-pair {
  margin-bottom: 12px;
  border: 1px solid rgba(239, 68, 68, 0.3);
  border-radius: 8px;
  padding: 10px;
  background: rgba(239, 68, 68, 0.03);
}

.conflict-files-title {
  font-size: 12px;
  font-family: monospace;
  font-weight: 600;
  margin-bottom: 8px;
  padding: 4px 8px;
  background: rgba(239, 68, 68, 0.08);
  border-radius: 4px;
  color: var(--text);
  word-break: break-all;
}

.conflict-comparison {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 10px;
}

.conflict-card {
  border: 1px solid var(--border);
  border-radius: 8px;
  background: var(--surface);
  display: flex;
  flex-direction: column;
}

.conflict-card-header {
  font-size: 11px;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.3px;
  color: var(--muted);
  padding: 8px 10px;
  border-bottom: 1px solid var(--border);
}

.conflict-card-body {
  padding: 8px 10px;
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.conflict-info-row {
  display: flex;
  gap: 6px;
  font-size: 12px;
  align-items: baseline;
}

.conflict-info-row .info-label {
  color: var(--muted);
  flex-shrink: 0;
  min-width: 4em;
}

.conflict-info-row .info-value {
  color: var(--text);
  word-break: break-all;
  font-family: monospace;
  font-size: 11px;
}

.conflict-card-actions {
  padding: 8px 10px;
  border-top: 1px solid var(--border);
  display: flex;
  justify-content: center;
}

@media (max-width: 600px) {
  .conflict-comparison {
    grid-template-columns: 1fr;
  }
}
</style>

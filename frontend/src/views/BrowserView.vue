<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { t } from '../i18n'
import AddRemoteModal from '../components/AddRemoteModal.vue'
import EditDescModal from '../components/EditDescModal.vue'
import * as api from '../api'
import type { RemoteTestState, FileItem } from '../types'

defineProps<{
  version: string
}>()

const emit = defineEmits<{
  navigate: [subview: string]
}>()

const remotes = ref<string[]>([])
const browserFs = ref('')
const browserPath = ref('')
const browserItems = ref<FileItem[]>([])
const browserError = ref('')
const testState = ref<Record<string, RemoteTestState>>({})
const remoteMenu = ref('')
const confirmModal = ref<{ show: boolean; title: string; message: string; onConfirm: () => void }>({
  show: false,
  title: '',
  message: '',
  onConfirm: () => {}
})

const descriptions = ref<Record<string, string>>(
  JSON.parse(localStorage.getItem('remoteDescriptions') || '{}')
)

// Clipboard for copy/move operations
const clipboardItem = ref<FileItem | null>(null)
const clipboardAction = ref<'copy' | 'move' | null>(null)
// Remember the source remote of the clipboard item for cross-remote ops
const clipboardSrcFs = ref<string>('')

// Context menu state
const contextMenu = ref({
  show: false,
  x: 0,
  y: 0,
  item: null as FileItem | null
})

// Delete confirmation modal
const showDeleteConfirm = ref(false)
const deletingItem = ref<FileItem | null>(null)

// Rename modal
const showRenameInput = ref(false)
const renamingItem = ref<FileItem | null>(null)
const renameInput = ref('')

const remoteOrder = ref<string[]>(JSON.parse(localStorage.getItem('remoteOrder') || '[]'))
const draggedRemote = ref('')

function getOrderedRemotes() {
  const remotesList = remotes.value
  const order = remoteOrder.value
  if (!order.length) return remotesList
  const ordered = order.filter(r => remotesList.includes(r))
  const newOnes = remotesList.filter(r => !order.includes(r))
  return [...ordered, ...newOnes]
}

function saveRemoteOrder() {
  remoteOrder.value = remotes.value
  localStorage.setItem('remoteOrder', JSON.stringify(remoteOrder.value))
}

function onDragStart(name: string) {
  draggedRemote.value = name
}

function onDragOver(e: DragEvent, _name: string) {
  e.preventDefault()
}

function onDrop(e: DragEvent, targetName: string) {
  e.preventDefault()
  if (draggedRemote.value === targetName) return
  
  const list = [...remotes.value]
  const fromIndex = list.indexOf(draggedRemote.value)
  const toIndex = list.indexOf(targetName)
  
  if (fromIndex !== -1 && toIndex !== -1) {
    list.splice(fromIndex, 1)
    list.splice(toIndex, 0, draggedRemote.value)
    remotes.value = list
    saveRemoteOrder()
  }
  
  draggedRemote.value = ''
}

function onDragEnd() {
  draggedRemote.value = ''
}

// 检查路径是否包含路径穿越风险
function hasPathTraversal(path: string): boolean {
  // 检查 .. 或 . 路径
  const normalized = path.replace(/\\/g, '/')
  if (normalized.includes('..') || normalized.includes('./') || normalized === '.') {
    return true
  }
  // 检查 URL 编码的 ..
  if (normalized.includes('%2e%2e') || normalized.includes('%2e.')) {
    return true
  }
  return false
}

// 验证路径安全性
function validatePath(path: string, name: string = '路径'): boolean {
  if (hasPathTraversal(path)) {
    alert(`${name}包含非法字符，存在路径穿越风险`)
    return false
  }
  return true
}

const showAddRemote = ref(false)
const isEditMode = ref(false)
const editRemoteName = ref('')
const showEditDesc = ref(false)
const editDescRemote = ref('')
const addRemoteModal = ref<InstanceType<typeof AddRemoteModal> | null>(null)

const subview = ref('explorer')

// Helpers for post-op stabilization: refresh until target entries disappear
function delay(ms: number) { return new Promise(res => setTimeout(res, ms)) }
function hasItemByName(name: string): boolean {
  return browserItems.value.some(it => it.Name === name)
}
async function refreshUntilGone(names: string[], timeoutMs = 5000, intervalMs = 400) {
  const deadline = Date.now() + timeoutMs
  // small initial delay to let backend settle
  await delay(200)
  while (Date.now() < deadline) {
    await refreshBrowser()
    const stillPresent = names.some(n => hasItemByName(n))
    if (!stillPresent) return
    await delay(intervalMs)
  }
}

const breadcrumbs = computed(() => {
  const parts = browserPath.value.split('/').filter(Boolean)
  const crumbs = [{ name: browserFs.value + ':', path: '' }]
  let current = ''
  for (const p of parts) {
    current += '/' + p
    crumbs.push({ name: p, path: current })
  }
  return crumbs
})

onMounted(async () => {
  await loadRemotes()
  // Close context menu on click outside
  document.addEventListener('click', () => {
    contextMenu.value.show = false
  })
})

async function loadRemotes() {
  try {
    const data = await api.listRemotes()
    remotes.value = data.remotes || []
    
    // Always use the first remote in sorted order as default after page load
    const orderedRemotes = getOrderedRemotes()
    if (orderedRemotes.length > 0) {
      await openRemote(orderedRemotes[0])
    }
    
    emit('navigate', 'explorer')
  } catch (e) {
    browserError.value = (e as Error).message
  }
}

async function openRemote(name: string) {
  browserFs.value = name
  browserPath.value = ''
  await refreshBrowser()
  subview.value = 'explorer'
}

async function refreshBrowser() {
  if (!browserFs.value) return
  browserError.value = ''
  try {
    const data = await api.listPath(browserFs.value, browserPath.value)
    browserItems.value = data.items || []
  } catch (e) {
    browserError.value = (e as Error).message
  }
}

function enterItem(item: FileItem) {
  if (!item.IsDir) return
  // item.Path is already the full path including remote prefix
  // e.g., "fnOS:/来自：百度网盘/测试/子文件夹"
  // Extract path without the remote prefix (e.g., "来自：百度网盘/测试/子文件夹")
  const remotePrefix = browserFs.value + ':'
  let newPath = item.Path
  if (newPath.startsWith(remotePrefix)) {
    newPath = newPath.substring(remotePrefix.length)
  }
  // Remove leading slash
  newPath = newPath.replace(/^\/+/, '')
  browserPath.value = newPath
  refreshBrowser()
}

function formatTime(time: string) {
  if (!time) return '-'
  try {
    return new Date(time).toLocaleString('zh-CN', {
      year: 'numeric', month: '2-digit', day: '2-digit',
      hour: '2-digit', minute: '2-digit'
    })
  } catch {
    return time
  }
}

function formatSize(size: string) {
  if (!size || size === '-') return '-'
  const bytes = parseInt(size)
  if (isNaN(bytes)) return size
  if (bytes < 1024) return bytes + ' B'
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' K'
  if (bytes < 1024 * 1024 * 1024) return (bytes / (1024 * 1024)).toFixed(1) + ' M'
  return (bytes / (1024 * 1024 * 1024)).toFixed(1) + ' G'
}

// Context menu handlers
function showContextMenu(e: MouseEvent, item: FileItem) {
  e.preventDefault()
  contextMenu.value = {
    show: true,
    x: e.clientX,
    y: e.clientY,
    item
  }
}

// Right-click on empty space shows paste option
function showBackgroundMenu(e: MouseEvent) {
  e.preventDefault()
  contextMenu.value = {
    show: true,
    x: e.clientX,
    y: e.clientY,
    item: null  // null means background click
  }
}

function closeContextMenu() {
  contextMenu.value.show = false
}

async function copyItem() {
  if (!contextMenu.value.item) return
  // item.Path is already the full path from remote root (e.g., "备份/file.txt")
  // NO need to add browserPath
  clipboardItem.value = {
    ...contextMenu.value.item,
    Path: contextMenu.value.item.Path
  }
  clipboardAction.value = 'copy'
  // 源 FS 记录：当前浏览的 remote 名称
  clipboardSrcFs.value = browserFs.value
  closeContextMenu()
}

async function moveItem() {
  if (!contextMenu.value.item) return
  // item.Path is already the full path from remote root
  clipboardItem.value = {
    ...contextMenu.value.item,
    Path: contextMenu.value.item.Path
  }
  clipboardAction.value = 'move'
  clipboardSrcFs.value = browserFs.value
  closeContextMenu()
}

async function pasteItem() {
  if (!clipboardItem.value || !clipboardAction.value) {
    alert('剪贴板为空，请先复制或剪切文件')
    return
  }
  if (!browserFs.value) {
    alert('请先选择一个存储节点')
    return
  }

  // 验证源路径
  // 规范源/目标 FS + 路径：源 FS 用 clipboardSrcFs，目标 FS 用当前浏览；路径取相对根
  const srcFs = clipboardSrcFs.value || browserFs.value
  const dstFs = browserFs.value
  const srcPath = clipboardItem.value.Path.replace(/^\/+/, '')
  const dstPath = (browserPath.value ? browserPath.value.replace(/^\/+/, '') + '/' : '') + clipboardItem.value.Name
  if (!validatePath(srcPath, '源路径')) return
  if (!validatePath(dstPath, '目标路径')) return

  try {
    const isDir = clipboardItem.value.IsDir
    const actionWasMove = clipboardAction.value === 'move'
    const oldName = clipboardItem.value.Name
    
    // For directory operations (copyDir/moveDir), use fs format: "remote:full/path"
    // For file operations (copyFile/moveFile), use fs + remote format: "remote:", "path"
    if (clipboardAction.value === 'copy') {
      if (isDir) {
        // Directory copy: copyDir(srcFs, srcPath, dstFs, dstPath)
        await api.copyDir(srcFs, srcPath, dstFs, dstPath)
      } else {
        // File copy: copyFile(srcFs, srcPath, dstFs, dstPath)
        await api.copyFile(srcFs, srcPath, dstFs, dstPath)
      }
    } else if (clipboardAction.value === 'move') {
      if (isDir) {
        // Directory move: moveDir(srcFs, srcPath, dstFs, dstPath)
        await api.moveDir(srcFs, srcPath, dstFs, dstPath)
      } else {
        // File move: moveFile(srcFs, srcPath, dstFs, dstPath)
        await api.moveFile(srcFs, srcPath, dstFs, dstPath)
      }
    }
    
    clipboardItem.value = null
    clipboardAction.value = null
    // After move/copy, do a stabilization refresh loop (especially for RC + SMB/WebDAV backends)
    if (actionWasMove) {
      await refreshUntilGone([oldName])
    } else {
      await refreshBrowser()
    }
  } catch (e) {
    alert('粘贴失败: ' + (e as Error).message)
  }
}

function startRename() {
  if (!contextMenu.value.item) return
  renamingItem.value = contextMenu.value.item
  renameInput.value = contextMenu.value.item.Name
  showRenameInput.value = true
  closeContextMenu()
}

async function confirmRename() {
  if (!renamingItem.value || !renameInput.value) return
  if (!browserFs.value) {
    alert('请先选择一个存储节点')
    return
  }
  try {
    const isDir = renamingItem.value.IsDir
    const srcPath = renamingItem.value.Path.replace(/^\/+/, '')
    // 构建目标路径: 当前目录/新名称 (确保路径格式一致)
    const dstPath = (browserPath.value ? browserPath.value.replace(/^\/+/, '') + '/' : '') + renameInput.value
    
    if (isDir) {
      // 目录移动使用 sync/move
      // moveDir(remote, srcPath, remote, dstPath) - 4个参数
      await api.moveDir(browserFs.value, srcPath, browserFs.value, dstPath)
    } else {
      // 文件移动使用 operations/movefile
      // moveFile(remote, srcPath, remote, dstPath) - 4个参数
      await api.moveFile(browserFs.value, srcPath, browserFs.value, dstPath)
    }
    
    showRenameInput.value = false
    const oldName = renamingItem.value.Name
    renamingItem.value = null
    // After rename (move), stabilize view until old name disappears
    await refreshUntilGone([oldName])
  } catch (e) {
    alert(`${t('browserView.renameFailed')}: ${(e as Error).message}`)
  }
}

function confirmDelete() {
  if (!contextMenu.value.item) return
  deletingItem.value = contextMenu.value.item
  showDeleteConfirm.value = true
  closeContextMenu()
}

async function executeDelete() {
  if (!deletingItem.value) return
  if (!browserFs.value) {
    alert(t('browserView.selectStorageFirst'))
    return
  }

  // 验证路径
  if (!validatePath(deletingItem.value.Path, t('browserView.deletePath'))) return

  try {
    if (deletingItem.value.IsDir) {
      // 目录使用 purge
      await api.purgeDir(browserFs.value, deletingItem.value.Path)
    } else {
      // 文件使用 deletefile
      await api.deleteFile(browserFs.value, deletingItem.value.Path)
    }
    showDeleteConfirm.value = false
    deletingItem.value = null
    await refreshBrowser()
  } catch (e) {
    alert(`${t('browserView.deleteFailed')}: ${(e as Error).message}`)
  }
}

async function testRemote(name: string) {
  if (testState.value[name] === 'testing') return
  testState.value[name] = 'testing'
  try {
    await api.testRemote(name)
    testState.value[name] = 'success'
  } catch {
    testState.value[name] = 'failed'
  }
  setTimeout(() => {
    testState.value[name] = 'idle'
  }, 5000)
}

function getTestText(name: string) {
  const s = testState.value[name]
  if (s === 'testing') return t('browserView.testing')
  if (s === 'success') return t('browserView.testSuccess')
  if (s === 'failed') return t('browserView.testFailed')
  return t('browserView.test')
}

async function deleteRemote(name: string) {
  confirmModal.value = {
    show: true,
    title: t('remote.deleteStorage'),
    message: t('remote.deleteStorageConfirm').replace('{name}', name),
    onConfirm: async () => {
      try {
        await api.deleteRemote(name)
        // 删除本地存储的介绍 - 同时更新localStorage和组件变量
        delete descriptions.value[name]
        localStorage.setItem('remoteDescriptions', JSON.stringify(descriptions.value))
        await loadRemotes()
      } catch (e) {
        alert((e as Error).message)
      }
    }
  }
}

function openEditDesc(name: string) {
  editDescRemote.value = name
  showEditDesc.value = true
}

function saveDesc(desc: string) {
  descriptions.value[editDescRemote.value] = desc
  localStorage.setItem('remoteDescriptions', JSON.stringify(descriptions.value))
}

function openManageStorage() {
  subview.value = 'manage-storage'
}

function openAddRemote() {
  isEditMode.value = false
  editRemoteName.value = ''
  showAddRemote.value = true
}

async function openEditRemote(name: string) {
  isEditMode.value = true
  editRemoteName.value = name
  showAddRemote.value = true
  setTimeout(async () => {
    if (addRemoteModal.value) {
      await addRemoteModal.value.loadConfig(name)
    }
  }, 100)
}
</script>

<template>
  <!-- Storage Panel -->
  <div class="card">
    <div class="card-header">
      <div style="display: flex; justify-content: space-between; align-items: center; width: 100%">
        <div>
          <div class="title">{{ t('remote.panelTitle') }}</div>
          <div class="subtitle">{{ t('remote.panelSubtitle') }}</div>
        </div>
        <div class="actions">
          <button class="ghost small" @click="openManageStorage">{{ t('remote.manageButton') }}</button>
          <button class="ghost small" @click="openAddRemote">{{ t('remote.addButton') }}</button>
        </div>
      </div>
    </div>
    <div class="tile-grid">
      <div
        v-for="name in getOrderedRemotes()"
        :key="name"
        class="tile"
        :class="{ active: browserFs === name }"
        draggable="true"
        @click="openRemote(name)"
        @dragstart="onDragStart(name)"
        @dragover="onDragOver($event, name)"
        @drop="onDrop($event, name)"
        @dragend="onDragEnd"
      >
        <div class="tile-header">
          <span class="tile-name">{{ name }}</span>
          <span class="tile-drag">⋮⋮</span>
        </div>
        <div v-if="descriptions[name]" class="tile-desc">
          {{ descriptions[name] }}
        </div>
      </div>
    </div>
  </div>

  <!-- Browser Panel -->
  <div v-if="subview === 'explorer'" class="card">
    <div class="card-header">
      <div class="title">{{ t('remote.browserTitle') }}</div>
    </div>
    <div class="pathbar">
      <template v-for="(crumb, i) in breadcrumbs" :key="crumb.path">
        <span v-if="i > 0" class="sep">/</span>
        <button
          class="crumb"
          :class="{ current: i === breadcrumbs.length - 1 }"
          @click="crumb.path !== browserPath && (browserPath = crumb.path, refreshBrowser())"
        >
          {{ crumb.name }}
        </button>
      </template>
    </div>
    <div class="list-header">
      <span class="col-name">{{ t('browserView.name') }}</span>
      <span class="col-time">{{ t('browserView.modifiedTime') }}</span>
      <span class="col-size">{{ t('browserView.size') }}</span>
    </div>
    <div class="list" @contextmenu.prevent="showBackgroundMenu($event)">
      <div
        v-for="item in browserItems"
        :key="item.Path"
        class="item"
        @click="enterItem(item)"
        @contextmenu.stop="showContextMenu($event, item)"
      >
        <div class="name">
          <span :class="item.IsDir ? 'folder' : 'icon'">{{ item.IsDir ? '📁' : '📄' }}</span>
          <span>{{ item.Name }}</span>
        </div>
        <div class="meta">
          <span class="time">{{ formatTime(item.ModTime) }}</span>
          <span class="size">{{ item.IsDir ? '-' : formatSize(item.Size) }}</span>
        </div>
      </div>
      <div v-if="browserError" class="item" style="color: #ef5350">
        {{ browserError }}
      </div>
      <div v-if="!browserItems.length && !browserError" class="empty">
        {{ t('remote.emptyDir') }}
      </div>
    </div>
  </div>

  <!-- Context Menu -->
  <div
    v-if="contextMenu.show"
    class="context-menu"
    :style="{ left: contextMenu.x + 'px', top: contextMenu.y + 'px' }"
  >
    <!-- Item context menu -->
    <template v-if="contextMenu.item">
      <button @click="copyItem">{{ t('browserView.copy') }}</button>
      <button @click="moveItem">{{ t('browserView.move') }}</button>
      <button @click="pasteItem" :disabled="!clipboardItem">{{ t('browserView.paste') }}</button>
      <button @click="startRename">{{ t('browserView.rename') }}</button>
      <button class="danger" @click="confirmDelete">{{ t('browserView.delete') }}</button>
    </template>
    <!-- Background context menu (empty area) -->
    <template v-else>
      <button @click="pasteItem" :disabled="!clipboardItem">{{ t('browserView.pasteToCurrent') }}</button>
    </template>
  </div>

  <!-- Delete Confirmation Modal -->
  <div v-if="showDeleteConfirm" class="modal-overlay" @click.self="showDeleteConfirm = false">
    <div class="modal delete-modal">
      <div class="modal-header">
        <h2>{{ t('browserView.confirmDelete') }}</h2>
        <button class="modal-close" @click="showDeleteConfirm = false">&times;</button>
      </div>
      <div class="modal-content">
        <p>{{ t('browserView.confirmDeleteText') }} <strong>{{ deletingItem?.Name }}</strong> ?</p>
        <p class="warning">{{ t('browserView.irreversible') }}</p>
      </div>
      <div class="modal-footer">
        <button class="ghost" @click="showDeleteConfirm = false">{{ t('common.cancel') }}</button>
        <button class="danger-btn" @click="executeDelete">{{ t('browserView.confirmDeleteAction') }}</button>
      </div>
    </div>
  </div>

  <!-- Rename Modal -->
  <div v-if="showRenameInput" class="modal-overlay" @click.self="showRenameInput = false">
    <div class="modal">
      <div class="modal-header">
        <h2>{{ t('browserView.rename') }}</h2>
        <button class="modal-close" @click="showRenameInput = false">&times;</button>
      </div>
      <div class="modal-content">
        <div class="field-item">
          <label>{{ t('browserView.newName') }}</label>
          <input v-model="renameInput" @keyup.enter="confirmRename" />
        </div>
      </div>
      <div class="modal-footer">
        <button class="ghost" @click="showRenameInput = false">{{ t('common.cancel') }}</button>
        <button class="primary" @click="confirmRename">{{ t('browserView.confirm') }}</button>
      </div>
    </div>
  </div>

  <!-- Manage Storage Subview -->
  <div v-if="subview === 'manage-storage'" class="card">
    <div class="card-header">
      <div style="display: flex; justify-content: space-between; align-items: center; width: 100%">
        <div class="title">{{ t('remote.manageTitle') }}</div>
        <div class="actions">
          <button class="ghost small" @click="subview = 'explorer'">{{ t('remote.backButton') }}</button>
          <button class="ghost small" @click="openAddRemote">{{ t('remote.addButton') }}</button>
        </div>
      </div>
    </div>
    <div class="list">
      <div v-for="name in remotes" :key="name" class="item" @click="openRemote(name)">
        <div class="name">
          <strong>{{ name }}</strong>
        </div>
        <div class="actions" @click.stop>
          <button class="ghost small" @click="openEditRemote(name)">{{ t('remote.editConfig') }}</button>
          <button class="ghost small" @click="openEditDesc(name)">{{ t('remote.editDesc') }}</button>
          <button
            class="ghost small"
            @click="testRemote(name)"
            :disabled="testState[name] === 'testing'"
          >
            {{ getTestText(name) }}
          </button>
          <div class="menu-area">
            <button
              class="menu-btn"
              @click="remoteMenu = remoteMenu === name ? '' : name"
            >
              ⋮
            </button>
            <div v-if="remoteMenu === name" class="menu-pop">
              <button class="danger-text" @click="deleteRemote(name); remoteMenu = ''">
                🗑️
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>

  <!-- 确认删除弹窗 -->
  <div v-if="confirmModal.show" class="modal-overlay" @click.self="confirmModal.show = false">
    <div class="modal-content confirm-modal">
      <div class="modal-header">
        <h3>{{ confirmModal.title }}</h3>
        <button class="close-btn" @click="confirmModal.show = false">×</button>
      </div>
      <div class="modal-body">
        <p>{{ confirmModal.message }}</p>
      </div>
      <div class="modal-footer">
        <button class="ghost" @click="confirmModal.show = false">{{ t('common.cancel') }}</button>
        <button class="primary danger" @click="() => { confirmModal.onConfirm(); confirmModal.show = false }">{{ t('browserView.confirm') }}</button>
      </div>
    </div>
  </div>

  <!-- Modals -->
  <AddRemoteModal
    ref="addRemoteModal"
    :show="showAddRemote"
    :edit-mode="isEditMode"
    :edit-name="editRemoteName"
    @close="showAddRemote = false"
    @success="loadRemotes"
  />
  <EditDescModal
    :show="showEditDesc"
    :remote-name="editDescRemote"
    :description="descriptions[editDescRemote] || ''"
    @close="showEditDesc = false"
    @save="saveDesc"
  />
</template>

<style scoped>
.list-header {
  display: flex;
  justify-content: space-between;
  padding: 10px 20px;
  background: var(--surface);
  font-size: 12px;
  color: var(--muted);
  border-bottom: 1px solid var(--border);
}

body.light .list-header {
  background: var(--surface);
  color: var(--muted);
  border-bottom: 1px solid var(--border);
}

.col-name { flex: 1; }
.col-time { width: 180px; text-align: right; }
.col-size { width: 100px; text-align: right; }

.item .name {
  color: var(--text);
}

body.light .item .name {
  color: var(--text);
}

body.light .item .meta .time,
body.light .item .meta .size {
  color: #888;
}

.item .meta {
  display: flex;
  gap: 24px;
}

.item .meta .time {
  width: 180px;
  text-align: right;
  color: var(--muted);
  font-size: 13px;
}

.item .meta .size {
  width: 100px;
  text-align: right;
  color: var(--muted);
  font-size: 13px;
}

.context-menu {
  position: fixed;
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: 10px;
  overflow: hidden;
  z-index: 1000;
  box-shadow: 0 8px 24px rgba(0,0,0,0.5);
  min-width: 140px;
}

body.light .context-menu {
  background: var(--surface);
  border-color: var(--border);
}

.context-menu button {
  display: block;
  width: 100%;
  text-align: left;
  padding: 12px 16px;
  background: transparent;
  border: none;
  color: #ccc;
  cursor: pointer;
  font-size: 14px;
}

body.light .context-menu button {
  color: #333;
}

.context-menu button:hover {
  background: #252525;
}

body.light .context-menu button:hover {
  background: #f5f5f5;
}

.context-menu button:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.context-menu button.danger {
  color: #ef5350;
}

.context-menu button.danger:hover {
  background: #3d2020;
}

.modal-content p {
  margin-bottom: 12px;
  color: #ccc;
}

body.light .modal-content p {
  color: #333;
}

.modal-content .warning {
  color: #ef5350;
  font-size: 13px;
}

.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
  margin-top: 20px;
  padding-top: 16px;
  border-top: 1px solid #333;
}

body.light .modal-footer {
  border-color: #eee;
}

.danger-btn {
  background: #d32f2f;
  color: #fff;
  padding: 8px 16px;
  border-radius: 8px;
  border: none;
  cursor: pointer;
  font-size: 13px;
  font-weight: 500;
}

.danger-btn:hover {
  background: #b71c1c;
}

.menu-area {
  position: relative;
}
</style>

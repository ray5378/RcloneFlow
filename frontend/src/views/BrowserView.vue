<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
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

const descriptions = ref<Record<string, string>>(
  JSON.parse(localStorage.getItem('remoteDescriptions') || '{}')
)

// Clipboard for copy/move operations
const clipboardItem = ref<FileItem | null>(null)
const clipboardAction = ref<'copy' | 'move' | null>(null)

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

// Copy/Move target modal
const showTargetPicker = ref(false)
const targetRemote = ref('')
const targetPath = ref('')
const pendingAction = ref<'copy' | 'move' | null>(null)

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

const showAddRemote = ref(false)
const isEditMode = ref(false)
const editRemoteName = ref('')
const showEditDesc = ref(false)
const editDescRemote = ref('')
const addRemoteModal = ref<InstanceType<typeof AddRemoteModal> | null>(null)

const subview = ref('explorer')

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
    
    if (!browserFs.value && remotes.value.length > 0) {
      await openRemote(remotes.value[0])
    }
    
    emit('navigate', 'explorer')
  } catch (e) {
    browserError.value = (e as Error).message
  }
}

async function openRemote(name: string) {
  browserFs.value = name
  browserPath.value = ''
  targetRemote.value = name
  targetPath.value = ''
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
  browserPath.value = item.Path
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

function closeContextMenu() {
  contextMenu.value.show = false
}

async function copyItem() {
  if (!contextMenu.value.item) return
  clipboardItem.value = contextMenu.value.item
  clipboardAction.value = 'copy'
  closeContextMenu()
}

async function moveItem() {
  if (!contextMenu.value.item) return
  pendingAction.value = 'move'
  targetRemote.value = browserFs.value
  targetPath.value = browserPath.value
  showTargetPicker.value = true
  closeContextMenu()
}

function pasteItem() {
  if (!clipboardItem.value || !clipboardAction.value) return
  if (clipboardAction.value === 'copy') {
    pendingAction.value = 'copy'
  } else {
    pendingAction.value = 'move'
  }
  targetRemote.value = browserFs.value
  targetPath.value = browserPath.value
  showTargetPicker.value = true
}

async function executePaste() {
  if (!clipboardItem.value || !pendingAction.value) return
  try {
    const srcPath = clipboardItem.value.Path
    const dstPath = targetPath.value ? targetPath.value + '/' + clipboardItem.value.Name : clipboardItem.value.Name
    
    if (pendingAction.value === 'copy') {
      await api.copyFile(browserFs.value, srcPath, targetRemote.value, dstPath)
    } else if (pendingAction.value === 'move') {
      await api.moveFile(browserFs.value, srcPath, targetRemote.value, dstPath)
    }
    
    showTargetPicker.value = false
    clipboardItem.value = null
    clipboardAction.value = null
    pendingAction.value = null
    await refreshBrowser()
  } catch (e) {
    alert((e as Error).message)
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
  try {
    const srcPath = renamingItem.value.Path
    const dstPath = browserPath.value ? browserPath.value + '/' + renameInput.value : renameInput.value
    await api.moveFile(browserFs.value, srcPath, browserFs.value, dstPath)
    showRenameInput.value = false
    renamingItem.value = null
    await refreshBrowser()
  } catch (e) {
    alert((e as Error).message)
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
  try {
    await api.deleteFile(browserFs.value, deletingItem.value.Path)
    showDeleteConfirm.value = false
    deletingItem.value = null
    await refreshBrowser()
  } catch (e) {
    alert('删除失败: ' + (e as Error).message)
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
  if (s === 'testing') return '测试中...'
  if (s === 'success') return '成功 ✓'
  if (s === 'failed') return '失败 ✗'
  return '测试'
}

async function deleteRemote(name: string) {
  if (!confirm(`确定删除存储 "${name}"？`)) return
  try {
    await api.deleteRemote(name)
    await loadRemotes()
  } catch (e) {
    alert((e as Error).message)
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
          <div class="title">存储节点</div>
          <div class="subtitle">选择浏览存储文件</div>
        </div>
        <div class="actions">
          <button class="ghost small" @click="openManageStorage">管理存储</button>
          <button class="ghost small" @click="openAddRemote">添加存储</button>
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
      <div class="title">文件浏览</div>
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
      <span class="col-name">名称</span>
      <span class="col-time">修改时间</span>
      <span class="col-size">大小</span>
    </div>
    <div class="list">
      <div
        v-for="item in browserItems"
        :key="item.Path"
        class="item"
        @click="enterItem(item)"
        @contextmenu="showContextMenu($event, item)"
      >
        <div class="name">
          <span :class="item.IsDir ? 'folder' : 'icon'">{{ item.IsDir ? '📁' : '📄' }}</span>
          <span>{{ item.Name }}</span>
        </div>
        <div class="meta">
          <span class="time">{{ formatTime(item.ModTime) }}</span>
          <span class="size">{{ item.IsDir ? '-' : item.Size }}</span>
        </div>
      </div>
      <div v-if="browserError" class="item" style="color: #ef5350">
        {{ browserError }}
      </div>
      <div v-if="!browserItems.length && !browserError" class="empty">
        空目录
      </div>
    </div>
  </div>

  <!-- Context Menu -->
  <div
    v-if="contextMenu.show"
    class="context-menu"
    :style="{ left: contextMenu.x + 'px', top: contextMenu.y + 'px' }"
  >
    <button @click="copyItem">复制</button>
    <button @click="moveItem">移动</button>
    <button @click="pasteItem" :disabled="!clipboardItem">粘贴</button>
    <button @click="startRename">重命名</button>
    <button class="danger" @click="confirmDelete">删除</button>
  </div>

  <!-- Delete Confirmation Modal -->
  <div v-if="showDeleteConfirm" class="modal-overlay" @click.self="showDeleteConfirm = false">
    <div class="modal delete-modal">
      <div class="modal-header">
        <h2>确认删除</h2>
        <button class="modal-close" @click="showDeleteConfirm = false">&times;</button>
      </div>
      <div class="modal-content">
        <p>确定要删除 <strong>{{ deletingItem?.Name }}</strong> 吗？</p>
        <p class="warning">⚠️ 此操作不可恢复！</p>
      </div>
      <div class="modal-footer">
        <button class="ghost" @click="showDeleteConfirm = false">取消</button>
        <button class="danger-btn" @click="executeDelete">确定删除</button>
      </div>
    </div>
  </div>

  <!-- Rename Modal -->
  <div v-if="showRenameInput" class="modal-overlay" @click.self="showRenameInput = false">
    <div class="modal">
      <div class="modal-header">
        <h2>重命名</h2>
        <button class="modal-close" @click="showRenameInput = false">&times;</button>
      </div>
      <div class="modal-content">
        <div class="field-item">
          <label>新名称</label>
          <input v-model="renameInput" @keyup.enter="confirmRename" />
        </div>
      </div>
      <div class="modal-footer">
        <button class="ghost" @click="showRenameInput = false">取消</button>
        <button class="primary" @click="confirmRename">确定</button>
      </div>
    </div>
  </div>

  <!-- Target Picker Modal (Copy/Move) -->
  <div v-if="showTargetPicker" class="modal-overlay" @click.self="showTargetPicker = false">
    <div class="modal">
      <div class="modal-header">
        <h2>{{ pendingAction === 'copy' ? '复制到' : '移动到' }}</h2>
        <button class="modal-close" @click="showTargetPicker = false">&times;</button>
      </div>
      <div class="modal-content">
        <div class="field-item">
          <label>目标存储</label>
          <select v-model="targetRemote">
            <option v-for="r in remotes" :key="r" :value="r">{{ r }}</option>
          </select>
        </div>
        <div class="field-item">
          <label>目标路径</label>
          <input v-model="targetPath" placeholder="留空表示根目录" />
        </div>
      </div>
      <div class="modal-footer">
        <button class="ghost" @click="showTargetPicker = false">取消</button>
        <button class="primary" @click="executePaste">确定</button>
      </div>
    </div>
  </div>

  <!-- Manage Storage Subview -->
  <div v-if="subview === 'manage-storage'" class="card">
    <div class="card-header">
      <div style="display: flex; justify-content: space-between; align-items: center; width: 100%">
        <div class="title">管理存储</div>
        <div class="actions">
          <button class="ghost small" @click="subview = 'explorer'">返回</button>
          <button class="ghost small" @click="openAddRemote">添加存储</button>
        </div>
      </div>
    </div>
    <div class="list">
      <div v-for="name in remotes" :key="name" class="item" @click="openRemote(name)">
        <div class="name">
          <strong>{{ name }}</strong>
        </div>
        <div class="actions" @click.stop>
          <button class="ghost small" @click="openEditRemote(name)">修改配置</button>
          <button class="ghost small" @click="openEditDesc(name)">自定义介绍</button>
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
              <button class="danger" @click="deleteRemote(name); remoteMenu = ''">
                删除
              </button>
            </div>
          </div>
        </div>
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
  background: #252525;
  font-size: 12px;
  color: #888;
  border-bottom: 1px solid #333;
}

.col-name { flex: 1; }
.col-time { width: 180px; text-align: right; }
.col-size { width: 100px; text-align: right; }

.item .name {
  color: #e0e0e0;
}

body.light .item .name {
  color: #1a1a1a;
}

.item .meta {
  display: flex;
  gap: 24px;
}

.item .meta .time {
  width: 180px;
  text-align: right;
  color: #888;
  font-size: 13px;
}

.item .meta .size {
  width: 100px;
  text-align: right;
  color: #888;
  font-size: 13px;
}

.context-menu {
  position: fixed;
  background: #1a1a1a;
  border: 1px solid #333;
  border-radius: 10px;
  overflow: hidden;
  z-index: 1000;
  box-shadow: 0 8px 24px rgba(0,0,0,0.5);
  min-width: 140px;
}

body.light .context-menu {
  background: #fff;
  border-color: #ddd;
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
</style>
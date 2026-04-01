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

// Drag and drop for storage order
const remoteOrder = ref<string[]>(JSON.parse(localStorage.getItem('remoteOrder') || '[]'))
const draggedRemote = ref('')

function getOrderedRemotes() {
  const remotesList = remotes.value
  const order = remoteOrder.value
  if (!order.length) return remotesList
  // Return ordered remotes, then any new remotes not in order
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
})

async function loadRemotes() {
  try {
    const data = await api.listRemotes()
    remotes.value = data.remotes || []
    
    // Auto-open first remote if none selected
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
  // Wait for modal to load providers, then load config
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
    <div class="card-header blue">
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
    <div class="card-header green">
      <div style="display: flex; justify-content: space-between; align-items: center; width: 100%">
        <div class="title">文件浏览</div>
        <div class="actions">
          <button class="ghost small" @click="openManageStorage">管理存储</button>
          <button class="ghost small" @click="openAddRemote">添加存储</button>
        </div>
      </div>
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
    <div class="list">
      <div
        v-for="item in browserItems"
        :key="item.Path"
        class="item"
        @click="enterItem(item)"
      >
        <div class="name">
          <span :class="item.IsDir ? 'folder' : 'icon'">{{ item.IsDir ? '📁' : '📄' }}</span>
          <span>{{ item.Name }}</span>
        </div>
        <div class="meta">
          <span class="size">{{ item.Size }}</span>
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

  <!-- Manage Storage Subview -->
  <div v-if="subview === 'manage-storage'" class="card">
    <div class="card-header yellow">
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

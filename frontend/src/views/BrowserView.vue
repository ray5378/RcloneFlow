<script setup lang="ts">
import { ref } from 'vue'
import AddRemoteModal from '../components/modals/AddRemoteModal.vue'
import EditDescModal from '../components/modals/EditDescModal.vue'
import StoragePanel from '../components/StoragePanel.vue'
import BrowserPanel from '../components/BrowserPanel.vue'
import ManageStoragePanel from '../components/ManageStoragePanel.vue'
import { useBrowser } from '../composables/useBrowser'

const emit = defineEmits<{
  navigate: [subview: string]
}>()

const addRemoteModal = ref<InstanceType<typeof AddRemoteModal> | null>(null)

const {
  // Core state
  browserFs,
  browserPath,
  browserItems,
  browserError,
  confirmModal,
  showAddRemote,
  isEditMode,
  editRemoteName,
  showEditDesc,
  editDescRemote,
  subview,

  // Composables
  clipboard,
  contextMenu,
  fileOps,
  remoteMgmt,

  // Computed
  breadcrumbs,

  // Methods
  refreshBrowser,
  loadRemotes,
  enterItem,
  formatTime,
  formatSize,
  copyItem,
  moveItem,
  startRename,
  confirmDelete,
  handleDeleteRemote,
  openManageStorage,
  openAddRemote,
  openEditRemote: openEditRemoteInternal,
  openEditDesc,
  saveDesc
} = useBrowser({ emit })

// 特殊处理 AddRemoteModal 的 openEditRemote 功能
async function openEditRemote(name: string) {
  await openEditRemoteInternal(name)
  setTimeout(async () => {
    if (addRemoteModal.value) {
      await addRemoteModal.value.loadConfig(name)
    }
  }, 100)
}
</script>

<template>
  <!-- Storage Panel -->
  <StoragePanel
    :remote-mgmt="remoteMgmt"
    @open-manage-storage="openManageStorage"
    @open-add-remote="openAddRemote"
  />

  <!-- Browser Panel -->
  <BrowserPanel
    v-if="subview === 'explorer'"
    :browser-fs="browserFs"
    :browser-path="browserPath"
    :browser-items="browserItems"
    :browser-error="browserError"
    :breadcrumbs="breadcrumbs"
    :clipboard="clipboard"
    :context-menu="contextMenu"
    :file-ops="fileOps"
    @refresh-browser="refreshBrowser"
    @enter-item="enterItem"
    @copy-item="copyItem"
    @move-item="moveItem"
    @start-rename="startRename"
    @confirm-delete="confirmDelete"
  />

  <!-- Manage Storage Panel -->
  <ManageStoragePanel
    v-if="subview === 'manage-storage'"
    :remote-mgmt="remoteMgmt"
    :file-ops="fileOps"
    @back="subview = 'explorer'"
    @open-add-remote="openAddRemote"
    @open-edit-remote="openEditRemote"
    @open-edit-desc="openEditDesc"
    @handle-delete-remote="handleDeleteRemote"
  />

  <!-- Delete Confirmation Modal -->
  <div v-if="fileOps.showDeleteConfirm.value" class="modal-overlay" @click.self="fileOps.showDeleteConfirm.value = false">
    <div class="modal delete-modal">
      <div class="modal-header">
        <h2>{{ $t('browserView.confirmDelete') }}</h2>
        <button class="modal-close" @click="fileOps.showDeleteConfirm.value = false">&times;</button>
      </div>
      <div class="modal-content">
        <p>{{ $t('browserView.confirmDeleteText') }} <strong>{{ fileOps.deletingItem.value?.Name }}</strong> ?</p>
        <p class="warning">{{ $t('browserView.irreversible') }}</p>
      </div>
      <div class="modal-footer">
        <button class="ghost" @click="fileOps.showDeleteConfirm.value = false">{{ $t('common.cancel') }}</button>
        <button class="danger-btn" @click="fileOps.executeDelete">{{ $t('browserView.confirmDeleteAction') }}</button>
      </div>
    </div>
  </div>

  <!-- Rename Modal -->
  <div v-if="fileOps.showRenameInput.value" class="modal-overlay" @click.self="fileOps.showRenameInput.value = false">
    <div class="modal">
      <div class="modal-header">
        <h2>{{ $t('browserView.rename') }}</h2>
        <button class="modal-close" @click="fileOps.showRenameInput.value = false">&times;</button>
      </div>
      <div class="modal-content">
        <div class="field-item">
          <label>{{ $t('browserView.newName') }}</label>
          <input v-model="fileOps.renameInput.value" @keyup.enter="fileOps.confirmRename" />
        </div>
      </div>
      <div class="modal-footer">
        <button class="ghost" @click="fileOps.showRenameInput.value = false">{{ $t('common.cancel') }}</button>
        <button class="primary" @click="fileOps.confirmRename">{{ $t('browserView.confirm') }}</button>
      </div>
    </div>
  </div>

  <!-- Confirm Modal -->
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
        <button class="ghost" @click="confirmModal.show = false">{{ $t('common.cancel') }}</button>
        <button class="primary danger" @click="confirmModal.onConfirm">{{ $t('browserView.confirm') }}</button>
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
    :description="remoteMgmt.descriptions[editDescRemote] || ''"
    @close="showEditDesc = false"
    @save="saveDesc"
  />
</template>

<style scoped>
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

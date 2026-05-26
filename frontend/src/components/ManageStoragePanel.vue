<script setup lang="ts">
import type { UseBrowserRemoteManagementReturn } from '../composables/useBrowserRemoteManagement'
import type { UseBrowserFileOpsReturn } from '../composables/useBrowserFileOps'

defineProps<{
  remoteMgmt: UseBrowserRemoteManagementReturn
  fileOps: UseBrowserFileOpsReturn
  onBack: () => void
  onOpenAddRemote: () => void
  onOpenEditRemote: (name: string) => Promise<void>
  onOpenEditDesc: (name: string) => void
  onHandleDeleteRemote: (name: string) => Promise<void>
}>()

defineEmits<{
  back: []
  openAddRemote: []
  openEditRemote: [name: string]
  openEditDesc: [name: string]
  handleDeleteRemote: [name: string]
}>()
</script>

<template>
  <div class="card">
    <div class="card-header">
      <div style="display: flex; justify-content: space-between; align-items: center; width: 100%">
        <div class="title">{{ $t('remote.manageTitle') }}</div>
        <div class="actions list-item-actions">
          <button class="ghost small" @click="$emit('back')">← {{ $t('remote.backButton') }}</button>
          <button class="ghost small" @click="$emit('openAddRemote')">➕ {{ $t('remote.addButton') }}</button>
        </div>
      </div>
    </div>
    <div class="list">
      <div v-for="name in remoteMgmt.remotes" :key="name" class="item" @click="remoteMgmt.openRemote(name)">
        <div class="name">
          <strong>{{ name }}</strong>
        </div>
        <div class="actions list-item-actions" @click.stop>
          <button class="ghost small" @click="$emit('openEditRemote', name)">✏️ {{ $t('remote.editConfig') }}</button>
          <button class="ghost small" @click="$emit('openEditDesc', name)">📝 {{ $t('remote.editDesc') }}</button>
          <button
            class="ghost small test-btn"
            :class="{ 'test-success': fileOps.testState.value[name] === 'success', 'test-failed': fileOps.testState.value[name] === 'failed' }"
            @click="fileOps.testRemote(name)"
            :disabled="fileOps.testState.value[name] === 'testing'"
          >
            🔗 {{ fileOps.getTestText(name) }}
          </button>
          <button class="ghost small danger-text" @click="$emit('handleDeleteRemote', name)">🗑️ {{ $t('remote.deleteStorage') }}</button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.actions.list-item-actions button.ghost.small:hover {
  background: #374151 !important;
  border-color: rgba(148, 163, 184, 0.42) !important;
  color: #fff !important;
}

:global(body.light) .actions.list-item-actions button.ghost.small:hover {
  background: #e5e7eb !important;
  border-color: rgba(148, 163, 184, 0.55) !important;
  color: #111827 !important;
}

.actions.list-item-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  align-items: center;
}

.actions.list-item-actions button.test-btn.test-success {
  background: #16a34a !important;
  border-color: #16a34a !important;
  color: #fff !important;
}
.actions.list-item-actions button.test-btn.test-failed {
  background: #dc2626 !important;
  border-color: #dc2626 !important;
  color: #fff !important;
}
:global(body.light) .actions.list-item-actions button.test-btn.test-success {
  background: #22c55e !important;
  border-color: #22c55e !important;
  color: #fff !important;
}
:global(body.light) .actions.list-item-actions button.test-btn.test-failed {
  background: #ef4444 !important;
  border-color: #ef4444 !important;
  color: #fff !important;
}
</style>

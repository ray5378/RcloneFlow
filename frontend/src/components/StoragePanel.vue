<script setup lang="ts">
import type { UseBrowserRemoteManagementReturn } from '../composables/useBrowserRemoteManagement'

defineProps<{
  remoteMgmt: UseBrowserRemoteManagementReturn
  selectedRemote: string
}>()

defineEmits<{
  openManageStorage: []
  openAddRemote: []
}>()
</script>

<template>
  <div class="card">
    <div class="card-header">
      <div style="display: flex; justify-content: space-between; align-items: center; width: 100%">
        <div>
          <div class="title">{{ $t('remote.panelTitle') }}</div>
          <div class="subtitle">{{ $t('remote.panelSubtitle') }}</div>
        </div>
        <div class="actions list-item-actions">
          <button class="ghost small" @click="$emit('openManageStorage')">⚙️ {{ $t('remote.manageButton') }}</button>
          <button class="ghost small" @click="$emit('openAddRemote')">➕ {{ $t('remote.addButton') }}</button>
        </div>
      </div>
    </div>
    <div class="tile-grid">
      <div
        v-for="name in remoteMgmt.getOrderedRemotes()"
        :key="name"
        class="tile"
        :class="{ active: name === selectedRemote }"
        draggable="true"
        @click="remoteMgmt.openRemote(name)"
        @dragstart="remoteMgmt.onDragStart(name)"
        @dragover="remoteMgmt.onDragOver($event, name)"
        @drop="remoteMgmt.onDrop($event, name)"
        @dragend="remoteMgmt.onDragEnd"
      >
        <div class="tile-header">
          <span class="tile-name">☁️ {{ name }}</span>
          <span class="tile-drag">⋮⋮</span>
        </div>
        <div v-if="remoteMgmt.descriptions[name]" class="tile-desc">
          {{ remoteMgmt.descriptions[name] }}
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.tile {
  border: 1px solid rgba(148, 163, 184, 0.22);
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.16);
  transition: background-color 0.18s ease, border-color 0.18s ease, box-shadow 0.18s ease, transform 0.18s ease;
}

.tile:hover {
  background: #252525;
  border-color: rgba(99, 102, 241, 0.42);
  box-shadow: 0 10px 24px rgba(0, 0, 0, 0.22);
  transform: translateY(-2px);
}

.tile.active {
  background: rgba(99, 102, 241, 0.18);
  border-color: rgba(99, 102, 241, 0.60);
  box-shadow: 0 4px 12px rgba(99, 102, 241, 0.20);
}

body.light .tile {
  border-color: rgba(15, 23, 42, 0.12);
  box-shadow: 0 1px 2px rgba(15, 23, 42, 0.06);
}

body.light .tile:hover {
  background: #f8f8f8;
  border-color: rgba(25, 118, 210, 0.30);
  box-shadow: 0 10px 24px rgba(15, 23, 42, 0.12);
}

body.light .tile.active {
  background: rgba(25, 118, 210, 0.12);
  border-color: rgba(25, 118, 210, 0.50);
  box-shadow: 0 4px 12px rgba(25, 118, 210, 0.15);
}
</style>

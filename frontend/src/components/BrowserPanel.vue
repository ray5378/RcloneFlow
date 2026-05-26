<script setup lang="ts">
import type { FileItem } from '../types'
import type { UseBrowserClipboardReturn } from '../composables/useBrowserClipboard'
import type { UseBrowserContextMenuReturn } from '../composables/useBrowserContextMenu'
import type { UseBrowserFileOpsReturn } from '../composables/useBrowserFileOps'

defineProps<{
  browserFs: string
  browserPath: string
  browserItems: FileItem[]
  browserError: string
  breadcrumbs: Array<{ name: string; path: string }>
  clipboard: UseBrowserClipboardReturn
  contextMenu: UseBrowserContextMenuReturn
  fileOps: UseBrowserFileOpsReturn
  onRefreshBrowser: () => Promise<void>
  onEnterItem: (item: FileItem) => void
  onFormatTime: (time: string) => string
  onFormatSize: (size: string) => string
  onCopyItem: () => void
  onMoveItem: () => void
  onStartRename: () => void
  onConfirmDelete: () => void
}>()

defineEmits<{
  refreshBrowser: []
  enterItem: [item: FileItem]
  copyItem: []
  moveItem: []
  startRename: []
  confirmDelete: []
}>()
</script>

<template>
  <div class="card">
    <div class="card-header">
      <div class="title">{{ $t('remote.browserTitle') }}</div>
    </div>
    <div class="pathbar">
      <template v-for="(crumb, i) in breadcrumbs" :key="crumb.path">
        <span v-if="i > 0" class="sep">/</span>
        <button
          class="crumb"
          :class="{ current: i === breadcrumbs.length - 1 }"
          @click="crumb.path !== browserPath && ($emit('refreshBrowser'))"
        >
          {{ crumb.name }}
        </button>
      </template>
    </div>
    <div class="list-header">
      <span class="col-name">{{ $t('browserView.name') }}</span>
      <span class="col-time">{{ $t('browserView.modifiedTime') }}</span>
      <span class="col-size">{{ $t('browserView.size') }}</span>
    </div>
    <div class="list" @contextmenu.prevent="contextMenu.showBackgroundMenuHandler($event)">
      <div
        v-for="item in browserItems"
        :key="item.Path"
        class="item"
        @click="$emit('enterItem', item)"
        @contextmenu.stop="contextMenu.showContextMenuHandler($event, item)"
      >
        <div class="name">
          <span :class="item.IsDir ? 'folder' : 'icon'">{{ item.IsDir ? '📁' : '📄' }}</span>
          <span>{{ item.Name }}</span>
        </div>
        <div class="meta">
          <span class="time">{{ onFormatTime(item.ModTime) }}</span>
          <span class="size">{{ item.IsDir ? '-' : onFormatSize(item.Size) }}</span>
        </div>
      </div>
      <div v-if="browserError" class="item" style="color: #ef5350">
        {{ browserError }}
      </div>
      <div v-if="!browserItems.length && !browserError" class="empty">
        {{ $t('remote.emptyDir') }}
      </div>
    </div>
  </div>

  <!-- Context Menu -->
  <div
    v-if="contextMenu.contextMenu.value.show"
    class="context-menu"
    :style="{ left: contextMenu.contextMenu.value.x + 'px', top: contextMenu.contextMenu.value.y + 'px' }"
  >
    <template v-if="contextMenu.contextMenu.value.item">
      <button @click="$emit('copyItem')">{{ $t('browserView.copy') }}</button>
      <button @click="$emit('moveItem')">{{ $t('browserView.move') }}</button>
      <button @click="clipboard.pasteItem" :disabled="!clipboard.clipboardItem.value">{{ $t('browserView.paste') }}</button>
      <button @click="$emit('startRename')">{{ $t('browserView.rename') }}</button>
      <button class="danger" @click="$emit('confirmDelete')">{{ $t('browserView.delete') }}</button>
    </template>
    <template v-else>
      <button @click="clipboard.pasteItem" :disabled="!clipboard.clipboardItem.value">{{ $t('browserView.pasteToCurrent') }}</button>
    </template>
  </div>
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
</style>

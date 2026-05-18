<script setup lang="ts">
import { t } from '../../../i18n'
import { formatBytes } from '../../../utils/format'
import type { ActiveTransferPendingFile, TrackingMode } from '../../../api/activeTransfer'
import { getTrackingLabels } from './transferringLabels'

const props = defineProps<{
  items: ActiveTransferPendingFile[]
  trackingMode: TrackingMode
  total?: number
  page: number
  totalPages: number
  jumpPage: number | null
}>()

const emit = defineEmits<{
  (e: 'prev-page'): void
  (e: 'next-page'): void
  (e: 'jump-page'): void
  (e: 'update:jump-page', value: number | null): void
}>()

function onJumpInput(event: Event) {
  const target = event.target as HTMLInputElement
  emit('update:jump-page', target.value === '' ? null : Number(target.value))
}

const PAGE_SIZE = 10

function getDisplayName(item: ActiveTransferPendingFile) {
  const raw = item.path || item.name || ''
  const normalized = raw.replace(/\\/g, '/')
  return normalized.split('/').filter(Boolean).pop() || raw
}

function placeholderRows() {
  return Array.from({ length: Math.max(0, PAGE_SIZE - props.items.length) }, (_, index) => index)
}
</script>

<template>
  <div class="list-box">
    <div class="title">{{ getTrackingLabels(props.trackingMode).pending }} <span v-if="props.total != null">({{ items.length }}/{{ props.total }})</span></div>
    <div class="list">
      <div v-for="item in items" :key="item.path || item.name" class="row">
        <span class="name" :title="item.path || item.name">{{ getDisplayName(item) }}</span>
        <span class="size">{{ formatBytes(item.sizeBytes || 0) }}</span>
      </div>
      <div v-for="slot in placeholderRows()" :key="`pending-placeholder-${slot}`" class="row placeholder-row">
        <span class="name placeholder-name">{{ t('activeTransfer.waitingFile') }}</span>
        <span class="size placeholder-size">—</span>
      </div>
    </div>
    <div v-if="(props.total || 0) > 10" class="pagination">
      <button class="page-btn" :disabled="page <= 1" @click="emit('prev-page')">{{ t('activeTransfer.prevPage') }}</button>
      <button class="page-btn" :disabled="page >= totalPages" @click="emit('next-page')">{{ t('activeTransfer.nextPage') }}</button>
      <input type="number" class="page-input" :value="jumpPage ?? ''" min="1" :max="totalPages" @input="onJumpInput" @keyup.enter="emit('jump-page')" />
      <button class="page-btn" @click="emit('jump-page')">{{ t('activeTransfer.jump') }}</button>
    </div>
  </div>
</template>

<style scoped>
.list-box { margin-bottom: 12px; width:100%; min-width:0; }
.title { font-size: 12px; color:#999; margin-bottom: 6px; }
.list { width:100%; box-sizing:border-box; border:1px solid #333; border-radius:8px; padding:8px; max-height:320px; overflow:auto; }
body.light .list { border-color:#ddd; }
.row { display:grid; grid-template-columns:minmax(0, 1fr) 110px; gap:12px; padding:6px 0; font-size:13px; align-items:center; min-height: 31px; box-sizing: border-box; }
.name { min-width:0; overflow:hidden; text-overflow:ellipsis; white-space:nowrap; }
.size { font-size:12px; color:#999; white-space:nowrap; text-align:right; }
.placeholder-row { color:#777; opacity:.55; }
.placeholder-name { font-style:italic; }
.placeholder-size { color:#777; }
.pagination { padding-top:8px; display:flex; gap:8px; align-items:center; flex-wrap:wrap; }

@media (max-width: 768px) {
  .list { padding: 8px 10px; }
  .row { grid-template-columns: minmax(0, 1fr) auto; gap: 10px; }
  .name { font-size: 13px; }
  .size { font-size: 11px; }
}
.page-btn { padding:4px 8px; border-radius:6px; border:1px solid #444; background:transparent; color:inherit; cursor:pointer; }
.page-btn:disabled { opacity:.5; cursor:not-allowed; }
.page-input { width:72px; min-width:72px; padding:4px 8px; border-radius:6px; border:1px solid #444; background:transparent; color:inherit; }
.empty { font-size:12px; color:#999; }
</style>

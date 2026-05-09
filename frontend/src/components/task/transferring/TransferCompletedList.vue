<script setup lang="ts">
import { t } from '../../../i18n'
import { formatBytes } from '../../../utils/format'
import type { ActiveTransferCompletedFile, TrackingMode } from '../../../api/activeTransfer'
import { getTrackingLabels, getTransferStatusLabel } from './transferringLabels'

const props = defineProps<{
  items: ActiveTransferCompletedFile[]
  trackingMode: TrackingMode
  total?: number
}>()

const emit = defineEmits<{
  (e: 'load-more'): void
}>()
</script>

<template>
  <div class="list-box">
    <div class="title">{{ getTrackingLabels(props.trackingMode).completed }} <span v-if="props.total != null">({{ items.length }}/{{ props.total }})</span></div>
    <div v-if="items.length" class="list">
      <div v-for="item in items" :key="`${item.path || item.name}-${item.at || ''}`" class="row">
        <div class="main">
          <span class="name">{{ item.path || item.name }}</span>
          <span class="size">{{ formatBytes(item.sizeBytes || 0) }}</span>
        </div>
        <span class="tag">{{ getTransferStatusLabel(item.status) }}</span>
      </div>
      <div v-if="(props.total || 0) > items.length" class="more-wrap">
        <button class="ghost small" @click="emit('load-more')">{{ t('activeTransfer.loadMore') }}</button>
      </div>
    </div>
    <div v-else class="empty">{{ t('activeTransfer.emptyList') }}</div>
  </div>
</template>

<style scoped>
.list-box { margin-bottom: 12px; }
.title { font-size: 12px; color:#999; margin-bottom: 6px; }
.list { border:1px solid #333; border-radius:8px; padding:8px; max-height:220px; overflow:auto; }
body.light .list { border-color:#ddd; }
.row { display:flex; justify-content:space-between; gap:12px; padding:6px 0; font-size:13px; align-items:flex-start; }
.main { display:flex; flex-direction:column; gap:2px; min-width:0; flex:1; }
.name { word-break:break-all; }
.size { font-size:12px; color:#999; }
.tag { color:#999; white-space:nowrap; }
.more-wrap { padding-top:8px; display:flex; justify-content:center; }
.empty { font-size:12px; color:#999; }
</style>

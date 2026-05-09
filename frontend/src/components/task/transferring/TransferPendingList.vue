<script setup lang="ts">
import { t } from '../../../i18n'
import type { ActiveTransferPendingFile, TrackingMode } from '../../../api/activeTransfer'
import { getTrackingLabels, getTransferStatusLabel } from './transferringLabels'

const props = defineProps<{
  items: ActiveTransferPendingFile[]
  trackingMode: TrackingMode
  total?: number
}>()

const emit = defineEmits<{
  (e: 'load-more'): void
}>()
</script>

<template>
  <div class="list-box">
    <div class="title">{{ getTrackingLabels(props.trackingMode).pending }} <span v-if="props.total != null">({{ items.length }}/{{ props.total }})</span></div>
    <div v-if="items.length" class="list">
      <div v-for="item in items" :key="item.path || item.name" class="row">
        <span class="name">{{ item.path || item.name }}</span>
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
.row { display:flex; justify-content:space-between; gap:12px; padding:6px 0; font-size:13px; }
.name { word-break:break-all; }
.tag { color:#999; white-space:nowrap; }
.more-wrap { padding-top:8px; display:flex; justify-content:center; }
.empty { font-size:12px; color:#999; }
</style>

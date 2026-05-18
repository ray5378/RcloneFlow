<script setup lang="ts">
import { computed } from 'vue'
import { formatBytes } from '../../utils/format'
import { getResolvedTotalCount, getUnifiedProgressText, type TaskProgressLike } from './progressText'
import { t } from '../../i18n'
import type { Run } from '../../types'

interface Summary {
  totalBytes?: number
  transferredBytes?: number
  counts?: { total?: number; copied?: number; deleted?: number; failed?: number; skipped?: number }
  message?: string
}

const props = defineProps<{ run: Run; progress?: TaskProgressLike; summary?: Summary }>()
const emit = defineEmits<{ click: [run: Run]; viewDetail: [run: Run]; viewLog: [run: Run]; clear: [id: number] }>()

const runTitle = computed(() => props.run.taskName || `${t('runItem.taskFallback')} #${props.run.taskId}`)
const statusClass = computed(() => getStatusClass(props.run.status || ''))
const statusText = computed(() => getStatusText(props.run.status || ''))
const triggerText = computed(() => props.run.trigger ? getTriggerText(props.run.trigger) : '')
const startedText = computed(() => formatTime(props.run.startedAt || ''))
const progressText = computed(() => props.run.status === 'running' ? getProgressText(props.run) : '-')
const normalizedSummary = computed(() => {
  const raw = props.summary as any
  return raw?.finalSummary && typeof raw.finalSummary === 'object' ? raw.finalSummary : raw
})
const summaryTotal = computed(() => {
  const summaryCount = Number(normalizedSummary.value?.counts?.total ?? 0)
  if (summaryCount > 0) return summaryCount
  return getResolvedTotalCount(props.progress)
})
const summarySuccess = computed(() => {
  const counts = normalizedSummary.value?.counts || {}
  return props.run.taskMode === 'move'
    ? (counts.copied || 0)
    : ((counts.copied || 0) + (counts.deleted || 0))
})
const summaryFailed = computed(() => normalizedSummary.value?.counts?.failed || 0)
const summaryTotalSize = computed(() => formatBytes(normalizedSummary.value?.totalBytes || 0))
const summaryTransferred = computed(() => formatBytes(normalizedSummary.value?.transferredBytes || 0))
const summaryMessage = computed(() => normalizedSummary.value?.message || '')

function getSuccessLabel(mode?: string) {
  if (mode === 'move') return t('runItem.moved')
  if (mode === 'sync') return t('runItem.synced')
  return t('runItem.copied')
}

function getStatusClass(status: string) {
  switch (status) {
    case 'finished': return 'success'
    case 'failed': return 'danger'
    case 'skipped': return 'warning'
    case 'running': return 'info'
    default: return ''
  }
}

function getStatusText(status: string) {
  const map: Record<string, string> = {
    finished: t('runItem.finished'),
    failed: t('runItem.failed'),
    skipped: t('runItem.skipped'),
    running: t('runItem.running'),
  }
  return map[status] || status
}

function formatTime(dateStr: string) {
  if (!dateStr) return '-'
  const d = new Date(dateStr)
  return d.toLocaleString('zh-CN', { month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit' })
}

function getTriggerText(trigger: string) {
  const map: Record<string, string> = {
    schedule: t('runItem.schedule'),
    webhook: 'Webhook',
    manual: t('runItem.manual'),
  }
  return map[trigger] || trigger
}

function getProgressText(run: Run): string {
  if (run.status === 'running') return getUnifiedProgressText(props.progress)
  return '-'
}
</script>

<template>
  <div class="item run-item" @click="emit('click', run)">
    <div class="name list-item-name">
      <strong>{{ runTitle }}</strong>
      <span v-if="run.taskMode" class="mode-tag list-item-tag">{{ run.taskMode }}</span>
      <span v-if="triggerText" class="trigger-tag list-item-tag">{{ triggerText }}</span>
    </div>

    <span :class="['status', statusClass, 'clickable']" @click.stop="emit('viewDetail', run)">
      {{ statusText }}
    </span>

    <div class="path-full">
      <span class="path-text list-item-secondary-text">{{ run.sourceRemote || '?' }}:{{ run.sourcePath || '/' }} → {{ run.targetRemote || '?' }}:{{ run.targetPath || '/' }}</span>
    </div>

    <span class="time list-item-tertiary-text">{{ startedText }}</span>

    <div class="summary-mini" v-if="run.status === 'running'">
      <span class="chip list-item-chip list-item-chip-meta">{{ progressText }}</span>
    </div>

    <div class="summary-mini" v-else-if="props.summary">
      <div v-if="summaryMessage" class="skipped-message">{{ summaryMessage }}</div>
      <template v-else>
        <span class="chip list-item-chip">{{ t('runItem.total') }} {{ summaryTotal }}</span>
        <span class="chip list-item-chip list-item-chip-success">
          {{ getSuccessLabel(run.taskMode) }}
          {{ summarySuccess }}
        </span>
        <span class="chip list-item-chip list-item-chip-failed">{{ t('runItem.failed') }} {{ summaryFailed }}</span>
        <span class="chip list-item-chip list-item-chip-meta">{{ t('runItem.totalSize') }} {{ summaryTotalSize }}</span>
        <span class="chip list-item-chip list-item-chip-meta">{{ t('runItem.transferred') }} {{ summaryTransferred }}</span>
      </template>
    </div>

    <div class="row-actions list-item-actions list-item-actions-right">
      <button class="ghost small" @click.stop="emit('viewDetail', run)">{{ t('runItem.viewDetail') }}</button>
      <button class="ghost small" @click.stop="emit('viewLog', run)">{{ t('runItem.viewLog') }}</button>
      <button class="ghost small danger-text" @click.stop="emit('clear', run.id)">{{ t('runItem.clear') }}</button>
    </div>
  </div>
</template>

<style scoped>
@import './listItemBase.css';
@import './listItemMeta.css';
@import './listItemActions.css';
@import './listItemSpacing.css';

.run-item { display: flex; flex-wrap: wrap; align-items: center; gap: 8px; }
.name { gap: 6px; }
.status { padding: 2px 8px; border-radius: 4px; font-size: 12px; }
.status.success { background: var(--success-bg, #22c55e33); color: var(--success, #22c55e); }
.status.danger { background: var(--danger-bg, #ef444433); color: var(--danger, #ef4444); }
.status.warning { background: var(--warning-bg, #f59e0b33); color: var(--warning, #f59e0b); }
.status.info { background: var(--accent-bg, #4f46e533); color: var(--accent, #4f46e5); }
.status.clickable { cursor: pointer; }
.path-full { flex: 1; min-width: 200px; overflow: hidden; }
.path-text { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.time { min-width: 80px; }
.summary-mini { width: 100%; display: flex; flex-wrap: wrap; gap: 6px; padding: 8px 0; }
.skipped-message { font-size: 12px; color: var(--warning, #f59e0b); padding: 4px 8px; background: #f59e0b22; border-radius: 4px; }
</style>

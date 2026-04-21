<script setup lang="ts">
import { formatBytes } from '../../utils/format'
import { getUnifiedProgressText } from './progressText'
import { t } from '../../i18n'

interface Progress {
  percentage?: number
  bytes?: number
  totalBytes?: number
  speed?: number
  completedFiles?: number
  totalCount?: number
}

interface Summary {
  files?: any[]
  totalBytes?: number
  transferredBytes?: number
  avgSpeedBps?: number
  counts?: { copied?: number; deleted?: number; failed?: number; skipped?: number }
  message?: string
}

interface RunRecord {
  id: number
  taskId: number
  taskName?: string
  taskMode?: string
  trigger?: string
  status?: string
  sourceRemote?: string
  sourcePath?: string
  targetRemote?: string
  targetPath?: string
  startedAt?: string
  bytesTransferred?: number
  speed?: string
  summary?: any
  progress?: Progress
}

const props = defineProps<{ run: RunRecord; progress?: Progress; summary?: Summary }>()
const emit = defineEmits<{ click: [run: RunRecord]; viewDetail: [run: RunRecord]; viewLog: [run: RunRecord]; clear: [id: number] }>()

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

function getProgressText(run: RunRecord): string {
  if (run.status === 'running') return getUnifiedProgressText(props.progress)
  return '-'
}
</script>

<template>
  <div class="item run-item" @click="emit('click', run)">
    <div class="name list-item-name">
      <strong>{{ run.taskName || `${t('runItem.taskFallback')} #${run.taskId}` }}</strong>
      <span v-if="run.taskMode" class="mode-tag list-item-tag">{{ run.taskMode }}</span>
      <span v-if="run.trigger" class="trigger-tag list-item-tag">{{ getTriggerText(run.trigger) }}</span>
    </div>

    <span :class="['status', getStatusClass(run.status || ''), 'clickable']" @click.stop="emit('viewDetail', run)">
      {{ getStatusText(run.status || '') }}
    </span>

    <div class="path-full">
      <span class="path-text list-item-secondary-text">{{ run.sourceRemote || '?' }}:{{ run.sourcePath || '/' }} → {{ run.targetRemote || '?' }}:{{ run.targetPath || '/' }}</span>
    </div>

    <span class="time list-item-tertiary-text">{{ formatTime(run.startedAt || '') }}</span>

    <div class="summary-mini" v-if="run.status === 'running'">
      <span class="chip list-item-chip list-item-chip-meta">{{ getProgressText(run) }}</span>
    </div>

    <div class="summary-mini" v-else-if="props.summary">
      <div v-if="props.summary.message" class="skipped-message">{{ props.summary.message }}</div>
      <template v-else>
        <span class="chip list-item-chip">{{ t('runItem.total') }} {{ props.summary.files?.length || 0 }}</span>
        <span class="chip list-item-chip list-item-chip-success">
          {{ run.taskMode === 'move' ? t('runItem.moved') : t('runItem.success') }}
          {{ (props.summary.counts?.copied || 0) + (props.summary.counts?.deleted || 0) }}
        </span>
        <span class="chip list-item-chip list-item-chip-failed">{{ t('runItem.failed') }} {{ props.summary.counts?.failed || 0 }}</span>
        <span class="chip list-item-chip list-item-chip-meta">{{ t('runItem.totalSize') }} {{ formatBytes(props.summary.totalBytes || 0) }}</span>
        <span class="chip list-item-chip list-item-chip-meta">{{ t('runItem.transferred') }} {{ formatBytes(props.summary.transferredBytes || 0) }}</span>
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

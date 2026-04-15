<script setup lang="ts">
import { formatBytes, formatBytesPerSec } from '../../utils/format'

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

const props = defineProps<{
  run: RunRecord
  progress?: Progress
  summary?: Summary
}>()

const emit = defineEmits<{
  click: [run: RunRecord]
  viewDetail: [run: RunRecord]
  viewLog: [run: RunRecord]
  clear: [id: number]
}>()

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
    finished: '完成',
    failed: '失败',
    skipped: '跳过',
    running: '运行中'
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
    schedule: '定时',
    webhook: 'Webhook',
    manual: '手动'
  }
  return map[trigger] || trigger
}

// 计算进度百分比
function getProgressPercent(run: RunRecord): string {
  if (run.status === 'running' && props.progress?.percentage) {
    return props.progress.percentage.toFixed(2) + '%'
  }
  return '-'
}

// 获取速度
function getSpeed(run: RunRecord): string {
  if (run.status === 'running' && props.progress?.speed) {
    return formatBytesPerSec(props.progress.speed)
  }
  return run.speed || '-'
}

// 获取已传输
function getTransferred(run: RunRecord): string {
  if (run.status === 'running' && props.progress?.bytes) {
    return formatBytes(props.progress.bytes)
  }
  return formatBytes(run.bytesTransferred || 0)
}
</script>

<template>
  <div class="item run-item" @click="emit('click', run)">
    <div class="name">
      <strong>{{ run.taskName || `任务 #${run.taskId}` }}</strong>
      <span class="mode-tag" v-if="run.taskMode">{{ run.taskMode }}</span>
      <span class="trigger-tag" v-if="run.trigger">{{ getTriggerText(run.trigger) }}</span>
    </div>
    
    <span :class="['status', getStatusClass(run.status || ''), 'clickable']" @click.stop="emit('viewDetail', run)">
      {{ getStatusText(run.status || '') }}
    </span>
    
    <div class="path-full">
      <span class="path-text">{{ run.sourceRemote || '?' }}:{{ run.sourcePath || '/' }} → {{ run.targetRemote || '?' }}:{{ run.targetPath || '/' }}</span>
    </div>
    
    <span class="time">{{ formatTime(run.startedAt || '') }}</span>
    
    <!-- 运行时进度 -->
    <div class="summary-mini" v-if="run.status === 'running'">
      <span class="chip">进度 {{ getProgressPercent(run) }}</span>
      <span class="chip meta">速度 {{ getSpeed(run) }}</span>
      <span class="chip meta">已传输 {{ getTransferred(run) }}</span>
    </div>
    
    <!-- 完成状态摘要 -->
    <div class="summary-mini" v-else-if="props.summary">
      <div v-if="props.summary.message" class="skipped-message">
        {{ props.summary.message }}
      </div>
      <template v-else>
        <span class="chip">总计 {{ props.summary.files?.length || 0 }}</span>
        <span class="chip success">
          {{ run.taskMode === 'move' ? '移动' : '成功' }}
          {{ (props.summary.counts?.copied || 0) + (props.summary.counts?.deleted || 0) }}
        </span>
        <span class="chip failed">失败 {{ props.summary.counts?.failed || 0 }}</span>
        <span class="chip meta">已传输 {{ formatBytes(props.summary.transferredBytes || 0) }}</span>
      </template>
    </div>
    
    <div class="row-actions">
      <button class="ghost small" @click.stop="emit('viewDetail', run)">运行详情</button>
      <button class="ghost small" @click.stop="emit('viewLog', run)">传输日志</button>
      <button class="ghost small danger-text" @click.stop="emit('clear', run.id)">清除</button>
    </div>
  </div>
</template>

<style scoped>
.run-item {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 8px;
  padding: 12px 16px;
  border-bottom: 1px solid #333;
  cursor: pointer;
}
.run-item:hover {
  background: rgba(255,255,255,0.03);
}
.name {
  display: flex;
  align-items: center;
  gap: 6px;
  min-width: 150px;
}
.name strong {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.mode-tag, .trigger-tag {
  font-size: 10px;
  padding: 2px 6px;
  border-radius: 4px;
  background: #333;
  color: #aaa;
}
.status {
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 12px;
}
.status.success { background: var(--success-bg, #22c55e33); color: var(--success, #22c55e); }
.status.danger { background: var(--danger-bg, #ef444433); color: var(--danger, #ef4444); }
.status.warning { background: var(--warning-bg, #f59e0b33); color: var(--warning, #f59e0b); }
.status.info { background: var(--accent-bg, #4f46e533); color: var(--accent, #4f46e5); }
.status.clickable { cursor: pointer; }
.path-full {
  flex: 1;
  min-width: 200px;
  overflow: hidden;
}
.path-text {
  font-size: 12px;
  color: #888;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.time {
  font-size: 12px;
  color: #666;
  min-width: 80px;
}
.summary-mini {
  width: 100%;
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  padding: 8px 0;
}
.chip {
  font-size: 11px;
  padding: 2px 8px;
  border-radius: 4px;
  background: #333;
  color: #ccc;
}
.chip.meta {
  background: #222;
  color: #999;
}
.chip.success { background: #22c55e33; color: #22c55e; }
.chip.failed { background: #ef444433; color: #ef4444; }
.skipped-message {
  font-size: 12px;
  color: var(--warning, #f59e0b);
  padding: 4px 8px;
  background: #f59e0b22;
  border-radius: 4px;
}
.row-actions {
  display: flex;
  gap: 8px;
  margin-left: auto;
}
</style>

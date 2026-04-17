<script setup lang="ts">
import { computed } from 'vue'
import { formatBytes, formatBytesPerSec, formatEta } from '../../utils/format'

interface Progress {
  percentage?: number
  bytes?: number
  totalBytes?: number
  speed?: number
  eta?: number  // 后端传来的 ETA（秒）
  completedFiles?: number
  totalCount?: number
  phase?: string
}

interface Schedule {
  enabled?: boolean
  spec?: string
}

interface ActiveRun {
  progress?: Progress
  stableProgress?: Progress
  runRecord?: { status?: string }
}

interface Task {
  id?: number
  name: string
  mode: string
  sourceRemote: string
  sourcePath: string
  targetRemote: string
  targetPath: string
  singleton?: boolean
  schedule?: string
  scheduleEnabled?: boolean
}

const props = defineProps<{
  task: Task
  schedule?: Schedule | null
  activeRun?: ActiveRun | null
  runningTaskId?: number | null
  stoppedTaskId?: number | null
}>()

const emit = defineEmits<{
  run: [task: Task]
  edit: [task: Task]
  delete: [task: Task]
  toggleSchedule: [task: Task]
  viewHistory: [taskId: number]
  stop: [taskId: number]
  setWebhook: [task: Task]
  setSingleton: [task: Task]
}>()

function getLiveProgress(): Progress | null {
  return props.activeRun?.progress || props.activeRun?.stableProgress || null
}

function getProgressPercent(): string {
  const p = getLiveProgress()
  if (!p) return '0.00'
  return (p.percentage || 0).toFixed(2)
}

function getProgressText(): string {
  const p = getLiveProgress()
  if (!p) return '-'
  if (p.phase === 'preparing') {
    return `准备中 · 已传 ${formatBytes(p.bytes || 0)} · 速度 ${formatBytesPerSec(p.speed || 0)}`
  }
  // 使用后端传来的 ETA
  let etaStr = ''
  if (p.eta && p.eta > 0) {
    etaStr = ` · 预计完成 ${formatEta(p.eta)}`
  }
  return `${getProgressPercent()}% · ${formatBytes(p.bytes || 0)} / ${formatBytes(p.totalBytes || 0)} · ${formatBytesPerSec(p.speed || 0)} · 总数量 ${p.totalCount || 0} ／ 已传输 ${p.completedFiles || 0}${etaStr}`
}

function formatSpec(spec: string): string {
  if (!spec) return '-'
  const parts = spec.split('|')
  if (parts.length !== 5) return spec
  const [min, hour, day, month, week] = parts
  const weekDay = ['周日', '周一', '周二', '周三', '周四', '周五', '周六'][parseInt(week) % 7] || ''
  const monthStr = month !== '*' ? `${month}月` : ''
  const dayStr = day !== '*' ? `${day}日` : ''
  return `${hour}:${min} ${weekDay} ${monthStr}${dayStr}`.trim()
}

function isRunning(): boolean {
  return props.runningTaskId === props.task.id
}

function isStopped(): boolean {
  return props.stoppedTaskId === props.task.id
}
</script>

<template>
  <div class="task-card" :class="{ active: activeRun }" @click="emit('viewHistory', task.id!)">
    <div class="task-main">
      <div class="name">
        <strong>{{ task.name }}</strong>
        <span class="mode-tag">{{ task.mode }}</span>
      </div>
      
      <div class="schedule-info">
        <template v-if="schedule">
          <span :class="['schedule-badge', schedule.enabled ? 'enabled' : 'disabled']">
            {{ schedule.enabled ? '已启用' : '已禁用' }}
          </span>
          <span class="schedule-rule">{{ formatSpec(schedule.spec || '') }}</span>
        </template>
        <span v-else class="no-schedule">未设置</span>
      </div>
      
      <div class="item-actions">
        <button class="ghost small" @click.stop="emit('viewHistory', task.id!)">📋 任务历史记录</button>
        <button class="ghost small" :class="{ 'danger-text': isStopped() }" @click.stop="emit('stop', task.id!)">
          {{ isStopped() ? '⏹ 已经停止' : '⏹ 停止传输' }}
        </button>
        <button v-if="schedule" class="ghost small" @click.stop="emit('toggleSchedule', task)">
          {{ schedule.enabled ? '⏸ 关闭定时' : '▶ 开启定时' }}
        </button>
        <button
          class="ghost small"
          :class="{ 'btn-running': isRunning() }"
          :disabled="isRunning()"
          @click.stop="emit('run', task)"
        >
          {{ isRunning() ? '运行成功' : '▶ 手动运行' }}
        </button>
        <button class="ghost small" @click.stop="emit('setWebhook', task)">🔗 Webhook</button>
        <button class="ghost small" @click.stop="emit('setSingleton', task)">🔒 单例</button>
        <button class="ghost small" @click.stop="emit('edit', task)">✏️</button>
        <button class="ghost small danger-text" @click.stop="emit('delete', task)">🗑️</button>
      </div>
    </div>
    
    <div class="task-paths">
      <div class="path-row">
        <span class="path-label">源:</span>
        <span class="path-value">{{ task.sourceRemote }}:{{ task.sourcePath || '根目录' }}</span>
      </div>
      <div class="path-row">
        <span class="path-label">目标:</span>
        <span class="path-value">{{ task.targetRemote }}:{{ task.targetPath || '根目录' }}</span>
      </div>
      <div class="path-row">
        <span class="path-label">进度:</span>
        <span class="path-value">{{ getProgressText() }}</span>
      </div>
      <div class="progress-bar-container" v-if="getLiveProgress() && getLiveProgress()!.phase !== 'preparing'">
        <div class="progress-bar" :style="{ width: getProgressPercent() + '%' }"></div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.task-card {
  position: relative;
  border-bottom: 1px solid #333;
  padding: 12px 16px;
  cursor: pointer;
  transition: background-color 0.18s ease, box-shadow 0.18s ease, border-left-color 0.18s ease;
  border-left: 3px solid transparent;
}
.task-card:hover {
  background: rgba(255,255,255,0.05);
  border-left-color: rgba(99,102,241,0.7);
  box-shadow: inset 0 0 0 1px rgba(255,255,255,0.04);
}
/* 在卡片分割线中间留一个轻微断点，帮助辨认这是两张独立卡片 */
.task-card::after {
  content: '';
  position: absolute;
  left: 50%;
  bottom: -1px;
  width: 18px;
  height: 3px;
  transform: translateX(-50%);
  background: var(--bg, #121212);
  border-radius: 999px;
  pointer-events: none;
  opacity: 0.95;
}
body.light .task-card::after {
  background: var(--bg, #f0f2f5);
}
body.light .task-card:hover {
  background: rgba(25,118,210,0.06);
  border-left-color: rgba(25,118,210,0.45);
  box-shadow: inset 0 0 0 1px rgba(25,118,210,0.08);
}
.task-card.active {
  border-left: 3px solid var(--accent, #4f46e5);
}
.task-main {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 12px;
}
.name {
  display: flex;
  align-items: center;
  gap: 8px;
  min-width: 150px;
}
.name strong {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.mode-tag {
  font-size: 10px;
  padding: 2px 6px;
  border-radius: 4px;
  background: #333;
  color: #aaa;
}
.schedule-info {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
}
.schedule-badge {
  padding: 2px 6px;
  border-radius: 4px;
  font-size: 10px;
}
.schedule-badge.enabled {
  background: #22c55e33;
  color: #22c55e;
}
.schedule-badge.disabled {
  background: #666633;
  color: #999;
}
.schedule-rule {
  color: #888;
}
.no-schedule {
  color: #666;
  font-size: 12px;
}
.item-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  margin-left: auto;
}
.task-paths {
  margin-top: 8px;
  display: flex;
  flex-direction: column;
  gap: 4px;
}
.path-row {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 12px;
}
.path-label {
  color: #888;
  min-width: 40px;
}
.path-value {
  color: #ccc;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.progress-bar-container {
  height: 4px;
  background: #333;
  border-radius: 2px;
  margin-top: 8px;
  overflow: hidden;
}
.progress-bar {
  height: 100%;
  background: var(--accent, #4f46e5);
  transition: width 0.3s;
}
.btn-running {
  color: #22c55e !important;
}
</style>

<script setup lang="ts">
import { getUnifiedProgressText } from './progressText'

interface Progress {
  percentage?: number
  bytes?: number
  totalBytes?: number
  speed?: number
  eta?: number
  completedFiles?: number
  totalCount?: number
  phase?: string
}

interface Schedule {
  enabled?: boolean
  spec?: string
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
  progress?: Progress | null
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
  return props.progress || null
}

function getProgressPercent(): string {
  const p = getLiveProgress()
  if (!p) return '0.00'
  return (p.percentage || 0).toFixed(2)
}

function getProgressText(): string {
  return getUnifiedProgressText(getLiveProgress())
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
  <div class="task-card" :class="{ active: !!progress }" @click="emit('viewHistory', task.id!)">
    <div class="task-main">
      <div class="name list-item-name">
        <strong>{{ task.name }}</strong>
        <span class="mode-tag list-item-tag">{{ task.mode }}</span>
      </div>

      <div class="schedule-info">
        <template v-if="schedule">
          <span :class="['schedule-badge', schedule.enabled ? 'enabled' : 'disabled']">
            {{ schedule.enabled ? '已启用' : '已禁用' }}
          </span>
          <span class="schedule-rule list-item-secondary-text">{{ formatSpec(schedule.spec || '') }}</span>
        </template>
        <span v-else class="no-schedule list-item-tertiary-text">未设置</span>
      </div>

      <div class="item-actions list-item-actions list-item-actions-right">
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
        <span class="path-label list-item-secondary-text">源:</span>
        <span class="path-value">{{ task.sourceRemote }}:{{ task.sourcePath || '根目录' }}</span>
      </div>
      <div class="path-row">
        <span class="path-label list-item-secondary-text">目标:</span>
        <span class="path-value">{{ task.targetRemote }}:{{ task.targetPath || '根目录' }}</span>
      </div>
      <div class="path-row">
        <span class="path-label list-item-secondary-text">进度:</span>
        <span class="path-value">{{ getProgressText() }}</span>
      </div>
      <div class="progress-bar-container" v-if="getLiveProgress() && getLiveProgress()!.phase !== 'preparing'">
        <div class="progress-bar" :style="{ width: getProgressPercent() + '%' }"></div>
      </div>
    </div>
  </div>
</template>

<style scoped>
@import './listItemBase.css';
@import './listItemMeta.css';
@import './listItemActions.css';

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
  gap: 8px;
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

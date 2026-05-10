<script setup lang="ts">
import { getUnifiedProgressText } from './progressText'
import { t } from '../../i18n'

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
  scheduleToggledTaskId?: number | null
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
  openTransferDetail: [taskId: number]
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
  const weekMap = [t('runtime.sunday'), t('runtime.monday'), t('runtime.tuesday'), t('runtime.wednesday'), t('runtime.thursday'), t('runtime.friday'), t('runtime.saturday')]
  const weekDay = weekMap[parseInt(week) % 7] || ''
  const monthStr = month !== '*' ? `${month}${t('schedule.monthSuffix')}` : ''
  const dayStr = day !== '*' ? `${day}` : ''
  return `${hour}:${min} ${weekDay} ${monthStr}${dayStr}`.trim()
}

function isRunning(): boolean {
  return props.runningTaskId === props.task.id
}

function isStopped(): boolean {
  return props.stoppedTaskId === props.task.id
}

function isScheduleToggled(): boolean {
  return props.scheduleToggledTaskId === props.task.id
}
</script>

<template>
  <div class="task-card" :class="{ active: !!progress }" @click="emit('viewHistory', task.id!)">
    <div class="task-main list-item-primary-group">
      <div class="name list-item-name">
        <strong>{{ task.name }}</strong>
        <span class="mode-tag list-item-tag">{{ task.mode }}</span>
      </div>

      <div class="schedule-info">
        <template v-if="schedule">
          <span :class="['schedule-badge', schedule.enabled ? 'enabled' : 'disabled']">
            {{ schedule.enabled ? t('taskCard.enabled') : t('taskCard.disabled') }}
          </span>
          <span class="schedule-rule list-item-secondary-text">{{ formatSpec(schedule.spec || '') }}</span>
        </template>
        <span v-else class="no-schedule list-item-tertiary-text">{{ t('taskCard.unset') }}</span>
      </div>

      <div class="item-actions list-item-actions list-item-actions-right">
        <button class="ghost small" @click.stop="emit('viewHistory', task.id!)">📋 {{ t('taskCard.history') }}</button>
        <button class="ghost small" @click.stop="emit('openTransferDetail', task.id!)">📦 {{ t('activeTransfer.title') }}</button>
        <button class="ghost small" :class="{ 'btn-stopped': isStopped() }" @click.stop="emit('stop', task.id!)">
          {{ isStopped() ? `⏹ ${t('taskCard.stopped')}` : `⏹ ${t('taskCard.stopTransfer')}` }}
        </button>
        <button v-if="schedule" class="ghost small" :class="{ 'btn-schedule-toggled': isScheduleToggled() }" @click.stop="emit('toggleSchedule', task)">
          {{ schedule.enabled ? `⏸ ${t('taskCard.disableSchedule')}` : `▶ ${t('taskCard.enableSchedule')}` }}
        </button>
        <button class="ghost small" :class="{ 'btn-running': isRunning() }" :disabled="isRunning()" @click.stop="emit('run', task)">
          {{ isRunning() ? t('taskCard.runSuccess') : `▶ ${t('taskCard.manualRun')}` }}
        </button>
        <button class="ghost small" @click.stop="emit('setWebhook', task)">🔗 {{ t('taskCard.webhook') }}</button>
        <button class="ghost small" @click.stop="emit('setSingleton', task)">🔒 {{ t('taskCard.singleton') }}</button>
        <button class="ghost small" @click.stop="emit('edit', task)">✏️ {{ t('taskCard.edit') }}</button>
        <button class="ghost small danger-text" @click.stop="emit('delete', task)">🗑️ {{ t('taskCard.delete') }}</button>
      </div>
    </div>

    <div class="task-paths list-item-secondary-group list-item-stack">
      <div class="path-row list-item-row">
        <span class="path-label list-item-secondary-text">{{ t('taskCard.source') }}:</span>
        <span class="path-value">{{ task.sourceRemote }}:{{ task.sourcePath || t('taskCard.rootDir') }}</span>
      </div>
      <div class="path-row list-item-row">
        <span class="path-label list-item-secondary-text">{{ t('taskCard.target') }}:</span>
        <span class="path-value">{{ task.targetRemote }}:{{ task.targetPath || t('taskCard.rootDir') }}</span>
      </div>
      <div class="path-row list-item-row">
        <span class="path-label list-item-secondary-text">{{ t('taskCard.progress') }}:</span>
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
@import './listItemSpacing.css';

.task-card { padding: 14px 18px; }
.task-card:hover { background: transparent; border-left-color: rgba(99, 102, 241, 0.55); }
body.light .task-card:hover { background: transparent; border-left-color: rgba(25, 118, 210, 0.38); }
.task-card.active { border-left: 3px solid var(--accent, #4f46e5); }
.task-main { display: flex; flex-wrap: wrap; align-items: center; }
.name { gap: 8px; }
.schedule-info { display: flex; align-items: center; gap: 6px; font-size: 12px; }
.schedule-badge { padding: 2px 6px; border-radius: 4px; font-size: 10px; }
.schedule-badge.enabled { background: #22c55e33; color: #22c55e; }
.schedule-badge.disabled { background: #666633; color: #999; }
.task-paths { padding: 8px 12px; background: #1a1a1a; border-radius: 6px; }
body.light .task-paths { background: #f5f5f5; }
.path-row { font-size: 12px; }
.path-label { min-width: 40px; }
.path-value { color: #ccc; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.progress-bar-container { height: 4px; background: #333; border-radius: 2px; margin-top: 8px; overflow: hidden; }
.progress-bar { height: 100%; background: var(--accent, #4f46e5); transition: width 0.3s; }
.btn-running { color: #22c55e !important; }
.btn-stopped { background: #b91c1c !important; border-color: #b91c1c !important; color: #fff !important; }
.btn-schedule-toggled { background: #2563eb !important; border-color: #2563eb !important; color: #fff !important; }
</style>

<script setup lang="ts">
import { getUnifiedProgressText, type TaskProgressLike } from './progressText'
import { formatScheduleSpec } from './scheduleOptions'
import { t } from '../../i18n'
import type { ActiveRun, Schedule, Task } from '../../types'

const props = defineProps<{
  task: Task
  schedule?: Schedule | null
  progress?: TaskProgressLike | null
  runningTaskId?: number | null
  stoppedTaskId?: number | null
  sorting?: boolean
  sortValue?: number | null
}>()

const emit = defineEmits<{
  run: [task: Task]
  edit: [task: Task]
  delete: [task: Task]
  openScheduleConfig: [task: Task]
  viewHistory: [taskId: number]
  stop: [taskId: number]
  setWebhook: [task: Task]
  setSingleton: [task: Task]
  openTransferDetail: [taskId: number]
  sortInput: [event: Event]
  sortEnter: [event: KeyboardEvent]
}>()

function getLiveProgress(): TaskProgressLike | null {
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

function isRunning(): boolean {
  return props.runningTaskId === props.task.id
}

function isStopped(): boolean {
  return props.stoppedTaskId === props.task.id
}

</script>

<template>
  <div class="task-card" :class="{ active: !!progress }" @click="emit('viewHistory', task.id!)">
    <div class="task-main list-item-primary-group">
      <div class="name list-item-name">
        <div v-if="sorting" class="sort-editor" @click.stop>
          <span class="sort-label">{{ t('taskUI.sortNumber') }}</span>
          <input class="sort-input" type="number" inputmode="numeric" step="1" :value="sortValue ?? ''" @click.stop @input="emit('sortInput', $event)" @keydown.enter.stop.prevent="emit('sortEnter', $event)" />
        </div>
        <strong>{{ task.name }}</strong>
        <span class="mode-tag list-item-tag">{{ task.mode }}</span>
      </div>

      <div class="schedule-info">
        <template v-if="schedule">
          <span :class="['schedule-badge', schedule.enabled ? 'enabled' : 'disabled']">
            {{ schedule.enabled ? t('taskCard.enabled') : t('taskCard.disabled') }}
          </span>
          <span class="schedule-rule list-item-secondary-text">{{ formatScheduleSpec(schedule.spec || '') }}</span>
        </template>
        <span v-else class="no-schedule list-item-tertiary-text">{{ t('taskCard.unset') }}</span>
      </div>

      <div v-if="!sorting" class="item-actions list-item-actions list-item-actions-right">
        <button class="ghost small action-info" @click.stop="emit('viewHistory', task.id!)">📋 {{ t('taskCard.history') }}</button>
        <button class="ghost small action-info" @click.stop="emit('openTransferDetail', task.id!)">📦 {{ t('activeTransfer.title') }}</button>
        <button class="ghost small action-danger" :class="{ 'btn-stopped': isStopped() }" @click.stop="emit('stop', task.id!)">
          {{ isStopped() ? `⏹ ${t('taskCard.stopped')}` : `⏹ ${t('taskCard.stopTransfer')}` }}
        </button>
        <button class="ghost small" @click.stop="emit('openScheduleConfig', task)">
          ⏰ {{ t('taskCard.scheduleConfig') }}
        </button>
        <button class="ghost small" :class="{ 'btn-running': isRunning() }" :disabled="isRunning()" @click.stop="emit('run', task)">
          {{ isRunning() ? t('taskCard.runSuccess') : `▶ ${t('taskCard.manualRun')}` }}
        </button>
        <button class="ghost small" @click.stop="emit('setWebhook', task)">🔗 {{ t('taskCard.webhook') }}</button>
        <button class="ghost small" @click.stop="emit('setSingleton', task)">🔒 {{ t('taskCard.singleton') }}</button>
        <button class="ghost small" @click.stop="emit('edit', task)">✏️ {{ t('taskCard.edit') }}</button>
        <button class="ghost small action-danger danger-text" @click.stop="emit('delete', task)">🗑️ {{ t('taskCard.delete') }}</button>
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
.task-card.active { background: rgba(99, 102, 241, 0.08); border-color: rgba(99, 102, 241, 0.34); }
body.light .task-card.active { background: #f5f7fa; border-color: rgba(25, 118, 210, 0.22); }
.task-main { display: flex; flex-wrap: wrap; align-items: center; }
.name { gap: 8px; }
.sort-editor { display: inline-flex; align-items: center; gap: 8px; margin-right: 10px; }
.sort-label { font-size: 12px; color: #999; }
.sort-input { width: 72px; height: 30px; border-radius: 6px; border: 1px solid #444; background: #111; color: #fff; padding: 0 8px; }
body.light .sort-input { background: #fff; color: #222; border-color: #ccc; }
.schedule-info { display: flex; align-items: center; gap: 6px; font-size: 12px; }
.schedule-badge { padding: 2px 6px; border-radius: 4px; font-size: 10px; }
.schedule-badge.enabled { background: #22c55e33; color: #22c55e; }
.schedule-badge.disabled { background: #666633; color: #999; }
.task-paths { padding: 8px 0 0; }
.path-row { font-size: 12px; }
.path-label { min-width: 40px; }
.path-value { color: #ccc; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.progress-bar-container { height: 4px; background: #333; border-radius: 2px; margin-top: 8px; overflow: hidden; }
.progress-bar { height: 100%; background: var(--accent, #4f46e5); transition: width 0.3s; }
.btn-running { color: #22c55e !important; }
.btn-stopped { background: #b91c1c !important; border-color: #b91c1c !important; color: #fff !important; }
</style>

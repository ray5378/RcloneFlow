<script setup lang="ts">
import type { Task } from './types'

const props = defineProps<{
  task: Task
  isActive: boolean
}>()

const emit = defineEmits<{
  run: [task: Task]
  edit: [task: Task]
  del: [task: Task]
  toggleSchedule: [task: Task]
}>()
</script>

<template>
  <div class="task-card" :class="{ active: isActive }">
    <div class="task-header">
      <span class="task-name">{{ task.name }}</span>
      <span class="chip mode">{{ task.mode }}</span>
      <span v-if="task.singleton" class="chip singleton" title="单例模式">🔒</span>
    </div>
    
    <div class="task-paths">
      <div class="path-row">
        <span class="label">源:</span>
        <span class="path">{{ task.sourceRemote }}:{{ task.sourcePath }}</span>
      </div>
      <div class="path-row">
        <span class="label">目标:</span>
        <span class="path">{{ task.targetRemote }}:{{ task.targetPath }}</span>
      </div>
    </div>

    <div v-if="task.scheduleEnabled" class="task-schedule">
      <span class="chip mini">📅 {{ task.schedule }}</span>
    </div>

    <div class="task-actions">
      <button class="btn primary small" @click.stop="emit('run', task)">▶ 运行</button>
      <button class="btn ghost small" @click.stop="emit('edit', task)">✏️</button>
      <button class="btn ghost small danger-text" @click.stop="emit('del', task)">🗑️</button>
      <button 
        class="btn ghost small" 
        :class="{ 'text-success': task.scheduleEnabled }"
        @click.stop="emit('toggleSchedule', task)"
        :title="task.scheduleEnabled ? '禁用定时' : '启用定时'"
      >
        📅
      </button>
    </div>
  </div>
</template>

<style scoped>
.task-card {
  background: var(--card-bg, #1a1a1a);
  border: 1px solid #333;
  border-radius: 12px;
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.task-card.active {
  border-color: var(--accent, #4f46e5);
}
.task-header {
  display: flex;
  align-items: center;
  gap: 8px;
}
.task-name {
  font-weight: 600;
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.task-paths {
  display: flex;
  flex-direction: column;
  gap: 4px;
  font-size: 12px;
}
.path-row {
  display: flex;
  gap: 4px;
  overflow: hidden;
}
.label {
  color: #888;
  flex-shrink: 0;
}
.path {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: #aaa;
}
.task-schedule {
  margin-top: 4px;
}
.task-actions {
  display: flex;
  gap: 8px;
  margin-top: 8px;
}
.chip.mini {
  padding: 2px 6px;
  font-size: 10px;
}
</style>

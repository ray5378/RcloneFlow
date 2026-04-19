<script setup lang="ts">
import TaskCard from './TaskCard.vue'
import TaskListHeader from './TaskListHeader.vue'
import TaskListPagination from './TaskListPagination.vue'

defineProps<{
  search: string
  filteredTasks: any[]
  getScheduleByTaskId: (taskId: number) => any
  getTaskCardProgressByTask: (taskId: number) => any
  runningTaskId: number | null
  stoppedTaskId: number | null
  tasksTotal: number
  tasksPageSize: number
  tasksPage: number
  currentTasksPages: number
  tasksJumpPage: number | null
}>()

const emit = defineEmits<{
  (e: 'update:search', value: string): void
  (e: 'add'): void
  (e: 'run', taskId: number): void
  (e: 'edit', task: any): void
  (e: 'delete', taskId: number): void
  (e: 'toggle-schedule', taskId: number): void
  (e: 'view-history', taskId: number): void
  (e: 'stop', taskId: number): void
  (e: 'set-webhook', task: any): void
  (e: 'set-singleton', task: any): void
  (e: 'prev-page'): void
  (e: 'next-page'): void
  (e: 'update:jump-page', value: number | null): void
  (e: 'jump-page'): void
}>()
</script>

<template>
  <div class="card">
    <TaskListHeader :search="search" @update:search="emit('update:search', $event)" @add="emit('add')" />

    <div class="list">
      <TaskCard
        v-for="task in filteredTasks"
        :key="task.id"
        :task="task"
        :schedule="getScheduleByTaskId(task.id)"
        :progress="getTaskCardProgressByTask(task.id)"
        :running-task-id="runningTaskId"
        :stopped-task-id="stoppedTaskId"
        @run="emit('run', task.id)"
        @edit="emit('edit', task)"
        @delete="emit('delete', task.id)"
        @toggle-schedule="emit('toggle-schedule', task.id)"
        @view-history="emit('view-history', task.id)"
        @stop="emit('stop', task.id)"
        @set-webhook="emit('set-webhook', task)"
        @set-singleton="emit('set-singleton', task)"
      />
      <div v-if="!filteredTasks.length" class="empty">暂无任务</div>
    </div>

    <TaskListPagination
      v-if="tasksTotal > tasksPageSize"
      :page="tasksPage"
      :total-pages="currentTasksPages"
      :jump-page="tasksJumpPage"
      @prev="emit('prev-page')"
      @next="emit('next-page')"
      @update:jump-page="emit('update:jump-page', $event)"
      @jump="emit('jump-page')"
    />
  </div>
</template>

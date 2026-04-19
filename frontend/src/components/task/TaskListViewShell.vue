<script setup lang="ts">
defineProps<{
  taskSearch: string
  filteredTasks: any[]
  getScheduleByTaskId: (taskId: number) => any
  getTaskCardProgressByTask: (task: any) => any
  runningTaskId: number | null
  stoppedTaskId: number | null
  tasksTotal: number
  tasksPageSize: number
  tasksPage: number
  currentTasksPages: number
  tasksJumpPage: number | null
  setTaskSearch: (value: string) => void
  goToAddTask: () => void
  runTask: (task: any) => void
  editTask: (task: any) => void
  deleteTask: (task: any) => void
  toggleSchedule: (task: any) => void
  viewTaskHistory: (task: any) => void
  stopTaskAny: (task: any) => void
  setWebhook: (task: any) => void
  setSingletonMode: (task: any) => void
  setTasksJumpPageValue: (value: number | null) => void
  jumpToTasksPage: () => void
}>()
</script>

<template>
  <TaskListSection
    :search="taskSearch"
    :filtered-tasks="filteredTasks"
    :get-schedule-by-task-id="getScheduleByTaskId"
    :get-task-card-progress-by-task="getTaskCardProgressByTask"
    :running-task-id="runningTaskId"
    :stopped-task-id="stoppedTaskId"
    :tasks-total="tasksTotal"
    :tasks-page-size="tasksPageSize"
    :tasks-page="tasksPage"
    :current-tasks-pages="currentTasksPages"
    :tasks-jump-page="tasksJumpPage"
    @update:search="setTaskSearch"
    @add="goToAddTask"
    @run="runTask"
    @edit="editTask"
    @delete="deleteTask"
    @toggle-schedule="toggleSchedule"
    @view-history="viewTaskHistory"
    @stop="stopTaskAny"
    @set-webhook="setWebhook"
    @set-singleton="setSingletonMode"
    @prev-page="tasksPage--"
    @next-page="tasksPage++"
    @update:jump-page="setTasksJumpPageValue"
    @jump-page="jumpToTasksPage"
  />
</template>

<script setup lang="ts">
import TaskListSection from './TaskListSection.vue'
import WebhookConfigModal from './WebhookConfigModal.vue'
import SingletonConfigModal from './SingletonConfigModal.vue'

defineProps<{
  taskSearch: string
  filteredTasks: any[]
  allTasks: any[]
  getScheduleByTaskId: (taskId: number) => any
  getTaskCardProgressByTask: (task: any) => any
  runningTaskId: number | null
  stoppedTaskId: number | null
  scheduleToggledTaskId: number | null
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
  saveTaskSortOrders: (orders: Record<number, number>) => Promise<boolean>
  openTransferDetail: (taskId: number) => void
  prevTasksPage: () => void
  nextTasksPage: () => void
  setTasksJumpPageValue: (value: number | null) => void
  jumpToTasksPage: () => void
  showWebhookModal: boolean
  webhookForm: any
  setWebhookTriggerId: (value: string) => void
  setWebhookPostUrl: (value: string) => void
  setWebhookWecomUrl: (value: string) => void
  setWebhookNotifyManual: (value: boolean) => void
  setWebhookNotifySchedule: (value: boolean) => void
  setWebhookNotifyWebhook: (value: boolean) => void
  setWebhookStatusSuccess: (value: boolean) => void
  setWebhookStatusFailed: (value: boolean) => void
  saveWebhook: () => void
  testWebhook: () => void
  closeWebhookModal: () => void
  showSingletonModal: boolean
  singletonForm: any
  setSingletonEnabled: (value: boolean) => void
  saveSingleton: () => void
  closeSingletonModal: () => void
}>()
</script>

<template>
  <TaskListSection
    :search="taskSearch"
    :filtered-tasks="filteredTasks"
    :all-tasks="allTasks"
    :get-schedule-by-task-id="getScheduleByTaskId"
    :get-task-card-progress-by-task="getTaskCardProgressByTask"
    :running-task-id="runningTaskId"
    :stopped-task-id="stoppedTaskId"
    :schedule-toggled-task-id="scheduleToggledTaskId"
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
    @open-transfer-detail="openTransferDetail"
    @save-sort="saveTaskSortOrders"
    @prev-page="prevTasksPage"
    @next-page="nextTasksPage"
    @update:jump-page="setTasksJumpPageValue"
    @jump-page="jumpToTasksPage"
  />

  <WebhookConfigModal
    :visible="showWebhookModal"
    :trigger-id="webhookForm?.triggerId ?? ''"
    :post-url="webhookForm?.postUrl ?? ''"
    :wecom-url="(webhookForm as any)?.wecomUrl ?? ''"
    :notify-manual="webhookForm?.notify?.manual ?? false"
    :notify-schedule="webhookForm?.notify?.schedule ?? false"
    :notify-webhook="webhookForm?.notify?.webhook ?? false"
    :status-success="(webhookForm as any)?.status?.success ?? false"
    :status-failed="(webhookForm as any)?.status?.failed ?? false"
    :can-test="!!(webhookForm?.postUrl) || !!((webhookForm as any)?.wecomUrl)"
    @update:trigger-id="setWebhookTriggerId"
    @update:post-url="setWebhookPostUrl"
    @update:wecom-url="setWebhookWecomUrl"
    @update:notify-manual="setWebhookNotifyManual"
    @update:notify-schedule="setWebhookNotifySchedule"
    @update:notify-webhook="setWebhookNotifyWebhook"
    @update:status-success="setWebhookStatusSuccess"
    @update:status-failed="setWebhookStatusFailed"
    @save="saveWebhook"
    @test="testWebhook"
    @close="closeWebhookModal"
  />

  <SingletonConfigModal
    :visible="showSingletonModal"
    :singleton-enabled="singletonForm?.singletonEnabled ?? false"
    @update:singleton-enabled="setSingletonEnabled"
    @save="saveSingleton"
    @close="closeSingletonModal"
  />
</template>

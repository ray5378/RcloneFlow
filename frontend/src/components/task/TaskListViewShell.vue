<script setup lang="ts">
import TaskListSection from './TaskListSection.vue'
import WebhookConfigModal from './WebhookConfigModal.vue'
import SingletonConfigModal from './SingletonConfigModal.vue'
import type { Schedule, Task } from '../../types'
import type { TaskProgressLike } from './progressText'

interface WebhookFormState {
  triggerId?: string
  matchText?: string
  postUrl?: string
  wecomUrl?: string
  notify?: {
    manual?: boolean
    schedule?: boolean
    webhook?: boolean
  }
  status?: {
    success?: boolean
    failed?: boolean
  }
}

interface SingletonFormState {
  singletonEnabled?: boolean
}

defineProps<{
  taskSearch: string
  filteredTasks: Task[]
  allTasks: Task[]
  getScheduleByTaskId: (taskId: number) => Schedule | null | undefined
  getTaskCardProgressByTask: (taskId: number) => TaskProgressLike | null
  runningTaskId: number | null
  stoppedTaskId: number | null
  openScheduleConfig: (task: Task) => void
  tasksTotal: number
  tasksPageSize: number
  tasksPage: number
  currentTasksPages: number
  tasksJumpPage: number | null
  setTaskSearch: (value: string) => void
  goToAddTask: () => void
  runTask: (taskId: number) => void
  editTask: (task: Task) => void
  deleteTask: (taskId: number) => void
  viewTaskHistory: (taskId: number) => void
  stopTaskAny: (taskId: number) => void
  setWebhook: (task: Task) => void
  setSingletonMode: (task: Task) => void
  saveTaskSortOrders: (orders: Record<number, number>, priorityTaskId?: number) => Promise<boolean>
  openTransferDetail: (taskId: number) => void
  prevTasksPage: () => void
  nextTasksPage: () => void
  setTasksJumpPageValue: (value: number | null) => void
  jumpToTasksPage: () => void
  showWebhookModal: boolean
  webhookForm: WebhookFormState | null
  setWebhookTriggerId: (value: string) => void
  setWebhookMatchText: (value: string) => void
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
  singletonForm: SingletonFormState | null
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
    :tasks-total="tasksTotal"
    :tasks-page-size="tasksPageSize"
    :tasks-page="tasksPage"
    :current-tasks-pages="currentTasksPages"
    :tasks-jump-page="tasksJumpPage"
    :save-task-sort-orders="saveTaskSortOrders"
    @update:search="setTaskSearch"
    @add="goToAddTask"
    @run="runTask"
    @edit="editTask"
    @delete="deleteTask"
    @open-schedule-config="openScheduleConfig"
    @view-history="viewTaskHistory"
    @stop="stopTaskAny"
    @set-webhook="setWebhook"
    @set-singleton="setSingletonMode"
    @open-transfer-detail="openTransferDetail"
    @prev-page="prevTasksPage"
    @next-page="nextTasksPage"
    @update:jump-page="setTasksJumpPageValue"
    @jump-page="jumpToTasksPage"
  />

  <WebhookConfigModal
    :visible="showWebhookModal"
    :trigger-id="webhookForm?.triggerId ?? ''"
    :match-text="webhookForm?.matchText ?? ''"
    :post-url="webhookForm?.postUrl ?? ''"
    :wecom-url="(webhookForm as any)?.wecomUrl ?? ''"
    :notify-manual="webhookForm?.notify?.manual ?? false"
    :notify-schedule="webhookForm?.notify?.schedule ?? false"
    :notify-webhook="webhookForm?.notify?.webhook ?? false"
    :status-success="(webhookForm as any)?.status?.success ?? false"
    :status-failed="(webhookForm as any)?.status?.failed ?? false"
    :can-test="!!(webhookForm?.postUrl) || !!((webhookForm as any)?.wecomUrl)"
    @update:trigger-id="setWebhookTriggerId"
    @update:match-text="setWebhookMatchText"
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

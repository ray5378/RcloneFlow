<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import TaskCard from './TaskCard.vue'
import TaskListHeader from './TaskListHeader.vue'
import TaskListPagination from './TaskListPagination.vue'
import { t } from '../../i18n'

const props = defineProps<{
  search: string
  filteredTasks: any[]
  allTasks: any[]
  getScheduleByTaskId: (taskId: number) => any
  getTaskCardProgressByTask: (taskId: number) => any
  runningTaskId: number | null
  stoppedTaskId: number | null
  scheduleToggledTaskId: number | null
  tasksTotal: number
  tasksPageSize: number
  tasksPage: number
  currentTasksPages: number
  tasksJumpPage: number | null
  savingSort?: boolean
  saveTaskSortOrders: (orders: Record<number, number>, priorityTaskId?: number) => Promise<boolean>
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
  (e: 'open-transfer-detail', taskId: number): void
  (e: 'prev-page'): void
  (e: 'next-page'): void
  (e: 'update:jump-page', value: number | null): void
  (e: 'jump-page'): void
}>()

const sorting = ref(false)
const sortInputs = ref<Record<number, number | null>>({})
const savingNow = ref(false)
const hasPendingChanges = ref(false)
const lastPriorityTaskId = ref<number | null>(null)

function buildSortMap(tasks: any[]) {
  const map: Record<number, number | null> = {}
  tasks.forEach((task, index) => {
    map[task.id] = typeof task.sortOrder === 'number' && Number.isFinite(task.sortOrder) ? task.sortOrder : index + 1
  })
  return map
}

function startSort() {
  sorting.value = true
  sortInputs.value = { ...buildSortMap(props.allTasks) }
  hasPendingChanges.value = false
  lastPriorityTaskId.value = null
}

function cancelSort() {
  sorting.value = false
  sortInputs.value = {}
  hasPendingChanges.value = false
  lastPriorityTaskId.value = null
}

function onSortInput(taskId: number, event: Event) {
  const target = event.target as HTMLInputElement
  const normalized = target.value.replace(/[^0-9-]/g, '')
  if (normalized !== target.value) {
    target.value = normalized
  }
  const raw = normalized.trim()
  if (raw === '' || raw === '-') {
    sortInputs.value[taskId] = null
    return
  }
  const value = Number(raw)
  sortInputs.value[taskId] = Number.isFinite(value) ? Math.trunc(value) : null
  if (sortInputs.value[taskId] !== null) {
    target.value = String(sortInputs.value[taskId])
  }
}

function buildResolvedSortMap(tasks: any[], priorityTaskId?: number) {
  const used = new Map<number, number>()
  const taskList = [...tasks]

  if (priorityTaskId !== undefined) {
    taskList.sort((a, b) => {
      if (a.id === priorityTaskId) return -1
      if (b.id === priorityTaskId) return 1
      return 0
    })
  }

  taskList.forEach((task, index) => {
    const requested = sortInputs.value[task.id]
    if (requested === null || requested === undefined || !Number.isFinite(requested)) return
    let current = requested
    while (used.has(current)) current += 1
    used.set(current, task.id)
  })

  taskList.forEach((task, index) => {
    if (hasTask(used, task.id)) return
    let current = typeof task.sortOrder === 'number' && Number.isFinite(task.sortOrder) ? task.sortOrder : index + 1
    while (used.has(current)) current += 1
    used.set(current, task.id)
  })

  const ordered = Array.from(used.entries())
    .sort((a, b) => a[0] - b[0])
    .map(([, taskId]) => taskId)

  const finalMap = new Map<number, number>()
  ordered.forEach((taskId, index) => {
    finalMap.set(taskId, index + 1)
  })

  return finalMap
}

function hasTask(used: Map<number, number>, taskId: number) {
  for (const usedTaskId of used.values()) {
    if (usedTaskId === taskId) return true
  }
  return false
}

const previewTasks = computed(() => {
  if (!sorting.value) return props.filteredTasks

  const tasks = [...props.allTasks]
  const finalMap = buildResolvedSortMap(tasks)

  return tasks
    .map(task => ({ ...task, __previewSortOrder: finalMap.get(task.id) ?? task.sortOrder ?? task.id }))
    .sort((a, b) => (a.__previewSortOrder - b.__previewSortOrder) || (a.id - b.id))
})

const saveDisabled = computed(() => !hasPendingChanges.value || savingNow.value || !!props.savingSort)

function applyPreviewSort(taskId: number) {
  const current = sortInputs.value[taskId]
  if (current === null || current === undefined || !Number.isFinite(current)) return

  const finalMap = buildResolvedSortMap([...props.allTasks], taskId)
  finalMap.forEach((sortOrder, id) => {
    sortInputs.value[id] = sortOrder
  })
  lastPriorityTaskId.value = taskId
  hasPendingChanges.value = true
}

async function saveSort() {
  if (saveDisabled.value) return
  const orders: Record<number, number> = {}
  Object.entries(sortInputs.value).forEach(([id, sortOrder]) => {
    if (sortOrder === null || sortOrder === undefined || !Number.isFinite(sortOrder)) return
    orders[Number(id)] = sortOrder
  })

  savingNow.value = true
  const ok = await props.saveTaskSortOrders(orders, lastPriorityTaskId.value ?? undefined)
  savingNow.value = false
  if (ok) {
    cancelSort()
  }
}

watch(() => props.allTasks, (tasks) => {
  if (!sorting.value) return
  if (savingNow.value) return
  if (hasPendingChanges.value) return
  sortInputs.value = { ...buildSortMap(tasks) }
}, { deep: true })
</script>

<template>
  <div class="card">
    <TaskListHeader
      :search="search"
      :sorting="sorting"
      :saving-sort="savingSort || savingNow"
      @update:search="emit('update:search', $event)"
      @add="emit('add')"
      @toggle-sort="startSort"
      @save-sort="saveSort"
      @cancel-sort="cancelSort"
    />

    <div v-if="sorting" class="sort-hint">{{ t('taskUI.sortAutoSaveHint') }}</div>

    <div class="list">
      <TaskCard
        v-for="task in previewTasks"
        :key="task.id"
        :task="task"
        :schedule="getScheduleByTaskId(task.id)"
        :progress="getTaskCardProgressByTask(task.id)"
        :running-task-id="runningTaskId"
        :stopped-task-id="stoppedTaskId"
        :schedule-toggled-task-id="scheduleToggledTaskId"
        :sorting="sorting"
        :sort-value="sortInputs[task.id] ?? null"
        @sort-input="onSortInput(task.id, $event)"
        @sort-enter="applyPreviewSort(task.id)"
        @run="emit('run', task.id)"
        @edit="emit('edit', task)"
        @delete="emit('delete', task.id)"
        @toggle-schedule="emit('toggle-schedule', task.id)"
        @view-history="emit('view-history', task.id)"
        @stop="emit('stop', task.id)"
        @set-webhook="emit('set-webhook', task)"
        @set-singleton="emit('set-singleton', task)"
        @open-transfer-detail="emit('open-transfer-detail', task.id)"
      />
      <div v-if="!previewTasks.length" class="empty">{{ t('taskUI.noTasks') }}</div>
    </div>

    <TaskListPagination
      v-if="!sorting && tasksTotal > tasksPageSize"
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

<style scoped>
.sort-hint {
  margin: 0 0 12px;
  padding: 10px 12px;
  border-radius: 8px;
  background: rgba(79, 70, 229, 0.14);
  border: 1px solid rgba(99, 102, 241, 0.28);
  color: #c7d2fe;
  font-size: 13px;
  line-height: 1.5;
}
body.light .sort-hint {
  background: rgba(25, 118, 210, 0.08);
  border-color: rgba(25, 118, 210, 0.2);
  color: #355070;
}
</style>

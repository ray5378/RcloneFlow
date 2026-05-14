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
  (e: 'save-sort', orders: Record<number, number>): void
}>()

const sorting = ref(false)
const savingSort = ref(false)
const sortInputs = ref<Record<number, number | null>>({})
const originalSortInputs = ref<Record<number, number | null>>({})

function buildSortMap(tasks: any[]) {
  const map: Record<number, number | null> = {}
  for (const task of tasks || []) {
    map[task.id] = typeof task.sortOrder === 'number' ? task.sortOrder : null
  }
  return map
}

function startSort() {
  const current = buildSortMap(props.allTasks)
  sortInputs.value = { ...current }
  originalSortInputs.value = { ...current }
  sorting.value = true
}

function cancelSort() {
  sortInputs.value = { ...originalSortInputs.value }
  sorting.value = false
  savingSort.value = false
}

watch(() => props.allTasks, (tasks) => {
  if (!sorting.value) return
  const next = buildSortMap(tasks)
  for (const task of tasks || []) {
    if (!(task.id in sortInputs.value)) {
      sortInputs.value[task.id] = next[task.id]
    }
  }
}, { deep: true })

const sortDirty = computed(() => {
  const ids = new Set([...Object.keys(originalSortInputs.value), ...Object.keys(sortInputs.value)])
  for (const id of ids) {
    const key = Number(id)
    if ((originalSortInputs.value[key] ?? null) !== (sortInputs.value[key] ?? null)) return true
  }
  return false
})

const saveDisabled = computed(() => !sortDirty.value || savingSort.value)
const sortHint = computed(() => `${t('taskUI.sortHint')}`)

const sortingPreviewTasks = computed(() => {
  if (!sorting.value) return props.filteredTasks

  const taskMap = new Map<number, any>()
  for (const task of props.allTasks || []) {
    taskMap.set(Number(task.id), task)
  }

  const used = new Map<number, number>()
  const explicitIds = new Set<number>()

  for (const task of props.allTasks || []) {
    const taskId = Number(task.id)
    const value = sortInputs.value[taskId]
    if (typeof value !== 'number' || !Number.isFinite(value)) continue
    explicitIds.add(taskId)
    let current = Math.trunc(value)
    while (used.has(current)) current += 1
    used.set(current, taskId)
  }

  for (const task of props.allTasks || []) {
    const taskId = Number(task.id)
    if (explicitIds.has(taskId)) continue
    let current = typeof task.sortOrder === 'number' && Number.isFinite(task.sortOrder)
      ? Math.trunc(task.sortOrder)
      : taskId
    while (used.has(current)) current += 1
    used.set(current, taskId)
  }

  return Array.from(used.entries())
    .sort((a, b) => a[0] - b[0])
    .map(([, taskId]) => taskMap.get(taskId))
    .filter(Boolean)
})

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

async function saveSort() {
  const orders: Record<number, number> = {}
  for (const task of props.allTasks || []) {
    const value = sortInputs.value[task.id]
    if (typeof value === 'number' && Number.isFinite(value)) {
      orders[task.id] = Math.trunc(value)
    }
  }
  savingSort.value = true
  try {
    emit('save-sort', orders)
    sorting.value = false
  } finally {
    savingSort.value = false
  }
}
</script>

<template>
  <div class="card">
    <TaskListHeader
      :search="search"
      :sorting="sorting"
      :save-disabled="saveDisabled"
      @update:search="emit('update:search', $event)"
      @add="emit('add')"
      @toggle-sort="startSort"
      @cancel-sort="cancelSort"
      @save-sort="saveSort"
    />

    <div v-if="sorting" class="sort-hint">{{ sortHint }}</div>

    <div class="list">
      <TaskCard
        v-for="task in (sorting ? sortingPreviewTasks : filteredTasks)"
        :key="task.id"
        :task="task"
        :schedule="getScheduleByTaskId(task.id)"
        :progress="getTaskCardProgressByTask(task.id)"
        :running-task-id="runningTaskId"
        :stopped-task-id="stoppedTaskId"
        :sorting="sorting"
        :sort-value="sortInputs[task.id] ?? null"
        @sort-input="onSortInput(task.id, $event)"
        @run="emit('run', task.id)"
        @edit="emit('edit', task)"
        @delete="emit('delete', task.id)"
        @toggle-schedule="emit('toggle-schedule', task.id)"
        @view-history="emit('view-history', task.id)"
        @stop="emit('stop', task.id)"
        @set-webhook="emit('set-webhook', task)"
        @set-singleton="emit('set-singleton', task)"
      />
      <div v-if="!(sorting ? sortingPreviewTasks : filteredTasks).length" class="empty">{{ t('taskUI.noTasks') }}</div>
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

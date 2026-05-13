<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import TaskCard from './TaskCard.vue'
import TaskListHeader from './TaskListHeader.vue'
import TaskListPagination from './TaskListPagination.vue'
import { t } from '../../i18n'

const props = defineProps<{
  search: string
  filteredTasks: any[]
  allTasks?: any[]
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
}>()

const emit = defineEmits<{
  (e: 'update:search', value: string): void
  (e: 'add'): void
  (e: 'enter-sort-mode'): void
  (e: 'cancel-sort'): void
  (e: 'save-sort', order: number[]): void
  (e: 'run', taskId: number): void
  (e: 'edit', task: any): void
  (e: 'delete', taskId: number): void
  (e: 'toggle-schedule', taskId: number): void
  (e: 'reorder', order: number[]): void
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

const sortingMode = ref(false)
const dragTasks = ref<any[]>([])
const dragTaskId = ref<number | null>(null)
const dragOverTaskId = ref<number | null>(null)
const longPressTimer = ref<number | null>(null)
const dragArmedTaskId = ref<number | null>(null)
const savingOrder = ref(false)
const sortDirty = ref(false)

watch(
  () => [props.filteredTasks, props.allTasks, sortingMode.value],
  () => {
    if (dragTaskId.value !== null) return
    const source = sortingMode.value ? (props.allTasks || props.filteredTasks) : props.filteredTasks
    dragTasks.value = Array.isArray(source) ? [...source] : []
  },
  { immediate: true, deep: true },
)

const orderedTasks = computed(() => dragTasks.value)

function clearLongPressTimer() {
  if (longPressTimer.value !== null) {
    window.clearTimeout(longPressTimer.value)
    longPressTimer.value = null
  }
}

function armDrag(taskId: number) {
  clearLongPressTimer()
  dragArmedTaskId.value = taskId
  longPressTimer.value = window.setTimeout(() => {
    dragTaskId.value = taskId
    longPressTimer.value = null
  }, 320)
}

function cancelArmDrag() {
  dragArmedTaskId.value = null
  clearLongPressTimer()
}

function enterSortMode() {
  sortingMode.value = true
  sortDirty.value = false
  dragTasks.value = Array.isArray(props.allTasks) ? [...props.allTasks] : [...props.filteredTasks]
  emit('enter-sort-mode')
}

function cancelSortMode() {
  sortingMode.value = false
  sortDirty.value = false
  dragTaskId.value = null
  dragOverTaskId.value = null
  dragArmedTaskId.value = null
  dragTasks.value = Array.isArray(props.filteredTasks) ? [...props.filteredTasks] : []
  emit('cancel-sort')
}

function moveTaskBefore(targetTaskId: number) {
  if (dragTaskId.value === null || dragTaskId.value === targetTaskId) return
  const fromIndex = dragTasks.value.findIndex(task => task.id === dragTaskId.value)
  const toIndex = dragTasks.value.findIndex(task => task.id === targetTaskId)
  if (fromIndex < 0 || toIndex < 0 || fromIndex === toIndex) return
  const next = [...dragTasks.value]
  const [moved] = next.splice(fromIndex, 1)
  next.splice(toIndex, 0, moved)
  dragTasks.value = next
  dragOverTaskId.value = targetTaskId
  sortDirty.value = true
}

async function finishDrag() {
  cancelArmDrag()
  if (dragTaskId.value === null) return
  dragTaskId.value = null
  dragOverTaskId.value = null
}

async function saveSortMode() {
  const order = dragTasks.value.map(task => Number(task.id)).filter(id => id > 0)
  if (!order.length || !sortDirty.value) {
    sortingMode.value = false
    dragTasks.value = Array.isArray(props.filteredTasks) ? [...props.filteredTasks] : []
    return
  }
  savingOrder.value = true
  try {
    emit('save-sort', order)
    sortingMode.value = false
    sortDirty.value = false
  } finally {
    savingOrder.value = false
  }
}
</script>

<template>
  <div class="card">
    <TaskListHeader
      :search="search"
      :sorting-mode="sortingMode"
      :sort-dirty="sortDirty"
      @update:search="emit('update:search', $event)"
      @add="emit('add')"
      @enter-sort-mode="enterSortMode"
      @cancel-sort="cancelSortMode"
      @save-sort="saveSortMode"
    />

    <div v-if="sortingMode" class="sort-mode-tip">
      <div>{{ t('taskUI.sortModeTip') }}</div>
      <div class="sort-mode-subtip">{{ t('taskUI.dragHandleHint') }}</div>
    </div>

    <div class="list">
      <TaskCard
        v-for="task in orderedTasks"
        :key="task.id"
        :task="task"
        :schedule="getScheduleByTaskId(task.id)"
        :progress="getTaskCardProgressByTask(task.id)"
        :running-task-id="runningTaskId"
        :stopped-task-id="stoppedTaskId"
        :schedule-toggled-task-id="scheduleToggledTaskId"
        :drag-enabled="dragTaskId === task.id || dragArmedTaskId === task.id"
        :dragging="dragTaskId === task.id"
        :drag-over="dragOverTaskId === task.id"
        :saving-order="savingOrder"
        :sorting-mode="sortingMode"
        draggable="true"
        @drag-start.prevent
        @dragenter.prevent="moveTaskBefore(task.id)"
        @dragover.prevent="moveTaskBefore(task.id)"
        @dragend="finishDrag"
        @drop.prevent="finishDrag"
        @pointercancel="cancelArmDrag"
        @pointerleave="cancelArmDrag"
        @drag-handle-down="sortingMode && armDrag(task.id)"
        @drag-handle-up="sortingMode && finishDrag()"
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
      <div v-if="savingOrder" class="reorder-hint">{{ t('common.saving') }}...</div>
      <div v-if="!orderedTasks.length" class="empty">{{ t('taskUI.noTasks') }}</div>
    </div>

    <TaskListPagination
      v-if="!sortingMode && tasksTotal > tasksPageSize"
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
.reorder-hint {
  padding: 8px 12px;
  font-size: 12px;
  color: #a5b4fc;
}

.sort-mode-tip {
  padding: 8px 12px 0;
  font-size: 12px;
  color: #c4b5fd;
}

.sort-mode-subtip {
  margin-top: 4px;
  color: #a5b4fc;
}
</style>

<script setup lang="ts">
import TaskHistoryPanel from './TaskHistoryPanel.vue'
import RunDetailModal from './RunDetailModal.vue'

defineProps<{
  currentTotal: number
  runsPage: number
  runsPageSize: number
  currentTotalPages: number
  jumpPage: number
  historyFilterTaskId: number | null
  historyStatusFilter: string
  filteredRuns: any[]
  getRunProgressFromSummary: (run: any) => any
  getRealtimeProgressByRun: (run: any) => any
  getFinalSummary: (run: any) => any
  showDetailModal: boolean
  runDetail: any
  getStatusClass: (status: string) => string
  getStatusText: (status: string) => string
  getPreflight: (run: any) => any
  formatBytes: (bytes: number) => string
  formatTime: (value: any) => string
  formatBps: (bps: number) => string
  finalCountAll: number
  finalCountSuccess: number
  finalCountFailed: number
  finalCountOther: number
  pagedRunFiles: any[]
  runFilesTotal: number
  runFilesPage: number
  totalRunFilesPages: number
}>()

const emit = defineEmits<{
  (e: 'back'): void
  (e: 'set-status-filter', value: string): void
  (e: 'prev-page'): void
  (e: 'next-page'): void
  (e: 'update-jump-page', value: number): void
  (e: 'jump-page'): void
  (e: 'clear-all'): void
  (e: 'view-detail', run: any): void
  (e: 'view-log', run: any): void
  (e: 'clear-run', runId: number): void
  (e: 'close-detail'): void
  (e: 'prev-files-page'): void
  (e: 'next-files-page'): void
}>()
</script>

<template>
  <TaskHistoryPanel
    :current-total="currentTotal"
    :runs-page="runsPage"
    :runs-page-size="runsPageSize"
    :current-total-pages="currentTotalPages"
    :jump-page="jumpPage"
    :history-filter-task-id="historyFilterTaskId"
    :history-status-filter="historyStatusFilter"
    :filtered-runs="filteredRuns"
    :get-run-progress-from-summary="getRunProgressFromSummary"
    :get-realtime-progress-by-run="getRealtimeProgressByRun"
    :get-final-summary="getFinalSummary"
    @back="emit('back')"
    @set-status-filter="emit('set-status-filter', $event)"
    @prev-page="emit('prev-page')"
    @next-page="emit('next-page')"
    @update-jump-page="emit('update-jump-page', $event)"
    @jump-page="emit('jump-page')"
    @clear-all="emit('clear-all')"
    @view-detail="emit('view-detail', $event)"
    @view-log="emit('view-log', $event)"
    @clear-run="emit('clear-run', $event)"
  />

  <RunDetailModal
    :visible="showDetailModal"
    :run-detail="runDetail"
    :get-status-class="getStatusClass"
    :get-status-text="getStatusText"
    :get-final-summary="getFinalSummary"
    :format-bytes="formatBytes"
    :format-time="formatTime"
    :format-bps="formatBps"
    :final-count-all="finalCountAll"
    :final-count-success="finalCountSuccess"
    :final-count-failed="finalCountFailed"
    :final-count-other="finalCountOther"
    :paged-run-files="pagedRunFiles"
    :run-files-total="runFilesTotal"
    :run-files-page="runFilesPage"
    :total-run-files-pages="totalRunFilesPages"
    @close="emit('close-detail')"
    @prev-files-page="emit('prev-files-page')"
    @next-files-page="emit('next-files-page')"
  />
</template>

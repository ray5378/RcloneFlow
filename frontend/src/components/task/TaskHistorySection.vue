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
  finalFilesTotal: number
  finalFilesPage: number
  totalFinalFilesPages: number
  finalFilesJump: number | null
  pagedFinalFiles: any[]
  finalFiles: any[]
  pagedRunFiles: any[]
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
  (e: 'set-final-filter', value: 'all' | 'success' | 'failed' | 'other'): void
  (e: 'prev-final-files-page'): void
  (e: 'next-final-files-page'): void
  (e: 'update-final-files-jump', value: number | null): void
  (e: 'jump-final-files-page'): void
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
    :final-files-total="finalFilesTotal"
    :final-files-page="finalFilesPage"
    :total-final-files-pages="totalFinalFilesPages"
    :final-files-jump="finalFilesJump"
    :paged-final-files="pagedFinalFiles"
    :final-files="finalFiles"
    :paged-run-files="pagedRunFiles"
    :run-files-page="runFilesPage"
    :total-run-files-pages="totalRunFilesPages"
    @close="emit('close-detail')"
    @set-final-filter="emit('set-final-filter', $event)"
    @prev-final-files-page="emit('prev-final-files-page')"
    @next-final-files-page="emit('next-final-files-page')"
    @update-final-files-jump="emit('update-final-files-jump', $event)"
    @jump-final-files-page="emit('jump-final-files-page')"
    @prev-files-page="emit('prev-files-page')"
    @next-files-page="emit('next-files-page')"
  />
</template>

<script setup lang="ts">
import TaskHistorySection from './TaskHistorySection.vue'
import RunLogModal from './RunLogModal.vue'
import RunningHintModal from './RunningHintModal.vue'
import type { Run } from '../../types'
import type { TaskProgressLike } from './progressText'

defineProps<{
  currentTotal: number
  runsPage: number
  runsPageSize: number
  currentTotalPages: number
  jumpPage: number
  historyFilterTaskId: number | null
  historyStatusFilter: string
  filteredRuns: Run[]
  getRunProgressFromSummary: (run: Run) => TaskProgressLike | null
  getRealtimeProgressByRun: (run: Run) => TaskProgressLike | null
  getFinalSummaryFromComposable: (run: Run) => any
  showDetailModal: boolean
  runDetail: any
  getStatusClass: (status: string) => string
  getStatusText: (status: string) => string
  formatBytes: (value: number) => string
  formatTime: (value: any) => string
  formatBps: (value: number) => string
  finalCountAll: number
  finalCountSuccess: number
  finalCountFailed: number
  finalCountOther: number
  pagedRunFiles: any[]
  runFilesTotal: number
  runFilesPage: number
  totalRunFilesPages: number
  backToTasks: () => void
  prevRunsPage: () => void
  nextRunsPage: () => void
  setHistoryStatusFilter: (value: string) => void
  setJumpPageValue: (value: number) => void
  jumpToPage: () => void
  clearAllRunsWithConfirm: () => void
  showRunDetail: (run: Run) => void
  openRunLog: (run: Run) => void
  clearRun: (run: number) => void
  closeRunDetail: () => void
  setFinalFilter: (filter: string) => void
  goPrevFilesPage: () => void
  goNextFilesPage: () => void
  showLogModal: boolean
  logModalTitle: string
  logContent: string
  closeLogModal: () => void
  runningHintVisible: boolean
  runningHintRun: any
  runningHintDebugOpen: boolean
  runningHintPhaseText: string
  runningHintProgressText: string
  runningHintDebugInfo: any
  closeRunningHint: () => void
  toggleRunningHintDebug: () => void
  openRunningHintLog: () => void
}>()
</script>

<template>
  <TaskHistorySection
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
    :get-final-summary="getFinalSummaryFromComposable"
    :show-detail-modal="showDetailModal"
    :run-detail="runDetail"
    :get-status-class="getStatusClass"
    :get-status-text="getStatusText"
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
    @back="backToTasks"
    @set-status-filter="setHistoryStatusFilter"
    @prev-page="prevRunsPage"
    @next-page="nextRunsPage"
    @update-jump-page="setJumpPageValue"
    @jump-page="jumpToPage"
    @clear-all="clearAllRunsWithConfirm"
    @view-detail="showRunDetail"
    @view-log="openRunLog"
    @clear-run="clearRun"
    @close-detail="closeRunDetail"
    @set-final-filter="setFinalFilter"
    @prev-files-page="goPrevFilesPage"
    @next-files-page="goNextFilesPage"
  />

  <RunLogModal
    :visible="showLogModal"
    :title="logModalTitle"
    :content="logContent"
    @close="closeLogModal"
  />

  <RunningHintModal
    :visible="runningHintVisible"
    :run="runningHintRun"
    :phase-text="runningHintPhaseText"
    :progress-text="runningHintProgressText"
    :debug-open="runningHintDebugOpen"
    :debug-check-text="runningHintDebugInfo.checkText"
    :debug-progress-line="runningHintDebugInfo.progressLine"
    :debug-progress-json="runningHintDebugInfo.progressJson"
    @close="closeRunningHint"
    @toggle-debug="toggleRunningHintDebug"
    @open-log="openRunningHintLog"
  />
</template>

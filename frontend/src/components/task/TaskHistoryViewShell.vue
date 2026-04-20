<script setup lang="ts">
import TaskHistorySection from './TaskHistorySection.vue'
import RunLogModal from './RunLogModal.vue'
import RunningHintModal from './RunningHintModal.vue'

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
  getFinalSummaryFromComposable: (run: any) => any
  showDetailModal: boolean
  runDetail: any
  getStatusClass: (status: string) => string
  getStatusText: (status: string) => string
  getPreflightFromComposable: (run: any) => any
  formatBytes: (value: number) => string
  formatTime: (value: any) => string
  formatBps: (value: number) => string
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
  backToTasks: () => void
  prevRunsPage: () => void
  nextRunsPage: () => void
  setHistoryStatusFilter: (value: string) => void
  setJumpPageValue: (value: number) => void
  jumpToPage: () => void
  clearAllRunsWithConfirm: () => void
  showRunDetail: (run: any) => void
  openRunLog: (run: any) => void
  clearRun: (run: any) => void
  closeRunDetail: () => void
  setFinalFilter: (filter: string) => void
  goPrevFinalFilesPage: () => void
  goNextFinalFilesPage: () => void
  setFinalFilesJumpValue: (value: number | null) => void
  jumpFinalFilesPage: () => void
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
    :final-files-total="finalFilesTotal"
    :final-files-page="finalFilesPage"
    :total-final-files-pages="totalFinalFilesPages"
    :final-files-jump="finalFilesJump"
    :paged-final-files="pagedFinalFiles"
    :final-files="finalFiles"
    :paged-run-files="pagedRunFiles"
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
    @prev-final-files-page="goPrevFinalFilesPage"
    @next-final-files-page="goNextFinalFilesPage"
    @update-final-files-jump="setFinalFilesJumpValue"
    @jump-final-files-page="jumpFinalFilesPage"
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

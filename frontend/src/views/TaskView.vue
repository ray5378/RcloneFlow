<script setup lang="ts">
import GlobalStatsModal from '../components/task/GlobalStatsModal.vue'
import ConfirmModal from '../components/task/ConfirmModal.vue'
import TaskListViewShell from '../components/task/TaskListViewShell.vue'
import TaskHistoryViewShell from '../components/task/TaskHistoryViewShell.vue'
import TaskEditorViewShell from '../components/task/TaskEditorViewShell.vue'
import ToastCenter from '../components/toast/ToastCenter.vue'
import ScheduleConfigModal from '../components/task/ScheduleConfigModal.vue'
import TransferringModal from '../components/task/transferring/TransferringModal.vue'
import TagManagerModal from '../components/task/TagManagerModal.vue'
import BisyncLstManagerModal from '../components/task/BisyncLstManagerModal.vue'
import { useTaskView } from '../composables/useTaskView'

const {
  toasts,
  currentModule,
  
  // Task List
  tasks,
  filteredTasks,
  tasksTotal,
  tasksPage,
  tasksPageSize,
  currentTasksPages,
  tasksJumpPage,
  taskSearch,
  filteredTasksRaw,
  setTaskSearch,
  setTasksJumpPageValue,
  prevTasksPage,
  nextTasksPage,
  jumpToTasksPage,
  
  // Tags
  actionTags,
  keywordTags,
  selectedKeywordTags,
  suggestedTags,
  tagManagerVisible,
  openTagManager,
  closeTagManager,
  handleCreateTag,
  handleSelectTag,
  handleUnselectTag,
  handleDeleteTag,
  
  // Bisync LST Manager
  bisyncLstManagerVisible,
  bisyncLstManagerTaskId,
  bisyncLstManagerOptions,
  openBisyncLstManager,
  closeBisyncLstManager,
  
  // Task List Runtime
  runningTaskId,
  stoppedTaskId,
  runTask,
  goToAddTask,
  editTask,
  deleteTask,
  stopTaskAny,
  saveTaskSortOrders,
  openScheduleConfigForTask,
  getScheduleByTaskId,
  getTaskCardProgressByTask,
  openMenuId,
  showConfirm,
  confirmModal,
  closeConfirm,
  confirmAndClose,
  clearAllRunsWithConfirm,
  
  // Load Data
  loadData,
  loadActiveRuns,
  
  // History
  runs,
  runsTotal,
  taskRuns,
  runsPage,
  runsPageSize,
  jumpPage,
  filteredRuns,
  currentTotal,
  currentTotalPages,
  historyFilterTaskId,
  historyStatusFilter,
  setHistoryStatusFilter,
  setJumpPageValue,
  viewTaskHistory,
  prevRunsPage,
  nextRunsPage,
  backToTasks,
  jumpToPage,
  clearRun,
  clearAllRuns,
  
  // Run Detail
  showDetailModal,
  runDetail,
  openRunDetailModal,
  closeRunDetailModal,
  runFilesTotal,
  runFilesPage,
  openRunDetailFiles,
  pagedRunFiles,
  totalRunFilesPages,
  goPrevFilesPage,
  goNextFilesPage,
  getFinalSummaryFromComposable,
  finalCountAll,
  finalCountSuccess,
  finalCountFailed,
  showRunDetail,
  closeRunDetail,
  getRunProgressFromSummary,
  getRealtimeProgressByRun,
  
  // Global Stats
  globalStats,
  showGlobalStatsModal,
  closeGlobalStatsModal,
  
  // Running Hint
  runningHintVisible,
  runningHintRun,
  runningHintPhaseText,
  runningHintProgressText,
  openRunningHint,
  closeRunningHint,
  openRunningHintLog,
  
  // Log Modal
  showLogModal,
  logModalTitle,
  logContent,
  openRunLog,
  closeLogModal,
  
  // Webhook Modal
  showWebhookModal,
  webhookForm,
  setWebhook,
  saveWebhook,
  testWebhook,
  regenerateWebhookSecret,
  setWebhookTriggerId,
  setWebhookMatchText,
  setWebhookSecret,
  setWebhookPostUrl,
  setWebhookWecomUrl,
  setWebhookNotifyManual,
  setWebhookNotifySchedule,
  setWebhookNotifyWebhook,
  setWebhookStatusSuccess,
  setWebhookStatusFailed,
  setWebhookStatusHasTransfer,
  closeWebhookModal,
  
  // Singleton Modal
  showSingletonModal,
  singletonForm,
  setSingletonMode,
  saveSingleton,
  setSingletonEnabled,
  closeSingletonModal,
  
  // Schedule Config Modal
  scheduleConfigVisible,
  scheduleConfigSaving,
  scheduleConfigDraft,
  scheduleConfigTitle,
  closeScheduleConfig,
  saveScheduleConfig,
  
  // Task Editor
  taskEditorVisible,
  taskEditorTitle,
  commandMode,
  commandText,
  createForm,
  remotes,
  editingTask,
  showAdvancedOptions,
  creatingState,
  createTask,
  sourcePathOptions,
  targetPathOptions,
  showSourcePathInput,
  showTargetPathInput,
  sourceCurrentPath,
  targetCurrentPath,
  sourceBreadcrumbs,
  targetBreadcrumbs,
  setShowSourcePathInput,
  setShowTargetPathInput,
  resetTaskPathBrowse,
  restoreTaskPathBrowse,
  onSourceRemoteChange,
  onTargetRemoteChange,
  onSourceBreadcrumbClick,
  onTargetBreadcrumbClick,
  onSourceClick,
  onSourceArrow,
  onTargetClick,
  onTargetArrow,
  setCommandMode,
  setCommandText,
  setShowAdvancedOptions,
  closeTaskEditorModal,
  
  // Active Transfer
  activeTransferVisible,
  activeTransferTrackingMode,
  activeTransferSummary,
  activeTransferCurrentFile,
  activeTransferCurrentFiles,
  activeTransferSlots,
  activeTransferCompletedItems,
  activeTransferPendingItems,
  activeTransferCompletedTotal,
  activeTransferPendingTotal,
  activeTransferCompletedPage,
  activeTransferPendingPage,
  activeTransferCompletedJumpPage,
  activeTransferPendingJumpPage,
  activeTransferCompletedTotalPages,
  activeTransferPendingTotalPages,
  activeTransferDegraded,
  activeTransferLoading,
  activeTransferError,
  openActiveTransfer,
  closeActiveTransfer,
  refreshActiveTransfer,
  prevActiveTransferCompletedPage,
  nextActiveTransferCompletedPage,
  jumpActiveTransferCompletedPage,
  prevActiveTransferPendingPage,
  nextActiveTransferPendingPage,
  jumpActiveTransferPendingPage,
  
  // Formatters
  formatTime,
  getStatusClass,
  getStatusText,
  formatBps,
  formatBytes,
  formatBytesPerSec,
  formatEta,
} = useTaskView()
</script>

<template>
  <ToastCenter :toasts="toasts" />

  <TaskListViewShell
    v-if="currentModule !== 'history'"
    :task-search="taskSearch"
    :all-tasks="tasks"
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
    :set-task-search="setTaskSearch"
    :go-to-add-task="goToAddTask"
    :run-task="runTask"
    :edit-task="editTask"
    :delete-task="deleteTask"
    :open-schedule-config="openScheduleConfigForTask"
    :view-task-history="viewTaskHistory"
    :stop-task-any="stopTaskAny"
    :set-webhook="setWebhook"
    :set-singleton-mode="setSingletonMode"
    :save-task-sort-orders="saveTaskSortOrders"
    :open-transfer-detail="openActiveTransfer"
    :prev-tasks-page="prevTasksPage"
    :next-tasks-page="nextTasksPage"
    :set-tasks-jump-page-value="setTasksJumpPageValue"
    :jump-to-tasks-page="jumpToTasksPage"
    :show-webhook-modal="showWebhookModal"
    :webhook-form="webhookForm"
    :set-webhook-trigger-id="setWebhookTriggerId"
    :set-webhook-match-text="setWebhookMatchText"
    :set-webhook-secret="setWebhookSecret"
    :set-webhook-post-url="setWebhookPostUrl"
    :set-webhook-wecom-url="setWebhookWecomUrl"
    :set-webhook-notify-manual="setWebhookNotifyManual"
    :set-webhook-notify-schedule="setWebhookNotifySchedule"
    :set-webhook-notify-webhook="setWebhookNotifyWebhook"
    :set-webhook-status-success="setWebhookStatusSuccess"
    :set-webhook-status-failed="setWebhookStatusFailed"
    :set-webhook-status-has-transfer="setWebhookStatusHasTransfer"
    :save-webhook="saveWebhook"
    :test-webhook="testWebhook"
    :regenerate-webhook-secret="regenerateWebhookSecret"
    :close-webhook-modal="closeWebhookModal"
    :show-singleton-modal="showSingletonModal"
    :singleton-form="singletonForm"
    :set-singleton-enabled="setSingletonEnabled"
    :save-singleton="saveSingleton"
    :close-singleton-modal="closeSingletonModal"
    :action-tags="actionTags"
    :selected-keyword-tags="selectedKeywordTags"
    @open-tag-manager="openTagManager"
    @open-bisync-lst-manager="openBisyncLstManager"
  />

  <TaskHistoryViewShell
    v-if="currentModule === 'history'"
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
    :get-final-summary-from-composable="getFinalSummaryFromComposable"
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
    :paged-run-files="pagedRunFiles"
    :run-files-total="runFilesTotal"
    :run-files-page="runFilesPage"
    :total-run-files-pages="totalRunFilesPages"
    :back-to-tasks="backToTasks"
    :prev-runs-page="prevRunsPage"
    :next-runs-page="nextRunsPage"
    :set-history-status-filter="setHistoryStatusFilter"
    :set-jump-page-value="setJumpPageValue"
    :jump-to-page="jumpToPage"
    :clear-all-runs-with-confirm="clearAllRunsWithConfirm"
    :show-run-detail="showRunDetail"
    :open-run-log="openRunLog"
    :clear-run="clearRun"
    :close-run-detail="closeRunDetail"
    :go-prev-files-page="goPrevFilesPage"
    :go-next-files-page="goNextFilesPage"
    :show-log-modal="showLogModal"
    :log-modal-title="logModalTitle"
    :log-content="logContent"
    :close-log-modal="closeLogModal"
    :running-hint-visible="runningHintVisible"
    :running-hint-run="runningHintRun"
    :running-hint-phase-text="runningHintPhaseText"
    :running-hint-progress-text="runningHintProgressText"
    :close-running-hint="closeRunningHint"
    :open-running-hint-log="openRunningHintLog"
  />

  <TaskEditorViewShell
    :visible="taskEditorVisible"
    :title="taskEditorTitle"
    :command-mode="commandMode"
    :command-text="commandText"
    :create-form="createForm"
    :remotes="remotes"
    :show-source-path-input="showSourcePathInput"
    :show-target-path-input="showTargetPathInput"
    :source-breadcrumbs="sourceBreadcrumbs"
    :source-current-path="sourceCurrentPath"
    :source-path-options="sourcePathOptions"
    :target-breadcrumbs="targetBreadcrumbs"
    :target-current-path="targetCurrentPath"
    :target-path-options="targetPathOptions"
    :show-advanced-options="showAdvancedOptions"
    :creating-state="creatingState"
    :editing-task="editingTask"
    :set-command-mode="setCommandMode"
    :set-command-text="setCommandText"
    :set-show-source-path-input="setShowSourcePathInput"
    :set-show-target-path-input="setShowTargetPathInput"
    :on-source-remote-change="onSourceRemoteChange"
    :on-target-remote-change="onTargetRemoteChange"
    :on-source-breadcrumb-click="onSourceBreadcrumbClick"
    :on-target-breadcrumb-click="onTargetBreadcrumbClick"
    :on-source-arrow="onSourceArrow"
    :on-source-click="onSourceClick"
    :on-target-arrow="onTargetArrow"
    :on-target-click="onTargetClick"
    :create-task="createTask"
    :close-editor-modal="closeTaskEditorModal"
  />

  <ScheduleConfigModal
    :visible="scheduleConfigVisible"
    :title="scheduleConfigTitle"
    :model-value="scheduleConfigDraft"
    :saving="scheduleConfigSaving"
    @update:model-value="scheduleConfigDraft = $event"
    @save="saveScheduleConfig"
    @close="closeScheduleConfig"
  />

  <TransferringModal
    :visible="activeTransferVisible"
    :tracking-mode="activeTransferTrackingMode"
    :summary="activeTransferSummary"
    :current-file="activeTransferCurrentFile"
    :current-files="activeTransferCurrentFiles"
    :transfer-slots="activeTransferSlots"
    :completed-items="activeTransferCompletedItems"
    :pending-items="activeTransferPendingItems"
    :completed-total="activeTransferCompletedTotal"
    :pending-total="activeTransferPendingTotal"
    :completed-page="activeTransferCompletedPage"
    :pending-page="activeTransferPendingPage"
    :completed-jump-page="activeTransferCompletedJumpPage"
    :pending-jump-page="activeTransferPendingJumpPage"
    :completed-total-pages="activeTransferCompletedTotalPages"
    :pending-total-pages="activeTransferPendingTotalPages"
    :degraded="activeTransferDegraded"
    :loading="activeTransferLoading"
    :error="activeTransferError"
    @close="closeActiveTransfer"
    @prev-completed-page="prevActiveTransferCompletedPage"
    @next-completed-page="nextActiveTransferCompletedPage"
    @jump-completed-page="jumpActiveTransferCompletedPage"
    @update:completed-jump-page="activeTransferCompletedJumpPage = $event"
    @prev-pending-page="prevActiveTransferPendingPage"
    @next-pending-page="nextActiveTransferPendingPage"
    @jump-pending-page="jumpActiveTransferPendingPage"
    @update:pending-jump-page="activeTransferPendingJumpPage = $event"
  />

  <!-- 全局实时数据弹窗 -->
  <GlobalStatsModal
    :visible="showGlobalStatsModal"
    :stats="globalStats"
    :format-bytes="formatBytes"
    :format-bytes-per-sec="formatBytesPerSec"
    :format-eta="formatEta"
    @close="closeGlobalStatsModal"
  />

  <TagManagerModal
    :visible="tagManagerVisible"
    :suggested-tags="suggestedTags"
    :selected-keyword-tags="selectedKeywordTags"
    :action-tags="actionTags"
    @create-tag="handleCreateTag"
    @select-tag="handleSelectTag"
    @unselect-tag="handleUnselectTag"
    @delete-tag="handleDeleteTag"
    @close="closeTagManager"
  />

  <BisyncLstManagerModal
    :visible="bisyncLstManagerVisible"
    :task-id="bisyncLstManagerTaskId"
    :bisync-options="bisyncLstManagerOptions"
    @close="closeBisyncLstManager"
  />

  <ConfirmModal
    :visible="confirmModal.show"
    :title="confirmModal.title"
    :message="confirmModal.message"
    @confirm="confirmAndClose"
    @close="closeConfirm"
  />
</template>

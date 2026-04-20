<script setup lang="ts">
const props = defineProps<{
  runningHintDebugEnabled?: boolean
}>()

import GlobalStatsModal from '../components/task/GlobalStatsModal.vue'
import ConfirmModal from '../components/task/ConfirmModal.vue'
import TaskListViewShell from '../components/task/TaskListViewShell.vue'
import TaskHistoryViewShell from '../components/task/TaskHistoryViewShell.vue'
import TaskEditorViewShell from '../components/task/TaskEditorViewShell.vue'
import ToastCenter from '../components/toast/ToastCenter.vue'
import { taskApi, remoteApi, runApi, jobApi, scheduleApi } from '../composables/useApi'
import { setErrorHandler } from '../composables/useError'
import { formatBytes, formatBytesPerSec, formatEta } from '../utils/format'
import { useRunningHintRuntime } from '../composables/useRunningHintRuntime'
import { useTaskHistoryRuntime } from '../composables/useTaskHistoryRuntime'
import { useTaskViewRuntime } from '../composables/useTaskViewRuntime'
import { useTaskViewState } from '../composables/useTaskViewState'
import { useTaskViewRuntimeState } from '../composables/useTaskViewRuntimeState'
import { useTaskViewAuxRuntime } from '../composables/useTaskViewAuxRuntime'
import { useRunDetailRuntime } from '../composables/useRunDetailRuntime'
import { useRunDetailEntry } from '../composables/useRunDetailEntry'
import { useTaskFormNormalize } from '../composables/useTaskFormNormalize'
import { useTaskFormRuntime } from '../composables/useTaskFormRuntime'
import { useTaskListRuntime } from '../composables/useTaskListRuntime'
import { useTaskListView } from '../composables/useTaskListView'
import { useTaskViewPagingBridge } from '../composables/useTaskViewPagingBridge'
import { useTaskViewModalBindings } from '../composables/useTaskViewModalBindings'
import { useToastCenter } from '../composables/useToastCenter'
import { parseRcloneCommand } from '../composables/useTaskCommandParse'

const { toasts, showToast } = useToastCenter()
const { normalizeTaskOptions } = useTaskFormNormalize()

// Set up global error handler for composables
setErrorHandler((message, type) => {
  showToast(message, type as 'info' | 'success' | 'error')
})

const {
  tasks,
  schedules,
  runs,
  runsTotal,
  taskRuns,
  runsPage,
  runsPageSize,
  jumpPage,
  remotes,
  currentModule,
  historyFilterTaskId,
  historyStatusFilter,
} = useTaskViewState()

const {
  activeRuns,
  globalStats,
  showGlobalStatsModal,
  activeRunLookup,
  lastNonDecreasingTotalsByTask,
  STUCK_MS,
} = useTaskViewRuntimeState()

const {
  tasksPage,
  tasksPageSize,
  tasksJumpPage,
  taskSearch,
  tasksTotal,
  currentTasksPages,
  filteredTasks,
  jumpToTasksPage,
} = useTaskListView(tasks)

const {
  showDetailModal,
  runDetail,
  openRunDetailModal,
  closeRunDetailModal,
  runFilesPage,
  openRunDetailFiles,
  pagedRunFiles,
  totalRunFilesPages,
  goPrevFilesPage,
  goNextFilesPage,
  getFinalSummary: getFinalSummaryFromComposable,
  getPreflight: getPreflightFromComposable,
  finalFiles,
  finalCountAll,
  finalCountSuccess,
  finalCountFailed,
  finalCountOther,
  setFinalFilter,
  finalFilesTotal,
  totalFinalFilesPages,
  pagedFinalFiles,
  finalFilesJump,
  goPrevFinalFilesPage,
  goNextFinalFilesPage,
  jumpFinalFilesPage,
} = useRunDetailRuntime({ runApi })

const {
  loadData,
  loadActiveRuns,
  getRunProgressFromSummary,
  getRealtimeProgressByRun,
  getTaskCardProgressByTask,
  formatBps,
} = useTaskViewRuntime({
  tasks,
  remotes,
  schedules,
  runs,
  runsTotal,
  runsPage,
  runsPageSize,
  activeRuns,
  globalStats,
  showGlobalStatsModal,
  activeRunLookup,
  lastNonDecreasingTotalsByTask,
  currentModule,
  stuckMs: STUCK_MS,
  taskApi,
  remoteApi,
  scheduleApi,
  runApi,
  jobApi,
})

const {
  setTaskSearch,
  setTasksJumpPageValue,
  setHistoryStatusFilter,
  setJumpPageValue,
  setFinalFilesJumpValue,
  prevTasksPage,
  nextTasksPage,
  backToTasks,
  prevRunsPage,
  nextRunsPage,
} = useTaskViewPagingBridge({
  taskSearch,
  tasksJumpPage,
  historyStatusFilter,
  jumpPage,
  finalFilesJump,
  tasksPage,
  runsPage,
  currentModule,
  loadData,
})

const {
  showWebhookModal,
  webhookForm,
  setWebhook,
  saveWebhook,
  testWebhook,
  showSingletonModal,
  singletonForm,
  setSingletonMode,
  saveSingleton,
  showLogModal,
  logModalTitle,
  logContent,
  openRunLog,
  openMenuId,
  showConfirm,
  confirmModal,
  closeConfirm,
  confirmAndClose,
  formatTime,
  getStatusClass,
  getStatusText,
} = useTaskViewAuxRuntime({
  loadData,
  showToast,
  taskApi,
  getFinalSummary: getFinalSummaryFromComposable,
})

const openRunLogFromHint = (run: any) => openRunLog(run)

// 已移除"实时进度"弹窗逻辑，卡片直接显示稳态进度。
// 注意：这里依赖 openRunLog，因此必须放在 useTaskViewAuxRuntime 之后。
const {
  runningHintVisible,
  runningHintRun,
  runningHintDebugOpen,
  runningHintPhaseText,
  runningHintProgressText,
  runningHintDebugInfo,
  openRunningHint,
  closeRunningHint,
  toggleRunningHintDebug,
  openRunningHintLog,
} = useRunningHintRuntime(activeRuns, openRunLogFromHint, props.runningHintDebugEnabled === true)

const {
  showRunDetail,
  closeRunDetail,
} = useRunDetailEntry({
  openRunningHint,
  openRunDetailModal,
  openRunDetailFiles,
  closeRunDetailModal,
})

// webhook / singleton / editor modal 绑定桥只负责字段级 UI 接线；
// 这里同样依赖 aux runtime 暴露出的表单 ref，因此必须放在其后。
const {
  closeWebhookModal,
  closeSingletonModal,
  closeLogModal,
  closeGlobalStatsModal,
  setWebhookTriggerId,
  setWebhookPostUrl,
  setWebhookWecomUrl,
  setWebhookNotifyManual,
  setWebhookNotifySchedule,
  setWebhookNotifyWebhook,
  setWebhookStatusSuccess,
  setWebhookStatusFailed,
  setSingletonEnabled,
  setCommandMode,
  setCommandText,
  setShowAdvancedOptions,
} = useTaskViewModalBindings({
  showWebhookModal,
  webhookForm,
  showSingletonModal,
  singletonForm,
  showLogModal,
  commandMode,
  commandText,
  showAdvancedOptions,
  showGlobalStatsModal,
})

// move 模式时，成功数量代表 Moved 条数；已在后端合并 Copied+Deleted 为 Moved

const {
  createForm,
  commandMode,
  commandText,
  editingTask,
  showAdvancedOptions,
  resetTaskFormForCreate,
  fillTaskFormForEdit,
  getScheduleByTaskId,
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
} = useTaskFormRuntime({
  schedules,
  currentModule,
  normalizeTaskOptions,
  loadData,
  taskApi,
  scheduleApi,
  showToast,
  parseRcloneCommand,
})

const {
  filteredRuns,
  currentTotal,
  currentTotalPages,
  viewTaskHistory,
  jumpToPage,
  clearRun,
  clearAllRuns,
} = useTaskHistoryRuntime({
  runs,
  runsTotal,
  taskRuns,
  historyFilterTaskId,
  historyStatusFilter,
  runsPage,
  runsPageSize,
  jumpPage,
  currentModule,
  getFinalSummary: getFinalSummaryFromComposable,
  loadData,
  runApi,
})

const {
  deleteTask,
  toggleSchedule,
  clearAllRunsWithConfirm,
  runningTaskId,
  stoppedTaskId,
  stopTaskAny,
  runTask,
  goToAddTask,
  editTask,
} = useTaskListRuntime({
  openMenuId,
  historyFilterTaskId,
  schedules,
  loadData,
  loadActiveRuns,
  showConfirm,
  showToast,
  clearAllRuns,
  currentModule,
  remotes,
  remoteApi,
  resetTaskFormForCreate,
  resetTaskPathBrowse,
  getScheduleByTaskId,
  fillTaskFormForEdit,
  restoreTaskPathBrowse,
  taskApi,
  scheduleApi,
})

</script>


<template>
  <ToastCenter :toasts="toasts" />

  <TaskListViewShell
    v-if="currentModule === 'tasks'"
    :task-search="taskSearch"
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
    :toggle-schedule="toggleSchedule"
    :view-task-history="viewTaskHistory"
    :stop-task-any="stopTaskAny"
    :set-webhook="setWebhook"
    :set-singleton-mode="setSingletonMode"
    :prev-tasks-page="prevTasksPage"
    :next-tasks-page="nextTasksPage"
    :set-tasks-jump-page-value="setTasksJumpPageValue"
    :jump-to-tasks-page="jumpToTasksPage"
    :show-webhook-modal="showWebhookModal"
    :webhook-form="webhookForm"
    :set-webhook-trigger-id="setWebhookTriggerId"
    :set-webhook-post-url="setWebhookPostUrl"
    :set-webhook-wecom-url="setWebhookWecomUrl"
    :set-webhook-notify-manual="setWebhookNotifyManual"
    :set-webhook-notify-schedule="setWebhookNotifySchedule"
    :set-webhook-notify-webhook="setWebhookNotifyWebhook"
    :set-webhook-status-success="setWebhookStatusSuccess"
    :set-webhook-status-failed="setWebhookStatusFailed"
    :save-webhook="saveWebhook"
    :test-webhook="testWebhook"
    :close-webhook-modal="closeWebhookModal"
    :show-singleton-modal="showSingletonModal"
    :singleton-form="singletonForm"
    :set-singleton-enabled="setSingletonEnabled"
    :save-singleton="saveSingleton"
    :close-singleton-modal="closeSingletonModal"
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
    :set-final-filter="setFinalFilter"
    :go-prev-final-files-page="goPrevFinalFilesPage"
    :go-next-final-files-page="goNextFinalFilesPage"
    :set-final-files-jump-value="setFinalFilesJumpValue"
    :jump-final-files-page="jumpFinalFilesPage"
    :go-prev-files-page="goPrevFilesPage"
    :go-next-files-page="goNextFilesPage"
    :show-log-modal="showLogModal"
    :log-modal-title="logModalTitle"
    :log-content="logContent"
    :close-log-modal="closeLogModal"
    :running-hint-visible="runningHintVisible"
    :running-hint-run="runningHintRun"
    :running-hint-debug-enabled="runningHintDebugEnabled"
    :running-hint-debug-open="runningHintDebugOpen"
    :running-hint-phase-text="runningHintPhaseText"
    :running-hint-progress-text="runningHintProgressText"
    :running-hint-debug-info="runningHintDebugInfo"
    :close-running-hint="closeRunningHint"
    :toggle-running-hint-debug="toggleRunningHintDebug"
    :open-running-hint-log="openRunningHintLog"
  />

  <TaskEditorViewShell
    v-if="currentModule === 'add'"
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
    :set-show-advanced-options="setShowAdvancedOptions"
    :on-source-remote-change="onSourceRemoteChange"
    :on-target-remote-change="onTargetRemoteChange"
    :on-source-breadcrumb-click="onSourceBreadcrumbClick"
    :on-target-breadcrumb-click="onTargetBreadcrumbClick"
    :on-source-arrow="onSourceArrow"
    :on-source-click="onSourceClick"
    :on-target-arrow="onTargetArrow"
    :on-target-click="onTargetClick"
    :create-task="createTask"
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

  <!-- 确认删除弹窗 -->
  <ConfirmModal
    :visible="confirmModal.show"
    :title="confirmModal.title"
    :message="confirmModal.message"
    @close="closeConfirm"
    @confirm="confirmAndClose"
  />
</template>

<style scoped src="./TaskView.css"></style>

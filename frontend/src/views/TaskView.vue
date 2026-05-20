<script setup lang="ts">
import GlobalStatsModal from '../components/task/GlobalStatsModal.vue'
import ConfirmModal from '../components/task/ConfirmModal.vue'
import TaskListViewShell from '../components/task/TaskListViewShell.vue'
import TaskHistoryViewShell from '../components/task/TaskHistoryViewShell.vue'
import TaskEditorViewShell from '../components/task/TaskEditorViewShell.vue'
import ToastCenter from '../components/toast/ToastCenter.vue'
import ScheduleConfigModal from '../components/task/ScheduleConfigModal.vue'
import TransferringModal from '../components/task/transferring/TransferringModal.vue'
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
import { useActiveTransferDetail } from '../composables/useActiveTransferDetail'
import { computed } from 'vue'
import { t } from '../i18n'
import { useScheduleConfigModal } from '../composables/useScheduleConfigModal'

// 1) 页面级基础能力
const { toasts, showToast } = useToastCenter()
const { normalizeTaskOptions } = useTaskFormNormalize()
const {
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
} = useActiveTransferDetail()

// 全局错误提示统一走 toast；不在各子块里各自弹。
setErrorHandler((message, type) => {
  showToast(message, type as 'info' | 'success' | 'error')
})

// 2) 页面主状态（tasks / history / add 三视图共享）
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

// 3) 运行时状态（active runs / 全局统计 / lookup）
const {
  activeRuns,
  globalStats,
  showGlobalStatsModal,
  activeRunLookup,
  lastNonDecreasingTotalsByTask,
  STUCK_MS,
} = useTaskViewRuntimeState()

// 4) 任务列表页视图态（分页 / 搜索）
const {
  tasksPage,
  tasksPageSize,
  tasksJumpPage,
  taskSearch,
  tasksTotal,
  currentTasksPages,
  filteredTasksRaw,
  filteredTasks,
  jumpToTasksPage,
} = useTaskListView(tasks)

// 5) 运行详情 / 最终总结链
const {
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
  getFinalSummary: getFinalSummaryFromComposable,
  finalCountAll,
  finalCountSuccess,
  finalCountFailed,
  finalCountOther,
  setFinalFilter,
} = useRunDetailRuntime({ runApi })

// 6) 主数据加载与进度相关派生
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

// 7) 页面级 bridge：分页 / 返回 / 跳页
const {
  setTaskSearch,
  setTasksJumpPageValue,
  setHistoryStatusFilter,
  setJumpPageValue,
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
  tasksPage,
  runsPage,
  currentModule,
  loadData,
})

// 8) 弹窗 / 辅助动作运行时
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

// 9) running hint 桥：这里只做 log 入口转发，不承担主进度链职责
const openRunLogFromHint = (run: any) => openRunLog(run)

// 已移除"实时进度"弹窗逻辑，卡片直接显示稳态进度。
// 注意：这里依赖 openRunLog，因此必须放在 useTaskViewAuxRuntime 之后。
const {
  runningHintVisible,
  runningHintRun,
  runningHintPhaseText,
  runningHintProgressText,
  openRunningHint,
  closeRunningHint,
  openRunningHintLog,
} = useRunningHintRuntime(activeRuns, openRunLogFromHint)

// 10) 运行详情入口编排
const {
  showRunDetail,
  closeRunDetail,
} = useRunDetailEntry({
  openRunningHint,
  openRunDetailModal,
  openRunDetailFiles,
  closeRunDetailModal,
})

// 11) 任务编辑器运行时
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

// 12) 页面级 bridge：modal 字段绑定与开关态
// webhook / singleton / editor modal 绑定桥只负责字段级 UI 接线；
// 除了 aux runtime 暴露出的表单 ref 外，还依赖 task form runtime 的 command/advanced 状态，
// 因此必须放在 useTaskFormRuntime 之后，避免压缩后触发 TDZ（before initialization）。
const {
  closeWebhookModal,
  closeSingletonModal,
  closeLogModal,
  closeGlobalStatsModal,
  setWebhookTriggerId,
  setWebhookMatchText,
  setWebhookPostUrl,
  setWebhookWecomUrl,
  setWebhookNotifyManual,
  setWebhookNotifySchedule,
  setWebhookNotifyWebhook,
  setWebhookStatusSuccess,
  setWebhookStatusFailed,
  setWebhookStatusHasTransfer,
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

// 13) 历史列表运行时
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

// 14) 任务列表动作运行时
const {
  deleteTask,
  clearAllRunsWithConfirm,
  runningTaskId,
  stoppedTaskId,
  stopTaskAny,
  runTask,
  goToAddTask,
  editTask,
  saveTaskSortOrders,
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

const {
  scheduleConfigVisible,
  scheduleConfigSaving,
  scheduleConfigDraft,
  scheduleConfigTitle,
  openScheduleConfigForTask,
  closeScheduleConfig,
  saveScheduleConfig,
} = useScheduleConfigModal({
  createForm,
  getScheduleByTaskId,
  scheduleApi,
  loadData,
  showToast,
})

const taskEditorVisible = computed(() => currentModule.value === 'add')
const taskEditorTitle = computed(() => editingTask.value ? t('taskEditor.editTitle') : t('taskEditor.createTitle'))

const taskEditorSnapshot = computed(() => JSON.stringify({
  commandMode: commandMode.value,
  commandText: commandText.value,
  createForm: createForm.value,
  showSourcePathInput: showSourcePathInput.value,
  showTargetPathInput: showTargetPathInput.value,
  showAdvancedOptions: showAdvancedOptions.value,
}))

const taskEditorBaseline = computed(() => JSON.stringify({
  commandMode: false,
  commandText: '',
  createForm: {
    name: editingTask.value?.name ?? '',
    mode: editingTask.value ? createForm.value.mode : 'copy',
    sourceRemote: editingTask.value?.sourceRemote ?? '',
    sourcePath: editingTask.value?.sourcePath ?? '',
    targetRemote: editingTask.value?.targetRemote ?? '',
    targetPath: editingTask.value?.targetPath ?? '',
    options: editingTask.value ? createForm.value.options : { enableStreaming: true },
  },
  showSourcePathInput: false,
  showTargetPathInput: false,
  showAdvancedOptions: false,
}))

const hasTaskEditorChanges = computed(() => taskEditorSnapshot.value !== taskEditorBaseline.value)

function doCloseTaskEditorModal() {
  creatingState.value = 'idle'
  currentModule.value = 'tasks'
}

function closeTaskEditorModal() {
  if (creatingState.value === 'loading') return
  if (!hasTaskEditorChanges.value) {
    doCloseTaskEditorModal()
    return
  }
  showConfirm(
    t('taskEditor.closeConfirmTitle'),
    t('taskEditor.closeConfirmMessage'),
    () => doCloseTaskEditorModal(),
  )
}

</script>


<template>
  <ToastCenter :toasts="toasts" />

  <TaskListViewShell
    v-if="currentModule !== 'history'"
    :task-search="taskSearch"
    :filtered-tasks="filteredTasks"
    :all-tasks="tasks"
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
    :final-count-other="finalCountOther"
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
    :set-final-filter="setFinalFilter"
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
    @refresh="refreshActiveTransfer"
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

  <!-- 确认删除弹窗 -->
  <ConfirmModal
    :visible="confirmModal.show"
    :title="confirmModal.title"
    :message="confirmModal.message"
    @close="closeConfirm"
    @confirm="confirmAndClose"
  />
</template>

<style scoped>
.list .run-item { align-items:flex-start }
.form-content { padding: 20px; }
.form-content .field-item { margin-bottom: 16px; }
.form-content label { display: block; margin-bottom: 6px; font-size: 13px; color: var(--muted); }
.form-content input,
.form-content select {
  width: 100%;
  padding: 10px 14px;
  border-radius: 8px;
  border: 1px solid #333;
  background: #252525;
  color: #e0e0e0;
  font-size: 14px;
  box-sizing: border-box;
}
.cmd-textarea{ width:100%; min-height:120px; padding:12px 14px; border-radius:10px; border:1px solid var(--border); background:var(--surface); color:var(--text); font-size:14px; box-sizing:border-box; resize:vertical; }
:global(body.light .cmd-textarea){ background:var(--surface); border-color:var(--border); color:var(--text) }
.form-content label.inline-label{ display:flex !important; align-items:center; gap:8px; margin:0 0 6px 0; }
.form-content label.inline-label input[type="checkbox"]{ width:16px; height:16px; }
:global(body.light .form-content input),
:global(body.light .form-content select) { background: #fff; border-color: #ddd; color: #333; }
.form-actions { margin-top: 20px; }
.btn-success { background: #2e7d32 !important; border-color: #2e7d32 !important; }
.btn-success:hover { background: #388e3c !important; }
.btn-running { background: #2e7d32 !important; border-color: #2e7d32 !important; color: #fff !important; }
.tile-grid { display: flex; flex-wrap: wrap; gap: 12px; padding: 16px 20px; }
.tile {
  min-width: 180px;
  padding: 14px 16px;
  border-radius: 10px;
  background: #252525;
  cursor: pointer;
  transition: all 0.2s;
  flex: 1 1 calc(25% - 12px);
  max-width: calc(25% - 12px);
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.tile:hover { background: #2a2a2a; }
:global(body.light .tile) { background: #f5f5f5; }
:global(body.light .tile:hover) { background: #e8e8e8; }
.tile-info { flex: 1; overflow: hidden; }
.tile-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 6px; }
.tile-name { font-weight: 600; font-size: 14px; color: #fff; }
:global(body.light .tile-name) { color: #1a1a1a; }
.tile-desc { font-size: 12px; color: #888; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.tile-actions { display: flex; align-items: center; gap: 4px; position: relative; }
.tile-actions .ghost.small { padding: 4px 8px; font-size: 12px; }
.menu-btn { font-size: 16px !important; padding: 4px 8px !important; }
.tile-menu {
  position: absolute;
  right: 0;
  top: 100%;
  background: #333;
  border: 1px solid #444;
  border-radius: 8px;
  padding: 4px;
  z-index: 100;
  min-width: 100px;
  box-shadow: 0 4px 12px rgba(0,0,0,0.3);
}
:global(body.light .tile-menu) { background: #fff; border-color: #ddd; }
.tile-menu button { width: 100%; text-align: left; padding: 8px 12px; }
.tile-menu button:hover { background: #444; }
:global(body.light .tile-menu button:hover) { background: #f0f0f0; }
.error-text { color: #ff6b6b; white-space: pre-wrap; }
.danger-hint { color: #ff6b6b; font-size: 13px; line-height: 1.5; }
.modal-content.log-modal{width:92vw !important; max-width:1200px !important; max-height:80vh; display:flex; flex-direction:column;}
.log-modal .modal-body{padding:12px 16px; width:100%; flex:1; overflow:hidden; display:flex;}
.log-box{width:100%; display:flex; justify-content:center;}
.log-pre{background:#0b1220;color:#e5e7eb;padding:12px;border-radius:8px;height:100%;overflow:auto;white-space:pre-wrap;width:calc(100% - 64px);max-width:1100px;box-sizing:border-box;margin:0;border:1px solid #334155}
.detail-modal{ width: 135% !important; max-width: 1200px !important; }
.summary-box{background:#111827;border:1px solid #333;border-radius:10px;padding:12px 14px;margin-top:6px;max-width:1200px}
.summary-title{font-weight:600;color:#e0e0e0;margin-bottom:8px}
.summary-grid{display:grid;grid-template-columns:repeat(4,1fr);gap:12px}
.summary-cell{background:#0f172a;border:1px solid #334155;border-radius:10px;padding:10px 12px}
.summary-key{font-size:12px;color:#94a3b8;margin-bottom:6px}
.summary-val{font-size:16px;color:#e2e8f0;font-weight:700}
:global(body.light .summary-box){background:#ffffff;border-color:#e5e7eb}
:global(body.light .summary-title){color:#111827}
:global(body.light .summary-cell){background:#f8fafc;border-color:#e5e7eb}
:global(body.light .summary-key){color:#64748b}
:global(body.light .summary-val){color:#111827}
.summary-cell.clickable{cursor:pointer;transition:background .15s ease,border-color .15s ease,box-shadow .15s ease}
.summary-cell.clickable:hover{background:#1f2937;border-color:#475569;box-shadow:0 0 0 1px #334155 inset}
:global(body.light .summary-cell.clickable:hover){background:#f0f4f8;border-color:#cbd5e1;box-shadow:0 0 0 1px #cbd5e1 inset}
.files-table{margin-top:14px;border:1px solid #333;border-radius:10px;overflow:hidden}
.files-table.large{max-width:1200px}
.files-header,.files-row{display:grid;grid-template-columns:1fr 140px 200px 140px;gap:18px;align-items:center}
.files-header{padding:12px 16px;background:#252525;color:#cbd5e1;font-size:13px}
.files-body{max-height:540px;overflow:auto}
:global(body.light .files-table){border-color:#e5e7eb}
:global(body.light .files-header){background:#f5f5f5;color:#4b5563}
.files-table .status{width:auto;padding:0;background:transparent;border-radius:0;font-weight:600}
.files-table .status.success{color:#34d399}
.files-table .status.failed{color:#f87171}
.files-table .status.skipped{color:#fbbf24}
.pager-inline{display:flex;align-items:center;gap:8px;flex-wrap:wrap}
.files-pager{display:flex;align-items:center;gap:8px;margin-top:10px;flex-wrap:wrap}
.pager-inline button,
.files-pager button,
.pager-inline span,
.files-pager span{white-space:nowrap}
.jump-input{width:72px;min-width:72px;padding:4px 8px}
@media (max-width: 1100px){
  .summary-grid{grid-template-columns:repeat(2,1fr)}
  .files-toolbar{flex-direction:column;align-items:flex-start}
}
@media (max-width: 768px) {
  .task-main {
    flex-direction: column;
    align-items: flex-start;
    gap: 8px;
  }
  .task-main .name,
  .task-main .schedule-info,
  .task-main .item-actions {
    width: 100%;
  }
  .task-main .item-actions button {
    min-width: 0;
  }
}
.toast-container {
  position: fixed;
  top: 20px;
  right: 20px;
  z-index: 10000;
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.toast {
  padding: 12px 20px;
  border-radius: 8px;
  font-size: 14px;
  min-width: 200px;
  max-width: 400px;
  box-shadow: 0 4px 12px rgba(0,0,0,0.15);
  animation: slideIn 0.2s ease;
}
.toast.info { background: #3b82f6; color: #fff; }
.toast.success { background: #10b981; color: #fff; }
.toast.error { background: #ef4444; color: #fff; }
@keyframes slideIn {
  from { transform: translateX(100%); opacity: 0; }
  to { transform: translateX(0); opacity: 1; }
}
.title.clickable, .card-header.clickable {
  cursor: pointer;
}
.title.clickable:hover, .card-header.clickable:hover {
  color: var(--accent, #4f46e5);
}
</style>

<script setup lang="ts">
import { RunningHintModal, GlobalStatsModal, RunLogModal, SingletonConfigModal, WebhookConfigModal, ConfirmModal, TaskListSection, TaskHistorySection, AddTaskForm } from '../components/task'
import { ToastItem } from '../components/toast'
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
import { useToastCenter } from '../composables/useToastCenter'
import { parseRcloneCommand } from '../composables/useTaskCommandParse'

const { toasts, showToast } = useToastCenter()
const { normalizeTaskOptions } = useTaskFormNormalize()

// Set up global error handler for composables
setErrorHandler((message, type) => {
  showToast(message, type as 'info' | 'success' | 'error')
})

const closeWebhookModal = () => { showWebhookModal.value = false }
const closeSingletonModal = () => { showSingletonModal.value = false }
const closeLogModal = () => { showLogModal.value = false }
const closeGlobalStatsModal = () => { showGlobalStatsModal.value = false }
const setTaskSearch = (value: string) => { taskSearch.value = value }
const setTasksJumpPageValue = (value: number | null) => { tasksJumpPage.value = value }
const setHistoryStatusFilter = (value: string) => { historyStatusFilter.value = value }
const setJumpPageValue = (value: number) => { jumpPage.value = value }
const setFinalFilesJumpValue = (value: number | null) => { finalFilesJump.value = value }
const setWebhookTriggerId = (value: string) => { webhookForm.triggerId = value }
const setWebhookPostUrl = (value: string) => { webhookForm.postUrl = value }
const setWebhookWecomUrl = (value: string) => { ;(webhookForm as any).wecomUrl = value }
const setWebhookNotifyManual = (value: boolean) => { webhookForm.notify.manual = value }
const setWebhookNotifySchedule = (value: boolean) => { webhookForm.notify.schedule = value }
const setWebhookNotifyWebhook = (value: boolean) => { webhookForm.notify.webhook = value }
const setWebhookStatusSuccess = (value: boolean) => { ;(webhookForm as any).status.success = value }
const setWebhookStatusFailed = (value: boolean) => { ;(webhookForm as any).status.failed = value }
const setSingletonEnabled = (value: boolean) => { singletonForm.singletonEnabled = value }
const setCommandMode = (value: boolean) => { commandMode.value = value }
const setCommandText = (value: string) => { commandText.value = value }
const setShowAdvancedOptions = (value: boolean) => { showAdvancedOptions.value = value }

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
  lastStableByTask,
  lastNonDecreasingTotalsByTask,
  LINGER_MS,
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

// 已移除"实时进度"弹窗逻辑，卡片直接显示稳态进度
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
} = useRunningHintRuntime(activeRuns, openRunLog)

const {
  showRunDetail,
  closeRunDetail,
} = useRunDetailEntry({
  openRunningHint,
  openRunDetailModal,
  openRunDetailFiles,
  closeRunDetailModal,
})
const {
  loadData,
  loadActiveRuns,
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
  lastStableByTask,
  lastNonDecreasingTotalsByTask,
  currentModule,
  lingerMs: LINGER_MS,
  stuckMs: STUCK_MS,
  taskApi,
  remoteApi,
  scheduleApi,
  runApi,
  jobApi,
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
  <!-- Toast 通知容器 -->
  <div class="toast-container">
    <ToastItem v-for="toast in toasts" :key="toast.id" :toast="toast" />
  </div>

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
    :prev-tasks-page="() => { tasksPage-- }"
    :next-tasks-page="() => { tasksPage++ }"
    :set-tasks-jump-page-value="setTasksJumpPageValue"
    :jump-to-tasks-page="jumpToTasksPage"
  />

  <WebhookConfigModal
    :visible="showWebhookModal"
    :trigger-id="webhookForm.triggerId"
    :post-url="webhookForm.postUrl"
    :wecom-url="(webhookForm as any).wecomUrl"
    :notify-manual="webhookForm.notify.manual"
    :notify-schedule="webhookForm.notify.schedule"
    :notify-webhook="webhookForm.notify.webhook"
    :status-success="(webhookForm as any).status.success"
    :status-failed="(webhookForm as any).status.failed"
    :can-test="!!webhookForm.postUrl || !!(webhookForm as any).wecomUrl"
    @update:trigger-id="setWebhookTriggerId"
    @update:post-url="setWebhookPostUrl"
    @update:wecom-url="setWebhookWecomUrl"
    @update:notify-manual="setWebhookNotifyManual"
    @update:notify-schedule="setWebhookNotifySchedule"
    @update:notify-webhook="setWebhookNotifyWebhook"
    @update:status-success="setWebhookStatusSuccess"
    @update:status-failed="setWebhookStatusFailed"
    @save="saveWebhook"
    @test="testWebhook"
    @close="closeWebhookModal"
  />

  <!-- 单例模式配置弹窗 -->
  <SingletonConfigModal
    :visible="showSingletonModal"
    :singleton-enabled="singletonForm.singletonEnabled"
    @update:singleton-enabled="setSingletonEnabled"
    @save="saveSingleton"
    @close="closeSingletonModal"
  />

  <!-- 传输日志弹窗 -->
  <RunLogModal
    :visible="showLogModal"
    :title="logModalTitle"
    :content="logContent"
    @close="closeLogModal"
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
    :get-realtime-progress-by-run="getRealtimeProgressByRun"
    :get-final-summary-from-composable="getFinalSummaryFromComposable"
    :show-detail-modal="showDetailModal"
    :run-detail="runDetail"
    :get-status-class="getStatusClass"
    :get-status-text="getStatusText"
    :get-preflight-from-composable="getPreflightFromComposable"
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
    :back-to-tasks="() => { currentModule = 'tasks' }"
    :prev-runs-page="() => { runsPage--; loadData() }"
    :next-runs-page="() => { runsPage++; loadData() }"
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
  />

  <!-- 运行中轻量提示小窗（不切主窗口） -->
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

<style scoped>
/* 让卡片更宽松，展示 chips */
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
body.light .cmd-textarea{ background:var(--surface); border-color:var(--border); color:var(--text) }
.form-content label.inline-label{ display:flex !important; align-items:center; gap:8px; margin:0 0 6px 0; }
.form-content label.inline-label input[type="checkbox"]{ width:16px; height:16px; }
body.light .form-content input,
body.light .form-content select { background: #fff; border-color: #ddd; color: #333; }
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
body.light .tile { background: #f5f5f5; }
body.light .tile:hover { background: #e8e8e8; }
.tile-info { flex: 1; overflow: hidden; }
.tile-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 6px; }
.tile-name { font-weight: 600; font-size: 14px; color: #fff; }
body.light .tile-name { color: #1a1a1a; }
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
body.light .tile-menu { background: #fff; border-color: #ddd; }
.tile-menu button { width: 100%; text-align: left; padding: 8px 12px; }
.tile-menu button:hover { background: #444; }
body.light .tile-menu button:hover { background: #f0f0f0; }
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
body.light .summary-box{background:#ffffff;border-color:#e5e7eb}
body.light .summary-title{color:#111827}
body.light .summary-cell{background:#f8fafc;border-color:#e5e7eb}
body.light .summary-key{color:#64748b}
body.light .summary-val{color:#111827}
.summary-cell.clickable{cursor:pointer;transition:background .15s ease,border-color .15s ease,box-shadow .15s ease}
.summary-cell.clickable:hover{background:#1f2937;border-color:#475569;box-shadow:0 0 0 1px #334155 inset}
body.light .summary-cell.clickable:hover{background:#f0f4f8;border-color:#cbd5e1;box-shadow:0 0 0 1px #cbd5e1 inset}
.files-table{margin-top:14px;border:1px solid #333;border-radius:10px;overflow:hidden}
.files-table.large{max-width:1200px}
.files-header,.files-row{display:grid;grid-template-columns:1fr 140px 200px 140px;gap:18px;align-items:center}
.files-header{padding:12px 16px;background:#252525;color:#cbd5e1;font-size:13px}
.files-body{max-height:540px;overflow:auto}
body.light .files-table{border-color:#e5e7eb}
body.light .files-header{background:#f5f5f5;color:#4b5563}
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

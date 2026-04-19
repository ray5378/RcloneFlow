<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import * as api from '../api'
import { TaskCard, RunItem, ScheduleOptions, AdvancedOptions, RunningHintModal, TaskHistoryPanel, RunDetailModal, AddTaskForm } from '../components/task'
import { getActiveProgress as getHintActiveProgress, getActiveProgressText as getHintActiveProgressText } from '../components/task/runningHint'
import { ToastItem } from '../components/toast'
import { FileItem } from '../components/files'
import { taskApi, remoteApi, runApi, jobApi, scheduleApi } from '../composables/useApi'
import { handleError, showSuccess, setErrorHandler } from '../composables/useError'
import { formatBytes, formatBytesPerSec, formatDuration, formatEta } from '../utils/format'
import { getToken } from '../api/auth'
import { useWebSocket, onWsMessage } from '../composables/useWebSocket'
import { useActiveRunLookup } from '../composables/useActiveRunLookup'
import { useRunningHint } from '../composables/useRunningHint'
import { useTaskHistoryComputed } from '../composables/useTaskHistoryComputed'
import { useTaskHistoryLoader } from '../composables/useTaskHistoryLoader'
import { useTaskHistoryActions } from '../composables/useTaskHistoryActions'
import { useTaskListActions } from '../composables/useTaskListActions'
import { useTaskRunActions } from '../composables/useTaskRunActions'
import { useTaskViewDataSync } from '../composables/useTaskViewDataSync'
import { useTaskProgressSync } from '../composables/useTaskProgressSync'
import { useTaskViewRefreshLifecycle } from '../composables/useTaskViewRefreshLifecycle'
import { useRunDetailComputed } from '../composables/useRunDetailComputed'
import { useRunDetailFiles } from '../composables/useRunDetailFiles'
import { useRunDetailState } from '../composables/useRunDetailState'
import { useRunDetailEntry } from '../composables/useRunDetailEntry'
import { useTaskFormState } from '../composables/useTaskFormState'
import { useTaskFormSubmit } from '../composables/useTaskFormSubmit'
import { useTaskFormPrepare } from '../composables/useTaskFormPrepare'
import { useTaskFormFlow } from '../composables/useTaskFormFlow'
import { useTaskPathBrowse } from '../composables/useTaskPathBrowse'
import { useTaskFormEntry } from '../composables/useTaskFormEntry'
import { useTaskListView } from '../composables/useTaskListView'
import { useTaskViewUi } from '../composables/useTaskViewUi'
import { parseRcloneCommand } from '../composables/useTaskCommandParse'
import { getDeNoisedStableByRun as buildDeNoisedStableByRun, getDeNoisedStableByTask as buildDeNoisedStableByTask } from '../composables/activeRunProgress'
import type { Task, Schedule, Run } from '../types'

// Toast 通知系统
interface Toast {
  id: number
  message: string
  type: 'info' | 'success' | 'error'
}
const toasts = ref<Toast[]>([])
let toastId = 0

function showToast(message: string, type: 'info' | 'success' | 'error' = 'info') {
  const id = ++toastId
  toasts.value.push({ id, message, type })
  setTimeout(() => {
    toasts.value = toasts.value.filter(t => t.id !== id)
  }, 3000)
}

// Set up global error handler for composables
setErrorHandler((message, type) => {
  showToast(message, type as 'info' | 'success' | 'error')
})

const tasks = ref<Task[]>([])
const schedules = ref<Schedule[]>([])
const runs = ref<Run[]>([])
const runsTotal = ref(0)
// 任务特定的历史记录（用于分页）
const taskRuns = ref<Run[]>([])
const runsPage = ref(1)
const runsPageSize = 50
const jumpPage = ref(1)

function jumpToPage() {
  const page = Math.min(Math.max(1, jumpPage.value || 1), currentTotalPages.value)
  runsPage.value = page
  jumpPage.value = page
  loadData()
}

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

const remotes = ref<string[]>([])
const currentModule = ref<'history' | 'add' | 'tasks'>('tasks')
const historyFilterTaskId = ref<number | null>(null)
const historyStatusFilter = ref<string>('all') // 'all' | 'finished' | 'failed' | 'skipped' | 'hasTransfer'
const { showDetailModal, runDetail, openRunDetailModal, closeRunDetailModal } = useRunDetailState()

const showCreateModal = ref(false)
const showGlobalStatsModal = ref(false)
const globalStats = ref<any>({}) // 全局实时数据保留为独立弹窗，与历史态无关
// 已移除"实时进度"弹窗逻辑，卡片直接显示稳态进度

const activeRuns = ref<any[]>([])
const activeRunLookup = useActiveRunLookup(activeRuns)
const {
  visible: runningHintVisible,
  run: runningHintRun,
  debugOpen: runningHintDebugOpen,
  phaseText: runningHintPhaseText,
  progressText: runningHintProgressText,
  debugInfo: runningHintDebugInfo,
  open: openRunningHint,
  close: closeRunningHint,
  toggleDebug: toggleRunningHintDebug,
  openLog: openRunningHintLog,
} = useRunningHint(activeRuns, openRunLog)
// 任务卡片：完成后保留最近稳态进度的观察期（默认 15s）
const LINGER_MS = 20000
const lastStableByTask = ref<Record<number, { sp:any; at:number }>>({})
const lastNonDecreasingTotalsByTask = ref<Record<number, { totalBytes:number; totalCount:number }>>({})
const {
  loadData,
  loadActiveRuns,
  loadGlobalStats,
  openGlobalStats,
  setupRealtimeSync,
} = useTaskViewDataSync({
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
  lastStableByTask,
  lastNonDecreasingTotalsByTask,
  taskApi,
  remoteApi,
  scheduleApi,
  runApi,
  jobApi,
})
// 监控帧是否停滞超过阈值（默认 25s），若是则强制刷新一次
const STUCK_MS = 25000
let lastRenderedSignature = ''
let stuckTimer: number | null = null
let activePollTimer: number | null = null
onMounted(() => {
  stuckTimer = window.setInterval(() => {
    try{
      // 用任务卡片可见的核心字段拼接签名：任务数、每个任务的 id+pct+completedFiles
      const sigParts: string[] = []
      for (const t of tasks.value||[]){
        const sp = getDeNoisedStableByTask((t as any).id) as any
        const pct = sp ? Number(sp.percentage||0).toFixed(3) : 'na'
        const c = sp ? Number(sp.completedFiles||0) : -1
        sigParts.push(`${(t as any).id}:${pct}:${c}`)
      }
      const sig = `${tasks.value?.length||0}|${sigParts.join(',')}`
      if (sig === lastRenderedSignature){
        // 帧未变化：触发一次性刷新
        const now = Date.now()
        const last = (window as any).__last_stuck_refresh || 0
        if (now - last > STUCK_MS){
          (window as any).__last_stuck_refresh = now
          // 轻量拉新
          loadData()
        }
      } else {
        lastRenderedSignature = sig
      }
    }catch{}
  }, 1000)
  // 兜底轮询：即使 ws 没推到，也每 3 秒刷新一次 activeRuns
  activePollTimer = window.setInterval(() => {
    if (document.visibilityState === 'visible') {
      loadActiveRuns().catch(console.error)
    }
  }, 3000)
})
onUnmounted(() => {
  if (stuckTimer) clearInterval(stuckTimer)
  if (activePollTimer) clearInterval(activePollTimer)
})
const {
  getDbProgressStable,
  getDeNoisedStableByRun,
  getDeNoisedStableByTask,
  formatBps,
  calcEtaFromAvg,
  triggerAutoRefresh,
} = useTaskProgressSync({
  activeRuns,
  activeRunLookup,
  lastStableByTask,
  loadData,
  loadActiveRuns,
  lingerMs: LINGER_MS,
})

// Webhook 配置（POST 地址 + 触发来源/状态勾选）
const showWebhookModal = ref(false)
const webhookForm = ref<{taskId:number|null, postUrl:string, triggerId:string, notify:{manual:boolean,schedule:boolean,webhook:boolean}, status:{success:boolean,failed:boolean}}>({ taskId: null, postUrl: '', triggerId:'', notify:{manual:false,schedule:false,webhook:false}, status:{success:true,failed:true} })
function setWebhook(task: Task){
  webhookForm.value.taskId = task.id
  try{
    const opts = (task.options||task.Options||{}) as any
    webhookForm.value.postUrl = (opts?.webhookPostUrl || '')
    ;(webhookForm.value as any).wecomUrl = (opts?.wecomPostUrl || '')
    webhookForm.value.triggerId = (opts?.webhookId || '')
    const n = opts?.webhookNotifyOn || {}
    webhookForm.value.notify = {
      manual: !!n.manual,
      schedule: !!n.schedule,
      webhook: !!n.webhook,
    }
    const s = opts?.webhookNotifyStatus || {}
    ;(webhookForm.value as any).status = { success: s.success!==false, failed: s.failed!==false }
  }catch{
    webhookForm.value.postUrl = ''
    ;(webhookForm.value as any).wecomUrl = ''
    webhookForm.value.triggerId = ''
    webhookForm.value.notify = { manual:false, schedule:false, webhook:false }
    ;(webhookForm.value as any).status = { success:true, failed:true }
  }
  showWebhookModal.value = true
}
async function saveWebhook(){
  if (!webhookForm.value.taskId){ showWebhookModal.value=false; return }
  try{
    const id = webhookForm.value.taskId
    const opts = { webhookId: webhookForm.value.triggerId, webhookPostUrl: webhookForm.value.postUrl, wecomPostUrl: (webhookForm.value as any).wecomUrl||'', webhookNotifyOn: webhookForm.value.notify, webhookNotifyStatus: (webhookForm.value as any).status }
    const fn:any = (api as any).updateTaskOptions
    if (typeof fn === 'function') {
      await fn(id, opts)
    } else {
      // 兼容旧 bundle：直接调用 PATCH /api/tasks
      await fetch('/api/tasks', { method:'PATCH', headers:{ 'Content-Type':'application/json', 'Authorization': `Bearer ${getToken()||''}` }, body: JSON.stringify({ id, options: opts }) })
    }
    showWebhookModal.value = false
    await loadData()
  }catch(e:any){
    showToast(e?.message||String(e), 'error')
  }
}

// 单例模式配置
const showSingletonModal = ref(false)
const singletonForm = ref<{taskId:number|null, singletonEnabled:boolean}>({ taskId: null, singletonEnabled: false })

function setSingletonMode(task: Task){
  singletonForm.value.taskId = task.id
  try{
    const opts = (task.options||task.Options||{}) as any
    singletonForm.value.singletonEnabled = !!opts.singletonMode
  }catch{
    singletonForm.value.singletonEnabled = false
  }
  showSingletonModal.value = true
}

async function saveSingleton(){
  if (!singletonForm.value.taskId){ showSingletonModal.value=false; return }
  try{
    const id = singletonForm.value.taskId
    const opts = { singletonMode: singletonForm.value.singletonEnabled }
    await taskApi.updateOptions(id, opts)
    showSingletonModal.value = false
    await loadData()
  }catch(e:any){
    showToast(e?.message||String(e), 'error')
  }
}

function buildWecomMarkdown(p:any){
  const taskName = p?.task?.name || '测试任务'
  const statusZh = p?.statusZh || '演示'
  const triggerZh = p?.triggerZh || '测试'
  const mode = (p?.task?.mode || '').toLowerCase()
  const okLabel = mode==='move' ? '移动' : '成功'
  const s = p?.summary||{}
  const total = s.totalCount ?? 0
  const ok = s.completedCount ?? 0
  const fail = s.failedCount ?? 0
  const skipped = s.skippedCount ?? 0
  const bytesFmt = formatBytes(Number(s.totalBytes||0))
  const txFmt = formatBytes(Number(s.transferredBytes||0))
  const spFmt = formatBytesPerSec(Number(s.avgSpeedBps||0))
  const dur = p?.run?.durationText || '-'
  let md = `**任务** <font color="info">${taskName}</font> 已${statusZh}（${triggerZh}）\n`
  md += `> 总计: ${total}  ${okLabel}: ${ok}  失败: ${fail}  其他: ${skipped}\n`
  md += `> 体积: ${bytesFmt} / 已传: ${txFmt}\n`
  md += `> 均速: ${spFmt}  耗时: ${dur}\n`
  if (Array.isArray(p?.files)){
    for (const f of p.files){ md += `> ${f}\n` }
  }
  return md
}
async function testWebhook(){
  try{
    const url1 = (webhookForm.value.postUrl||'').trim()
    const url2 = String((webhookForm.value as any)['wecomUrl']||'').trim()
    if (!url1 && !url2){ showToast('请先填写对外 POST 地址或企业微信地址', 'error'); return }
    const payload:any = {
      title: 'RcloneFlow 测试通知',
      triggerZh: '测试',
      statusZh: '演示',
      summaryZh: '这是一条测试消息，用于验证通知接收是否正常（自动识别企业微信/通用 Webhook）。',
      task: { id: 0, name: '测试任务', mode: 'test' },
      run: { id: 0, trigger: 'manual', status: 'success', startedAt: new Date().toISOString(), finishedAt: new Date().toISOString(), durationSeconds: 1, durationText: '1秒' },
      summary: { totalCount: 5, completedCount: 5, failedCount: 0, skippedCount: 0, totalBytes: 10485760, transferredBytes: 10485760, avgSpeedBps: 10485760 },
      files: ['test-a.mp4','test-b.mp4','test-c.mp4','test-d.mp4','test-e.mp4'],
      omittedCount: 0,
    }
    const send = async (url:string)=>{
      const isWecom = url.includes('qyapi.weixin.qq.com')
      let body:any
      if (isWecom){
        const md = buildWecomMarkdown(payload)
        body = JSON.stringify({ msgtype:'markdown', markdown: { content: md } })
      }else{
        body = JSON.stringify(payload)
      }
      const resp = await fetch(url, { method:'POST', headers:{ 'Content-Type':'application/json' }, body })
      if (!resp.ok){ throw new Error(`HTTP ${resp.status}`) }
    }
    const fails:string[] = []
    if (url1){ try{ await send(url1) } catch(e:any){ fails.push(`通用: ${e?.message||e}`) } }
    if (url2){ try{ await send(url2) } catch(e:any){ fails.push(`企业微信: ${e?.message||e}`) } }
    if (fails.length){ showToast(`测试部分失败：${fails.join('；')}`, 'error') } else { showToast('测试通知已发送（请在接收端查看）', 'success') }
  }catch(e:any){
    showToast(`测试发送失败：${e?.message||e}`, 'error')
  }
}

// 仅在 phase 前进时更新：记录每个 task 的最新 phase
const lastPhaseByTaskId: Record<number, string> = {}

// 传输日志弹窗
const showLogModal = ref(false)
const logModalTitle = ref('传输日志')
const logContent = ref('')
async function openRunLog(run:any){
  logModalTitle.value = `传输日志 #${run.id}`
  logContent.value = '加载中…'
  showLogModal.value = true
  try {
    const token = getToken() || ''
    const resp = await fetch(`/api/runs/${run.id}/log`, {
      headers: token ? { 'Authorization': `Bearer ${token}` } : {}
    })
    if (!resp.ok){
      // 兜底再试一次 ?auth= 兼容
      const fallback = await fetch(`/api/runs/${run.id}/log?auth=${token}`)
      if (!fallback.ok){
        const txt = await fallback.text()
        logContent.value = `加载失败：${fallback.status} ${txt}`
        return
      }
      const buf2 = await fallback.arrayBuffer()
      const dec2 = new TextDecoder('utf-8')
      logContent.value = dec2.decode(buf2)
      return
    }
    const buf = await resp.arrayBuffer()
    const dec = new TextDecoder('utf-8')
    logContent.value = dec.decode(buf)
  } catch(e:any){
    logContent.value = `加载异常：${e?.message||e}`
  }
}

// （重复定义已移除）

function showRunDetail(run:any){
  if (run.status === 'running'){
    // 运行中的记录不进入历史详情弹窗，而是走轻量提示小窗
    openRunningHint(run)
    return
  }

  // 历史详情入口：页面层只负责入口判断，详情状态与文件链已分别下沉
  openRunDetailModal(run)
  openRunDetailFiles(run)
}

function closeRunDetail(){
  // 历史详情出口：页面层只保留关闭入口，具体状态由 useRunDetailState 管理
  closeRunDetailModal()
}


// finalSummary.files 与统计计数已下沉到 useRunDetailComputed.ts
// move 模式时，成功数量代表 Moved 条数；已在后端合并 Copied+Deleted 为 Moved
// 筛选状态与 finalFilteredFiles 已下沉到 useRunDetailComputed.ts
// 分页（按筛选后的集）
const finalFilesPageSize = ref(Math.max(10, Math.floor((window.innerHeight - 420) / 34)))
const finalFilesPage = ref(1)
const {
  openMenuId,
  confirmModal,
  toggleMenu,
  closeMenus,
  showConfirm,
  closeConfirm,
  confirmAndClose,
} = useTaskViewUi()

const {
  createForm,
  commandMode,
  commandText,
  editingTask,
  showAdvancedOptions,
  resetTaskFormForCreate,
  fillTaskFormForEdit,
} = useTaskFormState()

const creatingState = ref<'idle' | 'loading' | 'done'>('idle')

function getScheduleByTaskId(taskId: number) {
  return schedules.value.find(s => s.taskId === taskId)
}

const {
  handleTaskFormDoneClick,
  validateTaskForm,
  executeTaskFormSubmit,
} = useTaskFormSubmit({
  createForm,
  editingTask,
  creatingState,
  currentModule,
  normalizeTaskOptions,
  getScheduleByTaskId,
  loadData,
  taskApi,
  scheduleApi,
})

const {
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
  loadSourcePath,
  loadTargetPath,
  onSourceClick,
  onSourceArrow,
  onTargetClick,
  onTargetArrow,
} = useTaskPathBrowse({
  createForm,
  listPath: api.listPath,
})

function normalizeTaskOptions(raw: Record<string, any> | undefined | null) {
  const options = { ...(raw || {}) }
  if (typeof options.enableStreaming === 'undefined') {
    options.enableStreaming = true
  }
  return options
}

onMounted(async () => {
  await loadData()
  await loadActiveRuns()
  setupRealtimeSync()
})

function getActiveRunByTaskId(taskId: number) {
  return activeRunLookup.getActiveRunByTaskId(taskId)
}

function formatTime(time: string | undefined) {
  if (!time) return '-'
  try {
    return new Date(time).toLocaleString('zh-CN', {
      year: 'numeric', month: '2-digit', day: '2-digit',
      hour: '2-digit', minute: '2-digit'
    })
  } catch {
    return time || '-'
  }
}

function formatDuration(startTime: string | undefined, endTime: string | undefined) {
  if (!startTime) return '-'
  const start = new Date(startTime).getTime()
  const end = endTime ? new Date(endTime).getTime() : Date.now()
  const diff = Math.max(0, end - start)
  const seconds = Math.floor(diff / 1000) % 60
  const minutes = Math.floor(diff / 60000) % 60
  const hours = Math.floor(diff / 3600000)
  if (hours > 0) return `${hours}h ${minutes}m ${seconds}s`
  if (minutes > 0) return `${minutes}m ${seconds}s`
  return `${seconds}s`
}

function getRunDurationText(run:any){
  const fs = getFinalSummaryFromComposable(run)
  if (fs && fs.durationText) return fs.durationText
  // fallback to local compute for running
  return formatDuration(run.startedAt, run.finishedAt)
}
function getStatusClass(status: string) {
  switch (status) {
    case 'running': return 'running'
    case 'finished': return 'success'
    case 'failed': return 'failed'
    case 'skipped': return 'skipped'
    default: return ''
  }
}

function getStatusText(status: string) {
  switch (status) {
    case 'running': return '运行中'
    case 'finished': return '已完成'
    case 'failed': return '失败'
    case 'skipped': return '已跳过'
    default: return status
  }
}


const { validateTaskFormBeforeSubmit } = useTaskFormPrepare({
  createForm,
  commandMode,
  commandText,
  normalizeTaskOptions,
  parseRcloneCommand,
  validateTaskForm,
})

const { runTaskFormFlow } = useTaskFormFlow({
  handleTaskFormDoneClick,
  validateTaskFormBeforeSubmit,
  executeTaskFormSubmit,
})

async function createTask() {
  const error = await runTaskFormFlow()
  if (error) {
    showToast(error, 'error')
  }
}

const {
  runFilesPage,
  openRunDetailFiles,
  pagedRunFiles,
  totalRunFilesPages,
  goPrevFilesPage,
  goNextFilesPage,
} = useRunDetailFiles({ runDetail, runApi })

const {
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
} = useRunDetailComputed({ runDetail, finalFilesPage, finalFilesPageSize })

const {
  filteredRuns,
  filteredRunsTotal,
  currentTotal,
  currentTotalPages,
} = useTaskHistoryComputed({
  runs,
  runsTotal,
  taskRuns,
  historyFilterTaskId,
  historyStatusFilter,
  runsPage,
  runsPageSize,
  getFinalSummary: getFinalSummaryFromComposable,
})

const {
  refreshTaskHistoryRuns,
  viewTaskHistory,
} = useTaskHistoryLoader({
  taskRuns,
  historyFilterTaskId,
  runsPage,
  jumpPage,
  currentModule,
  runApi,
})

const {
  clearRun,
  clearAllRuns,
} = useTaskHistoryActions({
  runs,
  taskRuns,
  historyFilterTaskId,
  runsPage,
  jumpPage,
  filteredRuns,
  loadData,
  refreshTaskHistoryRuns,
  runApi,
})

const {
  deleteTask,
  toggleSchedule,
  deleteSchedule,
  clearAllRunsWithConfirm,
} = useTaskListActions({
  openMenuId,
  historyFilterTaskId,
  schedules,
  loadData,
  showConfirm,
  showToast,
  clearAllRuns,
  taskApi,
  scheduleApi,
})

const {
  runningTaskId,
  stoppedTaskId,
  stopTaskAny,
  runTask,
} = useTaskRunActions({
  loadData,
  showToast,
  taskApi,
  jobApi,
})

const {
  goToAddTask,
  editTask,
} = useTaskFormEntry({
  currentModule,
  openMenuId,
  remotes,
  remoteApi,
  resetTaskFormForCreate,
  resetTaskPathBrowse,
  getScheduleByTaskId,
  fillTaskFormForEdit,
  restoreTaskPathBrowse,
})

function formatScheduleSpec(spec: string): string {
  if (!spec) return ''
  const parts = spec.split('|')
  if (parts.length !== 5) return spec

  // 标准cron格式: minute hour day month week
  // 例如: "43|17,19|*|**|*" 显示为 "43 17,19 * * *"
  const [minute, hour, day, month, week] = parts

  // 显示为标准cron格式
  return `${minute} ${hour} ${day} ${month} ${week}`
}

</script>


<template>
  <!-- Toast 通知容器 -->
  <div class="toast-container">
    <ToastItem v-for="toast in toasts" :key="toast.id" :toast="toast" />
  </div>

  <div v-if="currentModule === 'tasks'" class="card">
    <div class="card-header">
      <div class="title">任务列表</div>
      <div class="header-actions">
        <input v-model="taskSearch" type="text" placeholder="搜索任务..." class="search-input" />
        <button class="primary small" @click="goToAddTask">+ 添加任务</button>
      </div>
    </div>
    <div class="list">
      <TaskCard
        v-for="task in filteredTasks"
        :key="task.id"
        :task="task"
        :schedule="getScheduleByTaskId(task.id)"
        :active-run="getActiveRunByTaskId(task.id)"
        :running-task-id="runningTaskId"
        :stopped-task-id="stoppedTaskId"
        @run="runTask(task.id)"
        @edit="editTask(task)"
        @delete="deleteTask(task.id)"
        @toggle-schedule="toggleSchedule(task.id)"
        @view-history="viewTaskHistory(task.id)"
        @stop="stopTaskAny(task.id)"
        @set-webhook="setWebhook(task)"
        @set-singleton="setSingletonMode(task)"
      />
      <div v-if="!filteredTasks.length" class="empty">暂无任务</div>
    </div>
    <!-- 任务分页 -->
    <div class="pagination" v-if="tasksTotal > tasksPageSize">
      <span class="page-current">第 {{ tasksPage }} / {{ currentTasksPages }} 页</span>
      <button class="page-btn" :disabled="tasksPage <= 1" @click="tasksPage--">上一页</button>
      <button class="page-btn" :disabled="tasksPage >= currentTasksPages" @click="tasksPage++">下一页</button>
      <input type="number" class="page-input" v-model.number="tasksJumpPage" :min="1" :max="currentTasksPages" @keyup.enter="jumpToTasksPage" />
      <button class="page-btn" @click="jumpToTasksPage">跳转</button>
    </div>
  </div>

  <!-- Webhook 配置弹窗（POST 地址 + 触发来源） -->
  <div v-if="showWebhookModal" class="modal-overlay" @click.self="showWebhookModal=false">
    <div class="modal-content" style="max-width:560px">
      <div class="modal-header">
        <h3>Webhook 通知</h3>
        <button class="close-btn" @click="showWebhookModal=false">×</button>
      </div>
      <div class="modal-body">
        <div class="detail-item full-width">
          <label>Webhook 接收（触发ID）：</label>
          <input v-model="webhookForm.triggerId" type="text" placeholder="留空=使用任务ID（示例：/webhook/<任务ID> 或 /webhook/<你的ID>）" />
        </div>
        <div class="detail-item full-width">
          <label>对外 POST 地址：</label>
          <input v-model="webhookForm.postUrl" type="text" placeholder="https://example.com/hooks/endpoint" />
          <p class="hint">任务完成或失败后，将以 POST 通知该地址。</p>
        </div>
        <div class="detail-item full-width">
          <label>企业微信地址：</label>
          <input v-model="(webhookForm as any).wecomUrl" type="text" placeholder="https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=..." />
          <p class="hint">若填写，将同时向企业微信机器人发送 Markdown 通知。</p>
        </div>
        <div class="detail-item">
          <label>触发来源：</label>
          <div class="trigger-row">
            <label class="trigger-opt"><input type="checkbox" v-model="webhookForm.notify.manual" /><span>手动</span></label>
            <label class="trigger-opt"><input type="checkbox" v-model="webhookForm.notify.schedule" /><span>定时</span></label>
            <label class="trigger-opt"><input type="checkbox" v-model="webhookForm.notify.webhook" /><span>Webhook</span></label>
          </div>
        </div>
        <div class="detail-item">
          <label>状态过滤：</label>
          <div class="trigger-row">
            <label class="trigger-opt"><input type="checkbox" v-model="(webhookForm as any).status.success" /><span>成功</span></label>
            <label class="trigger-opt"><input type="checkbox" v-model="(webhookForm as any).status.failed" /><span>失败</span></label>
          </div>
          <p class="hint">仅当匹配状态时发送通知；默认两个都勾选。</p>
        </div>
      </div>
      <div class="modal-footer">
        <button class="primary" @click="saveWebhook">保存</button>
        <button class="ghost" @click="testWebhook" :disabled="!webhookForm.postUrl">发送测试</button>
        <button class="ghost" @click="showWebhookModal=false">取消</button>
      </div>
    </div>
  </div>

  <!-- 单例模式配置弹窗 -->
  <div v-if="showSingletonModal" class="modal-overlay" @click.self="showSingletonModal=false">
    <div class="modal-content" style="max-width:560px">
      <div class="modal-header">
        <h3>单例模式</h3>
        <button class="close-btn" @click="showSingletonModal=false">×</button>
      </div>
      <div class="modal-body">
        <div class="detail-item full-width">
          <label class="trigger-opt">
            <input type="checkbox" v-model="singletonForm.singletonEnabled" />
            <span>开启单例模式</span>
          </label>
          <p class="hint">开启后，该任务触发时会检测全局是否有其他传输任务在运行。有则放弃本次执行，不排队，不等待，不重试。</p>
        </div>
      </div>
      <div class="modal-footer">
        <button class="primary" @click="saveSingleton">保存</button>
        <button class="ghost" @click="showSingletonModal=false">取消</button>
      </div>
    </div>
  </div>

  <!-- 传输日志弹窗 -->
  <div v-if="showLogModal" class="modal-overlay" @click.self="showLogModal=false">
    <div class="modal-content log-modal">
      <div class="modal-header">
        <h3>{{ logModalTitle }}</h3>
        <button class="close-btn" @click="showLogModal=false">×</button>
      </div>
      <div class="modal-body">
        <div class="log-box">
          <pre class="log-pre">{{ logContent }}</pre>
        </div>
      </div>
      <div class="modal-footer">
        <button class="ghost" @click="showLogModal=false">关闭</button>
      </div>
    </div>
  </div>

  <div v-if="currentModule === 'history'">
    <TaskHistoryPanel
      :current-total="currentTotal"
      :runs-page="runsPage"
      :runs-page-size="runsPageSize"
      :current-total-pages="currentTotalPages"
      :jump-page="jumpPage"
      :history-filter-task-id="historyFilterTaskId"
      :history-status-filter="historyStatusFilter"
      :filtered-runs="filteredRuns"
      :get-db-progress-stable="getDbProgressStable"
      :get-final-summary="getFinalSummaryFromComposable"
      @back="currentModule = 'tasks'"
      @set-status-filter="historyStatusFilter = $event"
      @prev-page="runsPage--; loadData()"
      @next-page="runsPage++; loadData()"
      @update-jump-page="jumpPage = $event"
      @jump-page="jumpToPage"
      @clear-all="clearAllRunsWithConfirm"
      @view-detail="showRunDetail"
      @view-log="openRunLog"
      @clear-run="clearRun"
    />

    <!-- 运行详情弹窗 -->
    <RunDetailModal
      :visible="showDetailModal"
      :run-detail="runDetail"
      :get-status-class="getStatusClass"
      :get-status-text="getStatusText"
      :get-final-summary="getFinalSummaryFromComposable"
      :get-preflight="getPreflightFromComposable"
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
      @close="closeRunDetail()"
      @set-final-filter="setFinalFilter"
      @prev-final-files-page="goPrevFinalFilesPage()"
      @next-final-files-page="goNextFinalFilesPage()"
      @update-final-files-jump="finalFilesJump = $event"
      @jump-final-files-page="jumpFinalFilesPage()"
      @prev-files-page="goPrevFilesPage()"
      @next-files-page="goNextFilesPage()"
    />
  </div>

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

    <AddTaskForm
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
      @update:command-mode="commandMode = $event"
      @update:command-text="commandText = $event"
      @update:show-source-path-input="setShowSourcePathInput($event)"
      @update:show-target-path-input="setShowTargetPathInput($event)"
      @update:show-advanced-options="showAdvancedOptions = $event"
      @source-remote-change="onSourceRemoteChange"
      @target-remote-change="onTargetRemoteChange"
      @source-breadcrumb-click="onSourceBreadcrumbClick"
      @target-breadcrumb-click="onTargetBreadcrumbClick"
      @source-arrow="onSourceArrow"
      @source-click="onSourceClick"
      @target-arrow="onTargetArrow"
      @target-click="onTargetClick"
      @submit="createTask"
    />

  <!-- 全局实时数据弹窗 -->
  <div v-if="showGlobalStatsModal" class="modal-overlay" @click.self="showGlobalStatsModal = false">
    <div class="modal-content">
      <div class="modal-header">
        <h3>全局实时数据</h3>
        <button class="close-btn" @click="showGlobalStatsModal = false">×</button>
      </div>
      <div class="modal-body">
        <div class="detail-item">
          <label>已传输：</label>
          <span>{{ formatBytes(globalStats.bytes) || '-' }}</span>
        </div>
        <div class="detail-item">
          <label>总大小：</label>
          <span>{{ formatBytes(globalStats.totalBytes) || '-' }}</span>
        </div>
        <div class="detail-item">
          <label>当前速度：</label>
          <span>{{ formatBytesPerSec(globalStats.speed) || '-' }}</span>
        </div>
        <div class="detail-item">
          <label>平均速度：</label>
          <span>{{ formatBytesPerSec(globalStats.speedAvg) || '-' }}</span>
        </div>
        <div class="detail-item">
          <label>预计剩余时间：</label>
          <span>{{ formatEta(globalStats.eta) || '-' }}</span>
        </div>
        <div class="detail-item">
          <label>进度：</label>
          <span>{{ globalStats.percentage !== undefined ? globalStats.percentage.toFixed(2) + '%' : '-' }}</span>
        </div>
        <div class="progress-bar-container">
          <div class="progress-bar" :style="{ width: (globalStats.percentage || 0) + '%' }"></div>
        </div>
      </div>
    </div>
  </div>




  <!-- 确认删除弹窗 -->
  <div v-if="confirmModal.show" class="modal-overlay" @click.self="closeConfirm()">
    <div class="modal-content confirm-modal">
      <div class="modal-header">
        <h3>{{ confirmModal.title }}</h3>
        <button class="close-btn" @click="closeConfirm()">×</button>
      </div>
      <div class="modal-body">
        <p>{{ confirmModal.message }}</p>
      </div>
      <div class="modal-footer">
        <button class="ghost" @click="closeConfirm()">取消</button>
        <button class="primary danger" @click="confirmAndClose()">确认</button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.module-tabs {
  display: flex;
  gap: 8px;
  padding: 16px 20px;
}
.tab-btn {
  padding: 10px 24px;
  border-radius: 10px;
  background: var(--surface);
  border: 1px solid var(--border);
  color: var(--muted);
  cursor: pointer;
  font-size: 14px;
  font-weight: 500;
  transition: all 0.2s;
}
.tab-btn:hover { background: var(--surface); color: var(--text); }
.tab-btn.active { background: #1e3a5f; color: var(--accent); border-color: var(--accent-strong); }
body.light .tab-btn { background: var(--surface); border-color: var(--border); color: var(--muted); }
body.light .tab-btn:hover { background: #e8e8e8; color: var(--text); }
body.light .tab-btn.active { background: #e3f2fd; color: var(--accent); border-color: #bbdefb; }
.list-header {
  display: flex;
  justify-content: space-between;
  padding: 10px 20px;
  background: #252525;
  font-size: 12px;
  color: #888;
  border-bottom: 1px solid #333;
}
body.light .list-header { background: #f5f5f5; color: #666; border-bottom: 1px solid #e0e0e0; }
.list-header{ background:#0f172a; color:#cbd5e1; border-bottom:1px solid #334155 }
.col-name { flex: 1; }
.col-status { width: 80px; text-align: center; }
.col-time { width: 150px; text-align: right; }
.col-action { width: 80px; text-align: right; }
.col-info { width: 120px; }
/* 让卡片更宽松，展示 chips */
.list .run-item { align-items:flex-start }
.item {
  display: flex;
  align-items: center;
  padding: 12px 20px;
  border-bottom: 1px solid #252525;
  gap: 12px;
}
body.light .item { border-color: #f0f0f0; }
.item .name { flex: 1; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.status { width: 80px; padding: 4px 8px; border-radius: 6px; font-size: 12px; text-align: center; }
.status.running { background: var(--accent); color: #fff; }
.status.success { background: var(--success); color: #fff; }
.status.failed { background: var(--danger); color: #fff; }
.status.skipped { background: var(--warning); color: #fff; }
.status.clickable { cursor: pointer; }
.status.clickable:hover { opacity: 0.8; }
.time { width: 150px; text-align: right; color: var(--muted); font-size: 13px; }
.info { width: 120px; color: var(--muted); font-size: 13px; }
.summary-mini { display:flex; gap:8px; flex-wrap:wrap; margin:6px 0; }
.summary-mini .chip { padding:2px 8px; border-radius:999px; background:#1f2937; color:#e5e7eb; font-size:12px; border:1px solid #334155 }
.summary-mini .chip.success{ background:#0a2f22; border-color:#14532d; color:#34d399 }
.summary-mini .chip.failed{ background:#2b0a0a; border-color:#7f1d1d; color:#f87171 }
.summary-mini .chip.other{ background:#2b1a04; border-color:#92400e; color:#fbbf24 }
.summary-mini .chip.meta{ background:#111827; border-color:#374151; color:#cbd5e1 }
.skipped-message{ padding:8px 12px; background:#fef3c7; border:1px solid #f59e0b; border-radius:6px; color:#92400e; font-size:13px; margin-bottom:8px; }
body.light .skipped-message{ background:#fffbeb; color:#78350f; }
.item-actions { display: flex; gap: 8px; }
.danger-text { color: var(--danger) !important; }
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
.inline-logline{ display:block; white-space:pre-wrap; word-break:break-all; background:rgba(255,255,255,0.06); border:1px solid rgba(255,255,255,0.08); border-radius:8px; padding:8px 10px; font-size:12px; line-height:1.5; }
.debug-toggle{ width:100%; justify-content:center; }
.form-content label.inline-label{ display:flex !important; align-items:center; gap:8px; margin:0 0 6px 0; }
.form-content label.inline-label input[type="checkbox"]{ width:16px; height:16px; }
body.light .form-content input,
body.light .form-content select { background: #fff; border-color: #ddd; color: #333; }
.path-selector { display: flex; gap: 8px; align-items: flex-start; }
.path-browse { flex: 1; border: 1px solid var(--border); border-radius: 8px; background: var(--surface); overflow: hidden; }
body.light .path-browse { border-color: var(--border); background: var(--surface); }
.path-bar { display: flex; justify-content: space-between; align-items: center; padding: 8px 12px; background: var(--surface); border-bottom: 1px solid var(--border); }
body.light .path-bar { background: var(--surface); border-color: var(--border); }
.path-selected { font-size: 12px; color: var(--success); font-weight: 600; }
.path-label { font-size: 12px; color: var(--muted); }
.path-list { max-height: 200px; overflow-y: auto; padding: 8px; }
.path-item {
  padding: 6px 12px;
  border-radius: 6px;
  cursor: pointer;
  font-size: 13px;
  color: #e0e0e0;
  display: flex;
  align-items: center;
  gap: 8px;
}
body.light .path-item { color: #333; }
.path-item:hover { background: var(--border); }
body.light .path-item:hover { background: #f0f0f0; }
.folder-icon { margin-right: 4px; cursor: pointer; }
.file-icon { margin-right: 4px; }

/* Schedule Section */
.schedule-section { margin-top: 16px; padding-top: 16px; border-top: 1px solid #333; }
body.light .schedule-section { border-top-color: #ddd; }
.schedule-toggle { display: flex; align-items: center; gap: 8px; cursor: pointer; font-size: 14px; margin-bottom: 12px; }
.schedule-toggle input { width: 16px; height: 16px; }
.section-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 12px; }

/* 高级选项 */
.advanced-section { margin-top: 16px; padding-top: 16px; border-top: 1px solid #333; }
body.light .advanced-section { border-top-color: #ddd; }
.advanced-group { margin-bottom: 16px; padding-bottom: 16px; border-bottom: 1px solid #2a2a2a; }
body.light .advanced-group { border-bottom-color: #eee; }
.advanced-group:last-child { border-bottom: none; }
.advanced-group-title { font-weight: 600; font-size: 13px; color: #64b5f6; margin-bottom: 12px; }
.advanced-row { margin-bottom: 12px; }
.advanced-row label { display: block; font-size: 12px; color: #888; margin-bottom: 4px; }
.advanced-row input[type="text"],
.advanced-row input[type="number"] { width: 100%; padding: 8px 12px; border: 1px solid #333; border-radius: 8px; background: #252525; color: #e0e0e0; font-size: 13px; }
body.light .advanced-row input { background: #fff; border-color: #ddd; color: #1a1a1a; }
.advanced-row input:focus { outline: none; border-color: #64b5f6; }
.advanced-row.inline { display: flex; align-items: center; gap: 8px; }
.advanced-row.inline label { margin-bottom: 0; }
.advanced-row.inline input[type="checkbox"] { width: 16px; height: 16px; }
.advanced-row textarea { width: 100%; padding: 8px 12px; border: 1px solid #333; border-radius: 8px; background: #252525; color: #e0e0e0; font-size: 13px; resize: vertical; font-family: inherit; }
body.light .advanced-row textarea { background: #fff; border-color: #ddd; color: #1a1a1a; }
.advanced-row select { width: 100%; padding: 8px 12px; border: 1px solid #333; border-radius: 8px; background: #252525; color: #e0e0e0; font-size: 13px; }
body.light .advanced-row select { background: #fff; border-color: #ddd; color: #1a1a1a; }
.schedule-grid { display: grid; grid-template-columns: repeat(6, 1fr); gap: 12px; }
.schedule-item { display: flex; flex-direction: column; gap: 4px; }
.schedule-item label { font-size: 12px; color: #888; font-weight: 600; }
.schedule-item select { padding: 4px; border-radius: 4px; background: #2a2a2a; color: #fff; border: 1px solid #444; font-size: 12px; height: 150px; resize: none; }
body.light .schedule-item select { background: #fff; color: #333; border-color: #ccc; }
.schedule-item select option { padding: 2px 4px; }
.selected-val { font-size: 11px; color: #4caf50; min-height: 14px; word-break: break-all; }
.file-icon { font-size: 12px; }
.item-name { flex: 1; }
.path-item.is-dir .item-name { color: #64b5f6; font-weight: 500; }
.path-empty { padding: 20px; text-align: center; color: #666; font-size: 13px; }
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
/* 运行详情弹窗加宽 50% */
.detail-modal{ width: 135% !important; max-width: 1200px !important; }

/* 运行总结样式（深浅主题适配） */
.summary-box{background:#111827;border:1px solid #333;border-radius:10px;padding:12px 14px;margin-top:6px;max-width:1200px}
.summary-title{font-weight:600;color:#e0e0e0;margin-bottom:8px}
.summary-grid{display:grid;grid-template-columns:repeat(4,1fr);gap:12px}
.summary-cell{background:#0f172a;border:1px solid #334155;border-radius:10px;padding:10px 12px}
.summary-key{font-size:12px;color:#94a3b8;margin-bottom:6px}
.summary-val{font-size:16px;color:#e2e8f0;font-weight:700}
/* 浅色主题覆盖 */
body.light .summary-box{background:#ffffff;border-color:#e5e7eb}
body.light .summary-title{color:#111827}
body.light .summary-cell{background:#f8fafc;border-color:#e5e7eb}
body.light .summary-key{color:#64748b}
body.light .summary-val{color:#111827}
/* 可点击统计卡片的交互反馈 */
.summary-cell.clickable{cursor:pointer;transition:background .15s ease,border-color .15s ease,box-shadow .15s ease}
.summary-cell.clickable:hover{background:#1f2937;border-color:#475569;box-shadow:0 0 0 1px #334155 inset}
body.light .summary-cell.clickable:hover{background:#f0f4f8;border-color:#cbd5e1;box-shadow:0 0 0 1px #cbd5e1 inset}

/* 传输明细表更宽、更疏朗 */
.files-table{margin-top:14px;border:1px solid #333;border-radius:10px;}
.files-table.large{max-width:1200px}
.files-header,.files-row{display:grid;grid-template-columns:1fr 140px 200px 140px;gap:18px;align-items:center}
.files-header{padding:12px 16px;background:#252525;color:#cbd5e1;font-size:13px}
.files-body{max-height:540px;overflow:auto}
.files-row{padding:12px 16px;border-top:1px solid #333}
/* 修正黑底黑字：表格内文字统一亮色，浅色主题时再覆盖 */
.files-row .name,.files-row .status,.files-row .time,.files-row .size { color:#e5e7eb }
/* 文件名单行省略，避免换行导致高度抖动 */
.files-row .name{ white-space:nowrap; overflow:hidden; text-overflow:ellipsis }
body.light .files-header{ background:#f5f5f5; color:#4b5563 }
body.light .files-row .name,body.light .files-row .status,body.light .files-row .time,body.light .files-row .size { color:#1f2937 }
/* 表格内的 status 仅作为结果文本，不用徽章底色 */
.files-table .status { width:auto; padding:0; background:transparent; border-radius:0; font-weight:600 }
.files-table .status.success{ color:#34d399 }
.files-table .status.failed{ color:#f87171 }
.files-table .status.skipped{ color:#fbbf24 }
.pager-inline{display:flex;align-items:center;gap:8px}
.page-input{width:64px;padding:6px 8px;border:1px solid #333;border-radius:8px;background:#252525;color:#e0e0e0}
body.light .page-input{ background:#fff; color:#111827; border-color:#ddd }
.chip .est{ color:#ef4444 }
.chip .act{ color:#16a34a }
.summary-val.est{ color:#ef4444 }
.summary-val.act{ color:#16a34a }
.trigger-row{ display:flex; align-items:center; gap:16px; flex-wrap:wrap }
.trigger-opt{ display:inline-flex; align-items:center; gap:6px; cursor:pointer }
.trigger-opt input{ width:16px; height:16px }
/* 移动端强制列布局，彻底覆盖 global CSS */
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

/* Toast 通知 */
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

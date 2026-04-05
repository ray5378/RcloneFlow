<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import * as api from '../api'
import type { Task, Schedule, Run } from '../types'

const tasks = ref<Task[]>([])
const schedules = ref<Schedule[]>([])
const runs = ref<Run[]>([])
const taskSearch = ref('')

// 过滤后的任务列表
const filteredTasks = computed(() => {
  if (!taskSearch.value) return tasks.value
  const q = taskSearch.value.toLowerCase()
  return tasks.value.filter(t =>
    t.name.toLowerCase().includes(q) ||
    t.sourceRemote.toLowerCase().includes(q) ||
    t.targetRemote.toLowerCase().includes(q) ||
    t.mode.toLowerCase().includes(q)
  )
})

const remotes = ref<string[]>([])
const currentModule = ref<'history' | 'add' | 'tasks'>('tasks')
const historyFilterTaskId = ref<number | null>(null)
const showDetailModal = ref(false)
const runDetail = ref<any>({})
const showCreateModal = ref(false)
const showAdvancedOptions = ref(false)
const showGlobalStatsModal = ref(false)
const globalStats = ref<any>({})
const showTaskProgressModal = ref(false)
const taskProgressData = ref<any>({})
const activeRuns = ref<any[]>([])
let activeRunsTimer: number | null = null
const confirmModal = ref<{ show: boolean; title: string; message: string; onConfirm: () => void }>({
  show: false,
  title: '',
  message: '',
  onConfirm: () => {}
})

const createForm = ref({
  name: '',
  mode: 'copy',
  sourceRemote: '',
  sourcePath: '',
  targetRemote: '',
  targetPath: '',
  enableSchedule: false,
  scheduleMonth: '*',     // * = 每月, 或 "1,3,5"
  scheduleWeek: '',      // 空 = 不设置, 或 "1,3,5" (周一三五)
  scheduleDay: '',       // 空 = 不设置, 或 "1,15,28"
  scheduleHour: '00',   // "00,12,18"
  scheduleMinute: '00', // "00,30,59"
  options: { enableStreaming: true } as Record<string, any>,
})

const openMenuId = ref<number | null>(null)
const editingTask = ref<Task | null>(null)
const creatingState = ref<'idle' | 'loading' | 'done'>('idle')
const runningTaskId = ref<number | null>(null)

const sourcePathOptions = ref<any[]>([])
const targetPathOptions = ref<any[]>([])
const showSourcePathInput = ref(false)
const showTargetPathInput = ref(false)
const sourceCurrentPath = ref('')
const targetCurrentPath = ref('')

// 定时选项
const hourOptions = Array.from({length: 24}, (_, i) => ({ value: String(i).padStart(2,'0'), label: String(i).padStart(2,'0')+'时' }))
const minuteOptions = Array.from({length: 60}, (_, i) => ({ value: String(i).padStart(2,'0'), label: String(i).padStart(2,'0')+'分' }))

// 临时选择状态
const tempSchedule = ref({
  month: [] as string[],
  week: [] as string[],
  day: [] as string[],
  hour: [] as string[],
  minute: [] as string[],
})

// 确认选择
function confirmField(field: 'month' | 'week' | 'day' | 'hour' | 'minute') {
  const val = tempSchedule.value[field]
  if (val.length === 0) {
    // 空表示使用*代表任意
    createForm.value['schedule' + field.charAt(0).toUpperCase() + field.slice(1)] = '*'
  } else {
    createForm.value['schedule' + field.charAt(0).toUpperCase() + field.slice(1)] = val.join(',')
  }
}

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
  activeRunsTimer = window.setInterval(() => {
    loadActiveRuns().catch(console.error)
  }, 3000)
})

async function loadData() {
  try {
    const [taskData, remoteData, scheduleData, runData] = await Promise.all([
      api.listTasks(),
      api.listRemotes(),
      api.listSchedules(),
      api.listRuns(),
    ])
    tasks.value = taskData || []
    remotes.value = remoteData?.remotes || []
    schedules.value = scheduleData || []
    runs.value = runData || []
  } catch (e) {
    console.error(e)
  }
}

async function loadActiveRuns() {
  try {
    const data = await api.getActiveRuns()
    activeRuns.value = data || []
  } catch (e) {
    console.error(e)
  }
}

function getActiveRunByTaskId(taskId: number) {
  return activeRuns.value.find(item => item.runRecord?.taskId === taskId)
}

function getTaskRealtimeProgress(taskId: number) {
  const active = getActiveRunByTaskId(taskId)
  if (!active) return null
  const rt = active.realtimeStatus || {}
  const derived = (active.derivedProgress && typeof active.derivedProgress === 'object') ? active.derivedProgress : {}
  const groupStats = (active.groupStats && typeof active.groupStats === 'object') ? active.groupStats : {}
  const globalStats = (active.globalStats && typeof active.globalStats === 'object') ? active.globalStats : {}
  const progress = (rt.progress && typeof rt.progress === 'object') ? rt.progress : {}
  const group = (progress.group && typeof progress.group === 'object') ? progress.group : {}

  const bytes = Number(
    derived.bytes ?? groupStats.bytes ?? progress.bytes ?? group.bytes ?? globalStats.bytes ?? active.runRecord?.bytesTransferred ?? 0,
  )
  const totalBytes = Number(
    derived.totalBytes ?? derived.total_bytes ?? groupStats.totalBytes ?? groupStats.total_bytes ?? progress.totalBytes ?? progress.total_bytes ?? group.totalBytes ?? group.total_bytes ?? globalStats.totalBytes ?? globalStats.total_bytes ?? 0,
  )
  const speed = Number(
    derived.speed ?? groupStats.speed ?? progress.speed ?? group.speed ?? globalStats.speed ?? 0,
  )
  const eta = derived.eta ?? groupStats.eta ?? progress.eta ?? group.eta ?? globalStats.eta ?? null
  let percentage = Number(
    derived.percentage ?? groupStats.percentage ?? progress.percentage ?? group.percentage ?? globalStats.percentage ?? 0,
  )
  if ((!percentage || Number.isNaN(percentage)) && totalBytes > 0) {
    percentage = (bytes / totalBytes) * 100
  }
  return { bytes, totalBytes, speed, eta, percentage, raw: rt, derived, groupStats, globalStats, run: active.runRecord }
}

// 加载全局实时统计
async function loadGlobalStats() {
  try {
    const stats = await api.getGlobalStats()
    globalStats.value = stats || {}
  } catch (e) {
    console.error(e)
  }
}

// 打开全局实时数据弹窗
function openGlobalStats() {
  showGlobalStatsModal.value = true
  loadGlobalStats()
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
  
  if (hours > 0) {
    return `${hours}h ${minutes}m ${seconds}s`
  } else if (minutes > 0) {
    return `${minutes}m ${seconds}s`
  } else {
    return `${seconds}s`
  }
}

function getStatusClass(status: string) {
  switch (status) {
    case 'running': return 'running'
    case 'finished': return 'success'
    case 'failed': return 'failed'
    default: return ''
  }
}

function getStatusText(status: string) {
  switch (status) {
    case 'running': return '运行中'
    case 'finished': return '已完成'
    case 'failed': return '失败'
    default: return status
  }
}

async function createTask() {
  // 如果已完成，点击返回任务列表
  if (creatingState.value === 'done') {
    creatingState.value = 'idle'
    currentModule.value = 'tasks'
    return
  }

  if (!createForm.value.name) {
    alert('请输入任务名称')
    return
  }
  if (!createForm.value.sourceRemote || !createForm.value.targetRemote) {
    alert('请选择源和目标存储')
    return
  }

  creatingState.value = 'loading'
  try {
    // 构建任务数据
    const taskData = {
      name: createForm.value.name,
      mode: createForm.value.mode,
      sourceRemote: createForm.value.sourceRemote,
      sourcePath: createForm.value.sourcePath,
      targetRemote: createForm.value.targetRemote,
      targetPath: createForm.value.targetPath,
      options: normalizeTaskOptions(createForm.value.options),
    }

    if (editingTask.value) {
      // 更新任务
      await api.updateTask(editingTask.value.id, taskData)
      // 更新定时规则：先删除旧的在创建新的
      const oldSchedule = getScheduleByTaskId(editingTask.value.id)
      if (oldSchedule) {
        await api.deleteSchedule(oldSchedule.id)
      }
      if (createForm.value.enableSchedule) {
        const spec = [
          createForm.value.scheduleMinute || '00',
          createForm.value.scheduleHour || '*',
          createForm.value.scheduleDay || '*',
          createForm.value.scheduleMonth || '*',
          createForm.value.scheduleWeek || '*',
        ].join('|')
        await api.createSchedule({ taskId: editingTask.value.id, spec, enabled: true })
      }
      editingTask.value = null
    } else {
      // 新建任务
      const task = await api.createTask(taskData)
      // 如果启用了定时任务，创建定时规则
      if (createForm.value.enableSchedule) {
        const spec = [
          createForm.value.scheduleMinute || '00',
          createForm.value.scheduleHour || '*',
          createForm.value.scheduleDay || '*',
          createForm.value.scheduleMonth || '*',
          createForm.value.scheduleWeek || '*',
        ].join('|')
        await api.createSchedule({ taskId: task.id, spec, enabled: true })
      }
    }
    createForm.value = { 
      name: '', mode: 'copy', sourceRemote: '', sourcePath: '', targetRemote: '', targetPath: '',
      enableSchedule: false,
      scheduleMonth: '*', scheduleWeek: '*', scheduleDay: '*', scheduleHour: '*', scheduleMinute: '00',
      options: { enableStreaming: true }
    }
    sourcePathOptions.value = []
    targetPathOptions.value = []
    sourceCurrentPath.value = ''
    targetCurrentPath.value = ''
    tempSchedule.value = { month: [], week: [], day: [], hour: [], minute: [] }
    await loadData()
    creatingState.value = 'done'
  } catch (e) {
    creatingState.value = 'idle'
    alert((e as Error).message)
  }
}

// 过滤后的历史记录
const filteredRuns = computed(() => {
  if (historyFilterTaskId.value === null) return runs.value
  return runs.value.filter(r => r.taskId === historyFilterTaskId.value)
})

function viewTaskHistory(taskId: number) {
  historyFilterTaskId.value = taskId
  currentModule.value = 'history'
}

async function stopTaskTransfer(taskId: number) {
  showConfirm('停止传输', '确定停止该任务的当前传输？', async () => {
    try {
      await api.stopTaskTransfer(taskId)
      await refreshTaskViewData()
    } catch (e) {
      alert((e as Error).message)
    }
  })
}

async function viewTaskProgress(taskId: number) {
  try {
    const data = await api.getTaskProgress(taskId)
    taskProgressData.value = data || {}
    showTaskProgressModal.value = true
  } catch (e) {
    alert((e as Error).message)
  }
}

async function runTask(taskId: number) {
  if (runningTaskId.value !== null) return
  runningTaskId.value = taskId
  try {
    await api.runTask(taskId)
    // 5秒后恢复
    setTimeout(() => {
      if (runningTaskId.value === taskId) {
        runningTaskId.value = null
      }
    }, 5000)
  } catch (e) {
    runningTaskId.value = null
    alert((e as Error).message)
  }
}

async function goToAddTask() {
  // Reload remotes before switching to add mode
  try {
    const remoteData = await api.listRemotes()
    remotes.value = remoteData?.remotes || []
  } catch (e) {
    console.error('Failed to load remotes:', e)
  }
  currentModule.value = 'add'
  openMenuId.value = null
}

function editTask(task: Task) {
  editingTask.value = task
  
  // 查找该任务的定时规则
  const schedule = getScheduleByTaskId(task.id)
  
  if (schedule) {
    // 解析 spec: year,month,week,day,hour,minute
    const parts = schedule.spec.split(',')
    createForm.value = {
      name: task.name,
      mode: task.mode,
      sourceRemote: task.sourceRemote,
      sourcePath: task.sourcePath || '',
      targetRemote: task.targetRemote,
      targetPath: task.targetPath || '',
      enableSchedule: true,
      scheduleMinute: parts[0] || '00',
      scheduleHour: parts[1] || '*',
      scheduleDay: parts[2] || '*',
      scheduleMonth: parts[3] || '*',
      scheduleWeek: parts[4] || '*',
      options: task.options || {},
    }
    // 更新临时选择状态
    tempSchedule.value = {
      minute: parts[0] && parts[0] !== '*' ? parts[0].split(',') : [],
      hour: parts[1] && parts[1] !== '*' ? parts[1].split(',') : [],
      day: parts[2] && parts[2] !== '*' ? parts[2].split(',') : [],
      month: parts[3] && parts[3] !== '*' ? parts[3].split(',') : [],
      week: parts[4] && parts[4] !== '*' ? parts[4].split(',') : [],
    }
  } else {
    createForm.value = {
      name: task.name,
      mode: task.mode,
      sourceRemote: task.sourceRemote,
      sourcePath: task.sourcePath || '',
      targetRemote: task.targetRemote,
      targetPath: task.targetPath || '',
      enableSchedule: false,
      scheduleMonth: '*',
      scheduleWeek: '',
      scheduleDay: '',
      scheduleHour: '00',
      scheduleMinute: '00',
      options: normalizeTaskOptions(task.options as Record<string, any>),
    }
    tempSchedule.value = { month: [], week: [], day: [], hour: [], minute: [] }
  }
  
  // 加载源路径选项 - 加载父目录以便显示文件
  if (task.sourceRemote) {
    const parentPath = getParentPath(task.sourcePath || '')
    sourceCurrentPath.value = parentPath
    loadSourcePath(task.sourceRemote, parentPath)
  }
  // 加载目标路径选项
  if (task.targetRemote) {
    const parentPath = getParentPath(task.targetPath || '')
    targetCurrentPath.value = parentPath
    loadTargetPath(task.targetRemote, parentPath)
  }
  currentModule.value = 'add'
  openMenuId.value = null
}

function getParentPath(path: string): string {
  if (!path) return ''
  const parts = path.split('/')
  parts.pop()
  return parts.join('/')
}

function showConfirm(title: string, message: string, onConfirm: () => void) {
  confirmModal.value = { show: true, title, message, onConfirm }
}

async function deleteTask(id: number) {
  showConfirm('删除任务', '确定删除此任务？此操作不可恢复！', async () => {
    try {
      await api.deleteTask(id)
      openMenuId.value = null
      await loadData()
    } catch (e) {
      alert((e as Error).message)
    }
  })
}

function getScheduleByTaskId(taskId: number) {
  return schedules.value.find(s => s.taskId === taskId)
}

async function toggleSchedule(taskId: number) {
  const schedule = getScheduleByTaskId(taskId)
  if (!schedule) return
  try {
    await api.updateSchedule(schedule.id, !schedule.enabled)
    await loadData()
  } catch (e) {
    alert((e as Error).message)
  }
}

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

function toggleMenu(id: number) {
  openMenuId.value = openMenuId.value === id ? null : id
}

function closeMenus() {
  openMenuId.value = null
}

async function deleteSchedule(id: number) {
  if (!confirm('确定删除此定时任务？')) return
  try {
    await api.deleteSchedule(id)
    await loadData()
  } catch (e) {
    alert((e as Error).message)
  }
}

function showRunDetail(run: any) {
  runDetail.value = run
  showDetailModal.value = true
}

function formatSummary(summary: any): string {
  if (!summary) return '-'
  if (typeof summary === 'string') {
    try {
      summary = JSON.parse(summary)
    } catch {
      return summary
    }
  }
  if (typeof summary !== 'object') return String(summary)
  
  const parts: string[] = []
  if (summary.bytes !== undefined) {
    parts.push(`传输量: ${formatBytes(summary.bytes)}`)
  }
  if (summary.transferred !== undefined) {
    parts.push(`文件数: ${summary.transferred}`)
  }
  if (summary.deleted !== undefined) {
    parts.push(`删除: ${summary.deleted}`)
  }
  if (summary.errors !== undefined) {
    parts.push(`错误: ${summary.errors}`)
  }
  if (summary.speed !== undefined) {
    parts.push(`速度: ${summary.speed}`)
  }
  if (summary.elapsedTime !== undefined) {
    parts.push(`耗时: ${summary.elapsedTime}`)
  }
  if (summary.finished !== undefined) {
    parts.push(`完成: ${summary.finished ? '是' : '否'}`)
  }
  if (summary.success !== undefined) {
    parts.push(`成功: ${summary.success ? '是' : '否'}`)
  }
  if (summary.streamingEnabled !== undefined) {
    parts.push(`流式传输: ${summary.streamingEnabled ? '开启' : '关闭'}`)
  }
  if (summary.effectiveOptions && typeof summary.effectiveOptions === 'object') {
    const opts = summary.effectiveOptions
    const keyParts:string[] = []
    if (opts.transfers !== undefined) keyParts.push(`transfers=${opts.transfers}`)
    if (opts.multiThreadStreams !== undefined) keyParts.push(`multiThreadStreams=${opts.multiThreadStreams}`)
    if (opts.multiThreadCutoff !== undefined) keyParts.push(`multiThreadCutoff=${opts.multiThreadCutoff}`)
    if (opts.bufferSize !== undefined) keyParts.push(`bufferSize=${opts.bufferSize}`)
    if (opts.timeout !== undefined) keyParts.push(`timeout=${opts.timeout}`)
    if (keyParts.length) parts.push(`生效参数: ${keyParts.join(', ')}`)
  }
  return parts.length > 0 ? parts.join('\n') : JSON.stringify(summary, null, 2)
}

function formatBytes(bytes: number): string {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

function formatBytesPerSec(bytesPerSec: number): string {
  if (!bytesPerSec || bytesPerSec === 0) return '-'
  return formatBytes(bytesPerSec) + '/s'
}

function formatEta(seconds: number | null): string {
  if (!seconds || seconds <= 0) return '-'
  const hours = Math.floor(seconds / 3600)
  const minutes = Math.floor((seconds % 3600) / 60)
  const secs = Math.floor(seconds % 60)
  if (hours > 0) {
    return `${hours}小时${minutes}分${secs}秒`
  }
  if (minutes > 0) {
    return `${minutes}分${secs}秒`
  }
  return `${secs}秒`
}

async function clearRun(id: number) {
  try {
    await api.clearRun(id)
    await loadData()
  } catch (e) {
    alert((e as Error).message)
  }
}

async function clearAllRuns() {
  if (historyFilterTaskId.value === null) {
    alert('请先选择任务')
    return
  }
  showConfirm('删除所有历史', '确定删除该任务所有历史记录？此操作不可恢复！', async () => {
    try {
      await api.clearRunsByTask(historyFilterTaskId.value)
      await loadData()
    } catch (e) {
      alert((e as Error).message)
    }
  })
}

function onSourceRemoteChange() {
  sourceCurrentPath.value = ''
  if (createForm.value.sourceRemote) {
    loadSourcePath(createForm.value.sourceRemote, '')
  } else {
    sourcePathOptions.value = []
  }
  createForm.value.sourcePath = ''
}

function onTargetRemoteChange() {
  targetCurrentPath.value = ''
  if (createForm.value.targetRemote) {
    loadTargetPath(createForm.value.targetRemote, '')
  } else {
    targetPathOptions.value = []
  }
  createForm.value.targetPath = ''
}

async function loadSourcePath(remote: string, path: string) {
  try {
    const data = await api.listPath(remote, path)
    sourcePathOptions.value = data.items || []
    sourceCurrentPath.value = path
  } catch (e) {
    console.error(e)
  }
}

async function loadTargetPath(remote: string, path: string) {
  try {
    const data = await api.listPath(remote, path)
    targetPathOptions.value = data.items || []
    targetCurrentPath.value = path
  } catch (e) {
    console.error(e)
  }
}

function openSourceDir(item: any) {
  if (item.IsDir) {
    loadSourcePath(createForm.value.sourceRemote, item.Path)
  }
}

function openAndSetSource(item: any) {
  // 箭头点击：打开并设置为源路径
  if (item.IsDir) {
    createForm.value.sourcePath = item.Path
    loadSourcePath(createForm.value.sourceRemote, item.Path)
  }
}

function onSourceClick(item: any) {
  // 单击行：直接设置源路径
  createForm.value.sourcePath = item.Path
  showSourcePathInput.value = false
}

function onSourceArrow(item: any) {
  // 箭头点击：打开并设置为源路径
  createForm.value.sourcePath = item.Path
  loadSourcePath(createForm.value.sourceRemote, item.Path)
}

function openTargetDir(item: any) {
  if (item.IsDir) {
    loadTargetPath(createForm.value.targetRemote, item.Path)
  }
}

function openAndSetTarget(item: any) {
  // 箭头点击：打开并设置为目标路径
  if (item.IsDir) {
    createForm.value.targetPath = item.Path
    loadTargetPath(createForm.value.targetRemote, item.Path)
  }
}

function onTargetClick(item: any) {
  // 单击行：直接设置目标路径
  createForm.value.targetPath = item.Path
  showTargetPathInput.value = false
}

function onTargetArrow(item: any) {
  // 箭头点击：打开并设置目标路径
  createForm.value.targetPath = item.Path
  loadTargetPath(createForm.value.targetRemote, item.Path)
}

function selectSourceDir(item: any) {
  // 单击选中文件夹（填充路径但不进入）
  createForm.value.sourcePath = item.Path
  showSourcePathInput.value = false
}

function selectTargetDir(item: any) {
  // 单击选中文件夹（填充路径但不进入）
  createForm.value.targetPath = item.Path
  showTargetPathInput.value = false
}

function selectSourceFile(item: any) {
  if (!item.IsDir) {
    createForm.value.sourcePath = item.Path
    showSourcePathInput.value = false
  }
}

function selectTargetFile(item: any) {
  if (!item.IsDir) {
    createForm.value.targetPath = item.Path
    showTargetPathInput.value = false
  }
}

function goBackSource() {
  const parts = sourceCurrentPath.value.split('/')
  parts.pop()
  const parentPath = parts.join('/')
  loadSourcePath(createForm.value.sourceRemote, parentPath)
}

function goBackTarget() {
  const parts = targetCurrentPath.value.split('/')
  parts.pop()
  const parentPath = parts.join('/')
  loadTargetPath(createForm.value.targetRemote, parentPath)
}
</script>

<template>
  <div v-if="currentModule === 'tasks'" class="card">
    <div class="card-header">
      <div class="title">任务列表</div>
      <div class="header-actions">
        <input v-model="taskSearch" type="text" placeholder="搜索任务..." class="search-input" />
        <button class="secondary small" @click="openGlobalStats">全局实时数据</button>
        <button class="primary small" @click="goToAddTask">+ 添加任务</button>
      </div>
    </div>
    <div class="list-header">
      <span class="col-name">任务</span>
      <span class="col-schedule">定时</span>
      <span class="col-action">操作</span>
    </div>
    <div class="list">
      <div v-for="task in filteredTasks" :key="task.id" class="task-item">
        <div class="task-main">
          <div class="name">
            <strong>{{ task.name }}</strong>
            <span class="mode-tag">{{ task.mode }}</span>
          </div>
          <div class="schedule-info">
            <template v-if="getScheduleByTaskId(task.id)">
              <span :class="['schedule-badge', getScheduleByTaskId(task.id)?.enabled ? 'enabled' : 'disabled']">
                {{ getScheduleByTaskId(task.id)?.enabled ? '已启用' : '已禁用' }}
              </span>
              <span class="schedule-rule">{{ formatScheduleSpec(getScheduleByTaskId(task.id)?.spec || '') }}</span>
            </template>
            <span v-else class="no-schedule">未设置</span>
          </div>
          <div class="item-actions">
            <button class="ghost small" @click.stop="viewTaskHistory(task.id)">📋 历史</button>
            <button class="ghost small" :disabled="!getActiveRunByTaskId(task.id)" @click.stop="stopTaskTransfer(task.id)">⏹ 停止传输</button>
            <button class="ghost small" @click.stop="viewTaskProgress(task.id)">📊 实时进度</button>
            <button v-if="getScheduleByTaskId(task.id)" class="ghost small" @click.stop="toggleSchedule(task.id)">
              {{ getScheduleByTaskId(task.id)?.enabled ? '⏸ 关闭' : '▶ 开启' }}
            </button>
            <button 
              class="ghost small" 
              :class="{ 'btn-running': runningTaskId === task.id }"
              :disabled="runningTaskId === task.id"
              @click.stop="runTask(task.id)"
            >
              <template v-if="runningTaskId === task.id">运行成功</template>
              <template v-else>▶ 手动运行</template>
            </button>
            <button class="ghost small" @click.stop="editTask(task)">✏️</button>
            <button class="ghost small danger-text" @click.stop="deleteTask(task.id)">🗑️</button>
          </div>
        </div>
        <div class="task-paths">
          <div class="path-row">
            <span class="path-label">源:</span>
            <span class="path-value">{{ task.sourceRemote }}:{{ task.sourcePath || '根目录' }}</span>
          </div>
          <div class="path-row">
            <span class="path-label">目标:</span>
            <span class="path-value">{{ task.targetRemote }}:{{ task.targetPath || '根目录' }}</span>
          </div>
          <template v-if="getTaskRealtimeProgress(task.id)">
            <div class="path-row">
              <span class="path-label">实时:</span>
              <span class="path-value">运行中 · Job #{{ getTaskRealtimeProgress(task.id)?.run?.rcJobId }}</span>
            </div>
            <div class="path-row">
              <span class="path-label">进度:</span>
              <span class="path-value">{{ (getTaskRealtimeProgress(task.id)?.percentage || 0).toFixed(2) }}% · {{ formatBytes(getTaskRealtimeProgress(task.id)?.bytes || 0) }} / {{ formatBytes(getTaskRealtimeProgress(task.id)?.totalBytes || 0) }} · {{ formatBytesPerSec(getTaskRealtimeProgress(task.id)?.speed || 0) }} · ETA {{ formatEta(getTaskRealtimeProgress(task.id)?.eta || null) }}</span>
            </div>
            <div class="progress-bar-container" style="margin-top:8px">
              <div class="progress-bar" :style="{ width: ((getTaskRealtimeProgress(task.id)?.percentage || 0)) + '%' }"></div>
            </div>
          </template>
        </div>
      </div>
      <div v-if="!filteredTasks.length" class="empty">暂无任务</div>
    </div>
  </div>

  <div v-if="currentModule === 'history'" class="card">
    <div class="card-header">
      <div class="title">历史记录</div>
      <div class="header-actions">
        <button v-if="historyFilterTaskId !== null && filteredRuns.length > 0" class="ghost small danger-text" @click="clearAllRuns">删除所有</button>
        <button v-if="historyFilterTaskId !== null" class="ghost small" @click="currentModule = 'tasks'">
          ← 返回
        </button>
      </div>
    </div>
    <div class="list-header">
      <span class="col-name">任务</span>
      <span class="col-status">状态</span>
      <span class="col-path-full">路径</span>
      <span class="col-time">开始</span>
      <span class="col-time">耗时</span>
      <span class="col-action">操作</span>
    </div>
    <div class="list">
      <div v-for="run in filteredRuns" :key="run.id" class="item run-item">
        <div class="name">
          <strong>{{ run.taskName || `任务 #${run.taskId}` }}</strong>
          <span class="mode-tag" v-if="run.taskMode">{{ run.taskMode }}</span>
        </div>
        <span 
          :class="['status', getStatusClass(run.status), 'clickable']"
          @click="showRunDetail(run)"
        >{{ getStatusText(run.status) }}</span>
        <div class="path-full">
          <span class="path-text">{{ run.sourceRemote || '?' }}:{{ run.sourcePath || '/' }} → {{ run.targetRemote || '?' }}:{{ run.targetPath || '/' }}</span>
        </div>
        <span class="time">{{ formatTime(run.startedAt) }}</span>
        <span class="time">{{ formatDuration(run.startedAt, run.finishedAt) }}</span>
        <button class="ghost small" @click="showRunDetail(run)">详情</button>
        <button class="ghost small danger-text" @click="clearRun(run.id)">清除</button>
      </div>
      <div v-if="!filteredRuns.length" class="empty">暂无历史记录</div>
    </div>

    <!-- 运行详情弹窗 -->
    <div v-if="showDetailModal" class="modal-overlay" @click.self="showDetailModal = false">
      <div class="modal-content">
        <div class="modal-header">
          <h3>运行详情</h3>
          <button class="close-btn" @click="showDetailModal = false">×</button>
        </div>
        <div class="modal-body">
          <div class="detail-item">
            <label>任务名称：</label>
            <span>{{ runDetail.taskName || '-' }}</span>
          </div>
          <div class="detail-item">
            <label>执行模式：</label>
            <span>{{ runDetail.taskMode || '-' }}</span>
          </div>
          <div class="detail-item">
            <label>状态：</label>
            <span :class="['status', getStatusClass(runDetail.status)]">{{ getStatusText(runDetail.status) }}</span>
          </div>
          <div class="detail-item">
            <label>触发方式：</label>
            <span>{{ runDetail.trigger === 'schedule' ? '定时任务' : '手动执行' }}</span>
          </div>
          <div class="detail-item full-width">
            <label>源路径：</label>
            <span>{{ runDetail.sourceRemote }}:{{ runDetail.sourcePath || '/' }}</span>
          </div>
          <div class="detail-item full-width">
            <label>目标路径：</label>
            <span>{{ runDetail.targetRemote }}:{{ runDetail.targetPath || '/' }}</span>
          </div>
          <div class="detail-item">
            <label>开始时间：</label>
            <span>{{ formatTime(runDetail.startedAt) }}</span>
          </div>
          <div class="detail-item">
            <label>结束时间：</label>
            <span>{{ formatTime(runDetail.finishedAt) }}</span>
          </div>
          <div class="detail-item">
            <label>耗时：</label>
            <span>{{ formatDuration(runDetail.startedAt, runDetail.finishedAt) }}</span>
          </div>
          <div class="detail-item">
            <label>传输速度：</label>
            <span>{{ runDetail.speed || '-' }}</span>
          </div>
          <div class="detail-item full-width">
            <label>传输统计：</label>
            <pre class="summary-pre">{{ formatSummary(runDetail.summary) }}</pre>
          </div>
          <div v-if="runDetail.error" class="detail-item full-width">
            <label>错误信息：</label>
            <span class="error-text">{{ runDetail.error }}</span>
          </div>
        </div>
      </div>
    </div>
  </div>

    <div v-if="currentModule === 'add'" class="card">
    <div class="card-header"><div class="title">添加任务</div></div>
    <div class="form-content">
      <div class="field-item">
        <label>任务名称 <span style="color: #dc2626">*</span></label>
        <input v-model="createForm.name" type="text" placeholder="输入任务名称" />
      </div>
      <div class="field-item">
        <label>模式</label>
        <select v-model="createForm.mode">
          <option value="copy">复制 (copy)</option>
          <option value="sync">同步 (sync)</option>
          <option value="move">移动 (move)</option>
        </select>
      </div>
      <div class="field-item">
        <label>源存储 <span style="color: #dc2626">*</span></label>
        <select v-model="createForm.sourceRemote" @change="onSourceRemoteChange">
          <option value="">选择源存储</option>
          <option v-for="r in remotes" :key="r" :value="r">{{ r }}</option>
        </select>
      </div>
      <div class="field-item">
        <label>源路径</label>
        <div class="path-selector">
          <div class="path-browse">
            <div class="path-bar">
              <span class="path-selected">已选: /{{ createForm.sourcePath || '未选择' }}</span>
              <span class="path-label">当前: /{{ sourceCurrentPath || '根目录' }}</span>
              <button v-if="sourceCurrentPath" type="button" class="ghost small" @click="goBackSource">返回</button>
            </div>
            <div class="path-list">
              <div v-for="item in sourcePathOptions" :key="item.Path" class="path-item" :class="{ 'is-dir': item.IsDir }" @click="onSourceClick(item)">
                <span v-if="item.IsDir" class="folder-icon" @click.stop="onSourceArrow(item)">📁</span>
                <span v-else class="file-icon">📄</span>
                <span class="item-name">{{ item.Name }}</span>
              </div>
              <div v-if="!sourcePathOptions.length" class="path-empty">空目录</div>
            </div>
          </div>
          <button type="button" class="ghost small" @click="showSourcePathInput = !showSourcePathInput">手动输入</button>
        </div>
        <input v-if="showSourcePathInput" v-model="createForm.sourcePath" type="text" placeholder="手动输入路径" style="margin-top: 8px" />
      </div>
      <div class="field-item">
        <label>目标存储 <span style="color: #dc2626">*</span></label>
        <select v-model="createForm.targetRemote" @change="onTargetRemoteChange">
          <option value="">选择目标存储</option>
          <option v-for="r in remotes" :key="r" :value="r">{{ r }}</option>
        </select>
      </div>
      <div class="field-item">
        <label>目标路径</label>
        <div class="path-selector">
          <div class="path-browse">
            <div class="path-bar">
              <span class="path-selected">已选: /{{ createForm.targetPath || '未选择' }}</span>
              <span class="path-label">当前: /{{ targetCurrentPath || '根目录' }}</span>
              <button v-if="targetCurrentPath" type="button" class="ghost small" @click="goBackTarget">返回</button>
            </div>
            <div class="path-list">
              <div v-for="item in targetPathOptions" :key="item.Path" class="path-item" :class="{ 'is-dir': item.IsDir }" @click="onTargetClick(item)">
                <span v-if="item.IsDir" class="folder-icon" @click.stop="onTargetArrow(item)">📁</span>
                <span v-else class="file-icon">📄</span>
                <span class="item-name">{{ item.Name }}</span>
              </div>
              <div v-if="!targetPathOptions.length" class="path-empty">空目录</div>
            </div>
          </div>
          <button type="button" class="ghost small" @click="showTargetPathInput = !showTargetPathInput">手动输入</button>
        </div>
        <input v-if="showTargetPathInput" v-model="createForm.targetPath" type="text" placeholder="手动输入路径" style="margin-top: 8px" />
      </div>

      <!-- 定时任务设置 -->
      <div class="schedule-section">
        <div class="section-header">
          <label class="schedule-toggle">
            <input type="checkbox" v-model="createForm.enableSchedule" />
            <span>启用定时任务</span>
          </label>
        </div>
        <div v-if="createForm.enableSchedule" class="schedule-grid">
          <!-- 月 -->
          <div class="schedule-item">
            <label>月</label>
            <select v-model="tempSchedule.month" multiple size="6" @dblclick="confirmField('month')">
              <option value="*">每月</option>
              <option v-for="m in [1,2,3,4,5,6,7,8,9,10,11,12]" :key="m" :value="String(m)">{{ m }}月</option>
            </select>
            <button type="button" class="ghost small" @click="confirmField('month')">确定</button>
            <span class="selected-val">{{ createForm.scheduleMonth === '*' ? '每月' : (createForm.scheduleMonth || '每月') }}</span>
          </div>
          <!-- 周 -->
          <div class="schedule-item">
            <label>周</label>
            <select v-model="tempSchedule.week" multiple size="6" @dblclick="confirmField('week')">
              <option value="*">每日</option>
              <option value="">不设置</option>
              <option v-for="(w, idx) in ['周一','周二','周三','周四','周五','周六','周日']" :key="w" :value="String(idx+1)">{{ w }}</option>
            </select>
            <button type="button" class="ghost small" @click="confirmField('week')">确定</button>
            <span class="selected-val">{{ createForm.scheduleWeek === '*' ? '每日' : (createForm.scheduleWeek || '不设置') }}</span>
          </div>
          <!-- 日 -->
          <div class="schedule-item">
            <label>日</label>
            <select v-model="tempSchedule.day" multiple size="6" @dblclick="confirmField('day')">
              <option value="*">每日</option>
              <option value="">不设置</option>
              <option v-for="d in 31" :key="d" :value="String(d)">{{ d }}日</option>
            </select>
            <button type="button" class="ghost small" @click="confirmField('day')">确定</button>
            <span class="selected-val">{{ createForm.scheduleDay === '*' ? '每日' : (createForm.scheduleDay || '不设置') }}</span>
          </div>
          <!-- 时 -->
          <div class="schedule-item">
            <label>时</label>
            <select v-model="tempSchedule.hour" multiple size="6" @dblclick="confirmField('hour')">
              <option value="*">每时</option>
              <option v-for="h in 24" :key="h-1" :value="String(h-1).padStart(2,'0')">{{ String(h-1).padStart(2,'0') }}时</option>
            </select>
            <button type="button" class="ghost small" @click="confirmField('hour')">确定</button>
            <span class="selected-val">{{ createForm.scheduleHour === '*' ? '每时' : (createForm.scheduleHour || '00') + '时' }}</span>
          </div>
          <!-- 分 -->
          <div class="schedule-item">
            <label>分</label>
            <select v-model="tempSchedule.minute" multiple size="6" @dblclick="confirmField('minute')">
              <option value="*">每分</option>
              <option v-for="m in 60" :key="m-1" :value="String(m-1).padStart(2,'0')">{{ String(m-1).padStart(2,'0') }}分</option>
            </select>
            <button type="button" class="ghost small" @click="confirmField('minute')">确定</button>
            <span class="selected-val">{{ createForm.scheduleMinute === '*' ? '每分' : (createForm.scheduleMinute || '00') + '分' }}</span>
          </div>
        </div>
      </div>

      <!-- 高级选项 -->
      <button type="button" class="ghost small" @click="showAdvancedOptions = !showAdvancedOptions">
        {{ showAdvancedOptions ? '收起高级选项' : '+ 高级选项' }}
      </button>
      <div v-if="showAdvancedOptions" class="advanced-section">
        <div class="advanced-group">
          <div class="advanced-group-title">传输策略</div>
          <div class="advanced-row inline">
            <label>开启流式传输（推荐）</label>
            <input type="checkbox" v-model="createForm.options.enableStreaming" />
          </div>
        </div>

        <div class="advanced-group">
          <div class="advanced-group-title">过滤参数</div>
          <div class="advanced-row">
            <label>排除 (exclude)</label>
            <textarea v-model="createForm.options.exclude" placeholder="每行一个规则, 如: *.txt&#10;备份/**" rows="3"></textarea>
          </div>
          <div class="advanced-row">
            <label>包含 (include)</label>
            <textarea v-model="createForm.options.include" placeholder="每行一个规则, 如: *.pdf&#10;文档/**" rows="3"></textarea>
          </div>
          <div class="advanced-row">
            <label>过滤规则 (filter)</label>
            <textarea v-model="createForm.options.filter" placeholder="每行一个规则, 如: - *.tmp&#10;+ *.bak" rows="3"></textarea>
          </div>
          <div class="advanced-row inline">
            <label>忽略大小写</label>
            <input type="checkbox" v-model="createForm.options.ignoreCase" />
          </div>
          <div class="advanced-row inline">
            <label>忽略已存在的文件</label>
            <input type="checkbox" v-model="createForm.options.ignoreExisting" />
          </div>
          <div class="advanced-row inline">
            <label>删除被排除的文件</label>
            <input type="checkbox" v-model="createForm.options.deleteExcluded" />
          </div>
        </div>

        <div class="advanced-group">
          <div class="advanced-group-title">比较策略</div>
          <div class="advanced-row inline">
            <label>校验和比较</label>
            <input type="checkbox" v-model="createForm.options.checksum" />
          </div>
          <div class="advanced-row inline">
            <label>仅按大小</label>
            <input type="checkbox" v-model="createForm.options.sizeOnly" />
          </div>
          <div class="advanced-row inline">
            <label>忽略大小</label>
            <input type="checkbox" v-model="createForm.options.ignoreSize" />
          </div>
          <div class="advanced-row inline">
            <label>忽略时间</label>
            <input type="checkbox" v-model="createForm.options.ignoreTimes" />
          </div>
          <div class="advanced-row inline">
            <label>更新较新的</label>
            <input type="checkbox" v-model="createForm.options.update" />
          </div>
          <div class="advanced-row">
            <label>时间窗口</label>
            <input type="text" v-model="createForm.options.modifyWindow" placeholder="如: 1h2s" />
          </div>
        </div>

        <div class="advanced-group">
          <div class="advanced-group-title">路径策略</div>
          <div class="advanced-row inline">
            <label>不遍历</label>
            <input type="checkbox" v-model="createForm.options.noTraverse" />
          </div>
          <div class="advanced-row inline">
            <label>不检查目标</label>
            <input type="checkbox" v-model="createForm.options.noCheckDest" />
          </div>
          <div class="advanced-row">
            <label>比较目录</label>
            <input type="text" v-model="createForm.options.compareDest" placeholder="remote:path" />
          </div>
          <div class="advanced-row">
            <label>复制目录</label>
            <input type="text" v-model="createForm.options.copyDest" placeholder="remote:path" />
          </div>
        </div>

        <div class="advanced-group">
          <div class="advanced-group-title">传输控制</div>
          <div class="advanced-row">
            <label>并发传输数</label>
            <input type="number" v-model="createForm.options.transfers" min="1" max="100" />
          </div>
          <div class="advanced-row">
            <label>带宽限制</label>
            <input type="text" v-model="createForm.options.bwLimit" placeholder="如: 10M" />
          </div>
          <div class="advanced-row inline">
            <label>多线程传输</label>
            <input type="checkbox" v-model="createForm.options.multiThreadStreams" />
          </div>
          <div class="advanced-row">
            <label>最大传输</label>
            <input type="number" v-model="createForm.options.maxTransfer" min="0" placeholder="字节数, 0表示无限制" />
          </div>
          <div class="advanced-row">
            <label>最大时长</label>
            <input type="number" v-model="createForm.options.maxDuration" min="0" placeholder="秒, 0表示无限制" />
          </div>
        </div>

        <div class="advanced-group">
          <div class="advanced-group-title">同步选项</div>
          <div class="advanced-row">
            <label>删除时机</label>
            <select v-model="createForm.options.cutoffMode">
              <option value="">默认</option>
              <option value="before">删除前</option>
              <option value="during">删除中</option>
              <option value="after">删除后</option>
            </select>
          </div>
          <div class="advanced-row">
            <label>最大删除数</label>
            <input type="number" v-model="createForm.options.maxDelete" min="0" />
          </div>
          <div class="advanced-row inline">
            <label>跟踪重命名</label>
            <input type="checkbox" v-model="createForm.options.trackRenames" />
          </div>
          <div class="advanced-row inline">
            <label>忽略错误</label>
            <input type="checkbox" v-model="createForm.options.ignoreErrors" />
          </div>
        </div>

        <div class="advanced-group">
          <div class="advanced-group-title">其他选项</div>
          <div class="advanced-row inline">
            <label>模拟运行 (dry-run)</label>
            <input type="checkbox" v-model="createForm.options.dryRun" />
          </div>
          <div class="advanced-row inline">
            <label>交互模式</label>
            <input type="checkbox" v-model="createForm.options.interactive" />
          </div>
          <div class="advanced-row inline">
            <label>检查前先检查</label>
            <input type="checkbox" v-model="createForm.options.checkFirst" />
          </div>
          <div class="advanced-row inline">
            <label>服务器端跨配置</label>
            <input type="checkbox" v-model="createForm.options.serverSideAcrossConfigs" />
          </div>
          <div class="advanced-row">
            <label>检查器数</label>
            <input type="number" v-model="createForm.options.checkers" min="1" max="100" />
          </div>
          <div class="advanced-row">
            <label>重试次数</label>
            <input type="number" v-model="createForm.options.retries" min="0" />
          </div>
          <div class="advanced-row">
            <label>备份目录</label>
            <input type="text" v-model="createForm.options.backupDir" placeholder="remote:path" />
          </div>
          <div class="advanced-row">
            <label>日志文件</label>
            <input type="text" v-model="createForm.options.logFile" placeholder="/path/to/log" />
          </div>
        </div>
      </div>

      <div class="form-actions">
        <button 
          class="primary" 
          :class="{ 'btn-success': creatingState === 'done' }"
          :disabled="creatingState === 'loading'"
          @click="createTask"
        >
          <template v-if="creatingState === 'loading'">创建中...</template>
          <template v-else-if="creatingState === 'done'">完成（点击返回任务列表）</template>
          <template v-else-if="editingTask">保存修改</template>
          <template v-else>创建任务</template>
        </button>
      </div>
    </div>
  </div>

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

  <!-- 任务实时进度弹窗 -->
  <div v-if="showTaskProgressModal" class="modal-overlay" @click.self="showTaskProgressModal = false">
    <div class="modal-content">
      <div class="modal-header">
        <h3>任务实时进度 - {{ taskProgressData.name }}</h3>
        <button class="close-btn" @click="showTaskProgressModal = false">×</button>
      </div>
      <div class="modal-body">
        <div class="detail-item">
          <label>任务状态：</label>
          <span :class="taskProgressData.status === 'running' ? 'status-running' : taskProgressData.status === 'finished' ? 'status-success' : 'status-error'">
            {{ taskProgressData.status === 'running' ? '运行中' : taskProgressData.status === 'finished' ? '已完成' : taskProgressData.status === 'no_runs' ? '无运行记录' : taskProgressData.status === 'error' ? '获取失败' : taskProgressData.status }}
          </span>
        </div>
        <div class="detail-item">
          <label>进度：</label>
          <span>{{ taskProgressData.percentage !== undefined ? taskProgressData.percentage.toFixed(2) + '%' : '-' }}</span>
        </div>
        <div class="detail-item">
          <label>当前速度：</label>
          <span>{{ taskProgressData.speed || '-' }}</span>
        </div>
        <div class="detail-item">
          <label>已传输/总大小：</label>
          <span>{{ taskProgressData.bytes ? formatBytes(taskProgressData.bytes) : '-' }} / {{ taskProgressData.totalBytes ? formatBytes(taskProgressData.totalBytes) : '-' }}</span>
        </div>
        <div class="detail-item">
          <label>预计剩余时间：</label>
          <span>{{ taskProgressData.eta || '-' }}</span>
        </div>
        <div class="detail-item" v-if="taskProgressData.error">
          <label>错误信息：</label>
          <span class="error-text">{{ taskProgressData.error }}</span>
        </div>
        <div class="detail-item" v-if="taskProgressData.anomalyMessage">
          <label>异常提示：</label>
          <span class="error-text">{{ taskProgressData.anomalyMessage }}</span>
        </div>
        <div class="progress-bar-container">
          <div class="progress-bar" :style="{ width: (taskProgressData.percentage || 0) + '%' }"></div>
        </div>
      </div>
    </div>
  </div>

  <!-- 确认删除弹窗 -->
  <div v-if="confirmModal.show" class="modal-overlay" @click.self="confirmModal.show = false">
    <div class="modal-content confirm-modal">
      <div class="modal-header">
        <h3>{{ confirmModal.title }}</h3>
        <button class="close-btn" @click="confirmModal.show = false">×</button>
      </div>
      <div class="modal-body">
        <p>{{ confirmModal.message }}</p>
      </div>
      <div class="modal-footer">
        <button class="ghost" @click="confirmModal.show = false">取消</button>
        <button class="primary danger" @click="confirmModal.show = false; confirmModal.onConfirm()">确认</button>
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
  background: #252525;
  border: 1px solid #333;
  color: #888;
  cursor: pointer;
  font-size: 14px;
  font-weight: 500;
  transition: all 0.2s;
}
.tab-btn:hover { background: #2a2a2a; color: #e0e0e0; }
.tab-btn.active { background: #1e3a5f; color: #64b5f6; border-color: #2563a0; }
body.light .tab-btn { background: #f5f5f5; border-color: #ddd; color: #666; }
body.light .tab-btn:hover { background: #e8e8e8; color: #1a1a1a; }
body.light .tab-btn.active { background: #e3f2fd; color: #1976d2; border-color: #bbdefb; }
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
.col-name { flex: 1; }
.col-status { width: 80px; text-align: center; }
.col-time { width: 150px; text-align: right; }
.col-action { width: 80px; text-align: right; }
.col-info { width: 120px; }
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
.status.running { background: #1976d2; color: #fff; }
.status.success { background: #388e3c; color: #fff; }
.status.failed { background: #d32f2f; color: #fff; }
.status.clickable { cursor: pointer; }
.status.clickable:hover { opacity: 0.8; }
.time { width: 150px; text-align: right; color: #888; font-size: 13px; }
.info { width: 120px; color: #888; font-size: 13px; }
.item-actions { display: flex; gap: 8px; }
.danger-text { color: #ef5350 !important; }
.form-content { padding: 20px; }
.form-content .field-item { margin-bottom: 16px; }
.form-content label { display: block; margin-bottom: 6px; font-size: 13px; color: #888; }
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
body.light .form-content input,
body.light .form-content select { background: #fff; border-color: #ddd; color: #333; }
.path-selector { display: flex; gap: 8px; align-items: flex-start; }
.path-browse { flex: 1; border: 1px solid #333; border-radius: 8px; background: #252525; overflow: hidden; }
body.light .path-browse { border-color: #ddd; background: #fff; }
.path-bar { display: flex; justify-content: space-between; align-items: center; padding: 8px 12px; background: #1a1a1a; border-bottom: 1px solid #333; }
body.light .path-bar { background: #f5f5f5; border-color: #ddd; }
.path-selected { font-size: 12px; color: #4caf50; font-weight: 600; }
.path-label { font-size: 12px; color: #888; }
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
.path-item:hover { background: #333; }
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
</style>

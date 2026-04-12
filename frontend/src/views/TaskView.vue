<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import * as api from '../api'
import { getToken } from '../api/auth'
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
// 运行中提示小窗（不切换主窗口，不弹出完整详情）
const showRunningHint = ref(false)
const runningHintRun = ref<any>(null)

let runDetailTimer: any = null

// 运行详情 - 文件列表分页
const runFiles = ref<any[]>([])
const runFilesTotal = ref(0)
const runFilesPage = ref(1)
const runFilesPageSize = ref(Math.max(10, Math.floor((window.innerHeight - 380) / 32)))
const showCreateModal = ref(false)
const showAdvancedOptions = ref(false)
const showGlobalStatsModal = ref(false)
const globalStats = ref<any>({}) // 全局实时数据保留为独立弹窗，与历史态无关
// 已移除“实时进度”弹窗逻辑，卡片直接显示稳态进度

const activeRuns = ref<any[]>([])
// 任务卡片：完成后保留最近稳态进度的观察期（默认 15s）
const LINGER_MS = 15000
const lastStableByTask = ref<Record<number, { sp:any; at:number }>>({})
// DB 稳态进度去抖缓存（非响应式）：保证百分比/已传不回退，避免在渲染期写响应式导致重渲
const lastDbProgressByRunId: Record<number, { bytes:number; totalBytes:number; percentage:number; speed:number }> = {}
const webhookModal = ref<{show:boolean, id:number|null, value:string}>({show:false, id:null, value:''})
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

function setWebhook(task: Task){
  webhookModal.value = { show: true, id: task.id, value: (task.options as any)?.webhookId || '' }
}

async function saveWebhook(){
  const id = webhookModal.value.id
  if(!id) return
  const opts:any = {}
  if (webhookModal.value.value) opts.webhookId = webhookModal.value.value
  await api.updateTask(id, { options: opts } as any)
  webhookModal.value.show = false
  await loadData()
}

// （重复定义已移除）

async function reloadRunFiles(){
  try{
    if (!runDetail.value?.id) return
    const pageOffset = (runFilesPage.value-1) * runFilesPageSize.value
    const res = await api.getRunFiles(runDetail.value.id, pageOffset, runFilesPageSize.value)
    runFiles.value = res.items || []
    runFilesTotal.value = res.total || 0
  }catch(e){ console.error(e) }
}

function pickFilesFromRun(run:any){
  try{
    const sum = typeof run.summary === 'string' ? JSON.parse(run.summary) : run.summary
    if (sum?.finalSummary?.files && Array.isArray(sum.finalSummary.files)) return sum.finalSummary.files
  }catch{}
  return []
}

function showRunDetail(run:any){
  if (run.status === 'running'){
    // 不切主窗口：给出轻量提示小窗，引导去“任务日志”查看实时内容
    runningHintRun.value = run
    showRunningHint.value = true
    return
  }
  runDetail.value = run
  showDetailModal.value = true
  runFilesPage.value = 1
  runFiles.value = []
  runFilesTotal.value = 0
  reloadRunFiles()
}

function closeRunDetail(){
  showDetailModal.value = false
  if (runDetailTimer) { clearInterval(runDetailTimer); runDetailTimer = null }
}

onUnmounted(()=>{
  if (activeRunsTimer) { clearInterval(activeRunsTimer as any); activeRunsTimer = null }
  if (runsTimer) { clearInterval(runsTimer as any); runsTimer = null }
})

const pagedRunFiles = computed(()=> runFiles.value)
const totalRunFilesPages = computed(()=> Math.max(1, Math.ceil((runFilesTotal.value||0)/runFilesPageSize.value)))
function goPrevFilesPage(){ if (runFilesPage.value>1) { runFilesPage.value--; reloadRunFiles() } }
function goNextFilesPage(){ if (runFilesPage.value<totalRunFilesPages.value) { runFilesPage.value++; reloadRunFiles() } }
// finalSummary.files 分页（内存分页）
const finalFiles = computed(()=> (getFinalSummary(runDetail.value)?.files || []) as any[])
// 统计计数
const finalCountAll = computed(()=> finalFiles.value.length)
const finalCountSuccess = computed(()=> finalFiles.value.filter(it=> (it.status||'')==='success').length)
const finalCountFailed = computed(()=> finalFiles.value.filter(it=> (it.status||'')==='failed').length)
const finalCountOther = computed(()=> finalFiles.value.filter(it=> (it.status||'')==='skipped').length)
// move 模式时，成功数量代表 Moved 条数；已在后端合并 Copied+Deleted 为 Moved
// 预估总数（preflight），用于“需完成多少”
function getPreflight(run:any){
  try{
    const sum = typeof run?.summary === 'string' ? JSON.parse(run.summary) : run?.summary
    return sum?.preflight || null
  }catch{ return null }
}
// 可筛选标签：all|success|failed|other
const currentFinalFilter = ref<'all'|'success'|'failed'|'other'>('all')
function setFinalFilter(k:'all'|'success'|'failed'|'other') { currentFinalFilter.value = k; finalFilesPage.value = 1 }
const finalFilteredFiles = computed(()=> {
  if (currentFinalFilter.value==='success') return finalFiles.value.filter(it=> (it.status||'')==='success')
  if (currentFinalFilter.value==='failed') return finalFiles.value.filter(it=> (it.status||'')==='failed')
  if (currentFinalFilter.value==='other') return finalFiles.value.filter(it=> (it.status||'')==='skipped')
  return finalFiles.value
})
// 分页（按筛选后的集）
const finalFilesPageSize = ref(Math.max(10, Math.floor((window.innerHeight - 420) / 34)))
const finalFilesPage = ref(1)
const finalFilesTotal = computed(()=> finalFilteredFiles.value.length)
const totalFinalFilesPages = computed(()=> Math.max(1, Math.ceil((finalFilesTotal.value||0)/finalFilesPageSize.value)))
const pagedFinalFiles = computed(()=> {
  const start = (finalFilesPage.value-1) * finalFilesPageSize.value
  return finalFilteredFiles.value.slice(start, start + finalFilesPageSize.value)
})
const finalFilesJump = ref<number | null>(null)
function goPrevFinalFilesPage(){ if (finalFilesPage.value>1) finalFilesPage.value-- }
function goNextFinalFilesPage(){ if (finalFilesPage.value<totalFinalFilesPages.value) finalFilesPage.value++ }
function jumpFinalFilesPage(){ if (!finalFilesJump.value) return; const p = Math.min(Math.max(1, finalFilesJump.value), totalFinalFilesPages.value); finalFilesPage.value = p }
let activeRunsTimer: number | null = null
let runsTimer: number | null = null // 历史态仅为列表刷新，保留，其他实时逻辑已移除
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

// 命令行模式
const commandMode = ref(false)
const commandText = ref('')

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
  activeRunsTimer = window.setInterval(() => { loadActiveRuns().catch(console.error) }, 2000)
  // 历史列表也轮询，保证新 run 及时出现（避免必须手动刷新）
  runsTimer = window.setInterval(() => { api.listRuns().then(v=> runs.value = v||[]).catch(()=>{}) }, 3000)
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
    const now = Date.now()
    // 更新最后稳态快照
    for (const it of data || []){
      const tid = it.runRecord?.taskId
      const sp = it.stableProgress
      if (tid && sp && typeof sp === 'object'){
        lastStableByTask.value[tid] = { sp, at: now }
      }
    }
    activeRuns.value = data || []
  } catch (e) {
    console.error(e)
  }
}

function getActiveRunByTaskId(taskId: number) {
  // 优先返回活跃项
  const cur = activeRuns.value.find(item => item.runRecord?.taskId === taskId)
  if (cur) return cur
  // 否则在观察期内返回最后稳态快照（仅当确实完成：>=99.9%）
  const st = lastStableByTask.value[taskId]
  if (st && Date.now()-st.at <= LINGER_MS){
    const pct = Number((st.sp||{}).percentage || 0)
    if (!isNaN(pct) && pct >= 99.9) {
      return { runRecord: { taskId }, stableProgress: { ...(st.sp||{}), phase: 'completed', percentage: pct } }
    }
  }
  return undefined as any
}

function getDbProgressDejitter(run:any){
  try{
    const id = run?.id
    const p = getLiveSummaryFromDB(run)
    if (!p){
      // 没有新帧时，保留上一帧（非响应式读取）
      if (id && lastDbProgressByRunId[id]) return lastDbProgressByRunId[id]
      return null
    }
    if (!id) return p
    const last = lastDbProgressByRunId[id]
    // 仅在 phase 前进时更新缓存与展示
    const phase = (p as any).phase || ''
    const taskId = run?.taskId
    if (taskId){
      const lastPhase = lastPhaseByTaskId[taskId] || ''
      if (phase && lastPhase && phase === lastPhase){
        // phase 未前进：返回上一帧，避免在单一阶段内细粒度波动引发的“跳动/假死感”
        if (last) return last
      } else {
        lastPhaseByTaskId[taskId] = phase
      }
    }
    if (last){
      // 非递减守护
      p.bytes = Math.max(last.bytes, p.bytes||0)
      p.totalBytes = Math.max(last.totalBytes, p.totalBytes||0)
      p.percentage = Math.max(last.percentage, p.percentage||0)
    }
    lastDbProgressByRunId[id] = { bytes: p.bytes||0, totalBytes: p.totalBytes||0, percentage: p.percentage||0, speed: p.speed||0 }
    return p
  }catch{ return null }
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
  if (hours > 0) return `${hours}h ${minutes}m ${seconds}s`
  if (minutes > 0) return `${minutes}m ${seconds}s`
  return `${seconds}s`
}

function getFinalSummary(run: any){
  try{
    const sum = typeof run?.summary === 'string' ? JSON.parse(run.summary) : run?.summary
    if (sum && typeof sum === 'object' && sum.finalSummary) return sum.finalSummary
  }catch{}
  return null
}
function getRunDurationText(run:any){
  const fs = getFinalSummary(run)
  if (fs && fs.durationText) return fs.durationText
  // fallback to local compute for running
  return formatDuration(run.startedAt, run.finishedAt)
}
function formatBps(bps:number){
  if (!bps || bps<=0) return '-'
  return formatBytes(bps) + '/s'
}
// 从 DB 的 summary.progress 读取运行中实时（不落库也可兼容）
function getLiveSummaryFromDB(run:any){
  try{
    const sum = typeof run?.summary === 'string' ? JSON.parse(run.summary) : run?.summary
    const p = sum?.progress
    if (p && typeof p === 'object'){
      const bytes = Number(p.bytes || 0)
      const totalBytes = Number(p.totalBytes || 0)
      const speed = Number(p.speed || 0)
      let percentage = Number(p.percentage || 0)
      if ((!percentage || Number.isNaN(percentage)) && totalBytes>0) percentage = (bytes/totalBytes)*100
      return { bytes, totalBytes, speed, percentage }
    }
  }catch{}
  return null
}
// 以“开始时间 + 平均速度”计算更稳健的 ETA（剩余时间，秒）
function calcEtaFromAvg(run:any, live:any){
  try{
    if (!run?.startedAt || !live) return null
    const start = new Date(run.startedAt).getTime()
    const now = Date.now()
    const elapsedSec = Math.max(1, Math.floor((now - start)/1000))
    const bytes = Number(live.bytes||0)
    const total = Number(live.totalBytes||0)
    if (!total || bytes<=0) return null
    const avgBps = bytes / elapsedSec
    if (avgBps <= 0) return null
    const remaining = Math.max(0, total - bytes)
    return Math.floor(remaining / avgBps)
  }catch{ return null }
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

  // 命令行模式：解析 rclone 命令
  if (commandMode.value) {
    try {
      const parsed = parseRcloneCommand(commandText.value)
      createForm.value.mode = parsed.mode
      createForm.value.sourceRemote = parsed.src.remote
      createForm.value.sourcePath = parsed.src.path
      createForm.value.targetRemote = parsed.dst.remote
      createForm.value.targetPath = parsed.dst.path
      createForm.value.options = { ...normalizeTaskOptions(createForm.value.options), ...parsed.options }
    } catch (e) {
      alert('命令解析失败：' + (e as Error).message)
      return
    }
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
      if (createForm.value.enableSchedule) {
        const spec = [
          createForm.value.scheduleMinute || '00',
          createForm.value.scheduleHour || '*',
          createForm.value.scheduleDay || '*',
          createForm.value.scheduleMonth || '*',
          createForm.value.scheduleWeek || '*',
        ].join('|')
        if (oldSchedule) {
          await api.updateSchedule(oldSchedule.id, true, spec)
        } else {
          await api.createSchedule({ taskId: editingTask.value.id, spec, enabled: true })
        }
      } else if (oldSchedule) {
        // 关闭并保留记录
        await api.updateSchedule(oldSchedule.id, false)
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
    await loadData()
    currentModule.value = 'add'
    creatingState.value = 'done'
  } catch (e) {
    creatingState.value = 'idle'
    alert((e as Error).message)
  }
}

function parseRcloneCommand(cmd: string) {
  if (!cmd) throw new Error('命令为空')
  // 简单分词（支持引号包裹）
  const tokens = cmd.match(/(?:"[^"]*"|'[^']*'|\S)+/g) || []
  if (tokens.length < 3) throw new Error('缺少源/目标')
  // 查找子命令
  const sub = tokens[1]
  const mode = sub === 'sync' ? 'sync' : sub === 'move' ? 'move' : 'copy'
  // 源/目标
  const src = parseRemotePath(tokens[2])
  const dst = parseRemotePath(tokens[3])
  // 解析 flags
  const options: Record<string, any> = {}
  for (let i = 4; i < tokens.length; i++) {
    const t = tokens[i]
    if (t.startsWith('--')) {
      const key = t.replace(/^--/, '')
      const next = tokens[i + 1]
      switch (key) {
        case 'bwlimit': options.bwLimit = stripQuotes(next); i++; break
        case 'transfers': options.transfers = Number(next); i++; break
        case 'use-server-modtime': options.useServerModtime = true; break
        case 'size-only': options.sizeOnly = true; break
        case 'verbose': /* ignore */ break
        default:
          // 其他 boolean 开关
          if (!next || next.startsWith('--')) {
            options[toCamel(key)] = true
          } else {
            options[toCamel(key)] = stripQuotes(next); i++
          }
      }
    }
  }
  return { mode, src, dst, options }
}
function parseRemotePath(s: string){
  const m = s.split(':')
  if (m.length < 2) throw new Error('路径格式错误：'+s)
  return { remote: m[0], path: m.slice(1).join(':') || '' }
}
function stripQuotes(s?: string){ return s ? s.replace(/^['\"]|['\"]$/g, '') : s }
function toCamel(s: string){ return s.replace(/-([a-z])/g, (_,c)=>c.toUpperCase()) }

// 过滤后的历史记录
const filteredRuns = computed(() => {
  if (historyFilterTaskId.value === null) return runs.value
  return runs.value.filter(r => r.taskId === historyFilterTaskId.value)
})

function viewTaskHistory(taskId: number) {
  historyFilterTaskId.value = taskId
  currentModule.value = 'history'
}

const stoppedTaskId = ref<number|null>(null)
async function stopTaskAny(taskId: number) {
  try {
    // 直接按任务 ID kill（后端会定位最近 run 并发信号）
    await api.killTask(taskId)
    // 兼容 RC：如仍有 rcJobId，则再尝试停止
    const active = await api.getActiveRuns().catch(()=>[])
    const cur:any = Array.isArray(active) ? active.find(x => x?.runRecord?.taskId === taskId && x?.runRecord?.rcJobId) : null
    if (cur?.runRecord?.rcJobId) {
      await api.stopJob(cur.runRecord.rcJobId)
    }
    // 按钮状态反馈：红色“已经停止”，10秒后恢复
    stoppedTaskId.value = taskId
    setTimeout(()=>{ if (stoppedTaskId.value===taskId) stoppedTaskId.value=null }, 10000)
    // 轻量刷新列表
    await loadData()
  } catch (e) {
    stoppedTaskId.value = null
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
    // 解析 spec: minute|hour|day|month|week
    const parts = schedule.spec.split('|')
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

// duplicate old implementation removed (showRunDetail)

// 已弃用的历史态实时解析已移除，历史仅展示 finalSummary（冻结信息）

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
  const hp = historyProgressFromSummary(summary)
  if (hp){
    parts.push(`进度: ${(hp.percentage||0).toFixed(2)}%`)
    parts.push(`速度: ${formatBytesPerSec(hp.speed||0)}`)
    parts.push(`已传输/总大小: ${formatBytes(hp.bytes||0)} / ${formatBytes(hp.totalBytes||0)}`)
    parts.push(`ETA: ${formatEta(hp.eta||0)}`)
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
import TransferOptions from '../components/TransferOptions.vue'
</script>

<template>
  <div v-if="currentModule === 'tasks'" class="card">
    <div class="card-header">
      <div class="title">任务列表</div>
      <div class="header-actions">
        <input v-model="taskSearch" type="text" placeholder="搜索任务..." class="search-input" />
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
            <button class="ghost small" @click.stop="viewTaskHistory(task.id)">📋 任务历史记录</button>
            <button class="ghost small" :class="{ 'danger-text': stoppedTaskId===task.id }" @click.stop="stopTaskAny(task.id)">
              <template v-if="stoppedTaskId===task.id">⏹ 已经停止</template>
              <template v-else>⏹ 停止传输</template>
            </button>
            <button v-if="getScheduleByTaskId(task.id)" class="ghost small" @click.stop="toggleSchedule(task.id)">
              {{ getScheduleByTaskId(task.id)?.enabled ? '⏸ 关闭定时' : '▶ 开启定时' }}
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
            <button class="ghost small" @click.stop="() => setWebhook(task)">🔗 Webhook</button>
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
          <div class="path-row">
            <span class="path-label">进度:</span>
            <span class="path-value">
              <template v-if="getActiveRunByTaskId(task.id)?.stableProgress?.phase === 'preparing'">
                准备中 · 已传 {{ formatBytes(getActiveRunByTaskId(task.id)?.stableProgress?.bytes || 0) }} · 速度 {{ formatBytesPerSec(getActiveRunByTaskId(task.id)?.stableProgress?.speed || 0) }}
              </template>
              <template v-else>
                {{ ((getActiveRunByTaskId(task.id)?.stableProgress?.percentage) || 0).toFixed(2) }}% ·
                {{ formatBytes(getActiveRunByTaskId(task.id)?.stableProgress?.bytes || 0) }} /
                {{ formatBytes(getActiveRunByTaskId(task.id)?.stableProgress?.totalBytes || 0) }} ·
                {{ formatBytesPerSec(getActiveRunByTaskId(task.id)?.stableProgress?.speed || 0) }}
              </template>
            </span>
          </div>
          <div class="progress-bar-container" style="margin-top:8px" v-if="getActiveRunByTaskId(task.id)?.stableProgress && getActiveRunByTaskId(task.id)?.stableProgress?.phase !== 'preparing'">
            <div class="progress-bar" :style="{ width: ((getActiveRunByTaskId(task.id)?.stableProgress?.percentage || 0)) + '%' }"></div>
          </div>
        </div>
      </div>
      <div v-if="!filteredTasks.length" class="empty">暂无任务</div>
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

  <div v-if="currentModule === 'history'" class="card">
    <div class="card-header">
      <div class="title">任务历史记录</div>
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
        <div class="info" v-if="getFinalSummary(run)"></div>
        <!-- 运行中卡片的实时概览（优先读 active 的 stableProgress，缺失再用 DB 的 summary.progress） -->
        <div class="summary-mini" v-else-if="run.status==='running'">
          <!-- 全部用 DB：百分比/体量/速度/ETA 均取 DB 的 summary.progress；实时完成文件计数也取 DB（progress.completedFiles） -->
          <template v-if="getDbProgressDejitter(run) as dp">
            <span class="chip">进度 {{ (dp.percentage||0).toFixed(2) }}%</span>
            <span class="chip meta">速度 {{ formatBytesPerSec(dp.speed || 0) }}</span>
            <span class="chip meta">已传 {{ formatBytes(dp.bytes || 0) }}</span>
            <span class="chip meta">总量 {{ formatBytes(dp.totalBytes || 0) }}</span>
            <span class="chip meta" v-if="calcEtaFromAvg(run, dp)">ETA {{ formatEta(calcEtaFromAvg(run, dp)||0) }}</span>
            <span class="chip meta" v-if="getPreflight(run)">总数量 <span class="est">{{ getPreflight(run).totalCount }}</span> ／ <span class="act">已传输 {{ dp.completedFiles ?? 0 }}</span></span>
            <!-- 体量（运行中）：总体积/已传输 -->
            <span class="chip meta" v-if="getPreflight(run)">总体积 <span class="est">{{ formatBytes(getPreflight(run).totalBytes || 0) }}</span> ／ <span class="act">已传输 {{ formatBytes(dp.bytes || 0) }}</span></span>
          </template>
          <template v-else-if="getActiveRunByTaskId(run.taskId)?.stableProgress">
            <!-- DB 暂无时才回退 active -->
            <span class="chip">进度 {{ ((getActiveRunByTaskId(run.taskId)?.stableProgress?.percentage)||0).toFixed(2) }}%</span>
            <span class="chip meta">速度 {{ formatBytesPerSec(getActiveRunByTaskId(run.taskId)?.stableProgress?.speed || 0) }}</span>
            <span class="chip meta">已传 {{ formatBytes(getActiveRunByTaskId(run.taskId)?.stableProgress?.bytes || 0) }}</span>
            <span class="chip meta">总量 {{ formatBytes(getActiveRunByTaskId(run.taskId)?.stableProgress?.totalBytes || 0) }}</span>
            <span class="chip meta" v-if="getActiveRunByTaskId(run.taskId)?.stableProgress?.eta">ETA {{ formatEta(getActiveRunByTaskId(run.taskId)?.stableProgress?.eta) }}</span>
            <span class="chip meta" v-if="getPreflight(run)">总数量 <span class="est">{{ getPreflight(run).totalCount }}</span></span>
            <span class="chip meta" v-if="getPreflight(run)">总体积 <span class="est">{{ formatBytes(getPreflight(run).totalBytes || 0) }}</span> ／ <span class="act">已传输 {{ formatBytes(getActiveRunByTaskId(run.taskId)?.stableProgress?.bytes || 0) }}</span></span>
          </template>
        </div>
        <!-- 历史卡片内的一目了然统计概览（完成态） -->
        <div class="summary-mini" v-if="run.status==='finished' && getFinalSummary(run)">
          <span class="chip">总计 {{ (getFinalSummary(run).files?.length || 0) }}</span>
          <span class="chip success">{{ run.taskMode==='move' ? '移动' : '成功' }} {{ run.taskMode==='move' ? (getFinalSummary(run).counts?.copied || 0) : ((getFinalSummary(run).counts?.copied || 0) + (getFinalSummary(run).counts?.deleted || 0)) }}</span>
          <span class="chip failed">失败 {{ getFinalSummary(run).counts?.failed || 0 }}</span>
          <span class="chip other">其他 {{ getFinalSummary(run).counts?.skipped || 0 }}</span>
          <!-- 总数量/已传输、总体积/已传输（仅保留总体积，不再重复“总量”） -->
          <span class="chip meta" v-if="getPreflight(run)">总数量 <span class="est">{{ getPreflight(run).totalCount }}</span> ／ <span class="act">已传输 {{ run.taskMode==='move' ? (getFinalSummary(run).counts?.copied || 0) : ((getFinalSummary(run).counts?.copied || 0) + (getFinalSummary(run).counts?.deleted || 0)) }}</span></span>
          <span class="chip meta" v-if="getPreflight(run)">总体积 <span class="est">{{ formatBytes(getPreflight(run).totalBytes || 0) }}</span> ／ <span class="act">已传输 {{ formatBytes(getFinalSummary(run).transferredBytes || 0) }}</span></span>
          <span class="chip meta">均速 {{ formatBps(getFinalSummary(run).avgSpeedBps || 0) }}</span>
        </div>
        <button class="ghost small" @click="showRunDetail(run)">运行详情</button>
        <button class="ghost small" @click="openRunLog(run)">传输日志</button>
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
          <!-- 从统计概览中展示开始/结束/耗时/进度/速度（数据库无则省略） -->
          <!-- 运行总结（友好统计） -->
          <div class="detail-item full-width">
            <label>运行总结：</label>
            <div class="summary-box" v-if="getFinalSummary(runDetail)">
              <div class="summary-title">统计概览（可筛选）</div>
              <div class="summary-grid">
                <div class="summary-cell clickable" @click="setFinalFilter('all')">
                  <div class="summary-key">总计</div>
                  <div class="summary-val">{{ finalCountAll }}</div>
                </div>
                <div class="summary-cell clickable" @click="setFinalFilter('success')">
                  <div class="summary-key">{{ runDetail.taskMode==='move' ? '移动' : '成功' }}</div>
                  <div class="summary-val">{{ finalCountSuccess }}</div>
                </div>
                <div class="summary-cell clickable" @click="setFinalFilter('failed')">
                  <div class="summary-key">失败</div>
                  <div class="summary-val error-text">{{ finalCountFailed }}</div>
                </div>
                <div class="summary-cell clickable" @click="setFinalFilter('other')">
                  <div class="summary-key">其他</div>
                  <div class="summary-val">{{ finalCountOther }}</div>
                </div>
                <div class="summary-cell" v-if="getPreflight(runDetail)">
                  <div class="summary-key">总数量</div>
                  <div class="summary-val est">{{ getPreflight(runDetail).totalCount }}</div>
                </div>
                <div class="summary-cell">
                  <div class="summary-key">已传输数量</div>
                  <div class="summary-val act">{{ runDetail.taskMode==='move' ? (getFinalSummary(runDetail).counts?.copied || 0) : ((getFinalSummary(runDetail).counts?.copied || 0) + (getFinalSummary(runDetail).counts?.deleted || 0)) }}</div>
                </div>
                <div class="summary-cell" v-if="getPreflight(runDetail)">
                  <div class="summary-key">总体积</div>
                  <div class="summary-val est">{{ formatBytes(getPreflight(runDetail).totalBytes || 0) }}</div>
                </div>
                <div class="summary-cell">
                  <div class="summary-key">已传输体积</div>
                  <div class="summary-val act">{{ formatBytes(getFinalSummary(runDetail)?.transferredBytes || 0) }}</div>
                </div>
                <div class="summary-cell">
                  <div class="summary-key">开始时间</div>
                  <div class="summary-val">{{ formatTime(runDetail.startedAt) }}</div>
                </div>
                <div class="summary-cell">
                  <div class="summary-key">结束时间</div>
                  <div class="summary-val">{{ formatTime(runDetail.finishedAt) }}</div>
                </div>
                <div class="summary-cell">
                  <div class="summary-key">耗时</div>
                  <div class="summary-val">{{ getFinalSummary(runDetail)?.durationText || '-' }}</div>
                </div>
                <div class="summary-cell">
                  <div class="summary-key">平均速度</div>
                  <div class="summary-val">{{ formatBps(getFinalSummary(runDetail)?.avgSpeedBps || 0) }}</div>
                </div>
                <div class="summary-cell" v-if="getFinalSummary(runDetail)">
                  <div class="summary-key">已传输</div>
                  <div class="summary-val">{{ formatBytes(getFinalSummary(runDetail)?.transferredBytes || 0) }}</div>
                </div>
                <div class="summary-cell" v-if="getFinalSummary(runDetail)">
                  <div class="summary-key">总大小</div>
                  <div class="summary-val">{{ formatBytes(getFinalSummary(runDetail)?.totalBytes || 0) }}</div>
                </div>
                <div class="summary-cell" v-if="getLiveSummaryFromDB(runDetail)">
                  <div class="summary-key">（运行中）进度</div>
                  <div class="summary-val">{{ (getLiveSummaryFromDB(runDetail)?.percentage||0).toFixed(2) }}%</div>
                </div>
                <div class="summary-cell" v-if="getLiveSummaryFromDB(runDetail)">
                  <div class="summary-key">（运行中）速度</div>
                  <div class="summary-val">{{ formatBytesPerSec(getLiveSummaryFromDB(runDetail)?.speed || 0) }}</div>
                </div>
              </div>
            </div>
            <pre v-else class="summary-pre">{{ JSON.stringify({ note: '无总结，可查看传输日志' }, null, 2) }}</pre>
          </div>

          <!-- 传输明细（可分页、可跳页） -->
          <div class="detail-item full-width">
            <label>传输明细：</label>
            <div>
              <div class="files-toolbar">
                <span>共 {{ finalFilesTotal }} 条</span>
                <div class="pager-inline">
                  <button class="ghost small" :disabled="finalFilesPage<=1" @click="goPrevFinalFilesPage()">上一页</button>
                  <span>{{ finalFilesPage }}/{{ totalFinalFilesPages }}</span>
                  <button class="ghost small" :disabled="finalFilesPage>=totalFinalFilesPages" @click="goNextFinalFilesPage()">下一页</button>
                  <span>跳转</span>
                  <input class="page-input" v-model.number="finalFilesJump" type="number" min="1" :max="totalFinalFilesPages" />
                  <button class="ghost small" @click="jumpFinalFilesPage()">GO</button>
                </div>
                <button class="ghost small" @click="reloadRunFiles()">刷新</button>
              </div>
              <div class="files-table large">
                <div class="files-header">
                  <span class="name">文件</span>
                  <span class="status">结果</span>
                  <span class="time">时间</span>
                  <span class="size">大小</span>
                </div>
                <div class="files-body">
                  <template v-if="finalFiles && finalFiles.length">
                    <div v-for="it in pagedFinalFiles" :key="(it.path||it.name) + (it.at||'') + (it.status||'')" class="files-row">
                      <span class="name" :title="it.path||it.name">{{ it.path || it.name }}</span>
                      <span class="status" :class="it.status">{{ it.status }}</span>
                      <span class="time">{{ it.at || '-' }}</span>
                      <span class="size">{{ it.sizeBytes ? formatBytes(it.sizeBytes) : '-' }}</span>
                    </div>
                  </template>
                  <template v-else>
                    <div v-for="it in pagedRunFiles" :key="it.name + it.at + it.status" class="files-row">
                      <span class="name" :title="it.name">{{ it.name }}</span>
                      <span class="status" :class="it.status">{{ it.status }}</span>
                      <span class="time">{{ it.at || '-' }}</span>
                      <span class="size">{{ it.sizeBytes ? formatBytes(it.sizeBytes) : '-' }}</span>
                    </div>
                    <div v-if="!pagedRunFiles.length" class="path-empty">无明细（可能日志为空或历史记录较旧）</div>
                  </template>
                </div>
              </div>
              <div class="files-pager" v-if="!finalFiles || !finalFiles.length">
                <button class="ghost small" :disabled="runFilesPage<=1" @click="goPrevFilesPage()">上一页</button>
                <span>{{ runFilesPage }}/{{ totalRunFilesPages }}</span>
                <button class="ghost small" :disabled="runFilesPage>=totalRunFilesPages" @click="goNextFilesPage()">下一页</button>
              </div>
            </div>
          </div>
          <div v-if="runDetail.error" class="detail-item full-width">
            <label>错误信息：</label>
            <span class="error-text">{{ runDetail.error }}</span>
          </div>
        </div>
      </div>
    </div>
  </div>

  <!-- 运行中轻量提示小窗（不切主窗口） -->
  <div v-if="showRunningHint" class="modal-overlay" @click.self="showRunningHint = false">
    <div class="modal-content" style="max-width:520px">
      <div class="modal-header">
        <h3>任务运行中</h3>
        <button class="close-btn" @click="showRunningHint = false">×</button>
      </div>
      <div class="modal-body">
        <p>该任务仍在传输中，运行详情（历史）仅展示最终信息。</p>
        <p>实时日志与进度请点击“传输日志”或查看任务卡片上的实时进度。</p>
        <div class="hint-box">
          <div class="detail-item"><label>任务：</label><span>{{ runningHintRun?.taskName || `#${runningHintRun?.taskId}` }}</span></div>
          <div class="detail-item"><label>阶段：</label><span>{{ getActiveRunByTaskId(runningHintRun?.taskId)?.stableProgress?.phase || '-' }}</span></div>
          <div class="detail-item"><label>进度：</label><span>{{ ((getActiveRunByTaskId(runningHintRun?.taskId)?.stableProgress?.percentage)||0).toFixed(2) }}%</span></div>
        </div>
      </div>
      <div class="modal-footer">
        <button class="primary" @click="() => { openRunLog(runningHintRun); showRunningHint=false }">打开传输日志</button>
        <button class="ghost" @click="showRunningHint=false">我知道了</button>
      </div>
    </div>
  </div>

    <div v-if="currentModule === 'add'" class="card">
    <div class="card-header"><div class="title">添加任务</div></div>
    <div class="form-content">
      <div class="field-item">
        <label class="inline-label">
          <input type="checkbox" v-model="commandMode" />
          <span style="margin-left:8px">命令行模式（可粘贴 rclone 命令）</span>
        </label>
        <textarea v-if="commandMode" v-model="commandText" class="cmd-textarea" rows="4" placeholder='例如: rclone copy FNOS:/HDD/media openlist:/影音媒体/天翼5050 --bwlimit "07:30,2M;17:40,2M;23:00,2M" --use-server-modtime --size-only --verbose --transfers 2'></textarea>
        <p v-if="commandMode" class="hint">保存时将自动解析命令，填充“模式/源/目标/选项”。任务名称仍需手动填写。</p>
      </div>
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


  <!-- Webhook 设置弹窗 -->
  <div v-if="webhookModal.show" class="modal-overlay" @click.self="webhookModal.show = false">
    <div class="modal-content">
      <div class="modal-header">
        <h3>Webhook 触发设置</h3>
        <button class="close-btn" @click="webhookModal.show = false">×</button>
      </div>
      <div class="modal-body">
        <div class="detail-item">
          <label>触发 ID（可选，留空则使用任务 ID）</label>
          <input v-model="webhookModal.value" placeholder="例如：gate-front-01（可留空）" />
        </div>
        <div class="detail-item">
          <label>触发 URL 示例</label>
          <div class="hint">/webhook/&lt;任务ID&gt; 或 /webhook/{{ webhookModal.value || 'YOUR_ID' }}</div>
        </div>
      </div>
      <div class="modal-footer">
        <button class="ghost" @click="webhookModal.show = false">取消</button>
        <button class="primary" @click="saveWebhook">保存</button>
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
.status.running { background: #1976d2; color: #fff; }
.status.success { background: #388e3c; color: #fff; }
.status.failed { background: #d32f2f; color: #fff; }
.status.clickable { cursor: pointer; }
.status.clickable:hover { opacity: 0.8; }
.time { width: 150px; text-align: right; color: #888; font-size: 13px; }
.info { width: 120px; color: #888; font-size: 13px; }
.summary-mini { display:flex; gap:8px; flex-wrap:wrap; margin:6px 0; }
.summary-mini .chip { padding:2px 8px; border-radius:999px; background:#1f2937; color:#e5e7eb; font-size:12px; border:1px solid #334155 }
.summary-mini .chip.success{ background:#0a2f22; border-color:#14532d; color:#34d399 }
.summary-mini .chip.failed{ background:#2b0a0a; border-color:#7f1d1d; color:#f87171 }
.summary-mini .chip.other{ background:#2b1a04; border-color:#92400e; color:#fbbf24 }
.summary-mini .chip.meta{ background:#111827; border-color:#374151; color:#cbd5e1 }
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
.cmd-textarea{ width:100%; min-height:120px; padding:12px 14px; border-radius:10px; border:1px solid #333; background:#252525; color:#e0e0e0; font-size:14px; box-sizing:border-box; resize:vertical; }
body.light .cmd-textarea{ background:#fff; border-color:#ddd; color:#333 }
.form-content label.inline-label{ display:flex !important; align-items:center; gap:8px; margin:0 0 6px 0; }
.form-content label.inline-label input[type="checkbox"]{ width:16px; height:16px; }
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
.modal-content.log-modal{width:92vw !important; max-width:1200px !important; max-height:80vh; display:flex; flex-direction:column;}
.log-modal .modal-body{padding:12px 16px; width:100%; flex:1; overflow:hidden; display:flex;}
.log-box{width:100%; display:flex; justify-content:center;}
.log-pre{background:#0b1220;color:#e5e7eb;padding:12px;border-radius:8px;height:100%;overflow:auto;white-space:pre-wrap;width:calc(100% - 64px);max-width:1100px;box-sizing:border-box;margin:0;border:1px solid #334155}

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
</style>

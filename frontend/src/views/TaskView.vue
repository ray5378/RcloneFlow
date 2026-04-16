<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import * as api from '../api'
import { TaskCard, RunItem, ScheduleOptions, AdvancedOptions } from '../components/task'
import { ToastItem } from '../components/toast'
import { FileItem } from '../components/files'
import { PathItem } from '../components/path'
import { taskApi, remoteApi, runApi, queueApi, jobApi, scheduleApi } from '../composables/useApi'
import { handleError, showSuccess, setErrorHandler } from '../composables/useError'
import { formatBytes, formatBytesPerSec, formatDuration, formatEta } from '../utils/format'
import { getToken } from '../api/auth'
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

// 任务分页
const tasksPage = ref(1)
const tasksPageSize = 20
const tasksJumpPage = ref(1)
const tasksTotal = computed(() => filteredTasksRaw.value.length)
const currentTasksPages = computed(() => Math.max(1, Math.ceil(tasksTotal.value / tasksPageSize)))

function jumpToTasksPage() {
  const page = Math.min(Math.max(1, tasksJumpPage.value || 1), currentTasksPages.value)
  tasksPage.value = page
  tasksJumpPage.value = page
}

// 过滤后的任务列表（原始）
const filteredTasksRaw = computed(() => {
  if (!taskSearch.value) return tasks.value
  const q = taskSearch.value.toLowerCase()
  return tasks.value.filter(t =>
    t.name.toLowerCase().includes(q) ||
    t.sourceRemote.toLowerCase().includes(q) ||
    t.targetRemote.toLowerCase().includes(q) ||
    t.mode.toLowerCase().includes(q)
  )
})

// 过滤后的任务列表（分页后）
const filteredTasks = computed(() => {
  const start = (tasksPage.value - 1) * tasksPageSize
  const end = start + tasksPageSize
  return filteredTasksRaw.value.slice(start, end)
})

const remotes = ref<string[]>([])
const currentModule = ref<'history' | 'add' | 'tasks'>('tasks')
const historyFilterTaskId = ref<number | null>(null)
const historyStatusFilter = ref<string>('all') // 'all' | 'finished' | 'failed' | 'skipped' | 'hasTransfer'
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
// 已移除"实时进度"弹窗逻辑，卡片直接显示稳态进度

const activeRuns = ref<any[]>([])
// 任务卡片：完成后保留最近稳态进度的观察期（默认 15s）
const LINGER_MS = 20000
const lastStableByTask = ref<Record<number, { sp:any; at:number }>>({})
// 监控帧是否停滞超过阈值（默认 25s），若是则强制刷新一次
const STUCK_MS = 25000
let lastRenderedSignature = ''
let stuckTimer: any = null
onMounted(() => {
  stuckTimer = setInterval(() => {
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
})
onUnmounted(() => { if (stuckTimer) clearInterval(stuckTimer) })
// 仅用 DB 的 summary.progress；DB 暂无时保留上一帧，不再回退 active
const lastDbFrameByRunId: Record<number, any> = {}
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

async function reloadRunFiles(){
  try{
    if (!runDetail.value?.id) return
    const pageOffset = (runFilesPage.value-1) * runFilesPageSize.value
    const res = await runApi.getFiles(runDetail.value.id, pageOffset, runFilesPageSize.value)
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
    // 不切主窗口：给出轻量提示小窗，引导去"任务日志"查看实时内容
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
// 预估总数（preflight），用于"需完成多少"
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
  runsTimer = window.setInterval(() => { runApi.list(runsPage.value, runsPageSize).then(v=> { if(v?.runs) { runs.value = v.runs; runsTotal.value = typeof v.total === 'number' ? v.total : (v.runs?.length || 0) } }).catch(()=>{}) }, 3000)
})

let loadSeq = 0
async function loadData() {
  const seq = ++loadSeq
  try {
    const [taskData, remoteData, scheduleData, runResult] = await Promise.all([
      taskApi.list(),
      remoteApi.list(),
      scheduleApi.list(),
      runApi.list(runsPage.value, runsPageSize),
    ])
    if (seq !== loadSeq) return // 只接受最新一轮
    // 防御：只有明确有数据才更新，防止空数据覆盖
    if (Array.isArray(taskData) && taskData.length > 0) tasks.value = taskData
    if (Array.isArray(remoteData?.remotes) && remoteData.remotes.length > 0) remotes.value = remoteData.remotes
    if (Array.isArray(scheduleData) && scheduleData.length > 0) schedules.value = scheduleData
    if (runResult?.runs) {
      runs.value = runResult.runs
      runsTotal.value = typeof runResult.total === 'number' ? runResult.total : (runResult.runs?.length || 0)
    }
    // 本地快照：成功后保存
    try { localStorage.setItem('lastTasksSnapshot', JSON.stringify(tasks.value||[])) } catch {}
  } catch (e:any) {
    console.error(e)
    // 失败不覆写：保留上一帧；若当前为空，尝试用本地快照兜底
    if (!tasks.value || tasks.value.length===0) {
      try { const snap = JSON.parse(localStorage.getItem('lastTasksSnapshot')||'[]'); if (Array.isArray(snap)) tasks.value = snap } catch {}
    }
  }
}

async function loadActiveRuns() {
  try {
    const data = await jobApi.list()
    const now = Date.now()
    // 更新最后稳态快照（保持向前：非递减合并，避免抖动）
    for (const it of data || []){
      const tid = it.runRecord?.taskId
      const sp:any = it.stableProgress || {}
      if (tid && sp && typeof sp === 'object'){
        const prev = lastStableByTask.value[tid]?.sp || {}
        const merged:any = { ...prev, ...sp }
        // 非递减合并
        merged.bytes = Math.max(Number(prev.bytes||0), Number(sp.bytes||0))
        merged.totalBytes = Number(sp.totalBytes||prev.totalBytes||0)
        merged.percentage = Math.max(Number(prev.percentage||0), Number(sp.percentage||0))
        merged.completedFiles = Math.max(Number(prev.completedFiles||0), Number(sp.completedFiles||0))
        // totalCount：若本帧为 0，沿用上一帧
        merged.totalCount = Number(sp.totalCount||prev.totalCount||0)
        // 速度：若本帧为 0，则沿用上一帧的非零速度，减少闪烁
        merged.speed = Number(sp.speed||prev.speed||0)
        lastStableByTask.value[tid] = { sp: merged, at: now }
        // 同步回 it 以便本次渲染即用稳定值
        it.stableProgress = merged
      }
    }
    // 如果本帧没有 active，立即清空，不要维持旧进度
    const list:any[] = data || []
    if (list.length === 0) {
      // 直接清空，让进度条立即消失
      activeRuns.value = []
      return
    }
    activeRuns.value = list
  } catch (e) {
    console.error(e)
  }
}

function getActiveRunByTaskId(taskId: number) {
  // 只返回活跃项，不返回最后稳态快照
  const cur = activeRuns.value.find(item => item.runRecord?.taskId === taskId)
  if (cur) return cur
  return undefined as any
}

function getDbProgressStable(run:any){
  const db = getLiveSummaryFromDB(run)
  const id = run?.id
  if (db && id){ lastDbFrameByRunId[id] = db; return db }
  if (id && lastDbFrameByRunId[id]) return lastDbFrameByRunId[id]
  return db || null
}

// 历史"运行中"卡片也使用任务卡片的抗噪稳态：优先 lastStableByTask，再回退 DB
function getDeNoisedStableByRun(run:any){
  try{
    const tid = run?.taskId as number
    if (tid && lastStableByTask.value && lastStableByTask.value[tid] && lastStableByTask.value[tid].sp){
      return lastStableByTask.value[tid].sp
    }
  }catch{}
  return getDbProgressStable(run)
}

// 任务卡片的预估完成时间：使用卡片的抗噪稳态（sp.totalBytes/sp.bytes/sp.speed）
function calcEtaForTaskCard(taskId:number){
  try{
    const st = lastStableByTask.value?.[taskId]?.sp || getActiveRunByTaskId(taskId)?.stableProgress
    if (!st) return null
    const total = Number(st.totalBytes||0)
    const bytes = Number(st.bytes||0)
    if (!total || bytes<=0) return null
    let speed = Number(st.speed||0)
    // 仅当可渲染时采信，否则使用最近一次有效速度
    if (formatBytesPerSec(speed) === '-'){
      const last = lastNonZeroSpeedByTask[taskId] || 0
      if (!last || last<=0) return null
      speed = last
    } else {
      lastNonZeroSpeedByTask[taskId] = speed
    }
    const remaining = Math.max(0, total - bytes)
    const eta = Math.floor(remaining / speed)
    if (eta > 99*3600) return null
    return eta
  }catch{ return null }
}

// 抗噪稳态读取（任务卡片用）：优先 lastStableByTask，再回退当前 active 的 stableProgress
function getDeNoisedStableByTask(taskId:number){
  // 只返回活跃任务的进度，不返回已完成的数据
  const active = getActiveRunByTaskId(taskId)
  if (!active) return null
  const raw = active.stableProgress
  if (!raw) return null
  // clone & normalize for UI
  const st:any = { ...raw }
  const totalBytes = Number(st.totalBytes || 0)
  const bytes = Number(st.bytes || 0)
  const totalCount = Number(st.totalCount || 0)
  const completedFiles = Number(st.completedFiles || 0)
  let pct = Number(st.percentage || 0)
  // clamp bytes within [0,totalBytes]
  if (totalBytes > 0) st.bytes = Math.min(bytes, totalBytes)
  // when near 100% or files已达总数，钳制为完成，避免"卡最后1个"错觉
  if ((totalCount > 0 && completedFiles >= totalCount) || pct >= 99.999) {
    st.completedFiles = totalCount > 0 ? totalCount : completedFiles
    st.percentage = 100
  } else {
    // also cap percentage at 100
    st.percentage = Math.min(100, pct)
  }
  return st
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
      // 补充实时"已传输数量"：优先 progress.completedFiles，其次 stableProgress.completedFiles
      const completedFiles = Number(p.completedFiles ?? sum?.stableProgress?.completedFiles ?? 0)
      return { bytes, totalBytes, speed, percentage, completedFiles }
    }
    // 退路：直接从 stableProgress 取（DB-only 流）
    const sp = sum?.stableProgress
    if (sp && typeof sp === 'object'){
      const bytes = Number(sp.bytes || 0)
      const totalBytes = Number(sp.totalBytes || 0)
      const speed = Number(sp.speed || 0)
      let percentage = Number(sp.percentage || 0)
      if ((!percentage || Number.isNaN(percentage)) && totalBytes>0) percentage = (bytes/totalBytes)*100
      const completedFiles = Number(sp.completedFiles || 0)
      return { bytes, totalBytes, speed, percentage, completedFiles }
    }
  }catch{}
  return null
}
// 以"开始时间 + 平均速度"计算更稳健的 ETA（剩余时间，秒）
// 预估完成：使用"固定总量（preflight.totalBytes）- 抗噪后的已传 bytes" / 抗噪后的速度
const lastNonZeroSpeedByTask: Record<number, number> = {}
function calcEtaFromAvg(run:any, live:any){
  try{
    if (!run?.startedAt || !live) return null
    const tid = (run.taskId || run.taskID || run.task_id || run.runRecord?.taskId) as number
    // 固定总量：优先 preflight.totalBytes；无则用卡片稳态的 totalBytes；再无则返回 null
    const pf = getPreflight(run)
    let total = Number(pf?.totalBytes || 0)
    if (!total && tid && lastStableByTask.value?.[tid]?.sp){ total = Number(lastStableByTask.value[tid].sp.totalBytes || 0) }
    if (!total) return null
    // 抗噪后的已传 bytes
    let bytes = Number(live.bytes || 0)
    if (tid && lastStableByTask.value?.[tid]?.sp){ bytes = Number(lastStableByTask.value[tid].sp.bytes || bytes) }
    if (bytes<=0) return null
    const remaining = Math.max(0, total - bytes)
    // 抗噪后的速度（任务卡片显示），可渲染时才采信
    let speed = 0
    if (tid && lastStableByTask.value?.[tid]?.sp){ speed = Number(lastStableByTask.value[tid].sp.speed || 0) }
    if (speed<=0) speed = Number(live.speed || 0)
    if (formatBytesPerSec(speed) === '-') return null
    if (tid && speed>0) lastNonZeroSpeedByTask[tid] = speed
    const sp = tid ? (lastNonZeroSpeedByTask[tid] || 0) : speed
    if (!sp || sp<=0) return null
    const etaSec = Math.floor(remaining / sp)
    if (etaSec > 99*3600) return null
    return etaSec
  }catch{ return null }
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

// 当某任务的稳定进度达 100% 附近时，触发一次"延迟刷新"，拉取最终状态
let refreshLocks: Record<number, boolean> = {}
async function triggerAutoRefresh(taskId: number){
  if (refreshLocks[taskId]) return ''
  refreshLocks[taskId] = true
  try{
    // 达到 100% 后等待 20 秒再刷新，给用户留出"完成"可视停留时间
    await new Promise(r=>setTimeout(r, 20000))
    await Promise.all([loadActiveRuns(), loadData()])
    // 兜底：再延迟 1s 后再拉一次，避免偶发落库延迟卡一帧
    await new Promise(r=>setTimeout(r, 1000))
    await Promise.all([loadActiveRuns(), loadData()])
    // 再兜底：若依然停留，移除该任务的 lastStable，以强制下一帧不再使用占位
    const st = lastStableByTask.value?.[taskId]
    if (st && Date.now() - st.at > LINGER_MS) {
      delete lastStableByTask.value[taskId]
    }
  } finally {
    setTimeout(()=>{ delete refreshLocks[taskId] }, 5000)
  }
  return ''
}

async function createTask() {
  // 如果已完成，点击返回任务列表
  if (creatingState.value === 'done') {
    creatingState.value = 'idle'
    currentModule.value = 'tasks'
    return
  }

  if (!createForm.value.name) {
    showToast('请输入任务名称', 'error')
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
      showToast('命令解析失败：' + (e as Error).message, 'error')
      return
    }
  }

  if (!createForm.value.sourceRemote || !createForm.value.targetRemote) {
    showToast('请选择源和目标存储', 'error')
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
      await taskApi.update(editingTask.value.id, taskData)
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
          await scheduleApi.update(oldSchedule.id, true, spec)
        } else {
          await scheduleApi.create({ taskId: editingTask.value.id, spec, enabled: true })
        }
      } else if (oldSchedule) {
        // 关闭并保留记录
        await scheduleApi.update(oldSchedule.id, false)
      }
      editingTask.value = null
    } else {
      // 新建任务
      const task = await taskApi.create(taskData)
      // 如果启用了定时任务，创建定时规则
      if (createForm.value.enableSchedule) {
        const spec = [
          createForm.value.scheduleMinute || '00',
          createForm.value.scheduleHour || '*',
          createForm.value.scheduleDay || '*',
          createForm.value.scheduleMonth || '*',
          createForm.value.scheduleWeek || '*',
        ].join('|')
        await scheduleApi.create({ taskId: task.id, spec, enabled: true })
      }
    }
    await loadData()
    currentModule.value = 'add'
    creatingState.value = 'done'
  } catch (e) {
    creatingState.value = 'idle'
    showToast((e as Error).message, 'error')
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
  // 全局视图用 runs.value，任务视图用 taskRuns.value
  let source = historyFilterTaskId.value === null ? runs.value : taskRuns.value
  let result = [...source]
  if (historyStatusFilter.value === 'hasTransfer') {
    result = result.filter(r => {
      const sum = getFinalSummary(r)
      return sum && (sum.totalCount > 1 || sum.transferredBytes > 0)
    })
  } else if (historyStatusFilter.value !== 'all') {
    result = result.filter(r => r.status === historyStatusFilter.value)
  }
  // 客户端分页
  const start = (runsPage.value - 1) * runsPageSize
  const end = start + runsPageSize
  return result.slice(start, end)
})

// 筛选后的总数（用于分页）
const filteredRunsTotal = computed(() => {
  let source = historyFilterTaskId.value === null ? runs.value : taskRuns.value
  let result = [...source]
  if (historyStatusFilter.value === 'hasTransfer') {
    result = result.filter(r => {
      const sum = getFinalSummary(r)
      return sum && (sum.totalCount > 1 || sum.transferredBytes > 0)
    })
  } else if (historyStatusFilter.value !== 'all') {
    result = result.filter(r => r.status === historyStatusFilter.value)
  }
  return result.length
})

// 当前分页的总数（全局视图用 runsTotal，特定任务视图用 filteredRunsTotal）
const currentTotal = computed(() => {
  return historyFilterTaskId.value === null ? (runsTotal.value || 0) : filteredRunsTotal.value
})

// 当前总页数
const currentTotalPages = computed(() => Math.max(1, Math.ceil(currentTotal.value / runsPageSize)))

function viewTaskHistory(taskId: number) {
  runsPage.value = 1
  historyFilterTaskId.value = taskId
  currentModule.value = 'history'
  // 获取该任务的历史记录
  runApi.getRunsByTask(taskId).then(data => {
    if (data && Array.isArray(data)) {
      taskRuns.value = data
    }
  })
}

const stoppedTaskId = ref<number|null>(null)
async function stopTaskAny(taskId: number) {
  try {
    // 直接按任务 ID kill（后端会定位最近 run 并发信号）
    await taskApi.kill(taskId)
    // 兼容 RC：如仍有 rcJobId，则再尝试停止
    const active = await jobApi.list()
    const cur:any = Array.isArray(active) ? active.find(x => x?.runRecord?.taskId === taskId && x?.runRecord?.rcJobId) : null
    if (cur?.runRecord?.rcJobId) {
      await jobApi.stop(cur.runRecord.rcJobId)
    }
    // 按钮状态反馈：红色"已经停止"，10秒后恢复
    stoppedTaskId.value = taskId
    setTimeout(()=>{ if (stoppedTaskId.value===taskId) stoppedTaskId.value=null }, 10000)
    // 轻量刷新列表
    await loadData()
  } catch (e) {
    stoppedTaskId.value = null
    showToast((e as Error).message, 'error')
  }
}


async function runTask(taskId: number) {
  if (runningTaskId.value !== null) {
    showToast('单例模式：已有任务正在运行，跳过本次执行', 'error')
    return
  }
  runningTaskId.value = taskId
  const result = await taskApi.run(taskId)
  if (!result) {
    // handleError already showed a toast, just reset state
    runningTaskId.value = null
    return
  }
  // 5秒后恢复
  setTimeout(() => {
    if (runningTaskId.value === taskId) {
      runningTaskId.value = null
    }
  }, 5000)
  return result
}

async function goToAddTask() {
  // Reload remotes before switching to add mode
  const remoteData = await remoteApi.list()
  remotes.value = remoteData?.remotes || []
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

  // 加载源路径选项 - 直接加载配置的路径
  if (task.sourceRemote) {
    const sourcePath = task.sourcePath || ''
    sourceCurrentPath.value = sourcePath
    loadSourcePath(task.sourceRemote, sourcePath)
  }
  // 加载目标路径选项 - 直接加载配置的路径
  if (task.targetRemote) {
    const targetPath = task.targetPath || ''
    targetCurrentPath.value = targetPath
    loadTargetPath(task.targetRemote, targetPath)
  }
  currentModule.value = 'add'
  openMenuId.value = null
}

function showConfirm(title: string, message: string, onConfirm: () => void) {
  confirmModal.value = { show: true, title, message, onConfirm }
}

async function deleteTask(id: number) {
  showConfirm('删除任务', '确定删除此任务？此操作不可恢复！', async () => {
    const success = await taskApi.delete(id)
    if (success) {
      openMenuId.value = null
      await loadData()
    }
  })
}

function getScheduleByTaskId(taskId: number) {
  return schedules.value.find(s => s.taskId === taskId)
}

async function toggleSchedule(taskId: number) {
  const schedule = getScheduleByTaskId(taskId)
  if (!schedule) return
  await scheduleApi.update(schedule.id, !schedule.enabled)
  await loadData()
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
  await scheduleApi.delete(id)
  await loadData()
}

async function clearRun(id: number) {
  await runApi.delete(id)
  await loadData()
}

async function clearAllRuns() {
  if (historyFilterTaskId.value === null) {
    showToast('请先选择任务', 'error')
    return
  }
  showConfirm('删除所有历史', '确定删除该任务所有历史记录？此操作不可恢复！', async () => {
    await runApi.deleteByTask(historyFilterTaskId.value)
    await loadData()
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

// 源路径面包屑
const sourceBreadcrumbs = computed(() => {
  if (!createForm.value.sourceRemote) return []
  const parts = (sourceCurrentPath.value || '').split('/').filter(Boolean)
  const crumbs = [{ name: createForm.value.sourceRemote + ':', path: '' }]
  let current = ''
  for (const p of parts) {
    current += '/' + p
    crumbs.push({ name: p, path: current })
  }
  return crumbs
})

// 目标路径面包屑
const targetBreadcrumbs = computed(() => {
  if (!createForm.value.targetRemote) return []
  const parts = (targetCurrentPath.value || '').split('/').filter(Boolean)
  const crumbs = [{ name: createForm.value.targetRemote + ':', path: '' }]
  let current = ''
  for (const p of parts) {
    current += '/' + p
    crumbs.push({ name: p, path: current })
  }
  return crumbs
})
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

  <div v-if="currentModule === 'history'" class="card">
    <div class="card-header">
      <div class="title">任务历史记录</div>
      <div class="history-filters">
        <button :class="['filter-btn', historyStatusFilter==='all' && 'active']" @click="historyStatusFilter='all'">全部</button>
        <button :class="['filter-btn', historyStatusFilter==='finished' && 'active']" @click="historyStatusFilter='finished'">成功</button>
        <button :class="['filter-btn', historyStatusFilter==='failed' && 'active']" @click="historyStatusFilter='failed'">失败</button>
        <button :class="['filter-btn', historyStatusFilter==='skipped' && 'active']" @click="historyStatusFilter='skipped'">跳过</button>
        <button :class="['filter-btn', historyStatusFilter==='hasTransfer' && 'active']" @click="historyStatusFilter='hasTransfer'">有传输</button>
      </div>
      <!-- 历史记录分页 -->
      <div class="pagination" v-if="currentTotal > runsPageSize">
        <span class="page-current">第 {{ runsPage }} / {{ currentTotalPages }} 页</span>
        <button class="page-btn" :disabled="runsPage <= 1" @click="runsPage--; loadData()">上一页</button>
        <button class="page-btn" :disabled="runsPage >= currentTotalPages" @click="runsPage++; loadData()">下一页</button>
        <input type="number" class="page-input" v-model.number="jumpPage" :min="1" :max="currentTotalPages" @keyup.enter="jumpToPage" />
        <button class="page-btn" @click="jumpToPage">跳转</button>
      </div>
      <div class="header-actions">
        <button v-if="historyFilterTaskId !== null && filteredRuns.length > 0" class="ghost small danger-text" @click="clearAllRuns">删除所有</button>
        <button v-if="historyFilterTaskId !== null" class="ghost small" @click="currentModule = 'tasks'">
          ← 返回
        </button>
      </div>
    </div>
    <div class="list">
      <RunItem
        v-for="run in filteredRuns"
        :key="run.id"
        :run="run"
        :progress="run.status === 'running' ? getDbProgressStable(run) : undefined"
        :summary="getFinalSummary(run)"
        @click="showRunDetail(run)"
        @view-detail="showRunDetail(run)"
        @view-log="openRunLog(run)"
        @clear="clearRun(run.id)"
      />
      <div v-if="!filteredRuns.length" class="empty">暂无历史记录</div>
    </div>

    <!-- 运行详情弹窗 -->
    <div v-if="showDetailModal" class="modal-overlay" @click.self="showDetailModal = false">
      <div class="modal-content detail-modal">
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
            <span>{{ runDetail.trigger === 'schedule' ? '定时任务' : (runDetail.trigger === 'webhook' ? 'Webhook 触发' : '手动执行') }}</span>
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
                <!-- 运行中进度/运行中速度两项已移除：此处只展示完成态汇总，不含实时项 -->

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
                    <FileItem v-for="it in pagedFinalFiles" :key="(it.path||it.name) + (it.at||'') + (it.status||'')" :item="it" />
                  </template>
                  <template v-else>
                    <FileItem v-for="it in pagedRunFiles" :key="it.name + it.at + it.status" :item="it" />
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
        <p>实时日志与进度请点击"传输日志"或查看任务卡片上的实时进度。</p>
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
        <p v-if="commandMode" class="hint">保存时将自动解析命令，填充"模式/源/目标/选项"。任务名称仍需手动填写。</p>
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
            <div class="pathbar">
              <template v-for="(crumb, i) in sourceBreadcrumbs" :key="crumb.path">
                <span v-if="i > 0" class="sep">/</span>
                <button
                  class="crumb"
                  :class="{ current: i === sourceBreadcrumbs.length - 1 }"
                  @click="crumb.path !== sourceCurrentPath && loadSourcePath(createForm.sourceRemote, crumb.path)"
                >
                  {{ crumb.name }}
                </button>
              </template>
            </div>
            <div class="path-list">
              <PathItem 
                v-for="item in sourcePathOptions" 
                :key="item.Path" 
                :item="item"
                @enter="onSourceArrow(item)"
                @click="onSourceClick(item)"
              />
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
            <div class="pathbar">
              <template v-for="(crumb, i) in targetBreadcrumbs" :key="crumb.path">
                <span v-if="i > 0" class="sep">/</span>
                <button
                  class="crumb"
                  :class="{ current: i === targetBreadcrumbs.length - 1 }"
                  @click="crumb.path !== targetCurrentPath && loadTargetPath(createForm.targetRemote, crumb.path)"
                >
                  {{ crumb.name }}
                </button>
              </template>
            </div>
            <div class="path-list">
              <PathItem 
                v-for="item in targetPathOptions" 
                :key="item.Path" 
                :item="item"
                @enter="onTargetArrow(item)"
                @click="onTargetClick(item)"
              />
              <div v-if="!targetPathOptions.length" class="path-empty">空目录</div>
            </div>
          </div>
          <button type="button" class="ghost small" @click="showTargetPathInput = !showTargetPathInput">手动输入</button>
        </div>
        <input v-if="showTargetPathInput" v-model="createForm.targetPath" type="text" placeholder="手动输入路径" style="margin-top: 8px" />
      </div>

      <!-- 定时任务设置 -->
      <ScheduleOptions v-model="createForm" />

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
</style>

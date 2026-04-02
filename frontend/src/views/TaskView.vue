<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
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
const showCreateModal = ref(false)

const createForm = ref({
  name: '',
  mode: 'copy',
  sourceRemote: '',
  sourcePath: '',
  targetRemote: '',
  targetPath: '',
  enableSchedule: false,
  scheduleYear: '*',     // * = 每年, 或 "2026,2027"
  scheduleMonth: '*',     // * = 每月, 或 "1,3,5"
  scheduleWeek: '',      // 空 = 不设置, 或 "1,3,5" (周一三五)
  scheduleDay: '',       // 空 = 不设置, 或 "1,15,28"
  scheduleHour: '00',   // "00,12,18"
  scheduleMinute: '00', // "00,30,59"
})

const openMenuId = ref<number | null>(null)
const editingTask = ref<Task | null>(null)

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
  year: [] as string[],
  month: [] as string[],
  week: [] as string[],
  day: [] as string[],
  hour: [] as string[],
  minute: [] as string[],
})

// 确认选择
function confirmField(field: 'year' | 'month' | 'week' | 'day' | 'hour' | 'minute') {
  const val = tempSchedule.value[field]
  if (val.length === 0) {
    // 空表示不设置(周/日)或每年/每月(年/月)
    createForm.value['schedule' + field.charAt(0).toUpperCase() + field.slice(1)] = field === 'year' || field === 'month' ? '*' : ''
  } else {
    createForm.value['schedule' + field.charAt(0).toUpperCase() + field.slice(1)] = val.join(',')
  }
}

onMounted(async () => {
  await loadData()
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

function getStatusClass(status: string) {
  switch (status) {
    case 'running': return 'running'
    case 'finished': return 'success'
    case 'failed': return 'failed'
    default: return ''
  }
}

async function createTask() {
  if (!createForm.value.name) {
    alert('请输入任务名称')
    return
  }
  if (!createForm.value.sourceRemote || !createForm.value.targetRemote) {
    alert('请选择源和目标存储')
    return
  }
  try {
    if (editingTask.value) {
      // 更新任务
      await api.updateTask(editingTask.value.id, {
        name: createForm.value.name,
        mode: createForm.value.mode,
        sourceRemote: createForm.value.sourceRemote,
        sourcePath: createForm.value.sourcePath,
        targetRemote: createForm.value.targetRemote,
        targetPath: createForm.value.targetPath,
      })
      editingTask.value = null
    } else {
      // 新建任务
      const task = await api.createTask({
        name: createForm.value.name,
        mode: createForm.value.mode,
        sourceRemote: createForm.value.sourceRemote,
        sourcePath: createForm.value.sourcePath,
        targetRemote: createForm.value.targetRemote,
        targetPath: createForm.value.targetPath,
      })
      // 如果启用了定时任务，创建定时规则
      if (createForm.value.enableSchedule) {
        const spec = [
          createForm.value.scheduleYear || '*',
          createForm.value.scheduleMonth || '*',
          createForm.value.scheduleWeek || '',
          createForm.value.scheduleDay || '',
          createForm.value.scheduleHour || '00',
          createForm.value.scheduleMinute || '00',
        ].join(',')
        await api.createSchedule({ taskId: task.id, spec, enabled: true })
      }
    }
    createForm.value = { 
      name: '', mode: 'copy', sourceRemote: '', sourcePath: '', targetRemote: '', targetPath: '',
      enableSchedule: false,
      scheduleYear: '*', scheduleMonth: '*', scheduleWeek: '', scheduleDay: '', scheduleHour: '00', scheduleMinute: '00'
    }
    sourcePathOptions.value = []
    targetPathOptions.value = []
    sourceCurrentPath.value = ''
    targetCurrentPath.value = ''
    tempSchedule.value = { year: [], month: [], week: [], day: [], hour: [], minute: [] }
    await loadData()
  } catch (e) {
    alert((e as Error).message)
  }
}

async function runTask(taskId: number) {
  try {
    await api.runTask(taskId)
    alert('任务已启动')
    await loadData()
  } catch (e) {
    alert((e as Error).message)
  }
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
      scheduleYear: parts[0] || '*',
      scheduleMonth: parts[1] || '*',
      scheduleWeek: parts[2] || '',
      scheduleDay: parts[3] || '',
      scheduleHour: parts[4] || '00',
      scheduleMinute: parts[5] || '00',
    }
    // 更新临时选择状态
    tempSchedule.value = {
      year: parts[0] && parts[0] !== '*' ? parts[0].split(',') : [],
      month: parts[1] && parts[1] !== '*' ? parts[1].split(',') : [],
      week: parts[2] ? parts[2].split(',') : [],
      day: parts[3] ? parts[3].split(',') : [],
      hour: parts[4] && parts[4] !== '*' ? parts[4].split(',') : [],
      minute: parts[5] && parts[5] !== '*' ? parts[5].split(',') : [],
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
      scheduleYear: '*',
      scheduleMonth: '*',
      scheduleWeek: '',
      scheduleDay: '',
      scheduleHour: '00',
      scheduleMinute: '00',
    }
    tempSchedule.value = { year: [], month: [], week: [], day: [], hour: [], minute: [] }
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

async function deleteTask(id: number) {
  if (!confirm('确定删除此任务？')) return
  try {
    await api.deleteTask(id)
    openMenuId.value = null
    await loadData()
  } catch (e) {
    alert((e as Error).message)
  }
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
  const parts = spec.split(',')
  if (parts.length !== 6) return spec
  const [year, month, week, day, hour, minute] = parts
  const yearStr = year === '*' ? '每年' : year
  const monthStr = month === '*' ? '每月' : month + '月'
  const weekNames = ['周日', '周一', '周二', '周三', '周四', '周五', '周六']
  let weekStr = ''
  if (week === '*') {
    weekStr = '每天'
  } else if (week === '') {
    weekStr = ''
  } else {
    const days = week.split(',').map(d => weekNames[parseInt(d)] || d)
    weekStr = days.join(',')
  }
  const dayStr = day === '*' ? '每天' : (day ? day + '日' : '')
  const hourStr = hour === '*' ? '' : hour + '时'
  const minuteStr = minute === '*' ? '每分' : (minute ? minute + '分' : '')
  
  // 组合显示
  const parts2: string[] = []
  if (year !== '*') parts2.push(yearStr)
  if (month !== '*') parts2.push(monthStr)
  if (weekStr) parts2.push(weekStr)
  if (dayStr && day !== '*') parts2.push(dayStr)
  if (hourStr) parts2.push(hourStr)
  if (minuteStr && minute !== '*') parts2.push(minuteStr)
  
  return parts2.length > 0 ? parts2.join(' ') : '未设置'
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

async function clearRun(id: number) {
  if (!confirm('确定清除此记录？')) return
  try {
    await api.clearRun(id)
    await loadData()
  } catch (e) {
    alert((e as Error).message)
  }
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
  <div class="card">
    <div class="card-header">
      <div class="title">功能模块</div>
    </div>
    <div class="module-tabs">
      <button class="tab-btn" :class="{ active: currentModule === 'tasks' }" @click="currentModule = 'tasks'">
        任务列表
      </button>
      <button class="tab-btn" :class="{ active: currentModule === 'history' }" @click="currentModule = 'history'">
        历史记录
      </button>
      <button class="tab-btn" :class="{ active: currentModule === 'add' }" @click="currentModule = 'add'">
        添加任务
      </button>
    </div>
  </div>

  <div v-if="currentModule === 'tasks'" class="card">
    <div class="card-header">
      <div class="title">任务列表</div>
      <div class="header-actions">
        <input v-model="taskSearch" type="text" placeholder="搜索任务..." class="search-input" />
        <button class="primary small" @click="currentModule = 'add'">+ 添加任务</button>
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
            <button v-if="getScheduleByTaskId(task.id)" class="ghost small" @click.stop="toggleSchedule(task.id)">
              {{ getScheduleByTaskId(task.id)?.enabled ? '⏸ 关闭' : '▶ 开启' }}
            </button>
            <button class="ghost small" @click.stop="runTask(task.id)">▶ 执行</button>
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
        </div>
      </div>
      <div v-if="!filteredTasks.length" class="empty">暂无任务</div>
    </div>
  </div>

  <div v-if="currentModule === 'history'" class="card">
    <div class="card-header"><div class="title">历史记录</div></div>
    <div class="list-header">
      <span class="col-name">任务</span>
      <span class="col-status">状态</span>
      <span class="col-time">开始时间</span>
      <span class="col-time">结束时间</span>
      <span class="col-action">操作</span>
    </div>
    <div class="list">
      <div v-for="run in runs" :key="run.id" class="item">
        <div class="name"><strong>{{ tasks.find(t => t.id === run.taskId)?.name || `任务 #${run.taskId}` }}</strong></div>
        <span :class="['status', getStatusClass(run.status)]">{{ run.status }}</span>
        <span class="time">{{ formatTime(run.startedAt) }}</span>
        <span class="time">{{ formatTime(run.finishedAt) }}</span>
        <button class="ghost small" @click="clearRun(run.id)">清除</button>
      </div>
      <div v-if="!runs.length" class="empty">暂无历史记录</div>
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
        <label class="schedule-toggle">
          <input type="checkbox" v-model="createForm.enableSchedule" />
          <span>启用定时任务</span>
        </label>
        <div v-if="createForm.enableSchedule" class="schedule-grid">
          <!-- 年 -->
          <div class="schedule-item">
            <label>年</label>
            <select v-model="tempSchedule.year" multiple size="6" @dblclick="confirmField('year')">
              <option value="*">每年</option>
              <option v-for="y in [2026,2027,2028,2029,2030]" :key="y" :value="String(y)">{{ y }}</option>
            </select>
            <button type="button" class="ghost small" @click="confirmField('year')">确定</button>
            <span class="selected-val">{{ createForm.scheduleYear === '*' ? '每年' : (createForm.scheduleYear || '每年') }}</span>
          </div>
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

      <div class="form-actions">
        <button class="primary" @click="createTask">创建任务</button>
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
</style>
// DEBUG LINE

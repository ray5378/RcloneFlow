<script setup lang="ts">
import { ref, onMounted } from 'vue'
import * as api from '../api'
import type { Task, Schedule, Run } from '../types'

const tasks = ref<Task[]>([])
const schedules = ref<Schedule[]>([])
const runs = ref<Run[]>([])
const remotes = ref<string[]>([])
const currentModule = ref<'history' | 'schedule' | 'add' | 'tasks'>('tasks')
const showCreateModal = ref(false)

const createForm = ref({
  name: '',
  mode: 'copy',
  sourceRemote: '',
  sourcePath: '',
  targetRemote: '',
  targetPath: '',
})

const openMenuId = ref<number | null>(null)
const editingTask = ref<Task | null>(null)

const sourcePathOptions = ref<any[]>([])
const targetPathOptions = ref<any[]>([])
const showSourcePathInput = ref(false)
const showTargetPathInput = ref(false)
const sourceCurrentPath = ref('')
const targetCurrentPath = ref('')

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
      await api.createTask({
        name: createForm.value.name,
        mode: createForm.value.mode,
        sourceRemote: createForm.value.sourceRemote,
        sourcePath: createForm.value.sourcePath,
        targetRemote: createForm.value.targetRemote,
        targetPath: createForm.value.targetPath,
      })
    }
    createForm.value = { name: '', mode: 'copy', sourceRemote: '', sourcePath: '', targetRemote: '', targetPath: '' }
    sourcePathOptions.value = []
    targetPathOptions.value = []
    sourceCurrentPath.value = ''
    targetCurrentPath.value = ''
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
  createForm.value = {
    name: task.name,
    mode: task.mode,
    sourceRemote: task.sourceRemote,
    sourcePath: task.sourcePath || '',
    targetRemote: task.targetRemote,
    targetPath: task.targetPath || '',
  }
  // 加载源路径选项（不调用onSourceRemoteChange，因为它会重置path）
  if (task.sourceRemote) {
    sourceCurrentPath.value = task.sourcePath || ''
    loadSourcePath(task.sourceRemote, task.sourcePath || '')
  }
  // 加载目标路径选项
  if (task.targetRemote) {
    targetCurrentPath.value = task.targetPath || ''
    loadTargetPath(task.targetRemote, task.targetPath || '')
  }
  currentModule.value = 'add'
  openMenuId.value = null
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

function openTargetDir(item: any) {
  if (item.IsDir) {
    loadTargetPath(createForm.value.targetRemote, item.Path)
  }
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
      <button class="tab-btn" :class="{ active: currentModule === 'schedule' }" @click="currentModule = 'schedule'">
        定时任务
      </button>
      <button class="tab-btn" :class="{ active: currentModule === 'add' }" @click="currentModule = 'add'">
        添加任务
      </button>
    </div>
  </div>

  <div v-if="currentModule === 'tasks'" class="card">
    <div class="card-header"><div class="title">任务列表</div></div>
    <div class="tile-grid">
      <div v-for="task in tasks" :key="task.id" class="tile" @click="closeMenus">
        <div class="tile-info">
          <div class="tile-header"><span class="tile-name">{{ task.name }}</span></div>
          <div class="tile-desc">{{ task.mode }} {{ task.sourceRemote }}:{{ task.sourcePath || '根目录' }} → {{ task.targetRemote }}:{{ task.targetPath || '根目录' }}</div>
        </div>
        <div class="tile-actions">
          <button class="ghost small" @click.stop="runTask(task.id)">▶ 执行</button>
          <button class="ghost small" @click.stop="editTask(task)">✏️ 修改</button>
          <button class="ghost small menu-btn" @click.stop="toggleMenu(task.id)">⋮</button>
          <div v-if="openMenuId === task.id" class="tile-menu">
            <button class="danger-text" @click.stop="deleteTask(task.id)">🗑️ 删除</button>
          </div>
        </div>
      </div>
      <div v-if="!tasks.length" style="padding: 20px; color: #888; text-align: center; width: 100%">暂无任务</div>
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

  <div v-if="currentModule === 'schedule'" class="card">
    <div class="card-header"><div class="title">定时任务</div></div>
    <div class="list-header">
      <span class="col-name">任务</span>
      <span class="col-info">周期</span>
      <span class="col-action">操作</span>
    </div>
    <div class="list">
      <div v-for="s in schedules" :key="s.id" class="item">
        <div class="name"><strong>{{ tasks.find(t => t.id === s.taskId)?.name || `任务 #${s.taskId}` }}</strong></div>
        <span class="info">{{ s.spec }}</span>
        <div class="item-actions">
          <button class="ghost small" @click="runTask(s.taskId)">运行</button>
          <button class="ghost small danger-text" @click="deleteSchedule(s.id)">删除</button>
        </div>
      </div>
      <div v-if="!schedules.length" class="empty">暂无定时任务</div>
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
              <span class="path-label">当前: /{{ sourceCurrentPath || '根目录' }}</span>
              <button v-if="sourceCurrentPath" type="button" class="ghost small" @click="goBackSource">返回</button>
            </div>
            <div class="path-list">
              <div v-for="item in sourcePathOptions" :key="item.Path" class="path-item" :class="{ 'is-dir': item.IsDir }" @click="createForm.sourcePath = item.Path, showSourcePathInput = false" @dblclick="item.IsDir && openSourceDir(item)">
                <span v-if="item.IsDir" class="dir-arrow" @click.stop="openSourceDir(item)">▶</span>
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
              <span class="path-label">当前: /{{ targetCurrentPath || '根目录' }}</span>
              <button v-if="targetCurrentPath" type="button" class="ghost small" @click="goBackTarget">返回</button>
            </div>
            <div class="path-list">
              <div v-for="item in targetPathOptions" :key="item.Path" class="path-item" :class="{ 'is-dir': item.IsDir }" @click="createForm.targetPath = item.Path, showTargetPathInput = false" @dblclick="item.IsDir && openTargetDir(item)">
                <span v-if="item.IsDir" class="dir-arrow" @click.stop="openTargetDir(item)">▶</span>
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
  padding: 0 20px 16px;
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
.dir-arrow { color: #64b5f6; font-size: 10px; cursor: pointer; padding: 2px 4px; background: #333; border-radius: 3px; margin-right: 4px; }
.dir-arrow:hover { color: #90caf9; background: #444; }
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

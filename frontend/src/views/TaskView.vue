<script setup lang="ts">
import { ref, onMounted } from 'vue'
import * as api from '../api'
import type { Task, Schedule, Run } from '../types'

const tasks = ref<Task[]>([])
const schedules = ref<Schedule[]>([])
const runs = ref<Run[]>([])
const remotes = ref<string[]>([])
const selectedTaskId = ref<number | null>(null)
const activeTab = ref<'schedule' | 'history'>('history')
const taskMenu = ref<string | number>('')
const showCreateModal = ref(false)

const createForm = ref({
  name: '',
  mode: 'copy',
  sourceRemote: '',
  sourcePath: '',
  targetRemote: '',
  targetPath: '',
})

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

function getTaskSchedules(taskId: number) {
  return schedules.value.filter(s => s.taskId === taskId)
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
    await api.createTask({
      name: createForm.value.name,
      mode: createForm.value.mode,
      sourceRemote: createForm.value.sourceRemote,
      sourcePath: createForm.value.sourcePath,
      targetRemote: createForm.value.targetRemote,
      targetPath: createForm.value.targetPath,
    })
    showCreateModal.value = false
    createForm.value = { name: '', mode: 'copy', sourceRemote: '', sourcePath: '', targetRemote: '', targetPath: '' }
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

async function deleteTask(taskId: number) {
  if (!confirm('确定删除此任务？')) return
  try {
    await api.deleteTask(taskId)
    if (selectedTaskId.value === taskId) {
      selectedTaskId.value = null
    }
    await loadData()
  } catch (e) {
    alert((e as Error).message)
  }
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
</script>

<template>
  <!-- Task Header Card -->
  <div class="card">
    <div class="card-header">
      <div style="display: flex; justify-content: space-between; align-items: center; width: 100%">
        <div>
          <div class="title">任务管理</div>
          <div class="subtitle">创建和管理任务</div>
        </div>
        <div class="actions">
          <button class="ghost small" @click="showCreateModal = true">添加任务</button>
        </div>
      </div>
    </div>
    <div class="tile-grid">
      <div
        v-for="task in tasks"
        :key="task.id"
        class="tile"
        :class="{ active: selectedTaskId === task.id }"
        @click="selectedTaskId = task.id"
      >
        <div class="tile-header">
          <span class="tile-name">{{ task.name }}</span>
        </div>
        <div class="tile-desc">
          {{ task.mode }}: {{ task.sourceRemote }} → {{ task.targetRemote }}
        </div>
      </div>
      <div v-if="!tasks.length" style="padding: 20px; color: #888; text-align: center; width: 100%">
        暂无任务
      </div>
    </div>
  </div>

  <!-- Task Detail & Toggle Buttons -->
  <div v-if="selectedTaskId" class="card">
    <div class="card-header">
      <div class="title">{{ tasks.find(t => t.id === selectedTaskId)?.name }}</div>
      <div class="actions" @click.stop>
        <button class="ghost small" @click="runTask(selectedTaskId!)">运行</button>
        <div class="menu-area">
          <button class="menu-btn" @click="taskMenu = taskMenu === selectedTaskId ? '' : selectedTaskId">
            ⋮
          </button>
          <div v-if="taskMenu === selectedTaskId" class="menu-pop">
            <button class="danger" @click="deleteTask(selectedTaskId!); taskMenu = ''">删除</button>
          </div>
        </div>
      </div>
    </div>
    <div class="list">
      <div class="item">
        <span class="label">模式</span>
        <span class="value">{{ tasks.find(t => t.id === selectedTaskId)?.mode }}</span>
      </div>
      <div class="item">
        <span class="label">源存储</span>
        <span class="value">{{ tasks.find(t => t.id === selectedTaskId)?.sourceRemote }}</span>
      </div>
      <div class="item">
        <span class="label">源路径</span>
        <span class="value">{{ tasks.find(t => t.id === selectedTaskId)?.sourcePath || '/' }}</span>
      </div>
      <div class="item">
        <span class="label">目标存储</span>
        <span class="value">{{ tasks.find(t => t.id === selectedTaskId)?.targetRemote }}</span>
      </div>
      <div class="item">
        <span class="label">目标路径</span>
        <span class="value">{{ tasks.find(t => t.id === selectedTaskId)?.targetPath || '/' }}</span>
      </div>
    </div>
    <!-- Toggle Buttons -->
    <div class="toggle-bar">
      <button
        class="toggle-btn"
        :class="{ active: activeTab === 'schedule' }"
        @click="activeTab = 'schedule'"
      >
        定时任务
      </button>
      <button
        class="toggle-btn"
        :class="{ active: activeTab === 'history' }"
        @click="activeTab = 'history'"
      >
        历史记录
      </button>
    </div>
  </div>

  <!-- Schedule Card (Toggle Content) -->
  <div v-if="selectedTaskId && activeTab === 'schedule'" class="card">
    <div class="card-header">
      <div class="title">定时任务</div>
      <div class="actions">
        <button class="ghost small" @click="createSchedule">添加定时</button>
      </div>
    </div>
    <div class="list">
      <div v-for="s in getTaskSchedules(selectedTaskId!)" :key="s.id" class="item">
        <div class="name">
          <strong>周期: {{ s.spec }}</strong>
        </div>
        <button class="ghost small danger-text" @click="deleteSchedule(s.id)">删除</button>
      </div>
      <div v-if="!getTaskSchedules(selectedTaskId!).length" class="empty">暂无定时任务</div>
    </div>
  </div>

  <!-- History Card (Toggle Content) -->
  <div v-if="selectedTaskId && activeTab === 'history'" class="card">
    <div class="card-header">
      <div class="title">历史记录</div>
    </div>
    <div class="list-header">
      <span class="col-status">状态</span>
      <span class="col-time">开始时间</span>
      <span class="col-time">结束时间</span>
      <span class="col-action">操作</span>
    </div>
    <div class="list">
      <div v-for="run in runs" :key="run.id" class="item">
        <span :class="['status', getStatusClass(run.status)]">{{ run.status }}</span>
        <span class="time">{{ formatTime(run.startedAt) }}</span>
        <span class="time">{{ formatTime(run.finishedAt) }}</span>
        <button class="ghost small" @click="clearRun(run.id)">清除</button>
      </div>
      <div v-if="!runs.length" class="empty">暂无历史记录</div>
    </div>
  </div>

  <!-- Create Task Modal -->
  <div v-if="showCreateModal" class="modal-overlay" @click.self="showCreateModal = false">
    <div class="modal">
      <div class="modal-header">
        <h2>添加任务</h2>
        <button class="modal-close" @click="showCreateModal = false">&times;</button>
      </div>
      <div class="modal-content">
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
          <select v-model="createForm.sourceRemote">
            <option value="">选择源存储</option>
            <option v-for="r in remotes" :key="r" :value="r">{{ r }}</option>
          </select>
        </div>
        <div class="field-item">
          <label>源路径</label>
          <input v-model="createForm.sourcePath" type="text" placeholder="/路径，留空表示根目录" />
        </div>
        <div class="field-item">
          <label>目标存储 <span style="color: #dc2626">*</span></label>
          <select v-model="createForm.targetRemote">
            <option value="">选择目标存储</option>
            <option v-for="r in remotes" :key="r" :value="r">{{ r }}</option>
          </select>
        </div>
        <div class="field-item">
          <label>目标路径</label>
          <input v-model="createForm.targetPath" type="text" placeholder="/路径，留空表示根目录" />
        </div>
      </div>
      <div class="modal-footer">
        <button class="ghost" @click="showCreateModal = false">取消</button>
        <button class="primary" @click="createTask">创建</button>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
async function createSchedule() {
  const spec = prompt('输入执行周期（如：5m, 10m, 1h, 6h, 12h, 24h）')
  if (!spec) return
  // This needs to be defined in the parent setup
}
</script>

<style scoped>
.tile-grid {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  padding: 16px 20px;
}

.tile {
  min-width: 180px;
  padding: 14px 16px;
  border-radius: 10px;
  background: #252525;
  cursor: pointer;
  transition: all 0.2s;
  flex: 1 1 calc(25% - 12px);
  max-width: calc(25% - 12px);
}

.tile:hover {
  background: #2a2a2a;
}

.tile.active {
  background: #1e3a5f;
  border: 1px solid #2563a0;
}

body.light .tile {
  background: #f5f5f5;
}

body.light .tile:hover {
  background: #e8e8e8;
}

body.light .tile.active {
  background: #e3f2fd;
  border-color: #bbdefb;
}

.tile-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 6px;
}

.tile-name {
  font-weight: 600;
  font-size: 14px;
  color: #fff;
}

body.light .tile-name {
  color: #1a1a1a;
}

.tile-desc {
  font-size: 12px;
  color: #888;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.list-header {
  display: flex;
  justify-content: space-between;
  padding: 10px 20px;
  background: #252525;
  font-size: 12px;
  color: #888;
  border-bottom: 1px solid #333;
}

body.light .list-header {
  background: #f5f5f5;
  color: #666;
  border-bottom: 1px solid #e0e0e0;
}

.col-status { width: 80px; text-align: center; }
.col-time { width: 150px; text-align: right; }
.col-action { width: 60px; text-align: right; }

.item {
  display: flex;
  align-items: center;
  padding: 10px 20px;
  border-bottom: 1px solid #252525;
  gap: 12px;
}

body.light .item {
  border-color: #f0f0f0;
}

.item .name {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.item .label {
  width: 80px;
  color: #888;
  font-size: 13px;
}

.item .value {
  flex: 1;
  color: #e0e0e0;
  font-size: 13px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

body.light .item .value {
  color: #333;
}

.status {
  width: 80px;
  padding: 4px 8px;
  border-radius: 6px;
  font-size: 12px;
  text-align: center;
}

.status.running { background: #1976d2; color: #fff; }
.status.success { background: #388e3c; color: #fff; }
.status.failed { background: #d32f2f; color: #fff; }

.time {
  width: 150px;
  text-align: right;
  color: #888;
  font-size: 13px;
}

.danger-text {
  color: #ef5350 !important;
}

.toggle-bar {
  display: flex;
  gap: 8px;
  padding: 12px 20px;
  background: #1a1a1a;
  border-top: 1px solid #333;
}

body.light .toggle-bar {
  background: #f0f0f0;
  border-color: #e0e0e0;
}

.toggle-btn {
  padding: 8px 20px;
  border-radius: 8px;
  background: transparent;
  border: 1px solid #333;
  color: #888;
  cursor: pointer;
  font-size: 13px;
  transition: all 0.2s;
}

.toggle-btn:hover {
  background: #252525;
  color: #e0e0e0;
}

.toggle-btn.active {
  background: #1e3a5f;
  color: #64b5f6;
  border-color: #2563a0;
}

body.light .toggle-btn {
  color: #666;
  border-color: #ddd;
}

body.light .toggle-btn:hover {
  background: #e8e8e8;
  color: #1a1a1a;
}

body.light .toggle-btn.active {
  background: #e3f2fd;
  color: #1976d2;
  border-color: #bbdefb;
}

.modal-content .field-item {
  margin-bottom: 16px;
}

.modal-content label {
  display: block;
  margin-bottom: 6px;
  font-size: 13px;
  color: #888;
}

.modal-content input,
.modal-content select {
  width: 100%;
  padding: 10px 14px;
  border-radius: 8px;
  border: 1px solid #333;
  background: #252525;
  color: #e0e0e0;
  font-size: 14px;
  box-sizing: border-box;
}

body.light .modal-content input,
body.light .modal-content select {
  background: #fff;
  border-color: #ddd;
  color: #333;
}
</style>
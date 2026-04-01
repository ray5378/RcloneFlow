<script setup lang="ts">
import { ref, onMounted } from 'vue'
import * as api from '../api'
import type { Task, Schedule, Run } from '../types'

const tasks = ref<Task[]>([])
const schedules = ref<Schedule[]>([])
const runs = ref<Run[]>([])
const remotes = ref<string[]>([])
const taskMenu = ref<string | number>('')
const scheduleMenu = ref<string | number>('')
const runMenu = ref<string | number>('')

const taskForm = ref({
  name: '',
  mode: 'copy',
  sourceRemote: '',
  sourcePath: '',
  targetRemote: '',
  targetPath: '',
})

const scheduleForm = ref({
  taskId: 0,
  spec: '5m',
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

async function createTask() {
  if (!taskForm.value.name) {
    alert('请输入任务名称')
    return
  }
  if (!taskForm.value.sourceRemote || !taskForm.value.targetRemote) {
    alert('请选择源和目标存储')
    return
  }
  try {
    await api.createTask({
      name: taskForm.value.name,
      mode: taskForm.value.mode,
      sourceRemote: taskForm.value.sourceRemote,
      sourcePath: taskForm.value.sourcePath,
      targetRemote: taskForm.value.targetRemote,
      targetPath: taskForm.value.targetPath,
    })
    taskForm.value = { name: '', mode: 'copy', sourceRemote: '', sourcePath: '', targetRemote: '', targetPath: '' }
    await loadData()
  } catch (e) {
    alert((e as Error).message)
  }
}

async function runTask(taskId: number) {
  try {
    await api.runTask(taskId)
    alert('任务已启动')
  } catch (e) {
    alert((e as Error).message)
  }
}

async function deleteTask(taskId: number) {
  if (!confirm('确定删除此任务？')) return
  try {
    await api.deleteTask(taskId)
    await loadData()
  } catch (e) {
    alert((e as Error).message)
  }
}

async function createSchedule() {
  if (!scheduleForm.value.taskId) {
    alert('请选择任务')
    return
  }
  try {
    await api.createSchedule({
      taskId: scheduleForm.value.taskId,
      spec: '@every ' + scheduleForm.value.spec,
      enabled: true,
    })
    scheduleForm.value = { taskId: 0, spec: '5m' }
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

function getTaskName(taskId: number) {
  const task = tasks.value.find(t => t.id === taskId)
  return task?.name || `任务 #${taskId}`
}

function formatTime(time: string) {
  return new Date(time).toLocaleString('zh-CN')
}

function getStatusClass(status: string) {
  switch (status) {
    case 'running': return 'running'
    case 'finished': return 'success'
    case 'failed': return 'failed'
    default: return ''
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
  <div class="card">
    <div class="card-header">
      <div class="title">任务管理</div>
      <div class="subtitle">创建复制 / 同步 / 移动任务</div>
    </div>
    <div class="card-content">
      <div class="field-grid">
        <div class="field-item">
          <label>任务名称</label>
          <input v-model="taskForm.name" type="text" placeholder="输入任务名称" />
        </div>
        <div class="field-item">
          <label>模式</label>
          <select v-model="taskForm.mode">
            <option value="copy">复制 (copy)</option>
            <option value="sync">同步 (sync)</option>
            <option value="move">移动 (move)</option>
          </select>
        </div>
        <div class="field-item">
          <label>源存储</label>
          <select v-model="taskForm.sourceRemote">
            <option value="">选择源存储</option>
            <option v-for="r in remotes" :key="r" :value="r">{{ r }}</option>
          </select>
        </div>
        <div class="field-item">
          <label>源路径</label>
          <input v-model="taskForm.sourcePath" type="text" placeholder="/路径" />
        </div>
        <div class="field-item">
          <label>目标存储</label>
          <select v-model="taskForm.targetRemote">
            <option value="">选择目标存储</option>
            <option v-for="r in remotes" :key="r" :value="r">{{ r }}</option>
          </select>
        </div>
        <div class="field-item">
          <label>目标路径</label>
          <input v-model="taskForm.targetPath" type="text" placeholder="/路径" />
        </div>
      </div>
      <div style="margin-top: 16px">
        <button class="primary" @click="createTask">创建任务</button>
      </div>
    </div>
  </div>

  <!-- Task List -->
  <div class="card">
    <div class="card-header">
      <div class="title">任务列表</div>
    </div>
    <div class="list">
      <div v-for="task in tasks" :key="task.id" class="item">
        <div class="name">
          <strong>{{ task.name }}</strong>
          <div class="muted">
            {{ task.mode }}: {{ task.sourceRemote }}:{{ task.sourcePath || '/' }}
            →
            {{ task.targetRemote }}:{{ task.targetPath || '/' }}
          </div>
        </div>
        <div class="actions" @click.stop>
          <button class="ghost small" @click="runTask(task.id)">运行</button>
          <div class="menu-area">
            <button class="menu-btn" @click="taskMenu = taskMenu === task.id ? '' : task.id">
              ⋮
            </button>
            <div v-if="taskMenu === task.id" class="menu-pop">
              <button class="danger" @click="deleteTask(task.id); taskMenu = ''">删除</button>
            </div>
          </div>
        </div>
      </div>
      <div v-if="!tasks.length" class="empty">暂无任务</div>
    </div>
  </div>

  <!-- Schedule Section -->
  <div class="card">
    <div class="card-header">
      <div class="title">定时任务</div>
      <div class="subtitle">配置多个周期任务</div>
    </div>
    <div class="card-content">
      <div class="field-grid">
        <div class="field-item">
          <label>选择任务</label>
          <select v-model="scheduleForm.taskId">
            <option value="0">选择任务</option>
            <option v-for="t in tasks" :key="t.id" :value="t.id">
              {{ t.name }} ({{ t.mode }})
            </option>
          </select>
        </div>
        <div class="field-item">
          <label>执行周期</label>
          <select v-model="scheduleForm.spec">
            <option value="5m">每5分钟</option>
            <option value="10m">每10分钟</option>
            <option value="30m">每30分钟</option>
            <option value="1h">每1小时</option>
            <option value="6h">每6小时</option>
            <option value="12h">每12小时</option>
            <option value="24h">每24小时</option>
          </select>
        </div>
      </div>
      <div style="margin-top: 16px">
        <button class="primary" @click="createSchedule">创建定时任务</button>
      </div>
    </div>
  </div>

  <!-- Schedule List -->
  <div class="card">
    <div class="card-header">
      <div class="title">定时任务列表</div>
    </div>
    <div class="list">
      <div v-for="s in schedules" :key="s.id" class="item">
        <div class="name">
          <strong>{{ getTaskName(s.taskId) }}</strong>
          <div class="muted">周期: {{ s.spec }}</div>
        </div>
        <div class="actions" @click.stop>
          <button class="ghost small" @click="runTask(s.taskId)">运行</button>
          <div class="menu-area">
            <button class="menu-btn" @click="scheduleMenu = scheduleMenu === s.id ? '' : s.id">
              ⋮
            </button>
            <div v-if="scheduleMenu === s.id" class="menu-pop">
              <button class="danger" @click="deleteSchedule(s.id); scheduleMenu = ''">删除</button>
            </div>
          </div>
        </div>
      </div>
      <div v-if="!schedules.length" class="empty">暂无定时任务</div>
    </div>
  </div>

  <!-- Run Records Section -->
  <div class="card">
    <div class="card-header">
      <div class="title">运行记录</div>
    </div>
    <div class="list">
      <div v-for="run in runs" :key="run.id" class="item">
        <div class="name">
          <strong>{{ getTaskName(run.taskId) }}</strong>
          <div class="muted">
            {{ run.startedAt ? formatTime(run.startedAt) : '' }}
            <span v-if="run.finishedAt"> → {{ formatTime(run.finishedAt) }}</span>
          </div>
        </div>
        <div class="actions" @click.stop>
          <span :class="['status', getStatusClass(run.status)]">{{ run.status }}</span>
          <div class="menu-area">
            <button class="menu-btn" @click="runMenu = runMenu === run.id ? '' : run.id">
              ⋮
            </button>
            <div v-if="runMenu === run.id" class="menu-pop">
              <button class="danger" @click="clearRun(run.id); runMenu = ''">清除</button>
            </div>
          </div>
        </div>
      </div>
      <div v-if="!runs.length" class="empty">暂无运行记录</div>
    </div>
  </div>
</template>

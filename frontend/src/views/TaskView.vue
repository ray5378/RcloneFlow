<script setup lang="ts">
import { ref, onMounted } from 'vue'
import * as api from '../api'
import type { Task } from '../types'

const tasks = ref<Task[]>([])
const remotes = ref<string[]>([])
const taskMenu = ref<string | number>('')

const taskForm = ref({
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
    const [taskData, remoteData] = await Promise.all([
      api.listTasks(),
      api.listRemotes(),
    ])
    tasks.value = taskData
    remotes.value = remoteData.remotes || []
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
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import * as api from '../api'
import type { Schedule, Task } from '../types'

const schedules = ref<Schedule[]>([])
const tasks = ref<Task[]>([])
const scheduleMenu = ref<string | number>('')

const scheduleForm = ref({
  taskId: 0,
  spec: '5m',
})

onMounted(async () => {
  await loadData()
})

async function loadData() {
  try {
    const [scheduleData, taskData] = await Promise.all([
      api.listSchedules(),
      api.listTasks(),
    ])
    schedules.value = scheduleData
    tasks.value = taskData
  } catch (e) {
    console.error(e)
  }
}

function getTaskName(taskId: number) {
  const task = tasks.value.find(t => t.id === taskId)
  return task?.name || `任务 #${taskId}`
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
</script>

<template>
  <div class="card">
    <div class="topbar">
      <div>
        <div style="font-size: 18px; font-weight: 600">定时任务</div>
        <div class="muted">配置多个周期任务</div>
      </div>
    </div>
  </div>

  <!-- Create Schedule -->
  <div class="card">
    <h3 style="margin: 0 0 16px">创建定时任务</h3>
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
      <button @click="createSchedule">创建定时任务</button>
    </div>
  </div>

  <!-- Schedule List -->
  <div class="card">
    <h3 style="margin: 0 0 16px">定时任务列表</h3>
    <div class="list">
      <div v-for="s in schedules" :key="s.id" class="item">
        <div class="manage-row">
          <div>
            <strong>{{ getTaskName(s.taskId) }}</strong>
            <div class="muted">
              {{ s.spec.replace('@every ', '') }} / {{ s.enabled ? '已启用' : '已禁用' }}
            </div>
          </div>
          <div class="actions" @click.stop>
            <div class="menu-area">
              <button class="menu-btn ghost small" @click="scheduleMenu = scheduleMenu === s.id ? '' : s.id">
                ⋮
              </button>
              <div v-if="scheduleMenu === s.id" class="menu-pop">
                <button class="danger" @click="deleteSchedule(s.id); scheduleMenu = ''">删除</button>
              </div>
            </div>
          </div>
        </div>
      </div>
      <div v-if="!schedules.length" class="empty">暂无定时任务</div>
    </div>
  </div>
</template>
